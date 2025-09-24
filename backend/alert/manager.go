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

	log.Info("Alert settings saved for type: %s", alert.Type)

	m.Load()
	return nil
}

func (m *Manager) RemoveAlert(alertType string) error {
	if alertType == "" {
		return fmt.Errorf("alert type cannot be empty")
	}

	result := m.DB.Where("type = ?", alertType).Delete(&database.Alert{})

	if result.Error != nil {
		return fmt.Errorf("failed to remove alert: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		log.Warning("No alert found with type: %s", alertType)
		return fmt.Errorf("no alert found with type: %s", alertType)
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

func (m *Manager) SendTest(ctx context.Context, alertType, name, webhook string) error {
	err := m.SaveAlert(database.Alert{
		Type:    alertType,
		Enabled: true,
		Name:    name,
		Webhook: webhook,
	})
	if err != nil {
		log.Error("Failed to save test alert settings: %v", err)
		return err
	}

	for _, service := range m.services {
		if service.GetServiceName() == alertType {
			log.Info("Sending test alert via %s", service.GetServiceName())
			err = service.SendMessage(ctx, Message{
				Title:    "System",
				Content:  "This is a test alert from GoAway",
				Severity: "info",
			})
			if err != nil {
				log.Error("Failed to send test alert via %s: %v", service.GetServiceName(), err)
			}
			break
		}
	}

	err = m.RemoveAlert(alertType)
	if err != nil {
		log.Error("Failed to remove test alert settings: %v", err)
		return err
	}

	return nil
}
