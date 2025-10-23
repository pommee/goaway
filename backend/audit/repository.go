package audit

import (
	"goaway/backend/database"

	"gorm.io/gorm"
)

type Repository interface {
	CreateAudit(audit *Entry) error
	ReadAudits() ([]Entry, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateAudit(audit *Entry) error {
	dbAudit := database.Audit{
		Topic:     string(audit.Topic),
		Message:   audit.Message,
		CreatedAt: audit.CreatedAt,
	}

	result := r.db.Create(&dbAudit)
	return result.Error
}

func (r *repository) ReadAudits() ([]Entry, error) {
	var dbAudits []database.Audit
	result := r.db.Order("created_at DESC").Find(&dbAudits)
	if result.Error != nil {
		return nil, result.Error
	}

	audits := make([]Entry, len(dbAudits))
	for i, dbAudit := range dbAudits {
		audits[i] = Entry{
			ID:        dbAudit.ID,
			Topic:     Topic(dbAudit.Topic),
			Message:   dbAudit.Message,
			CreatedAt: dbAudit.CreatedAt,
		}
	}

	return audits, nil
}
