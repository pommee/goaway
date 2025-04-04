package setup

import (
	"goaway/internal/logging"
	"goaway/internal/settings"
	"os"

	"github.com/Masterminds/semver"
)

var log = logging.GetLogger()

type Flags struct {
	DnsPort             int
	WebserverPort       int
	LogLevel            int
	StatisticsRetention int
	DisableLogging      bool
	DisableAuth         bool
	DevMode             bool
}

func UpdateConfig(config *settings.Config, flags *Flags) {
	config.DNSServer.Port = flags.DnsPort
	config.DNSServer.LoggingDisabled = flags.DisableLogging
	config.DNSServer.StatisticsRetention = flags.StatisticsRetention
	config.APIServer.Port = flags.WebserverPort
	config.APIServer.Authentication = flags.DisableAuth
	config.LogLevel = logging.LogLevel(flags.LogLevel)
	config.DevMode = flags.DevMode
	log.SetLevel(logging.LogLevel(flags.LogLevel))
}

func GetVersionOrDefault(ver string) *semver.Version {
	versionObj, err := semver.NewVersion(ver)
	if err != nil {
		versionObj, _ = semver.NewVersion("0.0.0")
	}
	return versionObj
}

func InitializeSettings(flags *Flags) settings.Config {
	config, err := settings.LoadSettings()
	if err != nil {
		log.Error("Failed to load settings: %s", err)
		os.Exit(1)
	}

	UpdateConfig(&config, flags)
	config.Save()

	return config
}
