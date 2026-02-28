package blacklist

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"goaway/backend/database"
	"goaway/backend/logging"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

type Config struct {
	DefaultSources []BlocklistSource
	CacheTTL       time.Duration
	BatchSize      int
	UpdateInterval time.Duration
}

type Service struct {
	repository   Repository
	httpClient   HTTPClient
	cache        map[string]bool
	cacheMu      sync.RWMutex
	blocklistURL []BlocklistSource
	config       Config
}

var (
	log            = logging.GetLogger()
	invalidDomains = map[string]bool{
		"localhost":             true,
		"localhost.localdomain": true,
		"broadcasthost":         true,
		"local":                 true,
		blacklistedIP:           true,
	}
)

const (
	blacklistedIP    = "0.0.0.0"
	IPv4Loopback     = "127.0.0.1"
	defaultBatchSize = 1000
)

var defaultConfig = Config{
	DefaultSources: []BlocklistSource{
		{
			Name: "StevenBlack",
			URL:  "https://raw.githubusercontent.com/StevenBlack/hosts/refs/heads/master/hosts",
		},
	},
	CacheTTL:       24 * time.Hour,
	BatchSize:      defaultBatchSize,
	UpdateInterval: 24 * time.Hour,
}

func NewService(repo Repository) *Service {
	config := defaultConfig
	service := &Service{
		repository: repo,
		httpClient: http.DefaultClient,
		cache:      make(map[string]bool),
		config:     config,
	}

	if len(service.blocklistURL) == 0 {
		service.blocklistURL = config.DefaultSources
	}

	if err := service.initialize(context.Background()); err != nil {
		log.Error("Could not initialize blacklist: %v", err)
	}

	return service
}

func (s *Service) initialize(ctx context.Context) error {
	count, err := s.repository.CountDomains(ctx)
	if err != nil {
		return fmt.Errorf("failed to count domains: %w", err)
	}

	previouslyInitialized := s.repository.GetSourceExists(ctx, "Custom", "")

	if count == 0 && !previouslyInitialized {
		log.Info("No domains in blacklist. Running initialization...")
		if err := s.initializeBlockedDomains(ctx); err != nil {
			return fmt.Errorf("failed to initialize blocked domains: %w", err)
		}
	}

	if err := s.InitializeBlocklist(ctx, "Custom", ""); err != nil {
		return fmt.Errorf("failed to initialize custom blocklist: %w", err)
	}

	if _, err := s.GetBlocklistUrls(ctx); err != nil {
		log.Error("Failed to fetch blocklist URLs: %v", err)
		return fmt.Errorf("failed to fetch blocklist URLs: %w", err)
	}

	if err := s.PopulateCache(ctx); err != nil {
		log.Error("Failed to initialize blocklist cache: %v", err)
		return fmt.Errorf("failed to initialize blocklist cache: %w", err)
	}

	return nil
}

func (s *Service) initializeBlockedDomains(ctx context.Context) error {
	start := time.Now()
	for _, source := range s.blocklistURL {
		if source.Name == "Custom" {
			continue
		}
		if err := s.FetchAndLoadHosts(ctx, source.URL, source.Name); err != nil {
			return err
		}
	}

	log.Info("Blocked domains initialized in %.2fs", time.Since(start).Seconds())
	return nil
}

func (s *Service) PopulateCache(ctx context.Context) error {
	domains, err := s.repository.GetAllDomains(ctx)
	if err != nil {
		return fmt.Errorf("failed to populate cache: %w", err)
	}

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.cache = make(map[string]bool, len(domains))
	for _, domain := range domains {
		domain = strings.TrimSuffix(domain, ".")
		s.cache[domain] = true
	}

	return nil
}

func (s *Service) IsBlacklisted(domain string) bool {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	if exists, found := s.cache[domain]; found {
		return exists
	}
	return false
}

func (s *Service) GetBlocklistUrls(ctx context.Context) ([]BlocklistSource, error) {
	sources, err := s.repository.GetSources(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocklist URLs: %w", err)
	}

	blocklistURL := make([]BlocklistSource, len(sources))
	for i, source := range sources {
		blocklistURL[i] = BlocklistSource{
			Name: source.Name,
			URL:  source.URL,
		}
	}

	s.blocklistURL = blocklistURL
	return blocklistURL, nil
}

func (s *Service) NameExists(name, url string) bool {
	for _, source := range s.blocklistURL {
		if source.Name == name && source.URL == url {
			return true
		}
	}
	return false
}

func (s *Service) URLExists(url string) bool {
	for _, source := range s.blocklistURL {
		if source.URL == url {
			return true
		}
	}
	return false
}

func (s *Service) CheckIfUpdateAvailable(ctx context.Context, remoteListURL, listName string) (ListUpdateAvailable, error) {
	listUpdateAvailable := ListUpdateAvailable{}

	remoteDomains, remoteChecksum, err := s.FetchRemoteHostsList(ctx, remoteListURL)
	if err != nil {
		log.Warning("Failed to fetch remote hosts list: %v", err)
		return listUpdateAvailable, fmt.Errorf("failed to fetch remote hosts list: %w", err)
	}

	dbDomains, dbChecksum, err := s.FetchDBHostsList(ctx, listName)
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

func (s *Service) FetchRemoteHostsList(ctx context.Context, url string) ([]string, string, error) {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch hosts file from %s: %w", url, err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	domains, err := s.ExtractDomains(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to extract domains from %s: %w", url, err)
	}

	return domains, calculateDomainsChecksum(domains), nil
}

func (s *Service) FetchDBHostsList(ctx context.Context, name string) ([]string, string, error) {
	domains, err := s.repository.GetDomainsForSource(ctx, name)
	if err != nil {
		return nil, "", fmt.Errorf("could not fetch domains from database: %w", err)
	}

	return domains, calculateDomainsChecksum(domains), nil
}

func calculateDomainsChecksum(domains []string) string {
	sort.Strings(domains)
	data := strings.Join(domains, "\n")
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *Service) FetchAndLoadHosts(ctx context.Context, url, name string) error {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch hosts file from %s: %w", url, err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	domains, err := s.ExtractDomains(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to extract domains from %s: %w", url, err)
	}

	if err := s.InitializeBlocklist(ctx, name, url); err != nil {
		return fmt.Errorf("failed to initialize blocklist: %w", err)
	}

	if err := s.AddDomains(ctx, name, domains, url); err != nil {
		return fmt.Errorf("failed to add domains to database: %w", err)
	}

	log.Info("Added %d domains from list '%s' with url '%s'", len(domains), name, url)
	return nil
}

func (s *Service) isValidDomain(domain string) bool {
	return !invalidDomains[domain]
}

func (s *Service) ExtractDomains(body io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(body)
	domainSet := make(map[string]struct{})
	var domains []string

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}

		domain := fields[0]
		if (domain == blacklistedIP || domain == IPv4Loopback) && len(fields) > 1 {
			domain = fields[1]
			if !s.isValidDomain(domain) {
				continue
			}
		} else if domain == blacklistedIP || domain == IPv4Loopback {
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

func (s *Service) updateCache(domains []string, add bool) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	for _, domain := range domains {
		if add {
			s.cache[domain] = true
		} else {
			delete(s.cache, domain)
		}
	}
}

func (s *Service) AddBlacklistedDomain(ctx context.Context, domain string) error {
	blacklistEntry := &database.Blacklist{Domain: domain}

	if err := s.repository.CreateDomain(ctx, blacklistEntry); err != nil {
		return err
	}

	s.updateCache([]string{domain}, true)
	return nil
}

func (s *Service) AddDomains(ctx context.Context, name string, domains []string, url string) error {
	return s.repository.WithTransaction(ctx, func(tx *gorm.DB) error {
		currentTime := time.Now()

		if err := s.repository.UpdateSourceLastUpdated(ctx, url, currentTime); err != nil {
			return err
		}

		source, err := s.repository.GetSourceByNameAndURL(ctx, name, url)
		if err != nil {
			return err
		}

		blacklistEntries := make([]database.Blacklist, 0, len(domains))
		for _, domain := range domains {
			blacklistEntries = append(blacklistEntries, database.Blacklist{
				Domain:   domain,
				SourceID: source.ID,
			})
		}

		if len(blacklistEntries) > 0 {
			batchSize := s.config.BatchSize
			if batchSize == 0 {
				batchSize = defaultBatchSize
			}
			if err := s.repository.CreateDomainsInBatches(ctx, blacklistEntries, batchSize); err != nil {
				return err
			}
		}

		s.updateCache(domains, true)
		return nil
	})
}

func (s *Service) RemoveDomain(ctx context.Context, domain string) error {
	if err := s.repository.DeleteDomain(ctx, domain); err != nil {
		return err
	}

	s.updateCache([]string{domain}, false)
	return nil
}

func (s *Service) AddCustomDomains(ctx context.Context, domains []string) error {
	return s.repository.WithTransaction(ctx, func(tx *gorm.DB) error {
		currentTime := time.Now()

		source, err := s.repository.GetSourceByName(ctx, "Custom")
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				newSource := &database.Source{
					Name:        "Custom",
					LastUpdated: currentTime,
					Active:      true,
				}
				if err := s.repository.CreateOrUpdateSource(ctx, newSource); err != nil {
					return fmt.Errorf("failed to create custom source: %w", err)
				}
				source = newSource
			} else {
				return fmt.Errorf("failed to get custom source: %w", err)
			}
		} else {
			if err := s.repository.UpdateSourceLastUpdated(ctx, "", currentTime); err != nil {
				return fmt.Errorf("failed to update custom source: %w", err)
			}
		}

		for _, domain := range domains {
			entry := &database.Blacklist{
				Domain:   domain,
				SourceID: source.ID,
			}
			// Ignore duplicate errors
			if err := s.repository.CreateDomain(ctx, entry); err != nil {
				if errors.Is(err, gorm.ErrDuplicatedKey) {
					return fmt.Errorf("%s is already blacklisted", domain)
				}
				return err
			} else {
				s.updateCache([]string{domain}, true)
			}
		}

		return nil
	})
}

func (s *Service) RemoveCustomDomain(ctx context.Context, domain string) error {
	source, err := s.repository.GetSourceByName(ctx, "Custom")
	if err != nil {
		return fmt.Errorf("custom source not found: %w", err)
	}

	if err := s.repository.DeleteCustomDomain(ctx, domain, source.ID); err != nil {
		return err
	}

	s.updateCache([]string{domain}, false)

	currentTime := time.Now()
	if err := s.repository.UpdateSourceLastUpdated(ctx, "", currentTime); err != nil {
		log.Warning("Failed to update custom source timestamp: %v", err)
	}

	return nil
}

func (s *Service) InitializeBlocklist(ctx context.Context, name, url string) error {
	source := &database.Source{
		Name:        name,
		URL:         url,
		LastUpdated: time.Now(),
		Active:      true,
	}

	return s.repository.CreateOrUpdateSource(ctx, source)
}

func (s *Service) AddSource(ctx context.Context, name, url string) error {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(url) == "" {
		return fmt.Errorf("name and url cannot be empty")
	}

	source := &database.Source{
		Name:        name,
		URL:         url,
		LastUpdated: time.Now(),
		Active:      true,
	}

	if err := s.repository.UpsertSource(ctx, source); err != nil {
		return err
	}

	// Update in-memory list
	found := false
	for _, existing := range s.blocklistURL {
		if existing.Name == name && existing.URL == url {
			found = true
			break
		}
	}
	if !found {
		s.blocklistURL = append(s.blocklistURL, BlocklistSource{Name: name, URL: url})
	}

	return nil
}

func (s *Service) UpdateSourceName(ctx context.Context, oldName, newName, url string) error {
	if oldName == newName {
		return fmt.Errorf("new name is the same as the old name")
	}

	if err := s.repository.UpdateSourceName(ctx, oldName, newName, url); err != nil {
		return err
	}

	// Update in-memory list
	for i, source := range s.blocklistURL {
		if source.Name == oldName {
			s.blocklistURL[i].Name = newName
		}
	}

	log.Info("Updated blocklist name from '%s' to '%s'", oldName, newName)
	return nil
}

func (s *Service) ToggleBlocklistStatus(ctx context.Context, name string) error {
	return s.repository.ToggleSourceActive(ctx, name)
}

func (s *Service) RemoveSourceAndDomains(ctx context.Context, name, url string) error {
	return s.repository.WithTransaction(ctx, func(tx *gorm.DB) error {
		source, err := s.repository.GetSourceByNameAndURL(ctx, name, url)
		if err != nil {
			return err
		}

		if err := s.repository.DeleteDomainsBySourceID(ctx, source.ID); err != nil {
			return fmt.Errorf("failed to remove domains for source '%s': %w", name, err)
		}

		if err := s.repository.DeleteSource(ctx, name, url); err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) RemoveSourceByNameAndURL(name, url string) bool {
	for i := len(s.blocklistURL) - 1; i >= 0; i-- {
		if s.blocklistURL[i].Name == name && s.blocklistURL[i].URL == url {
			s.blocklistURL = append(s.blocklistURL[:i], s.blocklistURL[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Service) CountDomains(ctx context.Context) (int, error) {
	count, err := s.repository.CountDomains(ctx)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (s *Service) GetRequestMetrics(ctx context.Context) (allowed, blocked, cached int, err error) {
	stats, err := s.repository.GetRequestStats(ctx)
	if err != nil {
		return 0, 0, 0, err
	}
	for _, stat := range stats {
		if stat.Blocked {
			blocked = stat.Count
		} else {
			allowed = stat.Count
		}
		if stat.Cached {
			cached += stat.Count
		}
	}
	return allowed, blocked, cached, nil
}

func (s *Service) GetAllListStatistics(ctx context.Context) ([]SourceWithCount, error) {
	results, err := s.repository.GetAllSourceStats(ctx)
	if err != nil {
		return nil, err
	}

	stats := make([]SourceWithCount, len(results))
	for i, r := range results {
		stats[i] = SourceWithCount{
			Name:         r.Name,
			URL:          r.URL,
			BlockedCount: r.BlockedCount,
			LastUpdated:  r.LastUpdated,
			Active:       r.Active,
		}
	}

	return stats, nil
}

func (s *Service) GetListStatistics(ctx context.Context, listname string) (string, SourceWithCount, error) {
	result, err := s.repository.GetSourceStats(ctx, listname)
	if err != nil {
		return "", SourceWithCount{}, err
	}

	stats := SourceWithCount{
		URL:          result.URL,
		BlockedCount: result.BlockedCount,
		LastUpdated:  result.LastUpdated,
		Active:       result.Active,
	}

	return result.Name, stats, nil
}

func (s *Service) LoadPaginatedBlacklist(ctx context.Context, page, pageSize int, search string) ([]string, int, error) {
	blacklistEntries, total, err := s.repository.GetPaginatedDomains(ctx, page, pageSize, search)
	if err != nil {
		return nil, 0, err
	}

	domains := make([]string, len(blacklistEntries))
	for i, entry := range blacklistEntries {
		domains[i] = entry.Domain
	}

	return domains, int(total), nil
}

func (s *Service) Vacuum(ctx context.Context) {
	log.Debug("Vacuuming database...")
	if err := s.repository.Vacuum(ctx); err != nil {
		log.Warning("Error while vacuuming database: %v", err)
	}
}

func (s *Service) ScheduleAutomaticListUpdates() {
	ticker := time.NewTicker(s.config.UpdateInterval)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		log.Info("Starting automatic list updates...")

		for _, source := range s.blocklistURL {
			if source.Name == "Custom" {
				continue
			}
			log.Info("Checking for updates for blocklist %s from %s", source.Name, source.URL)

			availableUpdate, err := s.CheckIfUpdateAvailable(ctx, source.URL, source.Name)
			if err != nil {
				log.Warning("Failed to check for updates for %s: %v", source.Name, err)
				continue
			}

			if !availableUpdate.UpdateAvailable {
				log.Info("No updates available for %s", source.Name)
				continue
			}

			if err := s.RemoveSourceAndDomains(ctx, source.Name, source.URL); err != nil {
				log.Warning("Failed to remove old domains for %s: %v", source.Name, err)
				continue
			}

			if err := s.FetchAndLoadHosts(ctx, source.URL, source.Name); err != nil {
				log.Warning("Failed to fetch and load hosts for %s: %v", source.Name, err)
				continue
			}

			log.Info("Successfully updated %s with %d new domains", source.Name, len(availableUpdate.DiffAdded))
		}

		if err := s.PopulateCache(ctx); err != nil {
			log.Warning("Failed to populate blocklist cache after auto-update: %v", err)
		}
	}
}
