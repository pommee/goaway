package notification

import (
	"goaway/backend/database"
	"goaway/backend/logging"
)

type Service struct {
	repository Repository
}

type Severity string
type Category string

// Severity level of notification
// SeverityInfo: 	Server was upgraded, password changed...
// SeverityWarning: An error occurred on startup, database lock...
// SeverityError:   Server cant start, requests cant be handled...
const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

// Categories to describe what area the notification covers
const (
	CategoryServer Category = "server"
	CategoryDNS    Category = "dns"
	CategoryAPI    Category = "api"
)

var log = logging.GetLogger()

func NewService(repo Repository) *Service {
	return &Service{repository: repo}
}

func (s *Service) SendNotification(severity Severity, category Category, text string) {
	notification := &database.Notification{
		Severity: string(severity),
		Category: string(category),
		Text:     text,
		Read:     false,
	}

	err := s.repository.CreateNotification(notification)
	if err != nil {
		log.Warning("Could not send notification, %v", err)
		return
	}

	log.Info("New notification created, severity: %s", severity)
}

func (s *Service) GetNotifications() ([]database.Notification, error) {
	return s.repository.GetNotifications()
}

func (s *Service) MarkNotificationsAsRead(notificationIDs []int) error {
	log.Info("Notifications have been marked as read")
	return s.repository.MarkNotificationsAsRead(notificationIDs)
}
