package services

import (
	"crypto/tls"
	"fmt"
	"goaway/backend/database"
	"goaway/backend/dns/server"
	"goaway/backend/settings"

	"gorm.io/gorm"
)

type AppContext struct {
	Config      *settings.Config
	DBConn      *gorm.DB
	Certificate tls.Certificate
	DNSServer   *server.DNSServer
}

func NewAppContext(config *settings.Config) (*AppContext, error) {
	ctx := &AppContext{Config: config}
	if err := ctx.initialize(); err != nil {
		return nil, err
	}

	return ctx, nil
}

func (ctx *AppContext) initialize() error {
	ctx.DBConn = database.Initialize()

	cert, err := ctx.Config.GetCertificate()
	if err != nil {
		return fmt.Errorf("failed to get certificate: %w", err)
	}
	ctx.Certificate = cert

	dnsServer, err := server.NewDNSServer(
		ctx.Config,
		ctx.DBConn,
		cert,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize DNS server: %w", err)
	}
	ctx.DNSServer = dnsServer

	go dnsServer.ProcessLogEntries()

	return nil
}
