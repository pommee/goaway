package settings

import (
	"crypto/tls"
	"fmt"
	"goaway/backend/logging"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

var log = logging.GetLogger()

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
		return config, nil
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

func (config *Config) Update(updatedSettings Config) {
	config.API.Port = updatedSettings.API.Port
	config.API.Authentication = updatedSettings.API.Authentication
	config.API.RateLimit = updatedSettings.API.RateLimit

	config.DNS.Address = updatedSettings.DNS.Address
	config.DNS.Ports = updatedSettings.DNS.Ports
	config.DNS.UDPSize = updatedSettings.DNS.UDPSize
	config.DNS.CacheTTL = updatedSettings.DNS.CacheTTL
	config.DNS.TLS = updatedSettings.DNS.TLS
	config.DNS.Upstream = updatedSettings.DNS.Upstream

	config.Logging = updatedSettings.Logging
	config.Misc = updatedSettings.Misc

	log.ToggleLogging(config.Logging.Enabled)
	log.SetLevel(logging.LogLevel(config.Logging.Level))

	config.Save()
}

func createDefaultSettings(filePath string) (Config, error) {
	defaultConfig := Config{
		DNS: DNSConfig{
			Address:  "0.0.0.0",
			Gateway:  getDefaultGateway(),
			CacheTTL: 3600,
			UDPSize:  512,
			TLS: TLSConfig{
				Enabled: false,
				Cert:    "",
				Key:     "",
			},
			Upstream: UpstreamConfig{
				Preferred: "8.8.8.8:53",
				Fallback: []string{
					"1.1.1.1:53",
				},
			},
			Ports: PortsConfig{
				TCPUDP: getEnvAsIntWithDefault("DNS_PORT", 53),
				DoT:    getEnvAsIntWithDefault("DOT_PORT", 853),
				DoH:    getEnvAsIntWithDefault("DOH_PORT", 443),
			},
		},
		API: APIConfig{
			Port:           getEnvAsIntWithDefault("WEBSITE_PORT", 8080),
			Authentication: true,
			RateLimit: RateLimitConfig{
				Enabled:  true,
				MaxTries: 5,
				Window:   5,
			},
		},
		Logging: LoggingConfig{
			Enabled: true,
			Level:   int(logging.INFO),
		},
		Misc: MiscConfig{
			InAppUpdate:               false,
			StatisticsRetention:       7,
			Dashboard:                 true,
			ScheduledBlacklistUpdates: true,
		},
	}

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

func getEnvAsIntWithDefault(envVariable string, defaultValue int) int {
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

func (config *Config) GetCertificate() (tls.Certificate, error) {
	if config.DNS.TLS.Enabled && config.DNS.TLS.Cert != "" && config.DNS.TLS.Key != "" {
		cert, err := tls.LoadX509KeyPair(config.DNS.TLS.Cert, config.DNS.TLS.Key)
		if err != nil {
			return tls.Certificate{}, fmt.Errorf("failed to load TLS certificate: %w", err)
		}

		return cert, nil
	}

	return tls.Certificate{}, nil
}

func getDefaultGateway() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "192.168.0.1:53"
	}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	if localAddr.IP.IsPrivate() {
		ip := localAddr.IP.To4()
		if ip != nil {
			return fmt.Sprintf("%d.%d.%d.1:53", ip[0], ip[1], ip[2])
		}
	}

	return "192.168.0.1:53"
}
