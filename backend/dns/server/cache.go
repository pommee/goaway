package server

import (
	"time"

	"github.com/miekg/dns"
)

func (s *DNSServer) getCachedRecord(cached interface{}) ([]dns.RR, bool) {
	cachedRecord, ok := cached.(CachedRecord)
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
		s.DomainCache.Delete(cachedRecord.Key)
	}

	return nil, false
}

func (s *DNSServer) RemoveCachedDomain(domain string) {
	if domain == "" {
		return
	}

	s.DomainCache.Range(func(key, value interface{}) bool {
		cachedRecord, ok := value.(CachedRecord)
		if !ok || cachedRecord.Domain != domain+"." {
			return true
		}

		log.Debug("Removing cached record for domain %s", domain)
		s.DomainCache.Delete(key)
		return true
	})
}

func (s *DNSServer) CacheRecord(cacheKey, domain string, ipAddresses []dns.RR, ttl uint32) {
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
	s.DomainCache.Store(cacheKey, CachedRecord{
		IPAddresses: ipAddresses,
		ExpiresAt:   now.Add(cacheTTL),
		CachedAt:    now,
		OriginalTTL: ttl,
		Key:         cacheKey,
		Domain:      domain,
	})
}
