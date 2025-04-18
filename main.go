package main

import (
	"embed"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"goaway/internal/api"
	"goaway/internal/arp"
	"goaway/internal/asciiart"
	"goaway/internal/logging"
	"goaway/internal/server"
	"goaway/internal/settings"
	"goaway/internal/setup"

	"github.com/miekg/dns"
	"github.com/spf13/cobra"
)

var (
	version string
	log     = logging.GetLogger()

	//go:embed website/dist/*
	content embed.FS
)

func main() {
	if err := createRootCommand().Execute(); err != nil {
		log.Error("Command execution failed: %s", err)
		os.Exit(1)
	}
}

func createRootCommand() *cobra.Command {
	flags := setup.Flags{}

	cmd := &cobra.Command{
		Use:   "goaway",
		Short: "GoAway is a DNS sinkhole with a web interface",
		Run: func(cmd *cobra.Command, args []string) {
			startServer(setup.InitializeSettings(&flags))
		},
	}

	cmd.Flags().IntVar(&flags.DnsPort, "dnsport", 53, "Port for the DNS server")
	cmd.Flags().IntVar(&flags.WebserverPort, "webserverport", 8080, "Port for the web server")
	cmd.Flags().IntVar(&flags.LogLevel, "loglevel", 1, "0 = DEBUG | 1 = INFO | 2 = WARNING | 3 = ERROR")
	cmd.Flags().IntVar(&flags.StatisticsRetention, "statisticsRetention", 1, "Days to keep statistics")
	cmd.Flags().BoolVar(&flags.DisableLogging, "disablelogging", false, "Disable logging")
	cmd.Flags().BoolVar(&flags.DisableAuth, "auth", true, "Disable authentication for admin dashboard")
	cmd.Flags().BoolVar(&flags.DevMode, "dev", false, "Only use while developing goaway")

	return cmd
}

func startServer(config settings.Config) {
	dnsServer, err := server.NewDNSServer(config.DNSServer)
	if err != nil {
		log.Error("Failed to initialize server: %s", err)
		os.Exit(1)
	}

	go dnsServer.ProcessLogEntries()
	go arp.ProcessARPTable()
	go dnsServer.ClearOldEntries()

	blockedDomains, serverInstance := dnsServer.Init()
	currentVersion := setup.GetVersionOrDefault(version)

	asciiart.AsciiArt(config, blockedDomains, currentVersion.Original(), config.APIServer.Authentication)
	dnsServer.UpdateCounters()

	startServices(dnsServer, serverInstance, config)
}

func startServices(dnsServer *server.DNSServer, serverInstance *dns.Server, config settings.Config) {
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

	go func() {
		defer wg.Done()
		apiServer := api.API{
			Config: &settings.APIServerConfig{
				Port:           config.APIServer.Port,
				Authentication: config.APIServer.Authentication,
			},
			Version: version,
		}

		serveEmbedded := !config.DevMode
		if !serveEmbedded {
			log.Warning("No embedded content found, not serving")
		}

		apiServer.Start(serveEmbedded, content, dnsServer, errorChannel)
	}()

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
