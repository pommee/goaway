package asciiart

import (
	"fmt"
	"goaway/internal/server"
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

func AsciiArt(config *server.ServerConfig, blockedDomains int, version string, disableAuth bool) {
	const versionSpace = 7

	if version == "0.0.0" {
		version = "[DEV]"
	}
	versionFormatted := fmt.Sprintf("%-*s%s%s%-*s%s", (versionSpace-len(version))/2, "", Cyan, version, (versionSpace-len(version)+1)/2, "", Reset)

	adminPanelURL := fmt.Sprintf("http://localhost:%d/index.html", config.WebsitePort)
	portFormatted := fmt.Sprintf("%s%d%s", Green, config.Port, Reset)
	upstreamFormatted := fmt.Sprintf("%s%s%s", Cyan, config.PreferredUpstream, Reset)
	authFormatted := fmt.Sprintf("%s%v%s", Yellow, disableAuth, Reset)
	cacheTTLFormatted := fmt.Sprintf("%s%s%s", Blue, config.CacheTTL, Reset)
	blockedDomainsFormatted := fmt.Sprintf("%s%d%s", Magenta, blockedDomains, Reset)
	adminPanelFormatted := fmt.Sprintf("%s%s%s", Red, adminPanelURL, Reset)

	fmt.Printf(`
   __ _  ___   __ ___      ____ _ _   _   DNS port         %s
  / _' |/ _ \ / _' \ \ /\ / / _' | | | |  Upstream         %s
 | (_| | (_) | (_| |\ V  V / (_| | |_| |  Authentication   %s
  \__, |\___/ \__,_| \_/\_/ \__,_|\__, |  Cache TTL:       %s
   __/ |                           __/ |  Blocked Domains: %s
  |___/          %s          |___/   Admin Panel:     %s

`, portFormatted, upstreamFormatted, authFormatted, cacheTTLFormatted, blockedDomainsFormatted, versionFormatted, adminPanelFormatted)
}
