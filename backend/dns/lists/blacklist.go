package lists

import (
	"bufio"
	"context"
	"crypto/sha256"
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

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var log = logging.GetLogger()

type BlocklistSource struct {
	Name string
	URL  string
}

type Blacklist struct {
	DBManager      *database.DatabaseManager
	BlocklistURL   []BlocklistSource
	BlacklistCache map[string]bool
}

type SourceStats struct {
	Name         string `json:"name"`
	URL          string `json:"url"`
	BlockedCount int    `json:"blockedCount"`
	LastUpdated  int64  `json:"lastUpdated"`
	Active       bool   `json:"active"`
}

type ListUpdateAvailable struct {
	RemoteDomains   []string `json:"remoteDomains"`
	DBDomains       []string `json:"dbDomains"`
	RemoteChecksum  string   `json:"remoteChecksum"`
	DBChecksum      string   `json:"dbChecksum"`
	UpdateAvailable bool     `json:"updateAvailable"`
	DiffAdded       []string `json:"diffAdded"`
	DiffRemoved     []string `json:"diffRemoved"`
}

func InitializeBlacklist(dbManager *database.DatabaseManager) (*Blacklist, error) {
	b := &Blacklist{
		DBManager: dbManager,
		BlocklistURL: []BlocklistSource{
			{Name: "StevenBlack", URL: "https://raw.githubusercontent.com/StevenBlack/hosts/refs/heads/master/hosts"},
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
	for _, source := range b.BlocklistURL {
		if source.Name == "Custom" {
			continue
		}
		if err := b.FetchAndLoadHosts(source.URL, source.Name); err != nil {
			return err
		}
	}
	return nil
}

func (b *Blacklist) Vacuum() {
	b.DBManager.Mutex.Lock()
	tx := b.DBManager.Conn.Raw("VACUUM")
	if err := tx.Error; err != nil {
		log.Warning("Error while vacuuming database: %v", err)
	}
	err := tx.Commit().Error
	b.DBManager.Mutex.Unlock()
	if err != nil {
		log.Warning("Error while vacuuming database: %v", err)
	}
}

func (b *Blacklist) GetBlocklistUrls() ([]BlocklistSource, error) {
	var sources []database.Source

	result := b.DBManager.Conn.Where("name != ?", "Custom").Find(&sources)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query sources: %w", result.Error)
	}

	blocklistURL := make([]BlocklistSource, len(sources))
	for i, source := range sources {
		blocklistURL[i] = BlocklistSource{
			Name: source.Name,
			URL:  source.URL,
		}
	}

	b.BlocklistURL = blocklistURL
	return blocklistURL, nil
}

func (b *Blacklist) CheckIfUpdateAvailable(remoteListURL, listName string) (ListUpdateAvailable, error) {
	listUpdateAvailable := ListUpdateAvailable{}
	remoteDomains, remoteChecksum, err := b.FetchRemoteHostsList(remoteListURL)
	if err != nil {
		log.Warning("Failed to fetch remote hosts list: %v", err)
		return listUpdateAvailable, fmt.Errorf("failed to fetch remote hosts list: %w", err)
	}

	dbDomains, dbChecksum, err := b.FetchDBHostsList(listName)
	if err != nil {
		log.Warning("Failed to fetch database hosts list: %v", err)
		return listUpdateAvailable, fmt.Errorf("failed to fetch database hosts list: %w", err)
	}

	if remoteChecksum == dbChecksum {
		log.Debug("No updates available for %s", listName)
		return listUpdateAvailable, nil
	}

	diff := func(a, b []string) []string {
		mb := make(map[string]struct{}, len(b))
		for _, x := range b {
			mb[x] = struct{}{}
		}
		diff := make([]string, 0)
		for _, x := range a {
			if _, found := mb[x]; !found {
				diff = append(diff, x)
			}
		}
		return diff
	}

	return ListUpdateAvailable{
		RemoteDomains:   remoteDomains,
		DBDomains:       dbDomains,
		RemoteChecksum:  remoteChecksum,
		DBChecksum:      dbChecksum,
		UpdateAvailable: true,
		DiffAdded:       diff(remoteDomains, dbDomains),
		DiffRemoved:     diff(dbDomains, remoteDomains),
	}, nil
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

	log.Info("Added %d domains from list '%s' with url '%s'", len(domains), name, url)
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
	blacklistEntry := database.Blacklist{Domain: domain}

	result := b.DBManager.Conn.Create(&blacklistEntry)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "UNIQUE constraint failed") ||
			strings.Contains(result.Error.Error(), "duplicate key") {
			return fmt.Errorf("%s is already blacklisted", domain)
		}
		return fmt.Errorf("failed to add domain to blacklist: %w", result.Error)
	}

	b.BlacklistCache[domain] = true
	return nil
}

func (b *Blacklist) AddDomains(domains []string, url string) error {
	return b.DBManager.Conn.Transaction(func(tx *gorm.DB) error {
		var source database.Source
		currentTime := time.Now().Unix()

		result := tx.Model(&source).Where("url = ?", url).Update("last_updated", currentTime)
		if result.Error != nil {
			return fmt.Errorf("failed to update source: %w", result.Error)
		}

		if err := tx.Where("url = ?", url).First(&source).Error; err != nil {
			return fmt.Errorf("failed to find source: %w", err)
		}

		blacklistEntries := make([]database.Blacklist, 0, len(domains))
		for _, domain := range domains {
			blacklistEntries = append(blacklistEntries, database.Blacklist{
				Domain:   domain,
				SourceID: source.ID,
			})
		}

		if len(blacklistEntries) > 0 {
			if err := tx.CreateInBatches(blacklistEntries, 1000).Error; err != nil {
				if !strings.Contains(err.Error(), "UNIQUE constraint failed") &&
					!strings.Contains(err.Error(), "duplicate key") {
					return fmt.Errorf("failed to add domains: %w", err)
				}
			}
		}

		return nil
	})
}

func (b *Blacklist) PopulateBlocklistCache() (int, error) {
	var databaseDomains []string
	result := b.DBManager.Conn.Model(&database.Blacklist{}).
		Distinct("domain").
		Pluck("domain", &databaseDomains)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to query blacklist: %w", result.Error)
	}

	b.BlacklistCache = make(map[string]bool, len(databaseDomains))
	for _, domain := range databaseDomains {
		b.BlacklistCache[domain] = true
	}

	return len(b.BlacklistCache), nil
}

func (b *Blacklist) CountDomains() (int, error) {
	var count int64
	result := b.DBManager.Conn.Model(&database.Blacklist{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count domains: %w", result.Error)
	}
	return int(count), nil
}

func (b *Blacklist) GetAllowedAndBlocked() (allowed, blocked int, err error) {
	type RequestStats struct {
		Blocked bool
		Count   int
	}

	var stats []RequestStats
	result := b.DBManager.Conn.Model(&database.RequestLog{}).
		Select("blocked, COUNT(*) as count").
		Group("blocked").
		Scan(&stats)

	if result.Error != nil {
		return 0, 0, fmt.Errorf("failed to query request_logs: %w", result.Error)
	}

	for _, stat := range stats {
		if stat.Blocked {
			blocked = stat.Count
		} else {
			allowed = stat.Count
		}
	}

	return allowed, blocked, nil
}

func (b *Blacklist) RemoveDomain(domain string) error {
	result := b.DBManager.Conn.Where("domain = ?", domain).Delete(&database.Blacklist{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove domain from blacklist: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("%s is already whitelisted", domain)
	}

	delete(b.BlacklistCache, domain)
	return nil
}

func (b *Blacklist) UpdateSourceName(oldName, newName, url string) error {
	if strings.TrimSpace(newName) == "" {
		return fmt.Errorf("new name cannot be empty")
	}

	if oldName == newName {
		return fmt.Errorf("new name is the same as the old name")
	}

	result := b.DBManager.Conn.Model(&database.Source{}).
		Where("name = ? AND url = ?", oldName, url).
		Update("name", newName)

	if result.Error != nil {
		return fmt.Errorf("failed to update source name: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("list with name '%s' not found", oldName)
	}

	for i, source := range b.BlocklistURL {
		if source.Name == oldName {
			b.BlocklistURL[i].Name = newName
		}
	}

	log.Info("Updated blocklist name from '%s' to '%s'", oldName, newName)
	return nil
}

func (b *Blacklist) NameExists(name, url string) bool {
	for _, source := range b.BlocklistURL {
		if source.Name == name && source.URL == url {
			return true
		}
	}

	return false
}

func (b *Blacklist) URLExists(url string) bool {
	for _, source := range b.BlocklistURL {
		if source.URL == url {
			return true
		}
	}
	return false
}

func (b *Blacklist) IsBlacklisted(domain string) bool {
	return b.BlacklistCache[domain]
}

func (b *Blacklist) LoadPaginatedBlacklist(page, pageSize int, search string) ([]string, int, error) {
	searchPattern := "%" + search + "%"
	offset := (page - 1) * pageSize

	var blacklistEntries []database.Blacklist
	result := b.DBManager.Conn.Select("domain").
		Where("domain LIKE ?", searchPattern).
		Order("domain DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&blacklistEntries)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to query blacklist: %w", result.Error)
	}

	domains := make([]string, len(blacklistEntries))
	for i, entry := range blacklistEntries {
		domains[i] = entry.Domain
	}

	var total int64
	countResult := b.DBManager.Conn.Model(&database.Blacklist{}).
		Where("domain LIKE ?", searchPattern).
		Count(&total)

	if countResult.Error != nil {
		return nil, 0, fmt.Errorf("failed to count domains: %w", countResult.Error)
	}

	return domains, int(total), nil
}

func (b *Blacklist) InitializeBlocklist(name, url string) error {
	return b.DBManager.Conn.Transaction(func(tx *gorm.DB) error {
		source := database.Source{
			Name:        name,
			URL:         url,
			LastUpdated: time.Now().Unix(),
			Active:      true,
		}

		result := tx.Where(database.Source{Name: name, URL: url}).FirstOrCreate(&source)
		if result.Error != nil {
			return fmt.Errorf("failed to initialize new blocklist: %w", result.Error)
		}

		return nil
	})
}

func (b *Blacklist) AddCustomDomains(domains []string) error {
	return b.DBManager.Conn.Transaction(func(tx *gorm.DB) error {
		var source database.Source
		currentTime := time.Now().Unix()

		err := tx.Where("name = ?", "Custom").First(&source).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				source = database.Source{
					Name:        "Custom",
					LastUpdated: currentTime,
				}
				if err := tx.Create(&source).Error; err != nil {
					return fmt.Errorf("failed to insert custom source: %w", err)
				}
			} else {
				return fmt.Errorf("failed to get custom source ID: %w", err)
			}
		} else {
			if err := tx.Model(&source).Update("last_updated", currentTime).Error; err != nil {
				return fmt.Errorf("failed to update lastUpdated for custom source: %w", err)
			}
		}

		blacklistEntries := make([]database.Blacklist, 0, len(domains))
		for _, domain := range domains {
			blacklistEntries = append(blacklistEntries, database.Blacklist{
				Domain:   domain,
				SourceID: source.ID,
			})
		}

		for _, entry := range blacklistEntries {
			if err := tx.Where(database.Blacklist{Domain: entry.Domain, SourceID: entry.SourceID}).FirstOrCreate(&entry).Error; err != nil {
				return fmt.Errorf("failed to add custom domain '%s': %w", entry.Domain, err)
			}
			b.BlacklistCache[entry.Domain] = true
		}

		return nil
	})
}

func (b *Blacklist) RemoveCustomDomain(domain string) error {
	b.DBManager.Mutex.Lock()
	defer b.DBManager.Mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return b.DBManager.Conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var source database.Source
		err := tx.Where("name = ?", "Custom").First(&source).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("custom source not found")
			}
			return fmt.Errorf("failed to get custom source ID: %w", err)
		}

		result := tx.Where("domain = ? AND source_id = ?", domain, source.ID).Delete(&database.Blacklist{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete domain '%s': %w", domain, result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("domain '%s' not found in custom blacklist", domain)
		}

		delete(b.BlacklistCache, domain)

		currentTime := time.Now().Unix()
		if err := tx.Model(&source).Update("last_updated", currentTime).Error; err != nil {
			return fmt.Errorf("failed to update lastUpdated for custom source: %w", err)
		}

		return nil
	})
}

func (b *Blacklist) GetAllListStatistics() ([]SourceStats, error) {
	type SourceWithCount struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		URL          string `json:"url"`
		LastUpdated  int64  `json:"last_updated"`
		Active       bool   `json:"active"`
		BlockedCount int    `json:"blocked_count"`
	}

	var results []SourceWithCount
	result := b.DBManager.Conn.Table("sources s").
		Select("s.id, s.name, s.url, s.last_updated, s.active, COALESCE(bc.blocked_count, 0) as blocked_count").
		Joins("LEFT JOIN (SELECT source_id, COUNT(*) as blocked_count FROM blacklists GROUP BY source_id) bc ON s.id = bc.source_id").
		Order("s.name, s.id").
		Scan(&results)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query source statistics: %w", result.Error)
	}

	stats := make([]SourceStats, len(results))
	for i, r := range results {
		stats[i] = SourceStats{
			Name:         r.Name,
			URL:          r.URL,
			BlockedCount: r.BlockedCount,
			LastUpdated:  r.LastUpdated,
			Active:       r.Active,
		}
	}

	return stats, nil
}

func (b *Blacklist) GetListStatistics(listname string) (string, SourceStats, error) {
	type SourceWithCount struct {
		Name         string `json:"name"`
		URL          string `json:"url"`
		LastUpdated  int64  `json:"last_updated"`
		Active       bool   `json:"active"`
		BlockedCount int    `json:"blocked_count"`
	}

	var result SourceWithCount
	err := b.DBManager.Conn.Table("sources s").
		Select("s.name, s.url, s.last_updated, s.active, COALESCE(bc.blocked_count, 0) as blocked_count").
		Joins("LEFT JOIN (SELECT source_id, COUNT(*) as blocked_count FROM blacklists GROUP BY source_id) bc ON s.id = bc.source_id").
		Where("s.name = ?", listname).
		First(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", SourceStats{}, fmt.Errorf("list not found")
		}
		return "", SourceStats{}, fmt.Errorf("failed to query list statistics: %w", err)
	}

	stats := SourceStats{
		URL:          result.URL,
		BlockedCount: result.BlockedCount,
		LastUpdated:  result.LastUpdated,
		Active:       result.Active,
	}

	return result.Name, stats, nil
}

func (b *Blacklist) GetDomainsForList(list string) ([]string, error) {
	var blacklistEntries []database.Blacklist
	result := b.DBManager.Conn.Select("blacklists.domain").
		Joins("JOIN sources ON blacklists.source_id = sources.id").
		Where("sources.name = ?", list).
		Find(&blacklistEntries)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query domains for list: %w", result.Error)
	}

	domains := make([]string, len(blacklistEntries))
	for i, entry := range blacklistEntries {
		domains[i] = entry.Domain
	}

	return domains, nil
}

func (b *Blacklist) ToggleBlocklistStatus(name string) error {
	var source database.Source
	if err := b.DBManager.Conn.Where("name = ?", name).First(&source).Error; err != nil {
		return fmt.Errorf("failed to find source %s: %w", name, err)
	}

	result := b.DBManager.Conn.Model(&source).Update("active", !source.Active)
	if result.Error != nil {
		return fmt.Errorf("failed to toggle status for %s: %w", name, result.Error)
	}

	return nil
}

func (b *Blacklist) RemoveSourceAndDomains(name, url string) error {
	return b.DBManager.Conn.Transaction(func(tx *gorm.DB) error {
		var source database.Source
		err := tx.Where("name = ? AND url = ?", name, url).First(&source).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("source '%s' not found", name)
			}
			return fmt.Errorf("failed to get source ID: %w", err)
		}

		if err := tx.Where("source_id = ?", source.ID).Delete(&database.Blacklist{}).Error; err != nil {
			return fmt.Errorf("failed to remove domains for source '%s': %w", name, err)
		}

		if err := tx.Delete(&source).Error; err != nil {
			return fmt.Errorf("failed to remove source '%s': %w", name, err)
		}

		return nil
	})
}

func (b *Blacklist) RemoveSourceAndDomainsWithCacheRefresh(name, url string) error {
	if err := b.RemoveSourceAndDomains(name, url); err != nil {
		return err
	}

	if _, err := b.PopulateBlocklistCache(); err != nil {
		log.Warning("Failed to clear blocklist cache after removing source: %v", err)
	}

	log.Info("Removed all domains and source '%s'", name)
	return nil
}
func (b *Blacklist) ScheduleAutomaticListUpdates() {
	for {
		next := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)
		log.Info("Next auto-update for lists scheduled for: %s", next.Format(time.DateTime))
		time.Sleep(time.Until(next))

		for _, source := range b.BlocklistURL {
			if source.Name == "Custom" {
				continue
			}
			log.Info("Checking for updates for blocklist %s from %s", source.Name, source.URL)

			availableUpdate, err := b.CheckIfUpdateAvailable(source.URL, source.Name)
			if err != nil {
				log.Warning("Failed to check for updates for %s: %v", source.Name, err)
				continue
			}

			if !availableUpdate.UpdateAvailable {
				log.Info("No updates available for %s", source.Name)
				continue
			}

			if err := b.RemoveSourceAndDomains(source.Name, source.URL); err != nil {
				log.Warning("Failed to remove old domains for %s: %v", source.Name, err)
				continue
			}
			if err := b.FetchAndLoadHosts(source.URL, source.Name); err != nil {
				log.Warning("Failed to fetch and load hosts for %s: %v", source.Name, err)
				continue
			}

			log.Info("Successfully updated %s with %d new domains", source.Name, len(availableUpdate.DiffAdded))
		}
		if _, err := b.PopulateBlocklistCache(); err != nil {
			log.Warning("Failed to populate blocklist cache after auto-update: %v", err)
		}
	}
}

func (b *Blacklist) AddSource(name, url string) error {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(url) == "" {
		return fmt.Errorf("name and url cannot be empty")
	}

	if err := b.DBManager.Conn.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "url"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "last_updated", "active"}),
		},
	).Create(&database.Source{
		Name:        name,
		URL:         url,
		LastUpdated: time.Now().Unix(),
		Active:      true,
	}).Error; err != nil {
		return fmt.Errorf("failed to insert source: %w", err)
	}

	found := false
	for _, s := range b.BlocklistURL {
		if s.Name == name && s.URL == url {
			found = true
			break
		}
	}
	if !found {
		b.BlocklistURL = append(b.BlocklistURL, BlocklistSource{Name: name, URL: url})
	}

	return nil
}

func (b *Blacklist) RemoveSourceByNameAndURL(name, url string) bool {
	for i := len(b.BlocklistURL) - 1; i >= 0; i-- {
		if b.BlocklistURL[i].Name == name && b.BlocklistURL[i].URL == url {
			b.BlocklistURL = append(b.BlocklistURL[:i], b.BlocklistURL[i+1:]...)
			return true
		}
	}

	return false
}
