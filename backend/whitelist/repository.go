package whitelist

import (
	"errors"
	"fmt"
	"goaway/backend/database"

	"gorm.io/gorm"
)

type Repository interface {
	AddDomain(domain string) error
	GetDomains() (map[string]bool, error)
	RemoveDomain(domain string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) AddDomain(domain string) error {
	result := r.db.Create(&database.Whitelist{Domain: domain})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("%s is already whitelisted", domain)
		}

		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%s is already whitelisted", domain)
	}

	return nil
}

func (r *repository) GetDomains() (map[string]bool, error) {
	var records []database.Whitelist
	result := r.db.Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}

	domainMap := make(map[string]bool)
	for _, record := range records {
		domainMap[record.Domain] = true
	}

	return domainMap, nil
}

func (r *repository) RemoveDomain(domain string) error {
	result := r.db.Delete(&database.Whitelist{}, "domain = ?", domain)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%s does not exist", domain)
	}

	return nil
}
