package server

import (
	"time"

	"github.com/miekg/dns"
)

func (s *DNSServer) getCachedRecord(cached interface{}) ([]dns.RR, bool) {
	cachedRecord := cached.(cachedRecord)
	if time.Now().Before(cachedRecord.ExpiresAt) {
		return cachedRecord.IPAddresses, true
	}
	return nil, false
}

func (s *DNSServer) cacheRecord(domain string, ipAddresses []dns.RR, ttl uint32) {
	cacheTTL := s.Config.CacheTTL
	if ttl > 0 {
		cacheTTL = time.Duration(ttl) * time.Second
	}
	s.cache.Store(domain, cachedRecord{
		IPAddresses: ipAddresses,
		ExpiresAt:   time.Now().Add(cacheTTL),
	})
}
