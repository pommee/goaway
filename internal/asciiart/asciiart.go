package asciiart

import (
	"fmt"
	"goaway/internal/server"
	"log"
	"net"
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

func AsciiArt(config *server.ServerConfig, blockedDomains int, version string, disableAuth bool) {

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	if version == "0.0.0" {
		version = "[DEV]"
	}

	adminPanelURL := fmt.Sprintf("http://%s:%d/index.html", localAddr.IP, config.WebsitePort)

	const versionSpace = 7
	leftPadding := (versionSpace - len(version)) / 2
	rightPadding := versionSpace - len(version) - leftPadding
	versionPrinted := fmt.Sprintf("%s%s%s%s%s", strings.Repeat(" ", leftPadding), Cyan, version, Reset, strings.Repeat(" ", rightPadding))
	fmt.Printf("\n\n   __ _  ___   __ ___      ____ _ _   _   DNS port         %s%d%s\n"+
		"  / _` |/ _ \\ / _` \\ \\ /\\ / / _` | | | |  Upstream         %s%s%s\n"+
		" | (_| | (_) | (_| |\\ V  V / (_| | |_| |  Authentication   %s%v%s\n"+
		"  \\__, |\\___/ \\__,_| \\_/\\_/ \\__,_|\\__, |  Cache TTL:       %s%s%s\n"+
		"   __/ |                           __/ |  Blocked Domains: %s%d%s\n"+
		"  |___/          %s          |___/   Admin Panel:     %s%s%s\n\n\n",
		Green, config.Port, Reset,
		Cyan, config.PreferredUpstream, Reset,
		Yellow, disableAuth, Reset,
		Blue, config.CacheTTL, Reset,
		Magenta, blockedDomains, Reset,
		versionPrinted, Red, adminPanelURL, Reset)
}
