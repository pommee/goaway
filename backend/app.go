package app

import (
	"embed"
	"fmt"
	"goaway/backend/alert"
	"goaway/backend/api/key"
	"goaway/backend/asciiart"
	"goaway/backend/audit"
	"goaway/backend/lifecycle"
	"goaway/backend/logging"
	"goaway/backend/mac"
	"goaway/backend/prefetch"
	"goaway/backend/resolution"
	"goaway/backend/services"
	"goaway/backend/settings"
	"goaway/backend/setup"
	"goaway/backend/user"
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

	dbConn := a.context.DBManager.Conn
	resolutionService := resolution.NewService(resolution.NewRepository(dbConn))
	macService := mac.NewService(mac.NewRepository(dbConn))
	userService := user.NewService(user.NewRepository(dbConn))
	keyService := key.NewService(key.NewRepository(dbConn))
	auditService := audit.NewService(audit.NewRepository(dbConn))
	alertService := alert.NewService(alert.NewRepository(dbConn))
	prefetchService := prefetch.NewService(prefetch.NewRepository(dbConn), a.context.DNSServer)

	a.context.DNSServer.ResolutionService = resolutionService
	a.context.DNSServer.MACService = macService
	a.context.DNSServer.AuditService = auditService
	a.context.DNSServer.AlertService = alertService

	a.displayStartupInfo()

	a.services = services.NewServiceRegistry(a.context, a.version, a.commit, a.date, a.content)
	a.services.PrefetchService = prefetchService
	a.lifecycle = lifecycle.NewManager(a.services)

	runServices := a.lifecycle.Run()
	a.services.APIServer.ResolutionService = resolutionService
	a.services.APIServer.UserService = userService
	a.services.APIServer.KeyService = keyService
	return runServices
}

func (a *Application) displayStartupInfo() {
	domains, err := a.context.Blacklist.CountDomains()
	if err != nil {
		log.Warning("Failed to count blacklist domains: %v", err)
	}

	currentVersion := setup.GetVersionOrDefault(a.version)
	asciiart.AsciiArt(
		a.config,
		domains,
		currentVersion.Original(),
		a.config.API.Authentication,
	)
}
