package lifecycle

import (
	"context"
	"goaway/backend/api"
	"goaway/backend/jobs"
	"goaway/backend/logging"
	"goaway/backend/services"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var log = logging.GetLogger()

// Coordinates startup, shutdown, and signal handling
type Manager struct {
	services       *services.ServiceRegistry
	backgroundJobs *jobs.BackgroundJobs
	signalChan     chan os.Signal
}

func NewManager(registry *services.ServiceRegistry) *Manager {
	return &Manager{
		services:   registry,
		signalChan: make(chan os.Signal, 1),
	}
}

func (m *Manager) Run(restartCallback api.RestartApplicationCallback) error {
	if err := m.services.Initialize(); err != nil {
		return err
	}

	m.services.APIServer.RestartCallback = restartCallback

	m.backgroundJobs = jobs.NewBackgroundJobs(m.services)

	signal.Notify(m.signalChan, syscall.SIGINT, syscall.SIGTERM)

	m.services.StartAll()
	m.backgroundJobs.Start(m.services.ReadyChannel())

	go m.services.WaitGroup().Wait()

	return m.waitForTermination()
}

func (m *Manager) waitForTermination() error {
	select {
	case err := <-m.services.ErrorChannel():
		if m.services.APIServer.IsShuttingDown {
			log.Info("Ignoring error during controlled shutdown")
			return m.waitForTermination()
		}
		log.Error("%s server failed: %s", err.Service, err.Err)
		log.Fatal("Server failure detected. Exiting.")
		return err.Err
	case <-m.signalChan:
		log.Info("Received interrupt. Shutting down.")
		m.shutdown()
		return nil
	}
}

func (m *Manager) shutdown() {
	log.Info("Starting graceful shutdown...")

	m.services.Shutdown()
	m.services.APIServer.IsShuttingDown = true

	if err := m.services.APIServer.Stop(); err != nil {
		log.Error("Error stopping API server: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	m.services.UDPServer.Shutdown(ctx)
	log.Info("Stopped UDP server")

	m.services.TCPServer.Shutdown(ctx)
	log.Info("Stopped TCP server")

	if m.services.DoTServer != nil {
		m.services.DoTServer.Shutdown(ctx)
		log.Info("Stopped DNS-over-TLS server")
	}

	if m.services.DoHServer != nil {
		if err := m.services.DoHServer.Shutdown(ctx); err != nil && err != context.DeadlineExceeded {
			log.Error("Error stopping DoH server: %v", err)
		}
		log.Info("Stopped DNS-over-HTTPS server")
	}

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		m.services.WaitGroup().Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info("All services stopped gracefully")
	case <-time.After(15 * time.Second):
		log.Warning("Shutdown timeout exceeded, forcing exit")
	}

	os.Exit(0)
}
