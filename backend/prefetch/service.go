package prefetch

import (
	"fmt"
	"goaway/backend/database"
	"goaway/backend/dns/server"
	"goaway/backend/logging"
	"strconv"
	"time"

	"github.com/miekg/dns"
)

type Service struct {
	repository Repository
	DNS        *server.DNSServer
	Domains    map[string]database.Prefetch
}

var log = logging.GetLogger()

func NewService(repo Repository, dnsServer *server.DNSServer) *Service {
	service := &Service{
		repository: repo,
		DNS:        dnsServer,
		Domains:    make(map[string]database.Prefetch),
	}

	service.LoadPrefetchedDomains()
	return service
}

func (s *Service) Run() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		s.checkNewDomains()
		s.processExpiredEntries()
	}
}

func (s *Service) checkNewDomains() {
	for domain, prefetchDomain := range s.Domains {
		cacheKey := s.buildCacheKey(domain, dns.Type(prefetchDomain.QueryType))
		if _, exists := s.DNS.DomainCache.Load(cacheKey); !exists {
			log.Debug("Prefetching new/missing domain: %s", domain)
			s.prefetchDomain(prefetchDomain)
		}
	}
}

func (s *Service) processExpiredEntries() {
	now := time.Now()
	var expiredKeys []interface{}
	var removeFromDomains []string

	s.DNS.DomainCache.Range(func(key, value interface{}) bool {
		cachedDomain, ok := value.(server.CachedRecord)
		if !ok {
			log.Debug("Cache entry type assertion failed for key: %v", key)
			return true
		}

		if s.isExpired(cachedDomain, now) {
			expiredKeys = append(expiredKeys, key)

			if _, isPrefetched := s.Domains[cachedDomain.Domain]; !isPrefetched {
				removeFromDomains = append(removeFromDomains, cachedDomain.Domain)
				log.Debug("Non-prefetch entry '%v' expired and will be removed", key)
			} else {
				log.Debug("Prefetch entry '%v' expired and will be refreshed", key)
			}
		}
		return true
	})

	s.handleExpiredKeys(expiredKeys)
	s.removeNonPrefetchDomains(removeFromDomains)
}

func (s *Service) isExpired(record server.CachedRecord, now time.Time) bool {
	return now.After(record.ExpiresAt) || now.Equal(record.ExpiresAt)
}

func (s *Service) handleExpiredKeys(expiredKeys []interface{}) {
	for _, key := range expiredKeys {
		if value, exists := s.DNS.DomainCache.Load(key); exists {
			if cachedDomain, ok := value.(server.CachedRecord); ok {
				s.DNS.DomainCache.Delete(key)
				s.handleExpiredEntry(cachedDomain)
			}
		}
	}
}

func (s *Service) removeNonPrefetchDomains(domains []string) {
	for _, domain := range domains {
		delete(s.Domains, domain)
	}
}

func (s *Service) prefetchDomain(prefetchDomain database.Prefetch) {
	question := dns.Question{
		Name:   prefetchDomain.Domain,
		Qtype:  uint16(prefetchDomain.QueryType),
		Qclass: 1,
	}

	request := &server.Request{
		Msg:      &dns.Msg{Question: []dns.Question{question}},
		Question: question,
		Sent:     time.Now(),
		Prefetch: true,
	}

	answers, ttl, _ := s.DNS.QueryUpstream(request)
	cacheKey := s.buildCacheKey(question.Name, dns.Type(question.Qtype))
	s.DNS.CacheRecord(cacheKey, prefetchDomain.Domain, answers, ttl)
}

func (s *Service) buildCacheKey(domain string, qtype dns.Type) string {
	return domain + ":" + strconv.Itoa(int(qtype))
}

func (s *Service) handleExpiredEntry(record server.CachedRecord) {
	domain := record.IPAddresses[0].Header().Name
	prefetchDomain, exists := s.Domains[domain]

	if !exists {
		log.Debug("%s not set to be prefetched", domain)
		return
	}

	log.Debug("Prefetching expired domain: %s", domain)
	s.prefetchDomain(prefetchDomain)
}

func (s *Service) LoadPrefetchedDomains() {
	prefetched, err := s.repository.GetAll()
	if err != nil {
		log.Error("failed to load prefetched domains: %v", err)
		return
	}

	for _, p := range prefetched {
		s.Domains[p.Domain] = p
	}

	if len(s.Domains) > 0 {
		log.Info("Loaded %d prefetched domain(s)", len(s.Domains))
	}
}

func (s *Service) AddPrefetchedDomain(domain string, refresh, qtype int) error {
	prefetch := database.Prefetch{
		Domain:    domain,
		Refresh:   refresh,
		QueryType: qtype,
	}

	err := s.repository.Create(&prefetch)
	if err != nil {
		return fmt.Errorf("failed to add new domain to prefetch table: %w", err)
	}

	s.Domains[domain] = prefetch

	log.Info("%s was added as a prefetched domain", domain)
	return nil
}

func (s *Service) RemovePrefetchedDomain(domain string) error {
	err := s.repository.Delete(domain)
	if err != nil {
		return fmt.Errorf("failed to remove %s from prefetch table: %w", domain, err)
	}

	delete(s.Domains, domain)
	log.Info("%s was removed as a prefetched domain", domain)
	return nil
}
