package prefetch

import (
	"fmt"
	"goaway/backend/dns/database"
	"goaway/backend/dns/server"
	"goaway/backend/logging"
	"strconv"
	"time"

	"github.com/miekg/dns"
)

var log = logging.GetLogger()

type Manager struct {
	dbManager *database.DatabaseManager
	DNS       *server.DNSServer
	Domains   map[string]PrefetchedDomain
}

type PrefetchedDomain struct {
	Domain  string `json:"domain"`
	Refresh int    `json:"refresh"`
	Qtype   int    `json:"qtype"`
}

func (manager *Manager) LoadPrefetchedDomains() {
	var prefetched []database.Prefetch
	if err := manager.dbManager.Conn.Find(&prefetched).Error; err != nil {
		log.Warning("Failed to query prefetch table: %v", err)
		return
	}

	for _, p := range prefetched {
		manager.Domains[p.Domain] = PrefetchedDomain{
			Domain:  p.Domain,
			Refresh: p.Refresh,
			Qtype:   p.QType,
		}
	}

	if len(manager.Domains) > 0 {
		log.Info("Loaded %d prefetched domain(s)", len(manager.Domains))
	}
}

func (manager *Manager) AddPrefetchedDomain(domain string, refresh, qtype int) error {
	prefetch := database.Prefetch{
		Domain:  domain,
		Refresh: refresh,
		QType:   qtype,
	}

	result := manager.dbManager.Conn.FirstOrCreate(&prefetch, database.Prefetch{Domain: domain})
	if result.Error != nil {
		return fmt.Errorf("failed to add new domain to prefetch table: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("%s already exists", domain)
	}

	manager.Domains[domain] = PrefetchedDomain{
		Domain:  domain,
		Refresh: refresh,
		Qtype:   qtype,
	}

	log.Info("%s was added as a prefetched domain", domain)
	return nil
}

func (manager *Manager) RemovePrefetchedDomain(domain string) error {
	result := manager.dbManager.Conn.Delete(&database.Prefetch{}, "domain = ?", domain)
	if result.Error != nil {
		return fmt.Errorf("failed to remove %s from prefetch table: %w", domain, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("%s does not exist in the database", domain)
	}

	delete(manager.Domains, domain)
	log.Info("%s was removed as a prefetched domain", domain)
	return nil
}

func New(dnsServer *server.DNSServer) Manager {
	manager := Manager{
		dbManager: dnsServer.DBManager,
		DNS:       dnsServer,
		Domains:   make(map[string]PrefetchedDomain),
	}

	manager.LoadPrefetchedDomains()
	return manager
}

func (manager *Manager) Run() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		manager.checkNewDomains()
		manager.processExpiredEntries()
	}
}

func (manager *Manager) checkNewDomains() {
	for domain, prefetchDomain := range manager.Domains {
		cacheKey := manager.buildCacheKey(domain, dns.Type(prefetchDomain.Qtype))
		if _, exists := manager.DNS.Cache.Load(cacheKey); !exists {
			log.Debug("Prefetching new/missing domain: %s", domain)
			manager.prefetchDomain(prefetchDomain)
		}
	}
}

func (manager *Manager) processExpiredEntries() {
	now := time.Now()
	var expiredKeys []interface{}
	var removeFromDomains []string

	manager.DNS.Cache.Range(func(key, value interface{}) bool {
		cachedDomain, ok := value.(server.CachedRecord)
		if !ok {
			log.Debug("Cache entry type assertion failed for key: %v", key)
			return true
		}

		if manager.isExpired(cachedDomain, now) {
			expiredKeys = append(expiredKeys, key)

			if _, isPrefetched := manager.Domains[cachedDomain.Domain]; !isPrefetched {
				removeFromDomains = append(removeFromDomains, cachedDomain.Domain)
				log.Debug("Non-prefetch entry '%v' expired and will be removed", key)
			} else {
				log.Debug("Prefetch entry '%v' expired and will be refreshed", key)
			}
		}
		return true
	})

	manager.handleExpiredKeys(expiredKeys)
	manager.removeNonPrefetchDomains(removeFromDomains)
}

func (manager *Manager) isExpired(record server.CachedRecord, now time.Time) bool {
	return now.After(record.ExpiresAt) || now.Equal(record.ExpiresAt)
}

func (manager *Manager) handleExpiredKeys(expiredKeys []interface{}) {
	for _, key := range expiredKeys {
		if value, exists := manager.DNS.Cache.Load(key); exists {
			if cachedDomain, ok := value.(server.CachedRecord); ok {
				manager.DNS.Cache.Delete(key)
				manager.handleExpiredEntry(cachedDomain)
			}
		}
	}
}

func (manager *Manager) removeNonPrefetchDomains(domains []string) {
	for _, domain := range domains {
		delete(manager.Domains, domain)
	}
}

func (manager *Manager) prefetchDomain(prefetchDomain PrefetchedDomain) {
	question := dns.Question{
		Name:   prefetchDomain.Domain,
		Qtype:  uint16(prefetchDomain.Qtype),
		Qclass: 1,
	}

	request := &server.Request{
		Msg:      &dns.Msg{Question: []dns.Question{question}},
		Question: question,
		Sent:     time.Now(),
		Prefetch: true,
	}

	answers, ttl, _ := manager.DNS.QueryUpstream(request)
	cacheKey := manager.buildCacheKey(question.Name, dns.Type(question.Qtype))
	manager.DNS.CacheRecord(cacheKey, prefetchDomain.Domain, answers, ttl)
}

func (manager *Manager) handleExpiredEntry(record server.CachedRecord) {
	domain := record.IPAddresses[0].Header().Name
	prefetchDomain, exists := manager.Domains[domain]

	if !exists {
		log.Debug("%s not set to be prefetched", domain)
		return
	}

	log.Debug("Prefetching expired domain: %s", domain)
	manager.prefetchDomain(prefetchDomain)
}

func (manager *Manager) buildCacheKey(domain string, qtype dns.Type) string {
	return domain + ":" + strconv.Itoa(int(qtype))
}
