package cmd

import (
	"goaway/backend/setup"

	"github.com/spf13/cobra"
)

type Flags struct {
	DnsPort             int
	DoTPort             int
	DoHPort             int
	WebserverPort       int
	LogLevel            int
	StatisticsRetention int
	LoggingEnabled      bool
	Authentication      bool
	Dashboard           bool
	Ansi                bool
	JSON                bool
	InAppUpdate         bool
}

func NewFlags() *Flags {
	return &Flags{}
}

func (f *Flags) Register(cmd *cobra.Command) {
	cmd.Flags().IntVar(&f.DnsPort, "dns-port", 53, "Port for the DNS server")
	cmd.Flags().IntVar(&f.DoTPort, "dot-port", 853, "Port for the DoT (DNS-over-TCP) server")
	cmd.Flags().IntVar(&f.DoHPort, "doh-port", 443, "Port for the DoH (DNS-over-HTTPS) server")
	cmd.Flags().IntVar(&f.WebserverPort, "webserver-port", 8080, "Port for the web server")
	cmd.Flags().IntVar(&f.LogLevel, "log-level", 1, "0 = DEBUG | 1 = INFO | 2 = WARNING | 3 = ERROR")
	cmd.Flags().IntVar(&f.StatisticsRetention, "statistics-retention", 7, "Days to keep statistics")
	cmd.Flags().BoolVar(&f.LoggingEnabled, "logging", true, "Toggle logging")
	cmd.Flags().BoolVar(&f.Authentication, "auth", true, "Toggle authentication for admin dashboard")
	cmd.Flags().BoolVar(&f.Dashboard, "dashboard", true, "Serve dashboard")
	cmd.Flags().BoolVar(&f.Ansi, "ansi", true, "Toggle colorized logs. Only available in non-json formatted logs")
	cmd.Flags().BoolVar(&f.JSON, "json", false, "Toggle JSON formatted logs")
	cmd.Flags().BoolVar(&f.InAppUpdate, "in-app-update", false, "Toggle ability to update via dashboard")
}

func (f *Flags) GetSetFlags(cmd *cobra.Command) *setup.SetFlags {
	setFlags := &setup.SetFlags{}

	if cmd.Flags().Changed("dns-port") {
		setFlags.DNSPort = &f.DnsPort
	}
	if cmd.Flags().Changed("dot-port") {
		setFlags.DoTPort = &f.DoTPort
	}
	if cmd.Flags().Changed("doh-port") {
		setFlags.DoHPort = &f.DoHPort
	}
	if cmd.Flags().Changed("webserver-port") {
		setFlags.WebserverPort = &f.WebserverPort
	}
	if cmd.Flags().Changed("log-level") {
		setFlags.LogLevel = &f.LogLevel
	}
	if cmd.Flags().Changed("statistics-retention") {
		setFlags.StatisticsRetention = &f.StatisticsRetention
	}
	if cmd.Flags().Changed("logging") {
		setFlags.LoggingEnabled = &f.LoggingEnabled
	}
	if cmd.Flags().Changed("auth") {
		setFlags.Authentication = &f.Authentication
	}
	if cmd.Flags().Changed("dashboard") {
		setFlags.Dashboard = &f.Dashboard
	}
	if cmd.Flags().Changed("ansi") {
		setFlags.Ansi = &f.Ansi
	}
	if cmd.Flags().Changed("json") {
		setFlags.JSON = &f.JSON
	}
	if cmd.Flags().Changed("in-app-update") {
		setFlags.InAppUpdate = &f.InAppUpdate
	}

	return setFlags
}
