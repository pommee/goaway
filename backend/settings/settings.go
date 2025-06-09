package settings

import (
	"fmt"
	"goaway/backend/logging"
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
		Address           string   `yaml:"address" json:"address"`
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
	LoggingEnabled      bool             `yaml:"loggingEnabled" json:"loggingEnabled"`
	LogLevel            logging.LogLevel `yaml:"logLevel" json:"logLevel"`

	// settings not visible in config file
	DevMode    bool   `yaml:"-" json:"-"`
	BinaryPath string `yaml:"-" json:"-"`
}

func LoadSettings() (Config, error) {
	var config Config

	path, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf("could not determine current directory: %w", err)
	}
	path = filepath.Join(path, "settings.yaml")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Info("Settings file not found, creating from defaults...")
		config, err = createDefaultSettings(path)
		if err != nil {
			return Config{}, fmt.Errorf("failed to create default settings: %w", err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("could not read settings file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("invalid settings format: %w", err)
	}

	binaryPath, err := os.Executable()
	if err != nil {
		log.Warning("Unable to find installed binary path, err: %v", err)
	}
	config.BinaryPath = binaryPath

	return config, nil
}

func createDefaultSettings(filePath string) (Config, error) {
	defaultConfig := Config{
		StatisticsRetention: 7,
		LoggingEnabled:      true,
		LogLevel:            logging.INFO,
	}

	defaultConfig.DNS.Address = "0.0.0.0"
	defaultConfig.DNS.Port = 53
	defaultConfig.DNS.CacheTTL = 3600
	defaultConfig.DNS.PreferredUpstream = "8.8.8.8:53"
	defaultConfig.DNS.UpstreamDNS = []string{
		"1.1.1.1:53",
		"8.8.8.8:53",
	}

	defaultConfig.API.Port = 8080
	defaultConfig.API.Authentication = true

	data, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		return Config{}, fmt.Errorf("failed to marshal default config: %w", err)
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return Config{}, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return Config{}, fmt.Errorf("failed to create default settings file: %w", err)
	}

	log.Info("Default settings file created at: %s", filePath)
	return defaultConfig, nil
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
	config.LoggingEnabled = updatedSettings.LoggingEnabled

	log.ToggleLogging(config.LoggingEnabled)
	log.SetLevel(config.LogLevel)
	config.Save()
}
