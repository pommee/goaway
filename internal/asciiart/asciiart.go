package asciiart

import (
	"fmt"
	"goaway/internal/server"
	"strings"
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

func AsciiArt(config server.ServerConfig, blockedDomains int, version string) {

	adminPanelURL := fmt.Sprintf("http://localhost:%d/index.html", config.WebsitePort)

	if version == "0.0.0" {
		version = "[DEV]"
	}
	const versionSpace = 7
	leftPadding := (versionSpace - len(version)) / 2
	rightPadding := versionSpace - len(version) - leftPadding
	versionPrinted := fmt.Sprintf("%s%s%s%s%s", strings.Repeat(" ", leftPadding), Cyan, version, Reset, strings.Repeat(" ", rightPadding))
	fmt.Printf("\n\n   __ _  ___   __ ___      ____ _ _   _   DNS port         %s%d%s\n"+
		"  / _` |/ _ \\ / _` \\ \\ /\\ / / _` | | | |  Upstream         %s%s%s\n"+
		" | (_| | (_) | (_| |\\ V  V / (_| | |_| |  Blacklist        %s%s%s\n"+
		"  \\__, |\\___/ \\__,_| \\_/\\_/ \\__,_|\\__, |  Cache TTL:       %s%s%s\n"+
		"   __/ |                           __/ |  Admin Panel:     %s%s%s\n"+
		"  |___/          %s          |___/   Blocked Domains: %s%d%s\n\n\n",
		Green, config.Port, Reset,
		Cyan, config.UpstreamDNS, Reset,
		Yellow, config.BlacklistPath, Reset,
		Blue, config.CacheTTL, Reset,
		Magenta, adminPanelURL, Reset,
		versionPrinted, Red, blockedDomains, Reset)
}
