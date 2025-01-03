package blacklist

import (
	"bufio"
	"database/sql"
	"fmt"
	"goaway/internal/logging"
	"io"
	"net/http"
	"strings"
	"sync"
)

var log = logging.GetLogger()

type Blacklist struct {
	DB           *sql.DB
	BlocklistURL []string
}

type DNSResolver struct {
	hostsMap map[string]string
	mu       sync.RWMutex
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
        CREATE TABLE IF NOT EXISTS blacklist (
            domain TEXT PRIMARY KEY
        )
    `)
	return err
}

func (b *Blacklist) initializeBlockedDomains() error {
	for _, url := range b.BlocklistURL {
		if err := b.fetchAndLoadHosts(url); err != nil {
			return err
		}
	}
	return nil
}

func (b *Blacklist) fetchAndLoadHosts(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch hosts file from %s: %w", url, err)
	}
	defer resp.Body.Close()

	domains, err := b.extractDomains(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to extract domains from %s: %w", url, err)
	}

	if err := b.AddDomains(domains); err != nil {
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
		if len(parts) >= 2 {
			domains = append(domains, parts[1:2]...)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading hosts file: %w", err)
	}

	return domains, nil
}

func (b *Blacklist) AddDomain(domain string) error {
	result, err := b.DB.Exec(`INSERT OR IGNORE INTO blacklist (domain) VALUES (?)`, domain)

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("%s is already blacklisted.", domain)
	}
	if err != nil {
		return fmt.Errorf("failed to add domain to blacklist: %w", err)
	}
	return nil
}

func (b *Blacklist) AddDomains(domains []string) error {
	tx, err := b.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO blacklist (domain) VALUES (?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, domain := range domains {
		if _, err := stmt.Exec(domain); err != nil {
			tx.Rollback()
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

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("%s is already whitelisted.", domain)
	}
	if err != nil {
		return fmt.Errorf("failed to remove domain from blacklist: %w", err)
	}
	return nil
}

func (b *Blacklist) IsBlacklisted(domain string) (bool, error) {
	row := b.DB.QueryRow(`SELECT 1 FROM blacklist WHERE domain = ?`, domain)
	var exists int
	if err := row.Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if domain is blacklisted: %w", err)
	}
	return true, nil
}
