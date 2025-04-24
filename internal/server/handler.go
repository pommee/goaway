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

	lru "github.com/hashicorp/golang-lru/v2"
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

var hostCache, _ = lru.New[string, any](100)

func (s *DNSServer) processQuery(request *Request) model.RequestLogEntry {
	domainName := strings.TrimSuffix(request.question.Name, ".")

	if request.question.Qtype == dns.TypePTR || strings.HasSuffix(domainName, "in-addr.arpa.") {
		return s.handlePTRQuery(request)
	}

	if s.Status.Paused {
		if time.Since(s.Status.PausedAt).Seconds() >= float64(s.Status.PauseTime) {
			s.Status.Paused = false
		}
	}

	if !s.Status.Paused && s.Blacklist.IsBlacklisted(domainName) {
		log.Info("Blocked: %s", request.question.Name)
		return s.handleBlacklisted(request)
	}

	return s.handleStandardQuery(request)
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
	clientIP, _, _ := net.SplitHostPort(remoteAddr)

	if cachedInfo, found := hostCache.Get(clientIP); found {
		info := cachedInfo.([]string)
		return info[0], info[1], info[2]
	}

	macAddress := arp.GetMacAddress(clientIP)
	hostname := "unknown"
	resultIP := clientIP

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
		resultIP = localIP

		if h, err := os.Hostname(); err == nil {
			hostname = h
		} else {
			hostname = "localhost"
		}
	} else if hostnames, err := net.LookupAddr(clientIP); err == nil && len(hostnames) > 0 {
		hostname = strings.TrimSuffix(hostnames[0], ".")
	}

	hostCache.Add(clientIP, []string{resultIP, hostname, macAddress})
	return resultIP, hostname, macAddress
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

func (s *DNSServer) handlePTRQuery(request *Request) model.RequestLogEntry {
	ipParts := strings.TrimSuffix(request.question.Name, ".in-addr.arpa.")
	parts := strings.Split(ipParts, ".")

	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	ipStr := strings.Join(parts, ".")

	if ipStr == "127.0.0.1" {
		return s.respondWithLocalhost(request)
	}

	if !isPrivateIP(ipStr) {
		return s.forwardPTRQueryUpstream(request)
	}

	hostname := database.GetClientNameFromRequestLog(s.DB, ipStr)
	if hostname == "unknown" {
		if names, err := net.LookupAddr(ipStr); err == nil && len(names) > 0 {
			hostname = strings.TrimSuffix(names[0], ".")
		}
	}

	if hostname != "unknown" {
		return s.respondWithHostname(request, hostname)
	}

	return s.forwardPTRQueryUpstream(request)
}

func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	_, private24, _ := net.ParseCIDR("192.168.0.0/16")
	_, private20, _ := net.ParseCIDR("172.16.0.0/12")
	_, private16, _ := net.ParseCIDR("10.0.0.0/8")
	return private24.Contains(ip) || private20.Contains(ip) || private16.Contains(ip)
}

func (s *DNSServer) respondWithLocalhost(request *Request) model.RequestLogEntry {
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
		Timestamp:         request.sent,
		Domain:            request.question.Name,
		Status:            "NoError",
		IP:                []string{"localhost.lan"},
		Blocked:           false,
		Cached:            false,
		ResponseTime:      time.Since(request.sent),
		ClientInfo:        request.client,
		QueryType:         "PTR",
		ResponseSizeBytes: request.msg.Len(),
	}
}

func (s *DNSServer) respondWithHostname(request *Request, hostname string) model.RequestLogEntry {
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
		Domain:            request.question.Name,
		Status:            "NoError",
		QueryType:         "PTR",
		IP:                []string{hostname},
		ResponseSizeBytes: request.msg.Len(),
		Timestamp:         request.sent,
		ResponseTime:      time.Since(request.sent),
		Blocked:           false,
		Cached:            false,
		ClientInfo:        request.client,
	}
}

func (s *DNSServer) forwardPTRQueryUpstream(request *Request) model.RequestLogEntry {
	answers, _, status := s.queryUpstream(request.question.Name, request.question.Qtype)
	request.msg.Answer = append(request.msg.Answer, answers...)

	switch status {
	case "NXDomain":
		request.msg.Rcode = dns.RcodeNameError
	case "ServFail":
		request.msg.Rcode = dns.RcodeServerFailure
	}

	var resolvedHostnames []string
	for _, answer := range answers {
		if ptr, ok := answer.(*dns.PTR); ok {
			resolvedHostnames = append(resolvedHostnames, ptr.Ptr)
		}
	}

	_ = request.w.WriteMsg(request.msg)
	s.Counters.AllowedRequests++

	return model.RequestLogEntry{
		Domain:            request.question.Name,
		Status:            status,
		QueryType:         "PTR",
		IP:                resolvedHostnames,
		ResponseSizeBytes: request.msg.Len(),
		Timestamp:         request.sent,
		ResponseTime:      time.Since(request.sent),
		ClientInfo:        request.client,
	}
}

func (s *DNSServer) handleStandardQuery(request *Request) model.RequestLogEntry {
	queryStart := time.Now()
	answers, cached, status := s.resolve(request.question.Name, request.question.Qtype)
	upstreamQueryTime := time.Since(queryStart)

	resolvedAddresses := make([]string, 0, len(answers))
	if len(answers) > 0 {
		request.msg.Answer = append(request.msg.Answer, answers...)

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

	switch status {
	case "NXDomain":
		request.msg.Rcode = dns.RcodeNameError
	case "ServFail":
		request.msg.Rcode = dns.RcodeServerFailure
	}

	_ = request.w.WriteMsg(request.msg)
	s.Counters.AllowedRequests++

	processTimeNonQuery := time.Since(request.sent) - upstreamQueryTime
	log.Debug("Processing took: %s", processTimeNonQuery)

	return model.RequestLogEntry{
		Domain:            request.question.Name,
		Status:            status,
		QueryType:         dns.TypeToString[request.question.Qtype],
		IP:                resolvedAddresses,
		ResponseSizeBytes: request.msg.Len(),
		Timestamp:         request.sent,
		ResponseTime:      time.Since(request.sent),
		Cached:            cached,
		ClientInfo:        request.client,
	}
}

func (s *DNSServer) resolve(domain string, qtype uint16) ([]dns.RR, bool, string) {
	cacheKey := fmt.Sprintf("%s-%d", domain, qtype)
	if cached, found := s.cache.Load(cacheKey); found {
		if ipAddresses, valid := s.getCachedRecord(cached); valid {
			return ipAddresses, true, "NoError"
		}
	}

	answers, ttl, status := s.resolveResolution(domain)
	if len(answers) > 0 {
		s.cacheRecord(cacheKey, answers, ttl)
		return answers, false, status
	}

	answers, ttl, status = s.resolveCNAMEChain(domain, qtype, make(map[string]bool))
	if len(answers) > 0 {
		s.cacheRecord(cacheKey, answers, ttl)
	}

	return answers, false, status
}

func (s *DNSServer) resolveResolution(domain string) ([]dns.RR, uint32, string) {
	var (
		records []dns.RR
		ttl     = uint32(s.Config.CacheTTL.Seconds())
		status  = "NOERROR"
	)

	ipFound, err := database.FetchResolution(s.DB, domain)
	if err != nil {
		log.Error("Database lookup error for domain (%s): %v", domain, err)
		return nil, 0, "SERVFAIL"
	}

	if net.ParseIP(ipFound) != nil {
		var rr dns.RR
		if strings.Contains(ipFound, ":") {
			rr = &dns.AAAA{
				Hdr:  dns.RR_Header{Name: dns.Fqdn(domain), Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: ttl},
				AAAA: net.ParseIP(ipFound),
			}
		} else {
			rr = &dns.A{
				Hdr: dns.RR_Header{Name: dns.Fqdn(domain), Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ttl},
				A:   net.ParseIP(ipFound),
			}
		}
		records = append(records, rr)
	} else {
		status = "NXDOMAIN"
	}

	return records, ttl, status
}

func (s *DNSServer) resolveCNAMEChain(domain string, qtype uint16, visited map[string]bool) ([]dns.RR, uint32, string) {
	if visited[domain] {
		return nil, 0, "SERVFAIL"
	}
	visited[domain] = true

	answers, ttl, status := s.queryUpstream(domain, qtype)
	if len(answers) > 0 {
		return answers, ttl, status
	}

	cnameAnswers, cnameTTL, cnameStatus := s.queryUpstream(domain, dns.TypeCNAME)
	if len(cnameAnswers) > 0 {
		for _, answer := range cnameAnswers {
			if cname, ok := answer.(*dns.CNAME); ok {
				targetAnswers, targetTTL, targetStatus := s.resolveCNAMEChain(cname.Target, qtype, visited)
				if len(targetAnswers) > 0 {
					minTTL := min(targetTTL, cnameTTL)
					return append(cnameAnswers, targetAnswers...), minTTL, targetStatus
				}
				return cnameAnswers, cnameTTL, cnameStatus
			}
		}
	}

	return answers, ttl, status
}

func (s *DNSServer) queryUpstream(domain string, qtype uint16) ([]dns.RR, uint32, string) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), qtype)
	m.RecursionDesired = true
	m.SetEdns0(4096, false)

	resultCh := make(chan *dns.Msg, 1)
	errCh := make(chan error, 1)

	go func() {
		in, _, err := s.dnsClient.Exchange(m, s.Config.PreferredUpstream)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- in
	}()

	select {
	case in := <-resultCh:
		status := "SERVFAIL"
		if statusStr, ok := rcodes[in.Rcode]; ok {
			status = statusStr
		}

		var ttl uint32 = 3600
		if len(in.Answer) > 0 {
			ttl = in.Answer[0].Header().Ttl
			for _, a := range in.Answer {
				if a.Header().Ttl < ttl {
					ttl = a.Header().Ttl
				}
			}
		}

		return in.Answer, ttl, status

	case err := <-errCh:
		log.Error("Resolution error for domain (%s): %v", domain, err)
		return nil, 0, "SERVFAIL"

	case <-time.After(5 * time.Second):
		log.Warning("DNS lookup for %s timed out", domain)
		return nil, 0, "SERVFAIL"
	}
}

func (s *DNSServer) handleBlacklisted(request *Request) model.RequestLogEntry {
	request.msg.Rcode = dns.RcodeSuccess
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
		Domain:            request.question.Name,
		Status:            "Blacklisted",
		QueryType:         dns.TypeToString[request.question.Qtype],
		ResponseSizeBytes: request.msg.Len(),
		Timestamp:         request.sent,
		ResponseTime:      time.Since(request.sent),
		Blocked:           true,
		ClientInfo:        request.client,
	}
}
