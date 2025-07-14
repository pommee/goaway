package main

import (
	"crypto/tls"
	"embed"
	"fmt"
	"goaway/backend/api"
	"goaway/backend/asciiart"
	arp "goaway/backend/dns"
	"goaway/backend/dns/database"
	"goaway/backend/dns/lists"
	"goaway/backend/dns/server"
	"goaway/backend/dns/server/prefetch"
	"goaway/backend/logging"
	notification "goaway/backend/notifications"
	"goaway/backend/settings"
	"goaway/backend/setup"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"net/http"
	_ "net/http/pprof"

	"github.com/miekg/dns"
	"github.com/spf13/cobra"
)

var (
	version, commit, date string
	log                   = logging.GetLogger()

	//go:embed client/dist/*
	content embed.FS
)

type Flags struct {
	DnsPort             int
	DoTPort             int
	DoHPort             int
	WebserverPort       int
	LogLevel            int
	StatisticsRetention int
	LoggingEnabled      bool
	Authentication      bool
	Dashboard           bool
	Ansi                bool
	JSON                bool
	InAppUpdate         bool
}

func main() {
	if err := createRootCommand().Execute(); err != nil {
		log.Fatal("Command execution failed: %s", err)
	}
}

func createRootCommand() *cobra.Command {
	flags := Flags{}

	cmd := &cobra.Command{
		Use:   "goaway",
		Short: "GoAway is a DNS sinkhole with a web interface",
		Run: func(cmd *cobra.Command, args []string) {
			setFlags := getSetFlags(cmd, &flags)
			config := setup.InitializeSettings(setFlags)

			startServer(config, flags.Ansi)
		},
	}

	cmd.Flags().IntVar(&flags.DnsPort, "dns-port", 53, "Port for the DNS server")
	cmd.Flags().IntVar(&flags.DoTPort, "dot-port", 853, "Port for the DoT (DNS-over-TCP) server")
	cmd.Flags().IntVar(&flags.DoHPort, "doh-port", 443, "Port for the DoH (DNS-over-HTTPS) server")
	cmd.Flags().IntVar(&flags.WebserverPort, "webserver-port", 8080, "Port for the web server")
	cmd.Flags().IntVar(&flags.LogLevel, "log-level", 1, "0 = DEBUG | 1 = INFO | 2 = WARNING | 3 = ERROR")
	cmd.Flags().IntVar(&flags.StatisticsRetention, "statistics-retention", 7, "Days to keep statistics")
	cmd.Flags().BoolVar(&flags.LoggingEnabled, "logging", true, "Toggle logging")
	cmd.Flags().BoolVar(&flags.Authentication, "auth", true, "Toggle authentication for admin dashboard")
	cmd.Flags().BoolVar(&flags.Dashboard, "dashboard", true, "Serve dashboard")
	cmd.Flags().BoolVar(&flags.Ansi, "ansi", true, "Toggle colorized logs. Only available in non-json formatted logs")
	cmd.Flags().BoolVar(&flags.JSON, "json", false, "Toggle JSON formatted logs")
	cmd.Flags().BoolVar(&flags.InAppUpdate, "in-app-update", false, "Toggle ability to update via dashboard")

	return cmd
}

func getSetFlags(cmd *cobra.Command, flags *Flags) *setup.SetFlags {
	setFlags := &setup.SetFlags{}

	if cmd.Flags().Changed("dns-port") {
		setFlags.DnsPort = &flags.DnsPort
	}
	if cmd.Flags().Changed("dot-port") {
		setFlags.DoTPort = &flags.DoTPort
	}
	if cmd.Flags().Changed("doh-port") {
		setFlags.DoHPort = &flags.DoHPort
	}
	if cmd.Flags().Changed("webserver-port") {
		setFlags.WebserverPort = &flags.WebserverPort
	}
	if cmd.Flags().Changed("log-level") {
		setFlags.LogLevel = &flags.LogLevel
	}
	if cmd.Flags().Changed("statistics-retention") {
		setFlags.StatisticsRetention = &flags.StatisticsRetention
	}
	if cmd.Flags().Changed("logging") {
		setFlags.LoggingEnabled = &flags.LoggingEnabled
	}
	if cmd.Flags().Changed("auth") {
		setFlags.Authentication = &flags.Authentication
	}
	if cmd.Flags().Changed("dashboard") {
		setFlags.Dashboard = &flags.Dashboard
	}
	if cmd.Flags().Changed("ansi") {
		setFlags.Ansi = &flags.Ansi
	}
	if cmd.Flags().Changed("json") {
		setFlags.JSON = &flags.JSON
	}
	if cmd.Flags().Changed("in-app-update") {
		setFlags.InAppUpdate = &flags.InAppUpdate
	}

	return setFlags
}

func startServer(config *settings.Config, ansi bool) {
	if os.Getenv("GOAWAY_PROFILE") == "true" {
		log.Warning("GOAWAY_PROFILE was set, starting profiler...")
		go func() {
			fmt.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	dbManager, err := database.Initialize()
	if err != nil {
		log.Fatal("failed while initializing database: %v", err)
	}

	notificationManager := notification.NewNotificationManager(dbManager)

	dnsServer, err := server.NewDNSServer(config, dbManager, notificationManager)
	if err != nil {
		log.Fatal("Failed to initialize server: %s", err)
	}

	go dnsServer.ProcessLogEntries()

	dnsReadyChannel := make(chan struct{})
	errorChannel := make(chan struct{}, 1)

	notifyReady := func() {
		blacklistEntry, err := lists.InitializeBlacklist(dnsServer.DBManager)
		if err != nil {
			log.Error("Failed to initialize blacklist: %v", err)
			errorChannel <- struct{}{}
			return
		}
		dnsServer.Blacklist = blacklistEntry

		domains, err := blacklistEntry.CountDomains()
		if err != nil {
			log.Warning("Failed to count blacklist domains: %v", err)
		}

		currentVersion := setup.GetVersionOrDefault(version)
		asciiart.AsciiArt(config, domains, currentVersion.Original(), config.API.Authentication, ansi)

		log.Info("Started DNS server on: %s:%d", config.DNS.Address, config.DNS.Port)
		close(dnsReadyChannel)
	}

	udpServer, err := dnsServer.InitUDP(notifyReady)
	if err != nil {
		log.Fatal("Failed to initialize DNS server: %s", err)
	}

	tcpServer, _ := dnsServer.InitTCP()
	if err != nil {
		log.Fatal("Failed to initialize TCP server: %s", err)
	}

	startServices(dnsServer, udpServer, tcpServer, config, ansi, dnsReadyChannel, errorChannel)
}

func startServices(dnsServer *server.DNSServer, udpServer, tcpServer *dns.Server, config *settings.Config, ansi bool, dnsReadyChannel chan struct{}, errorChannel chan struct{}) {
	var (
		wg         sync.WaitGroup
		sigChannel = make(chan os.Signal, 1)
	)

	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)
	wg.Add(2)

	go func() {
		defer wg.Done()

		if err := udpServer.ListenAndServe(); err != nil {
			log.Error("UDP server failed: %s", err)
			errorChannel <- struct{}{}
		}
	}()

	go func() {
		defer wg.Done()

		if err := tcpServer.ListenAndServe(); err != nil {
			log.Error("TCP server failed: %s", err)
			errorChannel <- struct{}{}
		}
	}()

	if config.DNS.TLSCertFile != "" && config.DNS.TLSKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(config.DNS.TLSCertFile, config.DNS.TLSKeyFile)
		if err != nil {
			log.Fatal("Failed to load TLS certificate: %s", err)
		}

		dotServer, err := dnsServer.InitDoT(cert)
		if err != nil {
			log.Fatal("Failed to initialize DoT server: %s", err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := dotServer.ListenAndServe(); err != nil {
				log.Error("DoT server failed: %s", err)
				errorChannel <- struct{}{}
			}
		}()

		dohServer, err := dnsServer.InitDoH(cert)
		if err != nil {
			log.Fatal("Failed to initialize DoH server: %s", err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			if serverIP, err := api.GetServerIP(); err == nil {
				log.Info("DoH server running at https://%s:%d/dns-query", serverIP, config.DNS.DoHPort)
			} else {
				log.Info("DoH server running on port :%d", config.DNS.DoHPort)
			}

			if err := dohServer.ListenAndServeTLS(config.DNS.TLSCertFile, config.DNS.TLSKeyFile); err != nil {
				log.Error("DoH server failed: %s", err)
				errorChannel <- struct{}{}
			}
		}()
	}

	prefetcher := prefetch.New(dnsServer)

	go func() {
		defer wg.Done()

		<-dnsReadyChannel
		apiServer := api.API{
			Authentication:           config.API.Authentication,
			Config:                   config,
			DNSPort:                  config.DNS.Port,
			Version:                  version,
			Commit:                   commit,
			Date:                     date,
			DNSServer:                dnsServer,
			DBManager:                dnsServer.DBManager,
			Blacklist:                dnsServer.Blacklist,
			Whitelist:                dnsServer.Whitelist,
			PrefetchedDomainsManager: &prefetcher,
			Notifications:            dnsServer.Notifications,
			WSQueries:                dnsServer.WSQueries,
			WSCommunication:          dnsServer.WSCommunication,
		}

		apiServer.Start(content, errorChannel)
	}()

	go func() {
		<-dnsReadyChannel
		log.Debug("Starting ARP table processing...")
		arp.ProcessARPTable()
	}()

	go func() {
		<-dnsReadyChannel
		if config.ScheduledBlacklistUpdates {
			log.Debug("Starting scheduler for automatic list updates...")
			dnsServer.Blacklist.ScheduleAutomaticListUpdates()
		}
	}()

	go func() {
		<-dnsReadyChannel
		log.Debug("Starting cache cleanup routine...")
		dnsServer.ClearOldEntries()
	}()

	go func() {
		<-dnsReadyChannel
		log.Debug("Starting prefetcher...")
		prefetcher.Run()
	}()

	go func() { wg.Wait() }()

	select {
	case <-errorChannel:
		log.Fatal("Server failure detected. Exiting.")
	case <-sigChannel:
		log.Info("Received interrupt. Shutting down.")
		os.Exit(0)
	}
}
