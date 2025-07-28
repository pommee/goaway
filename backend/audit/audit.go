package audit

import (
	"database/sql"
	"goaway/backend/dns/database"
	"goaway/backend/logging"
	"time"
)

type Manager struct {
	dbManager *database.DatabaseManager
}

type Topic string

const (
	TopicServer     Topic = "server"
	TopicDNS        Topic = "dns"
	TopicAPI        Topic = "api"
	TopicResolution Topic = "resolution"
	TopicPrefetch   Topic = "prefetch"
	TopicUpstream   Topic = "upstream"
	TopicUser       Topic = "user"
	TopicList       Topic = "list"
	TopicLogs       Topic = "logs"
	TopicSettings   Topic = "settings"
	TopicDatabase   Topic = "database"
)

type Entry struct {
	Id        int       `json:"id"`
	Topic     Topic     `json:"topic"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

var logger = logging.GetLogger()

func NewAuditManager(dbManager *database.DatabaseManager) *Manager {
	return &Manager{dbManager: dbManager}
}

func (nm *Manager) CreateAudit(newAudit *Entry) {
	createdAt := time.Now()
	_, err := nm.dbManager.Conn.Exec(
		`INSERT INTO audit (topic, message, created_at) VALUES (?, ?, ?)`,
		newAudit.Topic, newAudit.Message, createdAt,
	)
	if err != nil {
		logger.Warning("Unable to create new audit, error: %v", err)
	}

	logger.Debug("Created new audit, %+v", newAudit)
}

func (nm *Manager) ReadAudits() ([]Entry, error) {
	rows, err := nm.dbManager.Conn.Query("SELECT id, topic, message, created_at FROM audit ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var audits = make([]Entry, 0)
	for rows.Next() {
		var id int
		var topic Topic
		var message string
		var createdAt time.Time
		if err := rows.Scan(&id, &topic, &message, &createdAt); err != nil {
			return nil, err
		}
		audits = append(audits, Entry{id, topic, message, createdAt})
	}

	return audits, nil
}
