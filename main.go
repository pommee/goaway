package main

import (
	"embed"
	"goaway/internal/asciiart"
	"goaway/internal/logger"
	"goaway/internal/server"
	"goaway/internal/settings"
	"goaway/internal/website"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"
)

//go:embed website/*
var content embed.FS
var version, commit, date string
var log = logger.GetLogger()

func main() {
	var dnsPort, webserverPort, logLevel int

	rootCmd := &cobra.Command{
		Use:   "goaway",
		Short: "GoAway is a DNS filtering tool with a web interface",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := settings.LoadSettings()
			if err != nil {
				log.Error("Failed to load config: %s", err)
			}

			config.Port = dnsPort
			config.WebsitePort = webserverPort
			config.LogLevel = logger.ToLogLevel(logLevel)
			log.SetLevel(logger.LogLevel(logLevel))
			settings.SaveSettings(&config)

			current, err := semver.NewVersion(version)
			if err != nil {
				current, _ = semver.NewVersion("0.0.0")
			}

			server, err := server.NewDNSServer(config)
			if err != nil {
				log.Error("Server initialization failed. %s", err)
				os.Exit(1)
			}

			blockedDomains, serverInstance := server.Init()

			var wg sync.WaitGroup
			errorChannel := make(chan struct{}, 1)

			asciiart.AsciiArt(&config, blockedDomains, current.Original())

			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := serverInstance.ListenAndServe(); err != nil {
					log.Error("DNS server failed to start. %s", err)
					errorChannel <- struct{}{}
				}
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				websiteInstance := website.API{}
				websiteInstance.Start(content, &server, webserverPort)
			}()

			go func() {
				wg.Wait()
			}()

			select {
			case <-errorChannel:
				log.Error("Exiting due to server failure")
				os.Exit(1)
			case <-waitForInterrupt():
				log.Error("Received interrupt, shutting down.")
				os.Exit(0)
			}
		},
	}

	rootCmd.Flags().IntVar(&dnsPort, "dnsport", 53, "Port for the DNS server")
	rootCmd.Flags().IntVar(&webserverPort, "webserverport", 8080, "Port for the web server")
	rootCmd.Flags().IntVar(&logLevel, "loglevel", 1, "0 = DEBUG | 1 = INFO | 2 = WARNING | 3 = ERROR\nSetting 'loglevel' to 1 will show all logs except for DEBUG.\n0 will activate all logs.")

	if err := rootCmd.Execute(); err != nil {
		log.Error("Command execution failed: %s", err)
	}
}

func waitForInterrupt() chan os.Signal {
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)
	return sigChannel
}
