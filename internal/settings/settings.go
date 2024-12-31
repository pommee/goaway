package settings

import (
	"encoding/json"
	"goaway/internal/server"
	"log"
	"os"
	"time"
)

type Config struct {
	ServerConfig struct {
		Port           int    `json:"Port"`
		WebsitePort    int    `json:"WebsitePort"`
		UpstreamDNS    string `json:"UpstreamDNS"`
		BlacklistPath  string `json:"BlacklistPath"`
		CountersFile   string `json:"CountersFile"`
		RequestLogFile string `json:"RequestLogFile"`
		CacheTTL       string `json:"CacheTTL"`
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
		log.Printf("Could not parse CacheTTL. Error: %s\n", err)
		cacheTTL = time.Minute
	}

	return server.ServerConfig{
		Port:           config.ServerConfig.Port,
		WebsitePort:    config.ServerConfig.WebsitePort,
		UpstreamDNS:    config.ServerConfig.UpstreamDNS,
		BlacklistPath:  config.ServerConfig.BlacklistPath,
		CountersFile:   config.ServerConfig.CountersFile,
		RequestLogFile: config.ServerConfig.RequestLogFile,
		CacheTTL:       cacheTTL,
	}, nil
}

func SaveSettings(config *server.ServerConfig) error {
	configData := Config{
		ServerConfig: struct {
			Port           int    `json:"Port"`
			WebsitePort    int    `json:"WebsitePort"`
			UpstreamDNS    string `json:"UpstreamDNS"`
			BlacklistPath  string `json:"BlacklistPath"`
			CountersFile   string `json:"CountersFile"`
			RequestLogFile string `json:"RequestLogFile"`
			CacheTTL       string `json:"CacheTTL"`
		}{
			Port:           config.Port,
			WebsitePort:    config.WebsitePort,
			UpstreamDNS:    config.UpstreamDNS,
			BlacklistPath:  config.BlacklistPath,
			CountersFile:   config.CountersFile,
			RequestLogFile: config.RequestLogFile,
			CacheTTL:       config.CacheTTL.String(),
		},
	}

	data, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("./settings.json", data, 0644)
}

func UpdateSettings(dnsServer *server.DNSServer, updatedSettings map[string]interface{}) {
	if ttl, ok := updatedSettings["cacheTTL"].(string); ok {
		updateCacheTTL(&dnsServer.Config, ttl)
	}
	SaveSettings(&dnsServer.Config)
}

func updateCacheTTL(config *server.ServerConfig, ttl string) {
	if parsedTTL, err := time.ParseDuration(ttl + "s"); err == nil {
		config.CacheTTL = parsedTTL
	}
}
