package alert

import (
	"context"
	"fmt"
	"goaway/backend/database"
	"goaway/backend/logging"
)

type Service struct {
	repository Repository
	services   []messageSender
}

type Message struct {
	Content  string
	Title    string
	Channel  string
	Severity string
}

type messageSender interface {
	SendMessage(ctx context.Context, msg Message) error
	IsEnabled() bool
	GetServiceName() string
}

var log = logging.GetLogger()

func NewService(repo Repository) *Service {
	return &Service{
		repository: repo,
		services:   make([]messageSender, 0),
	}
}

func (s *Service) reset() {
	s.services = make([]messageSender, 0)
}

func (s *Service) registerService(service messageSender) {
	log.Debug("Registering alert service: %s", service.GetServiceName())
	s.services = append(s.services, service)
}

func (s *Service) SaveAlert(alert database.Alert) error {
	err := s.repository.SaveAlert(alert)
	if err != nil {
		log.Error("Failed to save alert settings: %v", err)
		return err
	}
	s.Load()
	log.Info("Alert settings saved for type: %s", alert.Type)

	return nil
}

func (s *Service) GetAllAlerts() ([]database.Alert, error) {
	return s.repository.GetAllAlerts()
}

func (s *Service) Load() {
	discordService := NewDiscordService(DiscordConfig{})

	alerts, err := s.GetAllAlerts()

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

	s.reset()
	s.registerService(discordService)
	log.Debug("Alert Manager loaded with %d services", len(s.services))
}

func (s *Service) SendToAll(ctx context.Context, msg Message) error {
	var errors []error

	for _, service := range s.services {
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

func (s *Service) RemoveAlert(alertType string) error {
	err := s.repository.RemoveAlert(alertType)
	if err != nil {
		log.Error("Failed to remove alert: %v", err)
		return err
	}
	s.Load()
	log.Info("Alert removed for type: %s", alertType)

	return nil
}

func (s *Service) SendTest(ctx context.Context, alertType, name, webhook string) error {
	err := s.SaveAlert(database.Alert{
		Type:    alertType,
		Enabled: true,
		Name:    name,
		Webhook: webhook,
	})
	if err != nil {
		log.Error("Failed to save test alert settings: %v", err)
		return err
	}

	for _, service := range s.services {
		if service.GetServiceName() == alertType {
			log.Info("Sending test alert via %s", service.GetServiceName())
			err = service.SendMessage(ctx, Message{
				Title:    "System",
				Content:  "This is a test alert from GoAway",
				Severity: "info",
			})
			if err != nil {
				log.Error("Failed to send test alert via %s: %v", service.GetServiceName(), err)
				return err
			}
			break
		}
	}

	err = s.RemoveAlert(alertType)
	if err != nil {
		log.Error("Failed to remove test alert settings: %v", err)
		return err
	}

	return nil
}
