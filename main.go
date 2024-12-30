package main

import (
	"embed"
	"goaway/internal/asciiart"
	"goaway/internal/server"
	"goaway/internal/website"
	"log"
	"os"
	"strconv"
	"sync"
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
		BlacklistPath:  getEnv("BLACKLIST_PATH", "blacklist.txt"),
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
	var wg sync.WaitGroup

	if err != nil {
		log.Printf("Server initialization failed. %s", err)
		os.Exit(1)
	}

	blockedDomains, serverInstance := server.Init()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := serverInstance.ListenAndServe(); err != nil {
			log.Printf("Server failed to start. %s", err)
			os.Exit(1)
		}
	}()

	log.Printf("Starting DNS server, port: %d, upstream: %s, blacklist: %s, cache TTL: %v\n", config.Port, config.UpstreamDNS, config.BlacklistPath, config.CacheTTL)
	log.Printf("Loaded %d domains into blacklist\n", blockedDomains)

	wg.Add(1)
	go func() {
		defer wg.Done()
		websiteInstance := website.API{}
		websiteInstance.Start(content, &server, config.WebsitePort)
	}()

	asciiart.AsciiArt()
	wg.Wait()
}
