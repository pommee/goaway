package setup

import (
	"fmt"
	"goaway/backend/logging"
	"goaway/backend/settings"
	"os"
	"strconv"

	"github.com/Masterminds/semver"
)

var log = logging.GetLogger()

type SetFlags struct {
	DnsPort             *int
	DoTPort             *int
	DoHPort             *int
	WebserverPort       *int
	LogLevel            *int
	StatisticsRetention *int
	LoggingEnabled      *bool
	Authentication      *bool
	Dashboard           *bool
	Ansi                *bool
	JSON                *bool
	InAppUpdate         *bool
}

func UpdateConfig(config *settings.Config, flags *SetFlags) {
	if flags.JSON != nil && *flags.JSON {
		falseVal := false
		flags.Ansi = &falseVal
	}

	if flags.LogLevel != nil && (*flags.LogLevel > 3 || *flags.LogLevel < 0) {
		fmt.Println("Flag --log-level can't be greater than 3 or below 0.")
		os.Exit(1)
	}
	if flags.DnsPort != nil || os.Getenv("DNS_PORT") != "" {
		if port, found := os.LookupEnv("DNS_PORT"); found {
			dnsPort, err := strconv.Atoi(port)
			if err != nil {
				log.Fatal("Could not parse DNS_PORT environment variable")
			}
			config.DNS.Port = dnsPort
		} else {
			config.DNS.Port = *flags.DnsPort
		}
	}
	if flags.DoTPort != nil || os.Getenv("DOT_PORT") != "" {
		if port, found := os.LookupEnv("DOT_PORT"); found {
			dotPort, err := strconv.Atoi(port)
			if err != nil {
				log.Fatal("Could not parse DOT_PORT environment variable")
			}
			config.DNS.DoTPort = dotPort
		} else {
			config.DNS.DoTPort = *flags.DoTPort
		}
	}
	if flags.DoHPort != nil || os.Getenv("DOH_PORT") != "" {
		if port, found := os.LookupEnv("DOH_PORT"); found {
			dohPort, err := strconv.Atoi(port)
			if err != nil {
				log.Fatal("Could not parse DOH_PORT environment variable")
			}
			config.DNS.DoHPort = dohPort
		} else {
			config.DNS.DoHPort = *flags.DoHPort
		}
	}
	if flags.WebserverPort != nil || os.Getenv("WEBSITE_PORT") != "" {
		if port, found := os.LookupEnv("WEBSITE_PORT"); found {
			websitePort, err := strconv.Atoi(port)
			if err != nil {
				log.Fatal("Could not parse WEBSITE_PORT environment variable")
			}
			config.API.Port = websitePort
		} else {
			config.API.Port = *flags.WebserverPort
		}
	}
	if flags.StatisticsRetention != nil {
		config.StatisticsRetention = *flags.StatisticsRetention
	}
	if flags.Authentication != nil {
		config.API.Authentication = *flags.Authentication
	}
	if flags.Dashboard != nil {
		config.Dashboard = *flags.Dashboard
	}
	if flags.LoggingEnabled != nil {
		config.LoggingEnabled = *flags.LoggingEnabled
	}
	if flags.LogLevel != nil {
		config.LogLevel = logging.LogLevel(*flags.LogLevel)
	}
	if flags.InAppUpdate != nil {
		config.InAppUpdate = *flags.InAppUpdate
	}

	if flags.JSON != nil {
		log.JSON = *flags.JSON
		log.SetJson(log.JSON)
	} else {
		log.Ansi = flags.Ansi == nil || *flags.Ansi
		log.SetAnsi(log.Ansi)
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
	config.Save()

	return &config
}
