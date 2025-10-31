package prefetch

import (
	"fmt"
	"goaway/backend/database"

	"gorm.io/gorm"
)

type Repository interface {
	GetAll() ([]database.Prefetch, error)
	Create(prefetch *database.Prefetch) error
	Delete(domain string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(prefetch *database.Prefetch) error {
	result := r.db.Create(prefetch)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *repository) GetAll() ([]database.Prefetch, error) {
	var prefetched []database.Prefetch
	result := r.db.Model(&database.Prefetch{}).Find(&prefetched)
	if result.Error != nil {
		return nil, result.Error
	}
	return prefetched, nil
}

func (r *repository) Delete(domain string) error {
	result := r.db.Delete(&database.Prefetch{}, "domain = ?", domain)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("%s does not exist in the database", domain)
	}

	return nil
}
