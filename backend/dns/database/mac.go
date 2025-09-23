package database

import (
	"errors"

	"gorm.io/gorm"
)

func FindVendor(db *gorm.DB, mac string) (string, error) {
	var query MacAddress
	tx := db.Find(&query, "mac = ?", mac)

	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if tx.Error != nil {
		return "", tx.Error
	}

	return query.Vendor, nil
}

func SaveMacEntry(db *gorm.DB, clientIP, mac, vendor string) {
	entry := MacAddress{
		MAC:    mac,
		IP:     clientIP,
		Vendor: vendor,
	}
	tx := db.Create(&entry)

	if tx.Error != nil {
		log.Warning("Unable to save new MAC entry %v", tx.Error)
	}
}
