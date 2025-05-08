package settings

import (
	"fmt"
	"goaway/backend/logging"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

var log = logging.GetLogger()

type Status struct {
	Paused    bool      `json:"paused"`
	PausedAt  time.Time `json:"pausedAt"`
	PauseTime int       `json:"pauseTime"`
}

type Config struct {
	DNS struct {
		Port              int      `yaml:"port" json:"port"`
		CacheTTL          int      `yaml:"cacheTTL" json:"cacheTTL"`
		PreferredUpstream string   `yaml:"preferredUpstream" json:"preferredUpstream"`
		UpstreamDNS       []string `yaml:"upstreamDNS" json:"upstreamDNS"`
		Status            Status   `yaml:"-" json:"status"`
	} `yaml:"dns" json:"dns"`

	API struct {
		Port           int  `yaml:"port" json:"port"`
		Authentication bool `yaml:"authentication" json:"authentication"`
	} `yaml:"api" json:"api"`

	StatisticsRetention int              `yaml:"statisticsRetention" json:"statisticsRetention"`
	DevMode             bool             `yaml:"-" json:"-"`
	LoggingDisabled     bool             `yaml:"loggingDisabled" json:"loggingDisabled"`
	LogLevel            logging.LogLevel `yaml:"logLevel" json:"logLevel"`
}

func LoadSettings() (Config, error) {
	var config Config

	path, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf("could not determine current directory: %w", err)
	}
	path = filepath.Join(path, "settings.yaml")

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

	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("invalid settings format: %w", err)
	}

	return config, nil
}

func fetchAndSaveSettings(filePath string) error {
	url := "https://raw.githubusercontent.com/pommee/goaway/refs/heads/main/settings.yaml"

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch settings.yaml: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch settings.yaml: HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
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
	data, err := yaml.Marshal(config)
	if err != nil {
		log.Error("Could not parse settings %v", err)
		return
	}

	if err := os.WriteFile("./settings.yaml", data, 0644); err != nil {
		log.Error("Could not save settings %v", err)
	}
}

func (config *Config) UpdateSettings(updatedSettings Config) {

	config.DNS.CacheTTL = updatedSettings.DNS.CacheTTL
	config.LogLevel = updatedSettings.LogLevel
	config.StatisticsRetention = updatedSettings.StatisticsRetention
	config.LoggingDisabled = updatedSettings.LoggingDisabled

	log.ToggleLogging(config.LoggingDisabled)
	log.SetLevel(config.LogLevel)
	config.Save()
}
