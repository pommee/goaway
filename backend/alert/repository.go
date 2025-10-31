package alert

import (
	"fmt"
	"goaway/backend/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	SaveAlert(alert database.Alert) error
	GetAllAlerts() ([]database.Alert, error)
	RemoveAlert(alertType string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) SaveAlert(alert database.Alert) error {
	result := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "type"}},
		DoUpdates: clause.AssignmentColumns([]string{"enabled", "name", "webhook"}),
	}).Create(&alert)

	if result.Error != nil {
		return fmt.Errorf("failed to save alert: %w", result.Error)
	}

	return nil
}

func (r *repository) GetAllAlerts() ([]database.Alert, error) {
	var alerts []database.Alert
	result := r.db.Find(&alerts)
	return alerts, result.Error
}

func (r *repository) RemoveAlert(alertType string) error {
	if alertType == "" {
		return fmt.Errorf("alert type cannot be empty")
	}

	result := r.db.Where("type = ?", alertType).Delete(&database.Alert{})

	if result.Error != nil {
		return fmt.Errorf("failed to remove alert: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		log.Warning("No alert found with type: %s", alertType)
		return fmt.Errorf("no alert found with type: %s", alertType)
	}

	return nil
}
