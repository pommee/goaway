package mac

import (
	"errors"
	"fmt"
	"goaway/backend/database"

	"gorm.io/gorm"
)

type Repository interface {
	FindVendor(mac string) (string, error)
	SaveMac(clientIP, mac, vendor string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindVendor(mac string) (string, error) {
	var query database.MacAddress
	tx := r.db.Find(&query, "mac = ?", mac)

	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if tx.Error != nil {
		return "", tx.Error
	}

	return query.Vendor, nil
}

func (r *repository) SaveMac(clientIP, mac, vendor string) error {
	entry := database.MacAddress{
		MAC:    mac,
		IP:     clientIP,
		Vendor: vendor,
	}
	tx := r.db.Save(&entry)

	if tx.Error != nil {
		return fmt.Errorf("unable to save new MAC entry %v", tx.Error)
	}

	return nil
}
