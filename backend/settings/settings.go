package settings

import (
	"encoding/json"
	"fmt"
	"goaway/backend/logging"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var log = logging.GetLogger()

type Status struct {
	Paused    bool
	PausedAt  time.Time
	PauseTime int
}

type Config struct {
	DNSPort             int `json:"dnsPort"`
	APIPort             int `json:"apiPort"`
	StatisticsRetention int `json:"statisticsRetention"`

	Authentication  bool `json:"authentication"`
	DevMode         bool `json:"devMode"`
	LoggingDisabled bool `json:"loggingDisabled"`

	PreferredUpstream string   `json:"preferredUpstream"`
	UpstreamDNS       []string `json:"upstreamDNS"`
	DNSStatus         Status   `json:"dnsStatus,omitzero"`

	CacheTTL time.Duration    `json:"cacheTTL"`
	LogLevel logging.LogLevel `json:"logLevel"`
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

func (config *Config) UpdateDNSSettings(updatedSettings map[string]interface{}) {
	updateField := func(field string, updateFunc func(interface{})) {
		if value, ok := updatedSettings[field]; ok {
			updateFunc(value)
		}
	}

	updateField("disableLogging", func(value interface{}) {
		if disableLogging, ok := value.(bool); ok {
			log.ToggleLogging(disableLogging)
			config.LoggingDisabled = disableLogging
		}
	})

	updateField("cacheTTL", func(value interface{}) {
		if ttl, ok := value.(float64); ok {
			config.CacheTTL = time.Duration(ttl) * time.Second
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
			config.StatisticsRetention = int(days)
		}
	})
	config.Save()
}
