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

	fmt.Printf("\n   __ _  ___   __ ___      ____ _ _   _      DNS port         %s%d%s\n", Green, config.Port, Reset)
	fmt.Printf("  / _` |/ _ \\ / _` \\ \\ /\\ / / _` | | | |     Upstream         %s%s%s\n", Cyan, config.UpstreamDNS, Reset)
	fmt.Printf(" | (_| | (_) | (_| |\\ V  V / (_| | |_| |     Blacklist        %s%s%s\n", Yellow, config.BlacklistPath, Reset)
	fmt.Printf("  \\__, |\\___/ \\__,_| \\_/\\_/ \\__,_|\\__, |     Cache TTL:       %s%s%s\n", Blue, config.CacheTTL, Reset)
	fmt.Printf("   __/ |                           __/ |     Admin Panel:     %s%s%s\n", Magenta, adminPanelURL, Reset)
	fmt.Printf("  |___/          %s          |___/      Blocked Domains: %s%d%s\n", versionPrinted, Red, blockedDomains, Reset)
	fmt.Println()
}
