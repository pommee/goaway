package notification

import (
	"goaway/backend/dns/database"
	"goaway/backend/logging"
	"time"
)

type Manager struct {
	dbManager *database.DatabaseManager
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

type Notification struct {
	Id        int       `json:"id"`
	Severity  Severity  `json:"severity"`
	Category  Category  `json:"category"`
	Text      string    `json:"text"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"createdAt"`
}

var logger = logging.GetLogger()

func NewNotificationManager(dbManager *database.DatabaseManager) *Manager {
	return &Manager{dbManager: dbManager}
}

func (nm *Manager) CreateNotification(newNotification *Notification) {
	tx := nm.dbManager.Conn.Create(&database.Notification{
		Severity:  string(newNotification.Severity),
		Category:  string(newNotification.Category),
		Text:      newNotification.Text,
		Read:      false,
		CreatedAt: time.Now(),
	})
	if tx.Error != nil {
		logger.Warning("Unable to create new notification, error: %v", tx.Error)
		return
	}

	logger.Debug("Created new notification, %+v", newNotification)
}

func (nm *Manager) ReadNotifications() ([]database.Notification, error) {
	var notifications []database.Notification

	result := nm.dbManager.Conn.Where("read = ?", true).Find(&notifications)
	if result.Error != nil {
		return nil, result.Error
	}

	return notifications, nil
}

func (nm *Manager) MarkNotificationsAsRead(notificationIDs []int) error {
	if len(notificationIDs) == 0 {
		return nil
	}

	result := nm.dbManager.Conn.Model(&database.Notification{}).
		Where("id IN ?", notificationIDs).
		Update("read", true)

	if result.Error != nil {
		return result.Error
	}

	return nil
}
