package notification

import (
	"database/sql"
	"fmt"
	"goaway/backend/dns/database"
	"goaway/backend/logging"
	"strings"
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
	createdAt := time.Now()
	_, err := nm.dbManager.Conn.Exec(`INSERT INTO notifications (severity, category, text, read, created_at) VALUES (?, ?, ?, ?, ?)`, newNotification.Severity, newNotification.Category, newNotification.Text, false, createdAt)
	if err != nil {
		logger.Warning("Unable to create new notification, error: %v", err)
	}

	logger.Debug("Created new notification, %+v", newNotification)
}

func (nm *Manager) ReadNotifications() ([]Notification, error) {
	rows, err := nm.dbManager.Conn.Query(`SELECT id, severity, category, text, read, created_at FROM notifications WHERE read = 0`)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var notifications = make([]Notification, 0)
	for rows.Next() {
		var id int
		var severity Severity
		var category Category
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

func (nm *Manager) MarkNotificationsAsRead(notificationIDs []int) error {
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

	_, err := nm.dbManager.Conn.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}
