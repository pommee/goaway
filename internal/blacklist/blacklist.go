package blacklist

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"goaway/internal/logging"
	"io"
	"net/http"
	"strings"
	"time"
)

var log = logging.GetLogger()

type Blacklist struct {
	DB           *sql.DB
	BlocklistURL map[string]string
}

func (b *Blacklist) Initialize() error {
	if err := b.createBlacklistTable(); err != nil {
		return fmt.Errorf("failed to initialize blacklist table: %w", err)
	}

	if count, _ := b.CountDomains(); count == 0 {
		log.Info("No domains in blacklist. Running initialization...")
		if err := b.initializeBlockedDomains(); err != nil {
			return fmt.Errorf("failed to initialize blocked domains: %w", err)
		}
	}

	return nil
}

func (b *Blacklist) createBlacklistTable() error {
	_, err := b.DB.Exec(`
        CREATE TABLE IF NOT EXISTS sources (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE,
            url TEXT,
			active INTEGER,
            lastUpdated INTEGER
        );
        CREATE TABLE IF NOT EXISTS blacklist (
            domain TEXT,
            source_id INTEGER,
            PRIMARY KEY (domain, source_id),
            FOREIGN KEY (source_id) REFERENCES sources(id)
        )
    `)
	return err
}

func (b *Blacklist) initializeBlockedDomains() error {
	for source, url := range b.BlocklistURL {
		if source == "Custom" {
			continue
		}
		if err := b.FetchAndLoadHosts(url, source); err != nil {
			return err
		}
	}
	return nil
}

func (b *Blacklist) GetBlocklistUrls() (map[string]string, error) {
	rows, err := b.DB.Query(`SELECT name, url FROM sources WHERE name != 'Custom'`)
	if err != nil {
		return nil, fmt.Errorf("failed to query sources: %w", err)
	}
	defer rows.Close()

	blocklistURL := make(map[string]string)
	for rows.Next() {
		var name, url string
		if err := rows.Scan(&name, &url); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		blocklistURL[name] = url
	}
	return blocklistURL, rows.Err()
}

func (b *Blacklist) FetchAndLoadHosts(url, name string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch hosts file from %s: %w", url, err)
	}
	defer resp.Body.Close()

	domains, err := b.extractDomains(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to extract domains from %s: %w", url, err)
	}

	_ = b.InitializeBlocklist(name, url)

	if err := b.AddDomains(domains, url); err != nil {
		return fmt.Errorf("failed to add domains to database: %w", err)
	}

	log.Info("Added %d domains from %s", len(domains), url)
	return nil
}

func (b *Blacklist) extractDomains(body io.Reader) ([]string, error) {
	var domains []string
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		parts := strings.Fields(line)

		if len(parts) > 1 {
			switch parts[1] {
			case "localhost", "localhost.localdomain", "broadcasthost", "local":
				continue
			}
			domains = append(domains, parts[1:2]...)
		}
		domains = append(domains, parts[0])
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading hosts file: %w", err)
	}

	if len(domains) == 0 {
		return nil, errors.New("zero results when parsing")
	}

	return domains, nil
}

func (b *Blacklist) AddDomain(domain string) error {
	result, err := b.DB.Exec(`INSERT OR IGNORE INTO blacklist (domain) VALUES (?)`, domain)
	if err != nil {
		return fmt.Errorf("failed to add domain to blacklist: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("%s is already blacklisted", domain)
	}
	return nil
}

func (b *Blacklist) AddDomains(domains []string, url string) error {
	tx, err := b.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	var sourceID int
	currentTime := time.Now().Unix()
	err = tx.QueryRow(`UPDATE sources SET lastUpdated = (?) WHERE url IS (?) RETURNING id`, currentTime, url).Scan(&sourceID)
	if err != nil {
		return fmt.Errorf("failed to insert source: %w", err)
	}

	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO blacklist (domain, source_id) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, domain := range domains {
		if _, err := stmt.Exec(domain, sourceID); err != nil {
			err = tx.Rollback()
			return fmt.Errorf("failed to add domain '%s': %w", domain, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (b *Blacklist) LoadBlacklist() (map[string]bool, error) {
	rows, err := b.DB.Query("SELECT domain FROM blacklist")
	if err != nil {
		return nil, fmt.Errorf("failed to query blacklist: %w", err)
	}
	defer rows.Close()

	domains := make(map[string]bool)
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		domains[domain] = true
	}
	return domains, rows.Err()
}

func (b *Blacklist) CountDomains() (int, error) {
	var count int
	err := b.DB.QueryRow(`SELECT COUNT(*) FROM blacklist`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count domains: %w", err)
	}
	return count, nil
}

func (b *Blacklist) RemoveDomain(domain string) error {
	result, err := b.DB.Exec(`DELETE FROM blacklist WHERE domain = ?`, domain)
	if err != nil {
		return fmt.Errorf("failed to remove domain from blacklist: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("%s is already whitelisted", domain)
	}
	return nil
}

func (b *Blacklist) IsBlacklisted(domain string) (bool, error) {
	var query = "SELECT b.source_id FROM blacklist b JOIN sources s ON b.source_id = s.id WHERE b.domain = ? AND s.active = 1"
	row := b.DB.QueryRow(query, domain)
	var exists int
	if err := row.Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if domain is blacklisted: %w", err)
	}
	return true, nil
}

func (b *Blacklist) LoadPaginatedBlacklist(page, pageSize int, search string) ([]string, int, error) {
	query := `
		SELECT domain
		FROM blacklist
		WHERE domain LIKE ?
		ORDER BY domain DESC
		LIMIT ? OFFSET ?
	`
	searchPattern := "%" + search + "%"
	offset := (page - 1) * pageSize

	rows, err := b.DB.Query(query, searchPattern, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query blacklist: %w", err)
	}
	defer rows.Close()

	var domains []string
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			return nil, 0, fmt.Errorf("failed to scan row: %w", err)
		}
		domains = append(domains, domain)
	}

	countQuery := `SELECT COUNT(*) FROM blacklist WHERE domain LIKE ?`
	var total int
	err = b.DB.QueryRow(countQuery, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count domains: %w", err)
	}

	return domains, total, rows.Err()
}

func (b *Blacklist) InitializeBlocklist(name, url string) error {
	tx, err := b.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	_, err = tx.Exec(`INSERT OR IGNORE INTO sources (name, url, lastUpdated, active) VALUES (?, ?, ?, ?)`, name, url, time.Now().Unix(), true)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to initialize new blocklist: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (b *Blacklist) AddCustomDomains(domains []string) error {
	tx, err := b.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	var sourceID int
	currentTime := time.Now().Unix()
	err = tx.QueryRow(`SELECT id FROM sources WHERE name = ?`, "Custom").Scan(&sourceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = tx.QueryRow(`INSERT INTO sources (name, lastUpdated) VALUES (?, ?) RETURNING id`, "Custom", currentTime).Scan(&sourceID)
			if err != nil {
				return fmt.Errorf("failed to insert custom source: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get custom source ID: %w", err)
		}
	} else {
		_, err = tx.Exec(`UPDATE sources SET lastUpdated = ? WHERE id = ?`, currentTime, sourceID)
		if err != nil {
			return fmt.Errorf("failed to update lastUpdated for custom source: %w", err)
		}
	}

	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO blacklist (domain, source_id) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, domain := range domains {
		if _, err := stmt.Exec(domain, sourceID); err != nil {
			err = tx.Rollback()
			return fmt.Errorf("failed to add custom domain '%s': %w", domain, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (b *Blacklist) GetSourceStatistics() (map[string]map[string]interface{}, error) {
	query := `
		SELECT s.name, COUNT(b.domain) as blocked_count, s.lastUpdated, s.active
		FROM sources s
		LEFT JOIN blacklist b ON s.id = b.source_id
		GROUP BY s.name, s.lastUpdated, s.active
	`

	rows, err := b.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query source statistics: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]map[string]interface{})
	for rows.Next() {
		var sourceName string
		var blockedCount int
		var lastUpdated int64
		var active bool
		if err := rows.Scan(&sourceName, &blockedCount, &lastUpdated, &active); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		stats[sourceName] = map[string]interface{}{
			"blockedCount": blockedCount,
			"lastUpdated":  lastUpdated,
			"active":       active,
		}
	}

	return stats, rows.Err()
}

func (b *Blacklist) GetDomainsForList(list string) ([]string, error) {
	query := `
		SELECT domain
		FROM blacklist
		JOIN sources ON blacklist.source_id = sources.id
		WHERE sources.name = ?
	`
	rows, err := b.DB.Query(query, list)
	if err != nil {
		return nil, fmt.Errorf("failed to query domains for list: %w", err)
	}
	defer rows.Close()

	var domains []string
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		domains = append(domains, domain)
	}

	return domains, rows.Err()
}

func (b *Blacklist) ToggleBlocklistStatus(name string) error {
	query := `UPDATE sources SET active = NOT active WHERE name = ?`

	_, err := b.DB.Exec(query, name)
	if err != nil {
		return fmt.Errorf("failed to toggle status for %s: %w", name, err)
	}

	return nil
}

func (b *Blacklist) RemoveSourceAndDomains(source string) error {
	tx, err := b.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	var sourceID int
	err = tx.QueryRow(`SELECT id FROM sources WHERE name = ?`, source).Scan(&sourceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("source '%s' not found", source)
		}
		return fmt.Errorf("failed to get source ID: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM blacklist WHERE source_id = ?`, sourceID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to remove domains for source '%s': %w", source, err)
	}

	_, err = tx.Exec(`DELETE FROM sources WHERE id = ?`, sourceID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to remove source '%s': %w", source, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info("Removed all domains and source '%s'", source)
	return nil
}
