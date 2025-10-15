package services

import (
	"crypto/tls"
	"fmt"
	"goaway/backend/dns/database"
	"goaway/backend/dns/lists"
	"goaway/backend/dns/server"
	notification "goaway/backend/notifications"
	"goaway/backend/settings"
)

type AppContext struct {
	Config              *settings.Config
	DBManager           *database.Manager
	NotificationManager *notification.Manager
	Certificate         tls.Certificate
	DNSServer           *server.DNSServer
	Blacklist           *lists.Blacklist
	Whitelist           *lists.Whitelist
}

func NewAppContext(config *settings.Config) (*AppContext, error) {
	ctx := &AppContext{
		Config: config,
	}

	if err := ctx.initialize(); err != nil {
		return nil, err
	}

	return ctx, nil
}

func (ctx *AppContext) initialize() error {
	ctx.DBManager = database.Initialize()
	ctx.NotificationManager = notification.NewNotificationManager(ctx.DBManager)

	cert, err := ctx.Config.GetCertificate()
	if err != nil {
		return fmt.Errorf("failed to get certificate: %w", err)
	}
	ctx.Certificate = cert

	dnsServer, err := server.NewDNSServer(
		ctx.Config,
		ctx.DBManager,
		ctx.NotificationManager,
		cert,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize DNS server: %w", err)
	}
	ctx.DNSServer = dnsServer

	go dnsServer.ProcessLogEntries()

	blacklist, err := lists.InitializeBlacklist(ctx.DBManager)
	if err != nil {
		return fmt.Errorf("failed to initialize blacklist: %w", err)
	}
	ctx.Blacklist = blacklist
	ctx.DNSServer.Blacklist = blacklist

	ctx.Whitelist = ctx.DNSServer.Whitelist

	return nil
}
