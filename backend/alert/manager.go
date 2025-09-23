package alert

import (
	"context"
	"fmt"
	"goaway/backend/dns/database"
	"goaway/backend/logging"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var log = logging.GetLogger()

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
	DB       *gorm.DB
	services []MessageSender
}

func NewManager(db *gorm.DB) *Manager {
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

func (m *Manager) SaveAlert(alert database.Alert) error {
	result := m.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "type"}},
		DoUpdates: clause.AssignmentColumns([]string{"enabled", "name", "webhook"}),
	}).Create(&alert)

	if result.Error != nil {
		return fmt.Errorf("failed to save alert: %w", result.Error)
	}

	m.Load()
	return nil
}

func (m *Manager) GetAllAlerts() ([]database.Alert, error) {
	ctx := context.Background()
	return gorm.G[database.Alert](m.DB).Find(ctx)
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
