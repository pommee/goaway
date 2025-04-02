package server

import (
	"time"

	"github.com/miekg/dns"
)

func (s *DNSServer) getCachedRecord(cached interface{}) ([]dns.RR, bool) {
	cachedRecord, ok := cached.(cachedRecord)
	if !ok {
		return nil, false
	}

	if time.Now().Before(cachedRecord.ExpiresAt) {
		remainingSeconds := uint32(time.Until(cachedRecord.ExpiresAt).Seconds())
		updatedRecords := make([]dns.RR, 0, len(cachedRecord.IPAddresses))

		for _, rr := range cachedRecord.IPAddresses {
			clone := dns.Copy(rr)
			clone.Header().Ttl = remainingSeconds
			updatedRecords = append(updatedRecords, clone)
		}

		return updatedRecords, true
	}

	return nil, false
}

func (s *DNSServer) cacheRecord(domain string, ipAddresses []dns.RR, ttl uint32) {
	cacheTTL := s.Config.CacheTTL
	if ttl > 0 {
		cacheTTL = time.Duration(ttl) * time.Second
	}

	now := time.Now()
	s.cache.Store(domain, cachedRecord{
		IPAddresses: ipAddresses,
		ExpiresAt:   now.Add(cacheTTL),
		CachedAt:    now,
		OriginalTTL: ttl,
	})
}
