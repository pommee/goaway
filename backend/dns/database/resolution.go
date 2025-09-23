package database

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func FetchResolutions(db *gorm.DB) ([]Resolution, error) {
	var resolutions []Resolution
	if err := db.Find(&resolutions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch resolutions: %w", err)
	}
	return resolutions, nil
}

func FetchResolution(db *gorm.DB, domain string) (string, error) {
	log.Debug("Finding resolution for domain: %s", domain)
	var res Resolution

	db.Where("domain = ?", domain).Find(&res)
	if res.IP != "" {
		return res.IP, nil
	}

	parts := strings.Split(domain, ".")
	for i := 1; i < len(parts); i++ {
		wildcardDomain := "*." + strings.Join(parts[i:], ".")
		if err := db.Where("domain = ?", wildcardDomain).Find(&res).Error; err == nil {
			return res.IP, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", err
		}
	}

	return "", nil
}

func CreateNewResolution(db *gorm.DB, ip, domain string) error {
	res := Resolution{
		Domain: domain,
		IP:     ip,
	}

	if err := db.Create(&res).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return fmt.Errorf("domain already exists, must be unique")
		}
		return fmt.Errorf("could not create new resolution: %w", err)
	}
	return nil
}

func DeleteResolution(db *gorm.DB, ip, domain string) (int, error) {
	result := db.Where("ip = ? AND domain = ?", ip, domain).Delete(&Resolution{})
	if result.Error != nil {
		return 0, fmt.Errorf("could not delete resolution: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		log.Warning("No resolution found with IP: %s and Domain: %s", ip, domain)
	}

	return int(result.RowsAffected), nil
}
