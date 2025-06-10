package main

import (
	"embed"
	"fmt"
	"goaway/backend/api"
	"goaway/backend/asciiart"
	arp "goaway/backend/dns"
	"goaway/backend/dns/database"
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
	WebserverPort       int
	LogLevel            int
	StatisticsRetention int
	LoggingEnabled      bool
	Authentication      bool
	DevMode             bool
	Ansi                bool
	JSON                bool
}

func main() {
	if err := createRootCommand().Execute(); err != nil {
		log.Error("Command execution failed: %s", err)
		os.Exit(1)
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
	cmd.Flags().IntVar(&flags.WebserverPort, "webserver-port", 8080, "Port for the web server")
	cmd.Flags().IntVar(&flags.LogLevel, "log-level", 1, "0 = DEBUG | 1 = INFO | 2 = WARNING | 3 = ERROR")
	cmd.Flags().IntVar(&flags.StatisticsRetention, "statistics-retention", 7, "Days to keep statistics")
	cmd.Flags().BoolVar(&flags.LoggingEnabled, "logging", true, "Toggle logging")
	cmd.Flags().BoolVar(&flags.Authentication, "auth", true, "Toggle authentication for admin dashboard")
	cmd.Flags().BoolVar(&flags.DevMode, "dev", false, "Only used while developing goaway")
	cmd.Flags().BoolVar(&flags.Ansi, "ansi", true, "Toggle colorized logs. Only available in non-json formatted logs")
	cmd.Flags().BoolVar(&flags.JSON, "json", false, "Toggle JSON formatted logs")

	return cmd
}

func getSetFlags(cmd *cobra.Command, flags *Flags) *setup.SetFlags {
	setFlags := &setup.SetFlags{}

	if cmd.Flags().Changed("dns-port") {
		setFlags.DnsPort = &flags.DnsPort
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
	if cmd.Flags().Changed("dev") {
		setFlags.DevMode = &flags.DevMode
	}
	if cmd.Flags().Changed("ansi") {
		setFlags.Ansi = &flags.Ansi
	}
	if cmd.Flags().Changed("json") {
		setFlags.JSON = &flags.JSON
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
		log.Error("failed while initializing database: %v", err)
		os.Exit(1)
	}

	notificationManager := notification.NewNotificationManager(dbManager)

	dnsServer, err := server.NewDNSServer(config, dbManager, notificationManager)
	if err != nil {
		log.Error("Failed to initialize server: %s", err)
		os.Exit(1)
	}

	go dnsServer.ProcessLogEntries()

	blockedDomains, serverInstance, err := dnsServer.Init()
	if err != nil {
		log.Error("Failed to initialize DNS server: %s", err)
		os.Exit(1)
	}
	currentVersion := setup.GetVersionOrDefault(version)

	asciiart.AsciiArt(config, blockedDomains, currentVersion.Original(), config.API.Authentication, ansi)
	startServices(dnsServer, serverInstance, config)
}

func startServices(dnsServer *server.DNSServer, serverInstance *dns.Server, config *settings.Config) {
	var wg sync.WaitGroup
	errorChannel := make(chan struct{}, 1)
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(2)
	go func() {
		defer wg.Done()
		log.Info("Started DNS server on port %s", serverInstance.Addr)
		if err := serverInstance.ListenAndServe(); err != nil {
			log.Error("DNS server failed: %s", err)
			errorChannel <- struct{}{}
		}
	}()

	prefetcher := prefetch.New(dnsServer)

	go func() {
		defer wg.Done()
		apiServer := api.API{
			Authentication:           config.API.Authentication,
			Config:                   config,
			DBManager:                dnsServer.DBManager,
			Blacklist:                dnsServer.Blacklist,
			Whitelist:                dnsServer.Whitelist,
			WSQueries:                dnsServer.WSQueries,
			WSCommunication:          dnsServer.WSCommunication,
			DNSPort:                  config.DNS.Port,
			Version:                  version,
			Commit:                   commit,
			Date:                     date,
			Notifications:            dnsServer.Notifications,
			PrefetchedDomainsManager: &prefetcher,
		}

		apiServer.Start(content, dnsServer, errorChannel)
	}()

	go arp.ProcessARPTable()
	go dnsServer.ClearOldEntries()
	go prefetcher.Run()

	go func() { wg.Wait() }()

	select {
	case <-errorChannel:
		log.Error("Server failure detected. Exiting.")
		os.Exit(1)
	case <-sigChannel:
		log.Info("Received interrupt. Shutting down.")
		os.Exit(0)
	}
}
