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

	now := time.Now()
	if now.Before(cachedRecord.ExpiresAt) {
		remainingSeconds := uint32(cachedRecord.ExpiresAt.Sub(now).Seconds())
		updatedRecords := make([]dns.RR, len(cachedRecord.IPAddresses))

		for i, rr := range cachedRecord.IPAddresses {
			if rr.Header().Ttl != remainingSeconds {
				clone := dns.Copy(rr)
				clone.Header().Ttl = remainingSeconds
				updatedRecords[i] = clone
			} else {
				updatedRecords[i] = rr
			}
		}

		return updatedRecords, true
	}

	log.Debug("Cached response was expired")
	if cachedRecord.Key != "" {
		s.cache.Delete(cachedRecord.Key)
	}

	return nil, false
}

func (s *DNSServer) cacheRecord(domain string, ipAddresses []dns.RR, ttl uint32) {
	if len(ipAddresses) == 0 {
		return
	}

	// TODO:  Make use of the user set TTL
	// As of now we are using the DNS sent TTL time
	// cacheTTL := s.Config.CacheTTL * time.Microsecond

	now := time.Now()
	s.cache.Store(domain, cachedRecord{
		IPAddresses: ipAddresses,
		ExpiresAt:   now.Add(time.Duration(ttl) * time.Second),
		CachedAt:    now,
		OriginalTTL: ttl,
		Key:         domain,
	})
}
