package audit

import (
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
	audit := database.Audit{
		Topic:     string(newAudit.Topic),
		Message:   newAudit.Message,
		CreatedAt: time.Now(),
	}

	result := nm.dbManager.Conn.Create(&audit)
	if result.Error != nil {
		logger.Warning("Unable to create new audit, error: %v", result.Error)
		return
	}

	logger.Debug("Created new audit, %+v", newAudit)
}

func (nm *Manager) ReadAudits() ([]Entry, error) {
	var audits []database.Audit

	result := nm.dbManager.Conn.Order("created_at DESC").Find(&audits)
	if result.Error != nil {
		return nil, result.Error
	}

	entries := make([]Entry, len(audits))
	for i, audit := range audits {
		entries[i] = Entry{
			Id:        int(audit.ID),
			Topic:     Topic(audit.Topic),
			Message:   audit.Message,
			CreatedAt: audit.CreatedAt,
		}
	}

	return entries, nil
}
