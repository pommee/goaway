package lists

import (
	"bufio"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"goaway/backend/dns/database"
	"goaway/backend/logging"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

var log = logging.GetLogger()

type Blacklist struct {
	DBManager      *database.DatabaseManager
	BlocklistURL   map[string]string
	BlacklistCache map[string]bool
}

type SourceStats struct {
	URL          string `json:"url"`
	BlockedCount int    `json:"blockedCount"`
	LastUpdated  int64  `json:"lastUpdated"`
	Active       bool   `json:"active"`
}

func InitializeBlacklist(dbManager *database.DatabaseManager) (*Blacklist, error) {
	b := &Blacklist{
		DBManager: dbManager,
		BlocklistURL: map[string]string{
			"StevenBlack": "https://raw.githubusercontent.com/StevenBlack/hosts/refs/heads/master/hosts",
		},
		BlacklistCache: map[string]bool{},
	}

	if count, _ := b.CountDomains(); count == 0 {
		log.Info("No domains in blacklist. Running initialization...")
		if err := b.initializeBlockedDomains(); err != nil {
			return nil, fmt.Errorf("failed to initialize blocked domains: %w", err)
		}
	}

	if err := b.InitializeBlocklist("Custom", ""); err != nil {
		return nil, fmt.Errorf("failed to initialize custom blocklist: %w", err)
	}

	_, err := b.GetBlocklistUrls()
	if err != nil {
		log.Error("Failed to fetch blocklist URLs: %v", err)
		return nil, fmt.Errorf("failed to fetch blocklist URLs: %w", err)
	}
	_, err = b.PopulateBlocklistCache()
	if err != nil {
		log.Error("Failed to initialize blocklist cache")
		return nil, fmt.Errorf("failed to initialize blocklist cache: %w", err)
	}

	return b, nil
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

func (b *Blacklist) Vacuum() {
	b.DBManager.Mutex.Lock()
	_, err := b.DBManager.Conn.Exec("VACUUM")
	b.DBManager.Mutex.Unlock()
	if err != nil {
		log.Warning("Error while vacuuming database: %v", err)
	}
}

func (b *Blacklist) GetBlocklistUrls() (map[string]string, error) {
	rows, err := b.DBManager.Conn.Query(`SELECT name, url FROM sources WHERE name != 'Custom'`)
	if err != nil {
		return nil, fmt.Errorf("failed to query sources: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	blocklistURL := make(map[string]string)
	for rows.Next() {
		var name, url string
		if err := rows.Scan(&name, &url); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		blocklistURL[name] = url
	}

	b.BlocklistURL = blocklistURL
	return blocklistURL, nil
}

func (b *Blacklist) FetchRemoteHostsList(url string) ([]string, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch hosts file from %s: %w", url, err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	domains, err := b.ExtractDomains(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to extract domains from %s: %w", url, err)
	}

	return domains, calculateDomainsChecksum(domains), nil
}

func (b *Blacklist) FetchDBHostsList(name string) ([]string, string, error) {
	domains, err := b.GetDomainsForList(name)
	if err != nil {
		return nil, "", fmt.Errorf("could not fetch domains from database")
	}

	return domains, calculateDomainsChecksum(domains), nil
}

func calculateDomainsChecksum(domains []string) string {
	sort.Strings(domains)
	data := strings.Join(domains, "\n")

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (b *Blacklist) FetchAndLoadHosts(url, name string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch hosts file from %s: %w", url, err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	domains, err := b.ExtractDomains(resp.Body)
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

func (b *Blacklist) ExtractDomains(body io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(body)
	domainSet := make(map[string]struct{})
	var domains []string

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}

		domain := fields[0]
		if (domain == "0.0.0.0" || domain == "127.0.0.1") && len(fields) > 1 {
			domain = fields[1]
			switch domain {
			case "localhost", "localhost.localdomain", "broadcasthost", "local", "0.0.0.0":
				continue
			}
		} else if domain == "0.0.0.0" || domain == "127.0.0.1" {
			continue
		}

		if _, exists := domainSet[domain]; !exists {
			domainSet[domain] = struct{}{}
			domains = append(domains, domain)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading hosts file: %w", err)
	}
	if len(domains) == 0 {
		return nil, errors.New("zero results when parsing")
	}

	return domains, nil
}

func (b *Blacklist) AddBlacklistedDomain(domain string) error {
	result, err := b.DBManager.Conn.Exec(`INSERT OR IGNORE INTO blacklist (domain) VALUES (?)`, domain)
	if err != nil {
		return fmt.Errorf("failed to add domain to blacklist: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("%s is already blacklisted", domain)
	}
	b.BlacklistCache[domain] = true
	return nil
}

func (b *Blacklist) AddDomains(domains []string, url string) error {
	tx, err := b.DBManager.Conn.Begin()
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
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)

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

func (b *Blacklist) PopulateBlocklistCache() (int, error) {
	b.BlacklistCache = map[string]bool{}

	rows, err := b.DBManager.Conn.Query("SELECT domain FROM blacklist")
	if err != nil {
		return 0, fmt.Errorf("failed to query blacklist: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	domains := make(map[string]bool)
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			return 0, fmt.Errorf("failed to scan row: %w", err)
		}
		domains[domain] = true
	}
	b.BlacklistCache = domains

	return len(domains), nil
}

func (b *Blacklist) CountDomains() (int, error) {
	var count int
	err := b.DBManager.Conn.QueryRow(`SELECT COUNT(*) FROM blacklist`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count domains: %w", err)
	}
	return count, nil
}

func (b *Blacklist) GetAllowedAndBlocked() (allowed, blocked int, err error) {
	rows, err := b.DBManager.Conn.Query(`SELECT blocked, COUNT(*) FROM request_log GROUP BY blocked`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query request_log: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var blockedFlag bool
		var count int
		if err := rows.Scan(&blockedFlag, &count); err != nil {
			return 0, 0, fmt.Errorf("failed to scan row: %w", err)
		}
		if blockedFlag {
			blocked = count
		} else {
			allowed = count
		}
	}

	if err := rows.Err(); err != nil {
		return 0, 0, fmt.Errorf("row iteration error: %w", err)
	}

	return allowed, blocked, nil
}

func (b *Blacklist) RemoveDomain(domain string) error {
	result, err := b.DBManager.Conn.Exec(`DELETE FROM blacklist WHERE domain = ?`, domain)
	if err != nil {
		return fmt.Errorf("failed to remove domain from blacklist: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("%s is already whitelisted", domain)
	}
	b.BlacklistCache[domain] = true
	return nil
}

func (b *Blacklist) IsBlacklisted(domain string) bool {
	return b.BlacklistCache[domain]
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

	rows, err := b.DBManager.Conn.Query(query, searchPattern, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query blacklist: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

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
	err = b.DBManager.Conn.QueryRow(countQuery, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count domains: %w", err)
	}

	return domains, total, rows.Err()
}

func (b *Blacklist) InitializeBlocklist(name, url string) error {
	tx, err := b.DBManager.Conn.Begin()
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
	tx, err := b.DBManager.Conn.Begin()
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
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)

	for _, domain := range domains {
		if _, err := stmt.Exec(domain, sourceID); err != nil {
			err = tx.Rollback()
			return fmt.Errorf("failed to add custom domain '%s': %w", domain, err)
		}
		b.BlacklistCache[domain] = true
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (b *Blacklist) RemoveCustomDomain(domain string) error {
	tx, err := b.DBManager.Conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var sourceID int
	err = tx.QueryRow(`SELECT id FROM sources WHERE name = ?`, "Custom").Scan(&sourceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("custom source not found")
		}
		return fmt.Errorf("failed to get custom source ID: %w", err)
	}

	result, err := tx.Exec(`DELETE FROM blacklist WHERE domain = ? AND source_id = ?`, domain, sourceID)
	if err != nil {
		return fmt.Errorf("failed to delete domain '%s': %w", domain, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("domain '%s' not found in custom blacklist", domain)
	}

	delete(b.BlacklistCache, domain)

	currentTime := time.Now().Unix()
	_, err = tx.Exec(`UPDATE sources SET lastUpdated = ? WHERE id = ?`, currentTime, sourceID)
	if err != nil {
		return fmt.Errorf("failed to update lastUpdated for custom source: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (b *Blacklist) GetSourceStatistics() (map[string]SourceStats, error) {
	query := `
		SELECT s.name, s.url, s.lastUpdated, s.active, COALESCE(bc.blocked_count, 0) as blocked_count
		FROM sources s
		LEFT JOIN (
			SELECT source_id, COUNT(*) as blocked_count
			FROM blacklist
			GROUP BY source_id
		) bc ON s.id = bc.source_id
		ORDER BY s.name
	`

	rows, err := b.DBManager.Conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query source statistics: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	stats := make(map[string]SourceStats, 10)

	for rows.Next() {
		var name, url string
		var blockedCount int
		var lastUpdated int64
		var active bool

		if err := rows.Scan(&name, &url, &lastUpdated, &active, &blockedCount); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		stats[name] = SourceStats{
			URL:          url,
			BlockedCount: blockedCount,
			LastUpdated:  lastUpdated,
			Active:       active,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return stats, nil
}

func (b *Blacklist) GetDomainsForList(list string) ([]string, error) {
	query := `
		SELECT domain
		FROM blacklist
		JOIN sources ON blacklist.source_id = sources.id
		WHERE sources.name = ?
	`
	rows, err := b.DBManager.Conn.Query(query, list)
	if err != nil {
		return nil, fmt.Errorf("failed to query domains for list: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

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

	_, err := b.DBManager.Conn.Exec(query, name)
	if err != nil {
		return fmt.Errorf("failed to toggle status for %s: %w", name, err)
	}

	return nil
}

func (b *Blacklist) RemoveSourceAndDomains(source string) error {
	tx, err := b.DBManager.Conn.Begin()
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
