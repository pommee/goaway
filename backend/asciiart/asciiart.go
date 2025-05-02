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

func AsciiArt(config settings.Config, blockedDomains int, version string, disableAuth bool) {
	const versionSpace = 7

	versionFormatted := fmt.Sprintf("%-*s%s%s%-*s%s", (versionSpace-len(version))/2, "", Cyan, version, (versionSpace-len(version)+1)/2, "", Reset)
	portFormatted := fmt.Sprintf("%s%d%s", Green, config.DNSServer.Port, Reset)
	adminPanelFormatted := fmt.Sprintf("%s%d%s", Red, config.APIServer.Port, Reset)
	upstreamFormatted := fmt.Sprintf("%s%s%s", Cyan, config.DNSServer.PreferredUpstream, Reset)
	authFormatted := fmt.Sprintf("%s%v%s", Yellow, disableAuth, Reset)
	cacheTTLFormatted := fmt.Sprintf("%s%s%s", Blue, config.DNSServer.CacheTTL, Reset)
	blockedDomainsFormatted := fmt.Sprintf("%s%d%s", Magenta, blockedDomains, Reset)

	fmt.Printf(`
   __ _  ___   __ ___      ____ _ _   _   DNS port:         %s
  / _' |/ _ \ / _' \ \ /\ / / _' | | | |  Web port:         %s
 | (_| | (_) | (_| |\ V  V / (_| | |_| |  Upstream:         %s
  \__, |\___/ \__,_| \_/\_/ \__,_|\__, |  Authentication:   %s
   __/ |                           __/ |  Cache TTL:        %s
  |___/          %s          |___/   Blocked Domains:  %s

`, portFormatted, adminPanelFormatted, upstreamFormatted, authFormatted, cacheTTLFormatted, versionFormatted, blockedDomainsFormatted)
}
