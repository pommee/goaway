package audit

import (
	"goaway/backend/dns/database"
	"time"
)

type Manager struct {
	dbManager *database.Manager
}

func NewAuditManager(dbManager *database.Manager) *Manager {
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
			ID:        audit.ID,
			Topic:     Topic(audit.Topic),
			Message:   audit.Message,
			CreatedAt: audit.CreatedAt,
		}
	}

	return entries, nil
}
