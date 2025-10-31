package audit

import (
	"goaway/backend/database"
	"time"

	"gorm.io/gorm"
)

type Manager struct {
	dbConn *gorm.DB
}

func NewAuditManager(dbconn *gorm.DB) *Manager {
	return &Manager{dbConn: dbconn}
}

func (m *Manager) CreateAudit(newAudit *Entry) {
	audit := database.Audit{
		Topic:     string(newAudit.Topic),
		Message:   newAudit.Message,
		CreatedAt: time.Now(),
	}

	result := m.dbConn.Create(&audit)
	if result.Error != nil {
		log.Warning("Unable to create new audit, error: %v", result.Error)
		return
	}

	log.Debug("Created new audit, %+v", newAudit)
}

func (m *Manager) ReadAudits() ([]Entry, error) {
	var audits []database.Audit

	result := m.dbConn.Order("created_at DESC").Find(&audits)
	if result.Error != nil {
		return nil, result.Error
	}

	entries := make([]Entry, len(audits))
	for i, audit := range audits {
		entries[i] = Entry{
			ID:        audit.ID,
			Topic:     Topic(audit.Topic),
			Message:   audit.Message,
			CreatedAt: audit.CreatedAt,
		}
	}

	return entries, nil
}
