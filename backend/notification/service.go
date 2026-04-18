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

type NotificationPaginatedResult struct {
	Notifications []database.Notification `json:"notifications"`
	Total         int64                   `json:"total"`
	Page          int                     `json:"page"`
	Limit         int                     `json:"limit"`
	TotalPages    int                     `json:"totalPages"`
}

func (s *Service) GetNotifications(page, limit int) (*NotificationPaginatedResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}

	total, notifications, err := s.repository.GetNotifications(page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := max(int((total+int64(limit)-1)/int64(limit)), 1)
	return &NotificationPaginatedResult{
		Notifications: notifications,
		Total:         total,
		Page:          page,
		Limit:         limit,
		TotalPages:    totalPages,
	}, nil
}

func (s *Service) MarkNotificationsAsRead(notificationIDs []int) error {
	log.Info("Notifications have been marked as read")
	return s.repository.MarkNotificationsAsRead(notificationIDs)
}
