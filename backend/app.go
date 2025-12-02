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
	"net/http"
	"sync"
	"time"
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

func (a *Application) RestartApplication() {
	log.Warning("Restarting application...")

	a.services.APIServer.IsShuttingDown = true

	var wg sync.WaitGroup
	shutdownErrors := make([]error, 0)
	var mu sync.Mutex

	wg.Go(func() {
		if err := a.services.APIServer.Stop(); err != nil {
			mu.Lock()
			shutdownErrors = append(shutdownErrors, fmt.Errorf("API server: %w", err))
			mu.Unlock()
		}
	})
	wg.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wg.Add(4)
	go func() {
		defer wg.Done()
		if err := a.services.UDPServer.Shutdown(); err != nil {
			mu.Lock()
			shutdownErrors = append(shutdownErrors, fmt.Errorf("UDP server: %w", err))
			mu.Unlock()
		}
		log.Warning("Stopped UDP server")
	}()

	go func() {
		defer wg.Done()
		if err := a.services.TCPServer.Shutdown(); err != nil {
			mu.Lock()
			shutdownErrors = append(shutdownErrors, fmt.Errorf("TCP server: %w", err))
			mu.Unlock()
		}
		log.Warning("Stopped TCP server")
	}()

	go func() {
		defer wg.Done()
		if a.services.DoHServer != nil {
			if err := a.services.DoHServer.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
				mu.Lock()
				shutdownErrors = append(shutdownErrors, fmt.Errorf("DoH server: %w", err))
				mu.Unlock()
			}
			log.Warning("Stopped DNS-over-HTTPS server")
		}
	}()

	go func() {
		defer wg.Done()
		if a.services.DoTServer != nil {
			if err := a.services.DoTServer.Shutdown(); err != nil {
				mu.Lock()
				shutdownErrors = append(shutdownErrors, fmt.Errorf("DoT server: %w", err))
				mu.Unlock()
			}
			log.Warning("Stopped DNS-over-TLS server")
		}
	}()

	wg.Wait()

	if len(shutdownErrors) > 0 {
		log.Warning("Shutdown completed with errors:")
		for _, err := range shutdownErrors {
			log.Error("  - %v", err)
		}
	} else {
		log.Info("All servers stopped successfully")
	}

	time.Sleep(500 * time.Millisecond)
	a.services.APIServer.IsShuttingDown = false

	err := a.Start()
	if err != nil {
		log.Fatal("Unable to restart, manual intervention required. Reason: %v", err)
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

	runServices := a.lifecycle.Run(a.RestartApplication)
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
