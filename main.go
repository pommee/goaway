package main

import (
	"embed"
	"goaway/internal/asciiart"
	"goaway/internal/server"
	"goaway/internal/website"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

//go:embed website/*
var content embed.FS

func loadConfigFromEnv() server.ServerConfig {
	dnsPort, err := strconv.Atoi(getEnv("DNS_PORT", "53"))
	if err != nil {
		log.Printf("Invalid port number, got error: %s\n", err)
		log.Printf("Using default, %s", err)
		dnsPort = 53
	}
	websitePort, err := strconv.Atoi(getEnv("WEBSITE_PORT", "8080"))
	if err != nil {
		log.Printf("Invalid port number, got error: %s\n", err)
		log.Printf("Using default, %s", err)
		websitePort = 8080
	}

	cacheTTL, err := time.ParseDuration(getEnv("CACHE_TTL", "1m"))
	if err != nil {
		log.Printf("Invalid CACHE_TTL, using default: 1m. %s", err)
		cacheTTL = time.Hour
	}

	return server.ServerConfig{
		Port:           dnsPort,
		WebsitePort:    websitePort,
		UpstreamDNS:    getEnv("UPSTREAM_DNS", "8.8.8.8:53"),
		BlacklistPath:  getEnv("BLACKLIST_PATH", "blacklist.json"),
		CountersFile:   getEnv("COUNTERS_FILE", "counters.json"),
		RequestLogFile: getEnv("REQUEST_LOG_FILE", "requests.json"),
		CacheTTL:       cacheTTL,
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func main() {
	config := loadConfigFromEnv()
	server, err := server.NewDNSServer(config)
	if err != nil {
		log.Printf("Server initialization failed. %s", err)
		os.Exit(1)
	}

	blockedDomains, serverInstance := server.Init()

	var wg sync.WaitGroup
	errorChannel := make(chan struct{}, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := serverInstance.ListenAndServe(); err != nil {
			log.Printf("DNS server failed to start. %s", err)
			errorChannel <- struct{}{}
		}
	}()

	asciiart.AsciiArt(config, blockedDomains)

	wg.Add(1)
	go func() {
		defer wg.Done()
		websiteInstance := website.API{}
		websiteInstance.Start(content, &server, config.WebsitePort)
	}()

	go func() {
		wg.Wait()
	}()

	select {
	case <-errorChannel:
		log.Println("Exiting due to server failure.")
		os.Exit(1)
	case <-waitForInterrupt():
		log.Println("Received interrupt, shutting down.")
		os.Exit(0)
	}
}

func waitForInterrupt() chan os.Signal {
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)
	return sigChannel
}
