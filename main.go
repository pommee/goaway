package main

import (
	"embed"
	"goaway/internal/api"
	"goaway/internal/asciiart"
	"goaway/internal/logging"
	"goaway/internal/server"
	"goaway/internal/settings"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Masterminds/semver"
	"github.com/miekg/dns"
	"github.com/spf13/cobra"
)

//go:embed website/*
var content embed.FS

var (
	version, commit, date string
	log                   = logging.GetLogger()
)

func main() {
	var dnsPort, webserverPort, logLevel, statisticsRetention int
	var disableLogging, disableAuth bool

	rootCmd := createRootCommand(&dnsPort, &webserverPort, &logLevel, &statisticsRetention, &disableLogging, &disableAuth)
	if err := rootCmd.Execute(); err != nil {
		log.Error("Command execution failed: %s", err)
	}
}

func createRootCommand(dnsPort, webserverPort, logLevel, statisticsRetention *int, disableLogging, disableAuth *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "goaway",
		Short: "GoAway is a DNS sinkhole with a web interface",
		Run: func(cmd *cobra.Command, args []string) {
			runServer(*dnsPort, *webserverPort, *logLevel, *statisticsRetention, *disableLogging, *disableAuth)
		},
	}

	cmd.Flags().IntVar(dnsPort, "dnsport", 53, "Port for the DNS server")
	cmd.Flags().IntVar(webserverPort, "webserverport", 8080, "Port for the web server")
	cmd.Flags().IntVar(logLevel, "loglevel", 1, "0 = DEBUG | 1 = INFO | 2 = WARNING | 3 = ERROR")
	cmd.Flags().IntVar(statisticsRetention, "statisticsRetention", 1, "Number is amount of days to keep statistics")
	cmd.Flags().BoolVar(disableLogging, "disablelogging", false, "If true, then no logs will appear in the container")
	cmd.Flags().BoolVar(disableAuth, "auth", true, "If false, then no authentication is required for the admin dashboard")

	return cmd
}

func runServer(dnsPort, webserverPort, logLevel, statisticsRetention int, disableLogging, disableAuth bool) {
	config, err := settings.LoadSettings()

	if err != nil {
		log.Error("Failed to load settings: %s", err)
		exit(1)
	}

	updateConfig(&config, dnsPort, webserverPort, logLevel, statisticsRetention, disableLogging, disableAuth)
	config.Save()

	currentVersion := getVersionOrDefault()

	dnsServer, err := server.NewDNSServer(config.DNSServer)
	if err != nil {
		log.Error("Server initialization failed: %s", err)
		exit(1)
	}
	go dnsServer.ProcessLogEntries()

	blockedDomains, serverInstance := dnsServer.Init()
	asciiart.AsciiArt(&config, blockedDomains, currentVersion.Original(), disableAuth)
	dnsServer.UpdateCounters()

	go dnsServer.ClearOldEntries()
	//go update.PollUpdates(version)
	startServices(dnsServer, serverInstance, webserverPort, disableAuth)
}

func updateConfig(config *settings.Config, dnsPort, webserverPort, logLevel, statisticsRetention int, disableLogging, disableAuth bool) {
	config.DNSServer.Port = dnsPort
	config.DNSServer.LoggingDisabled = disableLogging
	config.DNSServer.StatisticsRetention = statisticsRetention
	config.APIServer.Port = webserverPort
	config.APIServer.Authentication = disableAuth
	config.LogLevel = logging.LogLevel(logLevel)
	log.SetLevel(logging.LogLevel(logLevel))
}

func getVersionOrDefault() *semver.Version {
	versionObj, err := semver.NewVersion(version)
	if err != nil {
		versionObj, _ = semver.NewVersion("0.0.0")
	}
	return versionObj
}

func startServices(dnsServer *server.DNSServer, serverInstance *dns.Server, webserverPort int, disableAuth bool) {
	var wg sync.WaitGroup
	errorChannel := make(chan struct{}, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := serverInstance.ListenAndServe(); err != nil {
			log.Error("DNS server failed to start: %s", err)
			errorChannel <- struct{}{}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		websiteInstance := api.API{Config: &settings.APIServerConfig{Port: webserverPort, Authentication: disableAuth}}
		websiteInstance.Start(content, dnsServer, errorChannel)
	}()

	go func() {
		wg.Wait()
	}()

	select {
	case <-errorChannel:
		log.Error("Exiting due to server failure")
		log.Info("Help can be provided using the --help flag")
		exit(1)
	case <-waitForInterrupt():
		log.Error("Received interrupt, shutting down.")
		exit(0)
	}
}

func waitForInterrupt() chan os.Signal {
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)
	return sigChannel
}

func exit(code int) {
	os.Exit(code)
}
