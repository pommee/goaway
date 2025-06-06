package asciiart

import (
	"fmt"
	"goaway/backend/settings"
)

const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Cyan    = "\033[36m"
	Magenta = "\033[35m"
)

func AsciiArt(config *settings.Config, blockedDomains int, version string, disableAuth bool, ansi bool) {
	const versionSpace = 7

	colorize := func(color, text string) string {
		if !ansi {
			return text
		}
		return color + text + Reset
	}

	versionFormatted := fmt.Sprintf("%-*s%s%-*s", (versionSpace-len(version))/2, "",
		colorize(Cyan, version), (versionSpace-len(version)+1)/2, "")

	portFormatted := colorize(Green, fmt.Sprintf("%d", config.DNS.Port))
	adminPanelFormatted := colorize(Red, fmt.Sprintf("%d", config.API.Port))
	upstreamFormatted := colorize(Cyan, config.DNS.PreferredUpstream)
	authFormatted := colorize(Yellow, fmt.Sprintf("%v", disableAuth))
	cacheTTLFormatted := colorize(Blue, fmt.Sprintf("%d", config.DNS.CacheTTL))
	blockedDomainsFormatted := colorize(Magenta, fmt.Sprintf("%d", blockedDomains))

	fmt.Printf(`
   __ _  ___   __ ___      ____ _ _   _   DNS port:         %s
  / _' |/ _ \ / _' \ \ /\ / / _' | | | |  Web port:         %s
 | (_| | (_) | (_| |\ V  V / (_| | |_| |  Upstream:         %s
  \__, |\___/ \__,_| \_/\_/ \__,_|\__, |  Authentication:   %s
   __/ |                           __/ |  Cache TTL:        %s
  |___/          %s          |___/   Blocked Domains:  %s

`, portFormatted, adminPanelFormatted, upstreamFormatted, authFormatted, cacheTTLFormatted, versionFormatted, blockedDomainsFormatted)
}
