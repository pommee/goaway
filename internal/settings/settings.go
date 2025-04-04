package settings

import (
	"encoding/json"
	"fmt"
	"goaway/internal/logging"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var log = logging.GetLogger()

type DNSServerConfig struct {
	Port                int           `json:"Port"`
	LoggingDisabled     bool          `json:"LoggingDisabled"`
	UpstreamDNS         []string      `json:"UpstreamDNS"`
	PreferredUpstream   string        `json:"PreferredUpstream"`
	CacheTTL            time.Duration `json:"CacheTTL"`
	StatisticsRetention int           `json:"StatisticsRetention"`
}

type APIServerConfig struct {
	Port           int  `json:"Port"`
	Authentication bool `json:"Authentication"`
}

type Config struct {
	DNSServer *DNSServerConfig `json:"dnsServer"`
	APIServer *APIServerConfig `json:"apiServer"`
	LogLevel  logging.LogLevel `json:"LogLevel"`
	DevMode   bool
}

func (c *DNSServerConfig) UnmarshalJSON(data []byte) error {
	type Alias DNSServerConfig
	aux := &struct {
		CacheTTL string `json:"CacheTTL"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.CacheTTL != "" {
		parsedTTL, err := time.ParseDuration(aux.CacheTTL)
		if err != nil {
			return fmt.Errorf("invalid CacheTTL: %w", err)
		}
		c.CacheTTL = parsedTTL
	}
	return nil
}

func (c *DNSServerConfig) MarshalJSON() ([]byte, error) {
	type Alias DNSServerConfig
	return json.Marshal(&struct {
		CacheTTL string `json:"CacheTTL"`
		*Alias
	}{
		CacheTTL: c.CacheTTL.String(),
		Alias:    (*Alias)(c),
	})
}

func LoadSettings() (Config, error) {
	var config Config

	path, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf("could not determine current directory: %w", err)
	}
	path = filepath.Join(path, "settings.json")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Info("Settings file not found. Fetching from remote source...")
		if err := fetchAndSaveSettings(path); err != nil {
			return Config{}, fmt.Errorf("failed to fetch settings: %w", err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("could not read settings file: %w", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("invalid settings format: %w", err)
	}

	config.DNSServer.PreferredUpstream, err = findBestDNS(config.DNSServer.UpstreamDNS)
	if err != nil {
		log.Error("Could not find best DNS: %v", err)
	}

	return config, nil
}

func fetchAndSaveSettings(filePath string) error {
	url := "https://raw.githubusercontent.com/pommee/goaway/refs/heads/main/settings.json"

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch settings.json: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch settings.json: HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create settings file: %w", err)
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to save settings file: %w", err)
	}

	return nil
}

func (config *Config) Save() {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Error("Could not parse settings %v", err)
		return
	}

	if err := os.WriteFile("./settings.json", data, 0644); err != nil {
		log.Error("Could not save settings %v", err)
	}
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

func (config *Config) UpdateDNSSettings(updatedSettings map[string]interface{}) {
	updateField := func(field string, updateFunc func(interface{})) {
		if value, ok := updatedSettings[field]; ok {
			updateFunc(value)
		}
	}

	updateField("disableLogging", func(value interface{}) {
		if disableLogging, ok := value.(bool); ok {
			log.ToggleLogging(disableLogging)
			config.DNSServer.LoggingDisabled = disableLogging
		}
	})

	updateField("cacheTTL", func(value interface{}) {
		if ttl, ok := value.(float64); ok {
			config.DNSServer.CacheTTL = time.Duration(ttl) * time.Second
		}
	})

	updateField("logLevel", func(value interface{}) {
		if logLevel, ok := value.(string); ok {
			config.LogLevel = logging.FromString(strings.ToUpper(logLevel))
			log.SetLevel(config.LogLevel)
		}
	})

	updateField("statisticsRetention", func(value interface{}) {
		if days, ok := value.(float64); ok {
			config.DNSServer.StatisticsRetention = int(days)
		}
	})
	config.Save()
}
