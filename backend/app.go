package app

import (
	"context"
	"embed"
	"fmt"
	"goaway/backend/alert"
	"goaway/backend/api/key"
	"goaway/backend/audit"
	"goaway/backend/blacklist"
	"goaway/backend/lifecycle"
	"goaway/backend/logging"
	"goaway/backend/mac"
	"goaway/backend/notification"
	"goaway/backend/prefetch"
	"goaway/backend/request"
	"goaway/backend/resolution"
	"goaway/backend/services"
	"goaway/backend/settings"
	"goaway/backend/setup"
	"goaway/backend/user"
	"goaway/backend/whitelist"
)

var log = logging.GetLogger()

type Application struct {
	config    *settings.Config
	context   *services.AppContext
	services  *services.ServiceRegistry
	lifecycle *lifecycle.Manager
	content   embed.FS
	version   string
	commit    string
	date      string
}

func New(setFlags *setup.SetFlags, version, commit, date string, content embed.FS) *Application {
	config := setup.InitializeSettings(setFlags)

	return &Application{
		config:  config,
		version: version,
		commit:  commit,
		date:    date,
		content: content,
	}
}

func (a *Application) Start() error {

	ctx, err := services.NewAppContext(a.config)
	if err != nil {
		return fmt.Errorf("failed to initialize application context: %w", err)
	}
	a.context = ctx

	dbConn := a.context.DBConn
	alertService := alert.NewService(alert.NewRepository(dbConn))
	auditService := audit.NewService(audit.NewRepository(dbConn))
	blacklistService := blacklist.NewService(blacklist.NewRepository(dbConn))
	keyService := key.NewService(key.NewRepository(dbConn))
	macService := mac.NewService(mac.NewRepository(dbConn))
	notificationService := notification.NewService(notification.NewRepository(dbConn))
	prefetchService := prefetch.NewService(prefetch.NewRepository(dbConn), a.context.DNSServer)
	requestService := request.NewService(request.NewRepository(dbConn))
	resolutionService := resolution.NewService(resolution.NewRepository(dbConn))
	userService := user.NewService(user.NewRepository(dbConn))
	whitelistService := whitelist.NewService(whitelist.NewRepository(dbConn))

	a.context.DNSServer.AlertService = alertService
	a.context.DNSServer.AuditService = auditService
	a.context.DNSServer.BlacklistService = blacklistService
	a.context.DNSServer.MACService = macService
	a.context.DNSServer.NotificationService = notificationService
	a.context.DNSServer.RequestService = requestService
	a.context.DNSServer.UserService = userService
	a.context.DNSServer.ResolutionService = resolutionService
	a.context.DNSServer.WhitelistService = whitelistService

	a.displayStartupInfo()

	a.services = services.NewServiceRegistry(a.context, a.version, a.commit, a.date, a.content)
	a.services.ResolutionService = resolutionService
	a.services.BlacklistService = blacklistService
	a.services.NotificationService = notificationService
	a.services.PrefetchService = prefetchService
	a.services.RequestService = requestService
	a.services.UserService = userService
	a.services.KeyService = keyService
	a.services.WhitelistService = whitelistService
	a.lifecycle = lifecycle.NewManager(a.services)

	runServices := a.lifecycle.Run()
	return runServices
}

func (a *Application) displayStartupInfo() {
	domains, err := a.context.DNSServer.BlacklistService.CountDomains(context.Background())
	if err != nil {
		log.Warning("Failed to count blacklist domains: %v", err)
	}

	currentVersion := setup.GetVersionOrDefault(a.version)
	ASCIIArt(
		a.config,
		domains,
		currentVersion.Original(),
		a.config.API.Authentication,
	)
}
