package key

import (
	"goaway/backend/database"

	"gorm.io/gorm"
)

// Repository handles all database operations for API keys
type Repository interface {
	Create(apiKey *database.APIKey) error
	FindByKey(key string) (*database.APIKey, error)
	FindAll() ([]database.APIKey, error)
	DeleteByName(name string) error
	CountByKey(key string) (int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(apiKey *database.APIKey) error {
	return r.db.Create(apiKey).Error
}

func (r *repository) FindByKey(key string) (*database.APIKey, error) {
	var apiKey database.APIKey
	err := r.db.Where("key = ?", key).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *repository) FindAll() ([]database.APIKey, error) {
	var apiKeys []database.APIKey
	err := r.db.Find(&apiKeys).Error
	return apiKeys, err
}

func (r *repository) DeleteByName(name string) error {
	return r.db.Where("name = ?", name).Delete(&database.APIKey{}).Error
}

func (r *repository) CountByKey(key string) (int64, error) {
	var count int64
	err := r.db.Model(&database.APIKey{}).Where("key = ?", key).Count(&count).Error
	return count, err
}
