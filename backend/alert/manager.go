package alert

import (
	"context"
	"database/sql"
	"fmt"
	"goaway/backend/logging"
)

var log = logging.GetLogger()

type Alert struct {
	Type    string
	Enabled bool
	Name    string
	Webhook string
}

type Message struct {
	Content  string
	Title    string
	Channel  string
	Severity string
}

type MessageSender interface {
	SendMessage(ctx context.Context, msg Message) error
	IsEnabled() bool
	GetServiceName() string
}

type Manager struct {
	DB       *sql.DB
	services []MessageSender
}

func NewManager(db *sql.DB) *Manager {
	return &Manager{
		DB:       db,
		services: make([]MessageSender, 0),
	}
}

func (m *Manager) Reset() {
	m.services = make([]MessageSender, 0)
}

func (m *Manager) Load() {
	discordService := NewDiscordService(DiscordConfig{})

	alerts, err := m.GetAllAlerts()

	if err != nil {
		log.Warning("Failed to load alerts from database: %v", err)
		return
	}

	for _, alert := range alerts {
		switch alert.Type {
		case "discord":
			discordService.config.Enabled = alert.Enabled
			discordService.config.WebhookURL = alert.Webhook
			discordService.config.Username = alert.Name
		default:
			log.Warning("Unknown alert type in database: %s", alert.Type)
		}
	}

	m.Reset()
	m.RegisterService(discordService)
	log.Debug("Alert Manager loaded with %d services", len(m.services))
}

func (m *Manager) SaveAlert(alert Alert) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
        INSERT INTO alert (type, enabled, name, webhook)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(type) DO UPDATE SET
            enabled=excluded.enabled,
            name=excluded.name,
            webhook=excluded.webhook;
    `
	_, err = tx.Exec(query, alert.Type, alert.Enabled, alert.Name, alert.Webhook)
	if err != nil {
		return fmt.Errorf("failed to save alert: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	m.Load()
	return nil
}

func (m *Manager) GetAllAlerts() ([]Alert, error) {
	query := "SELECT type, enabled, name, webhook FROM alert"
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var alert Alert
		err := rows.Scan(&alert.Type, &alert.Enabled, &alert.Name, &alert.Webhook)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}

func (m *Manager) RegisterService(service MessageSender) {
	log.Debug("Registering alert service: %s", service.GetServiceName())
	m.services = append(m.services, service)
}

func (m *Manager) SendToAll(ctx context.Context, msg Message) error {
	var errors []error

	for _, service := range m.services {
		if !service.IsEnabled() {
			continue
		}

		if err := service.SendMessage(ctx, msg); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", service.GetServiceName(), err))
		} else {
			log.Debug("Message sent successfully via %s", service.GetServiceName())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send to some services: %v", errors)
	}

	return nil
}
