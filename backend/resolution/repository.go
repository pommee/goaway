package resolution

import (
	"errors"
	"fmt"
	"goaway/backend/database"
	"strings"

	"gorm.io/gorm"
)

type Repository interface {
	CreateResolution(ip, domain string) error
	FindResolution(domain string) (string, error)
	FindResolutions() ([]database.Resolution, error)
	DeleteResolution(ip, domain string) (int, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateResolution(ip, domain string) error {
	res := database.Resolution{
		Domain: domain,
		IP:     ip,
	}

	if err := r.db.Create(&res).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return errors.New("domain already exists, must be unique")
		}
		return fmt.Errorf("could not create new resolution: %w", err)
	}
	return nil
}

func (r *repository) FindResolution(domain string) (string, error) {
	var res database.Resolution

	r.db.Where("domain = ?", domain).Find(&res)
	if res.IP != "" {
		return res.IP, nil
	}

	parts := strings.Split(domain, ".")
	for i := 1; i < len(parts); i++ {
		wildcardDomain := "*." + strings.Join(parts[i:], ".")
		if err := r.db.Where("domain = ?", wildcardDomain).Find(&res).Error; err == nil {
			return res.IP, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", err
		}
	}

	return "", nil
}

func (r *repository) FindResolutions() ([]database.Resolution, error) {
	var resolutions []database.Resolution
	if err := r.db.Find(&resolutions).Error; err != nil {
		return nil, err
	}
	return resolutions, nil
}

func (r *repository) DeleteResolution(ip, domain string) (int, error) {
	result := r.db.Where("domain = ? AND ip = ?", domain, ip).Delete(&database.Resolution{})
	if result.Error != nil {
		return 0, result.Error
	}
	return int(result.RowsAffected), nil
}
