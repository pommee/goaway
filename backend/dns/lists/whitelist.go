package lists

import (
	"fmt"
	"goaway/backend/dns/database"

	"gorm.io/gorm/clause"
)

type Whitelist struct {
	DBManager *database.DatabaseManager
	Cache     map[string]bool
}

func InitializeWhitelist(dbManager *database.DatabaseManager) (*Whitelist, error) {
	w := &Whitelist{
		DBManager: dbManager,
		Cache:     map[string]bool{},
	}

	_, err := w.GetDomains()
	if err != nil {
		log.Error("Failed to initialize whitelist cache")
	}

	return w, err
}

func (w *Whitelist) AddDomain(domain string) error {
	result := w.DBManager.Conn.Clauses(clause.OnConflict{DoNothing: true}).Create(&database.Whitelist{Domain: domain})

	if result.Error != nil {
		return fmt.Errorf("failed to add domain to whitelist: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%s is already whitelisted", domain)
	}

	w.Cache[domain] = true
	return nil
}

func (w *Whitelist) RemoveDomain(domain string) error {
	result := w.DBManager.Conn.Delete(&database.Whitelist{}, "domain = ?", domain)

	if result.Error != nil {
		return fmt.Errorf("failed to remove domain from whitelist: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%s does not exist", domain)
	}

	delete(w.Cache, domain)
	return nil
}

func (w *Whitelist) GetDomains() (map[string]bool, error) {
	var records []database.Whitelist
	if err := w.DBManager.Conn.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to query whitelist: %w", err)
	}

	domains := make(map[string]bool, len(records))
	for _, rec := range records {
		domains[rec.Domain] = true
	}

	return domains, nil
}

func (w *Whitelist) IsWhitelisted(domain string) bool {
	return w.Cache[domain]
}
