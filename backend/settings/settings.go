package settings

import (
	"fmt"
	"goaway/backend/api/ratelimit"
	"goaway/backend/logging"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

var log = logging.GetLogger()

type Status struct {
	Paused    bool      `json:"paused"`
	PausedAt  time.Time `json:"pausedAt"`
	PauseTime int       `json:"pauseTime"`
}

type DNSConfig struct {
	Address           string   `yaml:"address" json:"address"`
	Port              int      `yaml:"port" json:"port"`
	DoTPort           int      `yaml:"dotPort" json:"dotPort"`
	DoHPort           int      `yaml:"dohPort" json:"dohPort"`
	CacheTTL          int      `yaml:"cacheTTL" json:"cacheTTL"`
	PreferredUpstream string   `yaml:"preferredUpstream" json:"preferredUpstream"`
	UpstreamDNS       []string `yaml:"upstreamDNS" json:"upstreamDNS"`
	UDPSize           int      `yaml:"udpSize" json:"udpSize"`
	Status            Status   `yaml:"-" json:"status"`

	TLSCertFile string `yaml:"tlsCertFile" json:"tlsCertFile"`
	TLSKeyFile  string `yaml:"tlsKeyFile" json:"tlsKeyFile"`
}

type APIConfig struct {
	Port              int                         `yaml:"port" json:"port"`
	Authentication    bool                        `yaml:"authentication" json:"authentication"`
	RateLimiterConfig ratelimit.RateLimiterConfig `yaml:"rateLimit" json:"-"`
}

type Config struct {
	DNS DNSConfig `yaml:"dns" json:"dns"`
	API APIConfig `yaml:"api" json:"api"`

	Dashboard                 bool             `yaml:"dashboard" json:"-"`
	ScheduledBlacklistUpdates bool             `yaml:"scheduledBlacklistUpdates" json:"scheduledBlacklistUpdates"`
	StatisticsRetention       int              `yaml:"statisticsRetention" json:"statisticsRetention"`
	LoggingEnabled            bool             `yaml:"loggingEnabled" json:"loggingEnabled"`
	LogLevel                  logging.LogLevel `yaml:"logLevel" json:"logLevel"`
	InAppUpdate               bool             `yaml:"inAppUpdate" json:"inAppUpdate"`

	// settings not visible in config file
	BinaryPath string `yaml:"-" json:"-"`
}

func LoadSettings() (Config, error) {
	var config Config

	path, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf("could not determine current directory: %w", err)
	}
	path = filepath.Join(path, "config", "settings.yaml")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Info("Settings file not found, creating from defaults...")
		config, err = createDefaultSettings(path)
		if err != nil {
			return Config{}, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("could not read settings file: %w", err)
	}

	type configWithPtr struct {
		DNS                       DNSConfig        `yaml:"dns" json:"dns"`
		API                       APIConfig        `yaml:"api" json:"api"`
		Dashboard                 *bool            `yaml:"dashboard" json:"-"`
		ScheduledBlacklistUpdates *bool            `yaml:"scheduledBlacklistUpdates" json:"scheduledBlacklistUpdates"`
		StatisticsRetention       int              `yaml:"statisticsRetention" json:"statisticsRetention"`
		LoggingEnabled            bool             `yaml:"loggingEnabled" json:"loggingEnabled"`
		LogLevel                  logging.LogLevel `yaml:"logLevel" json:"logLevel"`
		InAppUpdate               bool             `yaml:"inAppUpdate" json:"inAppUpdate"`
	}

	var temp configWithPtr
	if err := yaml.Unmarshal(data, &temp); err != nil {
		return Config{}, fmt.Errorf("invalid settings format: %w", err)
	}

	config.DNS = temp.DNS
	config.API = temp.API
	config.StatisticsRetention = temp.StatisticsRetention
	config.LoggingEnabled = temp.LoggingEnabled
	config.LogLevel = temp.LogLevel
	config.InAppUpdate = temp.InAppUpdate

	if temp.Dashboard == nil {
		// true by default if the Dashboard field was not found in settings.yaml
		config.Dashboard = true
	} else {
		config.Dashboard = *temp.Dashboard
	}

	if temp.ScheduledBlacklistUpdates == nil {
		// false by default if the ScheduledBlacklistUpdates field was not found in settings.yaml
		config.ScheduledBlacklistUpdates = false
	} else {
		config.ScheduledBlacklistUpdates = *temp.ScheduledBlacklistUpdates
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
		InAppUpdate:         false,
	}

	defaultConfig.DNS.Address = "0.0.0.0"
	defaultConfig.DNS.Port = GetEnvAsIntWithDefault("DNS_PORT", 53)
	defaultConfig.DNS.DoTPort = GetEnvAsIntWithDefault("DOT_PORT", 853)
	defaultConfig.DNS.DoHPort = GetEnvAsIntWithDefault("DOH_PORT", 443)
	defaultConfig.DNS.CacheTTL = 3600
	defaultConfig.DNS.PreferredUpstream = "8.8.8.8:53"
	defaultConfig.DNS.UpstreamDNS = []string{
		"1.1.1.1:53",
		"8.8.8.8:53",
	}
	defaultConfig.DNS.UDPSize = 512
	defaultConfig.DNS.TLSCertFile = ""
	defaultConfig.DNS.TLSKeyFile = ""

	defaultConfig.Dashboard = true
	defaultConfig.ScheduledBlacklistUpdates = false
	defaultConfig.API.Port = GetEnvAsIntWithDefault("WEBSITE_PORT", 8080)
	defaultConfig.API.Authentication = true
	defaultConfig.API.RateLimiterConfig = ratelimit.RateLimiterConfig{Enabled: true, MaxTries: 5, Window: 5}

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

	if err := os.WriteFile("./config/settings.yaml", data, 0644); err != nil {
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

func GetEnvAsIntWithDefault(envVariable string, defaultValue int) int {
	val, found := os.LookupEnv(envVariable)
	if !found {
		return defaultValue
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}

	return intVal
}
