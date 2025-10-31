package lifecycle

import (
	"goaway/backend/jobs"
	"goaway/backend/logging"
	"goaway/backend/services"
	"os"
	"os/signal"
	"syscall"
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

func (m *Manager) Run() error {
	if err := m.services.Initialize(); err != nil {
		return err
	}

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
	// TODO: Add graceful shutdown logic
	os.Exit(0)
}
