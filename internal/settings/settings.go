package settings

import (
	"encoding/json"
	"fmt"
	"goaway/internal/logger"
	"goaway/internal/server"
	"net"
	"os"
	"strings"
	"time"
)

var log = logger.GetLogger()

type Config struct {
	ServerConfig struct {
		Port              int      `json:"Port"`
		WebsitePort       int      `json:"WebsitePort"`
		LogLevel          int      `json:"LogLevel"`
		LoggingDisabled   bool     `json:"LoggingDisabled"`
		UpstreamDNS       []string `json:"UpstreamDNS"`
		PreferredUpstream string   `json:"PreferredUpstream"`
		BlacklistPath     string   `json:"BlacklistPath"`
		CountersFile      string   `json:"CountersFile"`
		RequestLogFile    string   `json:"RequestLogFile"`
		CacheTTL          string   `json:"CacheTTL"`
	} `json:"serverConfig"`
}

func LoadSettings() (server.ServerConfig, error) {
	var config Config
	data, err := os.ReadFile("./settings.json")
	if err != nil {
		return server.ServerConfig{}, err
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return server.ServerConfig{}, err
	}

	cacheTTL, err := time.ParseDuration(config.ServerConfig.CacheTTL)
	if err != nil {
		log.Error("Could not parse CacheTTL. %s", err)
		cacheTTL = time.Minute
	}

	bestDNS, err := findBestDNS(config.ServerConfig.UpstreamDNS)
	if err != nil {
		log.Error("Could not find best DNS: %v", err)
	}

	return server.ServerConfig{
		Port:              config.ServerConfig.Port,
		WebsitePort:       config.ServerConfig.WebsitePort,
		LogLevel:          logger.ToLogLevel(config.ServerConfig.LogLevel),
		LoggingDisabled:   config.ServerConfig.LoggingDisabled,
		UpstreamDNS:       config.ServerConfig.UpstreamDNS,
		PreferredUpstream: bestDNS,
		BlacklistPath:     config.ServerConfig.BlacklistPath,
		CountersFile:      config.ServerConfig.CountersFile,
		RequestLogFile:    config.ServerConfig.RequestLogFile,
		CacheTTL:          cacheTTL,
	}, nil
}

func findBestDNS(dnsServers []string) (string, error) {
	var bestDNS string
	var bestTime time.Duration

	for _, dns := range dnsServers {
		duration, err := checkDNS(dns)
		if err != nil {
			log.Error("Error checking DNS %s: %v", dns, err)
			continue
		}

		if bestDNS == "" || duration < bestTime {
			bestDNS = dns
			bestTime = duration
		}
	}

	if bestDNS == "" {
		return "", fmt.Errorf("no DNS servers responded")
	}
	return bestDNS, nil
}

func checkDNS(ip string) (time.Duration, error) {
	start := time.Now()
	_, err := net.DialTimeout("tcp", ip, 2*time.Second)
	if err != nil {
		return 0, err
	}
	return time.Since(start), nil
}

func SaveSettings(config *server.ServerConfig) error {
	configData := Config{
		ServerConfig: struct {
			Port              int      `json:"Port"`
			WebsitePort       int      `json:"WebsitePort"`
			LogLevel          int      `json:"LogLevel"`
			LoggingDisabled   bool     `json:"LoggingDisabled"`
			UpstreamDNS       []string `json:"UpstreamDNS"`
			PreferredUpstream string   `json:"PreferredUpstream"`
			BlacklistPath     string   `json:"BlacklistPath"`
			CountersFile      string   `json:"CountersFile"`
			RequestLogFile    string   `json:"RequestLogFile"`
			CacheTTL          string   `json:"CacheTTL"`
		}{
			Port:              config.Port,
			WebsitePort:       config.WebsitePort,
			LogLevel:          logger.ToInteger(config.LogLevel),
			LoggingDisabled:   config.LoggingDisabled,
			UpstreamDNS:       config.UpstreamDNS,
			PreferredUpstream: config.PreferredUpstream,
			BlacklistPath:     config.BlacklistPath,
			CountersFile:      config.CountersFile,
			RequestLogFile:    config.RequestLogFile,
			CacheTTL:          config.CacheTTL.String(),
		},
	}

	data, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("./settings.json", data, 0644)
}

func UpdateSettings(dnsServer *server.DNSServer, updatedSettings map[string]interface{}) {

	if disableLogging, ok := updatedSettings["disableLogging"].(bool); ok {
		log.ToggleLogging(disableLogging)
		dnsServer.Config.LoggingDisabled = disableLogging
	}

	if ttl, ok := updatedSettings["cacheTTL"].(string); ok {
		if parsedTTL, err := time.ParseDuration(ttl + "s"); err == nil {
			dnsServer.Config.CacheTTL = parsedTTL
		}
	}

	if logLevel, ok := updatedSettings["logLevel"].(string); ok {
		dnsServer.Config.LogLevel = logger.FromString(strings.ToUpper(logLevel))
		log.SetLevel(dnsServer.Config.LogLevel)
	}

	SaveSettings(&dnsServer.Config)
}
