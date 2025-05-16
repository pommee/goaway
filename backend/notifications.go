package notification

import (
	"database/sql"
	"fmt"
	"goaway/backend/logging"
	"strings"
	"time"
)

type NotificationManager struct {
	db *sql.DB
}

type NotificationSeverity string
type NotificationCategory string

// Severity level of notification
// SeverityInfo: 	Server was upgraded, password changed...
// SeverityWarning: An error occured on startup, database lock...
// SeverityError:   Server cant start, requests cant be handled...
const (
	SeverityInfo    NotificationSeverity = "info"
	SeverityWarning NotificationSeverity = "warning"
	SeverityError   NotificationSeverity = "error"
)

// Categories to describe what area the notification covers
const (
	CategoryServer NotificationCategory = "server"
	CategoryDNS    NotificationCategory = "dns"
	CategoryAPI    NotificationCategory = "api"
)

type Notification struct {
	Id        int                  `json:"id"`
	Severity  NotificationSeverity `json:"severity"`
	Category  NotificationCategory `json:"category"`
	Text      string               `json:"text"`
	Read      bool                 `json:"read"`
	CreatedAt time.Time            `json:"createdAt"`
}

var logger = logging.GetLogger()

func NewNotificationManager(db *sql.DB) *NotificationManager {
	return &NotificationManager{db}
}

func (nm *NotificationManager) CreateNotification(newNotification *Notification) {
	createdAt := time.Now()
	_, err := nm.db.Exec(`INSERT INTO notifications (severity, category, text, read, created_at) VALUES (?, ?, ?, ?, ?)`, newNotification.Severity, newNotification.Category, newNotification.Text, false, createdAt)
	if err != nil {
		logger.Warning("Unable to create new notification, error: %v", err)
	}

	logger.Debug("Created new notification, %+v", newNotification)
}

func (nm *NotificationManager) ReadNotifications() ([]Notification, error) {
	rows, err := nm.db.Query(`SELECT id, severity, category, text, read, created_at FROM notifications WHERE read = 0`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications = make([]Notification, 0)
	for rows.Next() {
		var id int
		var severity NotificationSeverity
		var category NotificationCategory
		var text string
		var read bool
		var createdAt time.Time
		if err := rows.Scan(&id, &severity, &category, &text, &read, &createdAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, Notification{id, severity, category, text, read, createdAt})
	}

	return notifications, nil
}

func (nm *NotificationManager) MarkNotificationsAsRead(notificationIDs []int) error {
	if len(notificationIDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(notificationIDs))
	args := make([]any, len(notificationIDs))

	for i, id := range notificationIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`UPDATE notifications SET read = true WHERE id IN (%s)`,
		strings.Join(placeholders, ","))

	_, err := nm.db.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}
