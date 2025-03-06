package server

import (
	"fmt"
	"goaway/internal/arp"
	"goaway/internal/database"
	model "goaway/internal/server/models"
	"net"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
)

var rcodes = map[int]string{
	dns.RcodeSuccess:        "NoError",
	dns.RcodeFormatError:    "FormErr",
	dns.RcodeServerFailure:  "ServFail",
	dns.RcodeNameError:      "NXDomain",
	dns.RcodeNotImplemented: "NotImp",
	dns.RcodeRefused:        "Refused",
	dns.RcodeYXDomain:       "YXDomain",
	dns.RcodeYXRrset:        "YXRRSet",
	dns.RcodeNXRrset:        "NXRRSet",
	dns.RcodeNotAuth:        "NotAuth",
	dns.RcodeNotZone:        "NotZone",
	dns.RcodeBadSig:         "BADSIG",
	dns.RcodeBadKey:         "BADKEY",
	dns.RcodeBadTime:        "BADTIME",
	dns.RcodeBadMode:        "BADMODE",
	dns.RcodeBadName:        "BADNAME",
	dns.RcodeBadAlg:         "BADALG",
	dns.RcodeBadTrunc:       "BADTRUNC",
	dns.RcodeBadCookie:      "BADCOOKIE",
}

func (s *DNSServer) GetVendor(mac string) (string, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	return database.FindVendor(s.DB, mac)
}

func (s *DNSServer) SaveMacVendor(clientIP, mac, vendor string) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	log.Debug("Saving new MAC address: %s %s", mac, vendor)
	database.SaveMacEntry(s.DB, clientIP, mac, vendor)
}

func (s *DNSServer) getClientInfo(remoteAddr string) (string, string, string) {
	clientIP := strings.Split(remoteAddr, ":")[0]
	macAddress := arp.GetMacAddress(clientIP)

	vendor, err := s.GetVendor(macAddress)
	if macAddress != "unknown" {
		if err != nil || vendor == "" {
			log.Debug("Lookup vendor for mac %s", macAddress)
			vendor, err = arp.GetMacVendor(macAddress)
			if err == nil {
				s.SaveMacVendor(clientIP, macAddress, vendor)
			} else {
				log.Warning("Error while lookup mac address vendor: %v", err)
			}
		}
	}

	if clientIP == "127.0.0.1" || clientIP == "::1" || clientIP == "[" {
		localIP, err := getLocalIP()
		if err != nil {
			log.Warning("Failed to get local IP: %v", err)
			localIP = "127.0.0.1"
		}

		if h, err := os.Hostname(); err == nil {
			return localIP, h, macAddress
		}
		return localIP, "localhost", macAddress
	}

	if hostnames, err := net.LookupAddr(clientIP); err == nil && len(hostnames) > 0 {
		return clientIP, strings.TrimSuffix(hostnames[0], "."), macAddress
	}

	return clientIP, "unknown", macAddress
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "127.0.0.1", fmt.Errorf("no non-loopback IPv4 address found")
}

func (s *DNSServer) processQuery(request *Request) model.RequestLogEntry {
	if request.question.Qtype == dns.TypePTR && strings.HasSuffix(request.question.Name, ".in-addr.arpa.") {
		return s.handlePTRQuery(request)
	}

	isBlacklisted, err := s.Blacklist.IsBlacklisted(strings.TrimSuffix(request.question.Name, "."))
	if err != nil {
		log.Error("%v", err)
	}
	if isBlacklisted {
		return s.handleBlacklisted(request)
	}
	return s.handleQuery(request)
}

func (s *DNSServer) handlePTRQuery(request *Request) model.RequestLogEntry {
	ptrName := request.question.Name
	ipParts := strings.TrimSuffix(ptrName, ".in-addr.arpa.")
	parts := strings.Split(ipParts, ".")

	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	ipStr := strings.Join(parts, ".")

	if ipStr == "127.0.0.1" {
		ptr := &dns.PTR{
			Hdr: dns.RR_Header{
				Name:   request.question.Name,
				Rrtype: dns.TypePTR,
				Class:  dns.ClassINET,
				Ttl:    3600,
			},
			Ptr: "localhost.lan.",
		}

		request.msg.Answer = append(request.msg.Answer, ptr)
		_ = request.w.WriteMsg(request.msg)

		return model.RequestLogEntry{
			Timestamp:    request.sent,
			Domain:       request.question.Name,
			IP:           []string{"localhost.lan"},
			Blocked:      false,
			Cached:       false,
			ResponseTime: time.Since(request.sent),
			ClientInfo:   request.client,
			Status:       "NoError",
			QueryType:    "PTR",
		}
	}

	hostname := database.GetClientNameFromRequestLog(s.DB, ipStr)

	if hostname == "unknown" {
		if names, err := net.LookupAddr(ipStr); err == nil && len(names) > 0 {
			hostname = strings.TrimSuffix(names[0], ".")
		}
	}

	if hostname != "unknown" {
		ptr := &dns.PTR{
			Hdr: dns.RR_Header{
				Name:   request.question.Name,
				Rrtype: dns.TypePTR,
				Class:  dns.ClassINET,
				Ttl:    3600,
			},
			Ptr: hostname + ".",
		}

		request.msg.Answer = append(request.msg.Answer, ptr)
		_ = request.w.WriteMsg(request.msg)

		return model.RequestLogEntry{
			Timestamp:    request.sent,
			Domain:       request.question.Name,
			IP:           []string{hostname},
			Blocked:      false,
			Cached:       false,
			ResponseTime: time.Since(request.sent),
			ClientInfo:   request.client,
			Status:       "NoError",
			QueryType:    "PTR",
		}
	}

	return s.handleQuery(request)
}

func (s *DNSServer) handleQuery(request *Request) model.RequestLogEntry {
	answers, cached, status := s.resolve(request.question.Name, request.question.Qclass)
	request.msg.Answer = append(request.msg.Answer, answers...)

	if status == "NXDomain" {
		request.msg.Rcode = dns.RcodeNameError
	} else if status == "ServFail" {
		request.msg.Rcode = dns.RcodeServerFailure
	}

	var resolvedAddresses []string
	if len(answers) > 0 {
		for _, answer := range answers {
			switch rec := answer.(type) {
			case *dns.A:
				resolvedAddresses = append(resolvedAddresses, rec.A.String())
			case *dns.AAAA:
				resolvedAddresses = append(resolvedAddresses, rec.AAAA.String())
			case *dns.PTR:
				resolvedAddresses = append(resolvedAddresses, rec.Ptr)
			case *dns.CNAME:
				resolvedAddresses = append(resolvedAddresses, rec.Target)
			}
		}
	}

	_ = request.w.WriteMsg(request.msg)
	s.Counters.AllowedRequests++

	return model.RequestLogEntry{
		Timestamp:    request.sent,
		Domain:       request.question.Name,
		IP:           resolvedAddresses,
		Blocked:      false,
		Cached:       cached,
		ResponseTime: time.Since(request.sent),
		ClientInfo:   request.client,
		Status:       status,
		QueryType:    dns.TypeToString[request.question.Qtype],
	}
}

func (s *DNSServer) resolve(domain string, qtype uint16) ([]dns.RR, bool, string) {
	cacheKey := fmt.Sprintf("%s-%d", domain, qtype)
	if cached, found := s.cache.Load(cacheKey); found {
		if ipAddresses, valid := s.getCachedRecord(cached); valid {
			return ipAddresses, true, "NoError"
		}
	}

	answers, ttl, status := s.resolveCNAMEChain(domain, qtype, make(map[string]bool))
	if len(answers) > 0 {
		s.cacheRecord(cacheKey, answers, ttl)
	}

	return answers, false, status
}

func (s *DNSServer) resolveCNAMEChain(domain string, qtype uint16, visited map[string]bool) ([]dns.RR, uint32, string) {
	if visited[domain] {
		return nil, 0, "SERVFAIL"
	}
	visited[domain] = true

	answers, ttl, status := s.queryUpstream(domain, qtype)
	if len(answers) == 0 {
		cnameAnswers, cnameTTL, cnameStatus := s.queryUpstream(domain, dns.TypeCNAME)
		if len(cnameAnswers) > 0 {
			for _, answer := range cnameAnswers {
				if cname, ok := answer.(*dns.CNAME); ok {
					targetAnswers, targetTTL, targetStatus := s.resolveCNAMEChain(cname.Target, qtype, visited)
					if len(targetAnswers) > 0 {
						minTTL := cnameTTL
						if targetTTL < minTTL {
							minTTL = targetTTL
						}
						return append(cnameAnswers, targetAnswers...), minTTL, targetStatus
					}
					return cnameAnswers, cnameTTL, cnameStatus
				}
			}
		}
	}
	return answers, ttl, status
}

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

func (s *DNSServer) queryUpstream(domain string, qtype uint16) ([]dns.RR, uint32, string) {
	var (
		ipAddresses []dns.RR
		ttl         uint32
		status      = "SERVFAIL"
	)
	done := make(chan struct{})

	go func() {
		defer close(done)
		c := new(dns.Client)

		m := new(dns.Msg)
		m.SetQuestion(dns.Fqdn(domain), qtype)
		m.RecursionDesired = true

		in, _, err := c.Exchange(m, s.Config.PreferredUpstream)
		if err != nil {
			log.Error("Resolution error for domain (%s): %v", domain, err)
			return
		}

		if statusStr, ok := rcodes[in.Rcode]; ok {
			status = statusStr
		}

		ipAddresses = append(ipAddresses, in.Answer...)
		if len(in.Answer) > 0 && ttl == 0 {
			ttl = in.Answer[0].Header().Ttl
		}

		ipAddresses = append(ipAddresses, in.Ns...)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		log.Warning("DNS lookup for %s timed out", domain)
		status = "SERVFAIL"
	}

	return ipAddresses, ttl, status
}

func (s *DNSServer) handleBlacklisted(request *Request) model.RequestLogEntry {
	log.Info("Blocked: %s", request.question.Name)

	request.msg.Rcode = dns.RcodeSuccess
	var status = "Blacklisted"

	rr4 := &dns.A{
		Hdr: dns.RR_Header{
			Name:   request.question.Name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    uint32(s.Config.CacheTTL.Seconds()),
		},
		A: net.ParseIP("0.0.0.0"),
	}

	rr6 := &dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   request.question.Name,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    uint32(s.Config.CacheTTL.Seconds()),
		},
		AAAA: net.ParseIP("::"),
	}

	request.msg.Answer = append(request.msg.Answer, rr4, rr6)
	_ = request.w.WriteMsg(request.msg)

	s.Counters.BlockedRequests++

	return model.RequestLogEntry{
		Timestamp:    request.sent,
		Domain:       request.question.Name,
		IP:           []string{""},
		Blocked:      true,
		Cached:       false,
		ResponseTime: time.Since(request.sent),
		ClientInfo:   request.client,
		Status:       status,
	}
}

func (s *DNSServer) ProcessLogEntries() {
	var batch []model.RequestLogEntry
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case entry := <-s.logEntryChannel:
			batch = append(batch, entry)
			if len(batch) >= batchSize {
				s.saveBatch(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				s.saveBatch(batch)
				batch = nil
			}
		}
	}
}

func (s *DNSServer) saveBatch(entries []model.RequestLogEntry) {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	database.SaveRequestLog(s.DB, entries)
}

func (s *DNSServer) ClearOldEntries() {
	const (
		maxRetries      = 10
		retryDelay      = 150 * time.Millisecond
		cleanupInterval = 1 * time.Minute
	)

	for {
		requestThreshold := ((60 * 60) * 24) * s.StatisticsRetention
		log.Debug("Next cleanup running at %s", time.Now().Add(cleanupInterval).Format(time.DateTime))
		time.Sleep(cleanupInterval)

		database.DeleteRequestLogsTimebased(s.DB, requestThreshold, maxRetries, retryDelay)
		s.UpdateCounters()
	}
}

func (s *DNSServer) UpdateCounters() {
	blockedCount, allowedCount, err := database.CountAllowedAndBlockedRequest(s.DB)
	if err != nil {
		log.Error("%s %v", "failed to get counters: ", err)
	}
	s.Counters.BlockedRequests = blockedCount
	s.Counters.AllowedRequests = allowedCount
}
