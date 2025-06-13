package lists

import (
	"database/sql"
	"fmt"
	"goaway/backend/dns/database"
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

	err := w.getDomains()
	if err != nil {
		log.Error("Failed to initialize whitelist cache")
	}

	return w, err
}

func (w *Whitelist) AddDomain(domain string) error {
	w.DBManager.Mutex.Lock()
	defer w.DBManager.Mutex.Unlock()

	result, err := w.DBManager.Conn.Exec(`INSERT OR IGNORE INTO whitelist (domain) VALUES (?)`, domain)
	if err != nil {
		return fmt.Errorf("failed to add domain to whitelist: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("%s is already whitelisted", domain)
	}
	w.Cache[domain] = true
	return nil
}

func (w *Whitelist) RemoveDomain(domain string) error {
	result, err := w.DBManager.Conn.Exec(`DELETE FROM whitelist WHERE domain = ?`, domain)
	if err != nil {
		return fmt.Errorf("failed to remove domain from whitelist: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("%s does not exist", domain)
	}

	delete(w.Cache, domain)
	return nil
}

func (w *Whitelist) getDomains() error {
	w.Cache = map[string]bool{}

	rows, err := w.DBManager.Conn.Query("SELECT domain FROM whitelist")
	if err != nil {
		return fmt.Errorf("failed to query whitelist: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	domains := make(map[string]bool)
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		domains[domain] = true
	}
	w.Cache = domains
	return nil
}

func (w *Whitelist) IsWhitelisted(domain string) bool {
	return w.Cache[domain]
}
