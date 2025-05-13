package setup

import (
	"fmt"
	"goaway/backend/logging"
	"goaway/backend/settings"
	"os"

	"github.com/Masterminds/semver"
)

var log = logging.GetLogger()

type Flags struct {
	DnsPort             int
	WebserverPort       int
	LogLevel            int
	StatisticsRetention int
	LoggingEnabled      bool
	Authentication      bool
	DevMode             bool
	Ansi                bool
	JSON                bool
}

func UpdateConfig(config *settings.Config, flags *Flags) {

	if flags.JSON {
		flags.Ansi = false
	}
	if flags.LogLevel > 3 || flags.LogLevel < 0 {
		fmt.Println("Flag --log-level can't be greater than 3 or below 0.")
		os.Exit(1)
	}

	config.DNS.Port = flags.DnsPort
	config.API.Port = flags.WebserverPort
	config.StatisticsRetention = flags.StatisticsRetention
	config.API.Authentication = flags.Authentication
	config.DevMode = flags.DevMode
	config.LoggingEnabled = flags.LoggingEnabled
	config.LogLevel = logging.LogLevel(flags.LogLevel)

	log.JSON = flags.JSON
	log.Ansi = flags.Ansi
	log.SetLevel(logging.LogLevel(flags.LogLevel))
	log.ToggleLogging(config.LoggingEnabled)
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
