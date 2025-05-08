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

	if cachedRecord.Key != "" {
		log.Debug("Cached entry has expired, removing %s from cache", cachedRecord.Key)
		s.cache.Delete(cachedRecord.Key)
	}

	return nil, false
}

func (s *DNSServer) cacheRecord(domain string, ipAddresses []dns.RR, ttl uint32) {
	if len(ipAddresses) == 0 {
		return
	}

	cacheTTL := time.Duration(s.Config.DNS.CacheTTL) * time.Second
	if ttl > 0 {
		recordTTL := time.Duration(ttl) * time.Second
		if recordTTL < cacheTTL {
			cacheTTL = recordTTL
		}
	}

	now := time.Now()
	s.cache.Store(domain, cachedRecord{
		IPAddresses: ipAddresses,
		ExpiresAt:   now.Add(cacheTTL),
		CachedAt:    now,
		OriginalTTL: ttl,
		Key:         domain,
	})
}
