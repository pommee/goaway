package setup

import (
	"fmt"
	"goaway/backend/logging"
	"goaway/backend/settings"
	"os"

	"github.com/Masterminds/semver"
)

var log = logging.GetLogger()

type SetFlags struct {
	DnsPort             *int
	WebserverPort       *int
	LogLevel            *int
	StatisticsRetention *int
	LoggingEnabled      *bool
	Authentication      *bool
	DevMode             *bool
	Ansi                *bool
	JSON                *bool
}

func UpdateConfig(config *settings.Config, flags *SetFlags) {
	if flags.JSON != nil && *flags.JSON {
		falseVal := false
		flags.Ansi = &falseVal
	}
	if flags.LogLevel != nil {
		if *flags.LogLevel > 3 || *flags.LogLevel < 0 {
			fmt.Println("Flag --log-level can't be greater than 3 or below 0.")
			os.Exit(1)
		}
	}
	if flags.DnsPort != nil {
		config.DNS.Port = *flags.DnsPort
	}
	if flags.WebserverPort != nil {
		config.API.Port = *flags.WebserverPort
	}
	if flags.StatisticsRetention != nil {
		config.StatisticsRetention = *flags.StatisticsRetention
	}
	if flags.Authentication != nil {
		config.API.Authentication = *flags.Authentication
	}
	if flags.DevMode != nil {
		config.DevMode = *flags.DevMode
	}
	if flags.LoggingEnabled != nil {
		config.LoggingEnabled = *flags.LoggingEnabled
	}
	if flags.LogLevel != nil {
		config.LogLevel = logging.LogLevel(*flags.LogLevel)
	}
	if flags.JSON != nil {
		log.JSON = *flags.JSON
	}
	if flags.Ansi != nil {
		log.Ansi = *flags.Ansi
	}
	if flags.LogLevel != nil {
		log.SetLevel(logging.LogLevel(*flags.LogLevel))
	}
	if flags.LoggingEnabled != nil {
		log.ToggleLogging(*flags.LoggingEnabled)
	}
}

func GetVersionOrDefault(ver string) *semver.Version {
	versionObj, err := semver.NewVersion(ver)
	if err != nil {
		versionObj, _ = semver.NewVersion("0.0.0")
	}
	return versionObj
}

func InitializeSettings(flags *SetFlags) *settings.Config {
	config, err := settings.LoadSettings()
	if err != nil {
		log.Error("Failed to load settings: %s", err)
		os.Exit(1)
	}

	UpdateConfig(&config, flags)

	return &config
}
