package server

import (
	"fmt"
	notification "goaway/backend"
	arp "goaway/backend/dns"
	"goaway/backend/dns/database"
	model "goaway/backend/dns/server/models"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

func (s *DNSServer) processQuery(request *Request) model.RequestLogEntry {
	domainName := strings.TrimRight(request.question.Name, ".")

	if request.question.Qtype == dns.TypePTR || strings.HasSuffix(domainName, "in-addr.arpa.") {
		return s.handlePTRQuery(request)
	}

	if s.Status.Paused && time.Since(s.Status.PausedAt).Seconds() >= float64(s.Status.PauseTime) {
		s.Status.Paused = false
	}

	if !s.Status.Paused && s.Blacklist.IsBlacklisted(domainName) {
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

func (s *DNSServer) getClientInfo(remoteAddr string) *model.Client {
	clientIP, _, _ := net.SplitHostPort(remoteAddr)

	if cachedClient, ok := s.clientCache.Load(clientIP); ok {
		return cachedClient.(*model.Client)
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

	client := model.Client{IP: resultIP, Name: hostname, MAC: macAddress}
	s.clientCache.Store(clientIP, &client)
	return &client
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
		Status:            dns.RcodeToString[dns.RcodeSuccess],
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
		Status:            dns.RcodeToString[dns.RcodeSuccess],
		QueryType:         dns.TypeToString[request.question.Qtype],
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
	answers, _, status := s.queryUpstream(request)
	request.msg.Answer = append(request.msg.Answer, answers...)
	request.msg.Rcode = dns.RcodeNameError

	var resolvedHostnames []string
	for _, answer := range answers {
		if ptr, ok := answer.(*dns.PTR); ok {
			resolvedHostnames = append(resolvedHostnames, ptr.Ptr)
		}
	}

	_ = request.w.WriteMsg(request.msg)

	return model.RequestLogEntry{
		Domain:            request.question.Name,
		Status:            status,
		QueryType:         dns.TypeToString[request.question.Qclass],
		IP:                resolvedHostnames,
		ResponseSizeBytes: request.msg.Len(),
		Timestamp:         request.sent,
		ResponseTime:      time.Since(request.sent),
		ClientInfo:        request.client,
	}
}

func (s *DNSServer) handleStandardQuery(req *Request) model.RequestLogEntry {
	answers, cached, status := s.resolve(req)

	resolved := make([]string, 0, len(answers))
	req.msg.Answer = answers

	for _, a := range answers {
		switch rr := a.(type) {
		case *dns.A:
			resolved = append(resolved, rr.A.String())
		case *dns.AAAA:
			resolved = append(resolved, rr.AAAA.String())
		case *dns.PTR:
			resolved = append(resolved, rr.Ptr)
		case *dns.CNAME:
			resolved = append(resolved, rr.Target)
		}
	}

	if len(answers) == 0 {
		req.msg.Rcode = dns.RcodeServerFailure
	}

	_ = req.w.WriteMsg(req.msg)

	return model.RequestLogEntry{
		Domain:            req.question.Name,
		Status:            status,
		QueryType:         dns.TypeToString[req.question.Qtype],
		IP:                resolved,
		ResponseSizeBytes: req.msg.Len(),
		Timestamp:         req.sent,
		ResponseTime:      time.Since(req.sent),
		Cached:            cached,
		ClientInfo:        req.client,
	}
}

func (s *DNSServer) resolve(req *Request) ([]dns.RR, bool, string) {
	cacheKey := req.question.Name + ":" + strconv.Itoa(int(req.question.Qtype))
	if cached, found := s.cache.Load(cacheKey); found {
		if ipAddresses, valid := s.getCachedRecord(cached); valid {
			return ipAddresses, true, dns.RcodeToString[dns.RcodeSuccess]
		}
	}

	if answers, ttl, status := s.queryUpstream(req); len(answers) > 0 {
		s.cacheRecord(cacheKey, answers, ttl)
		return answers, false, status
	}

	if answers, ttl, status := s.resolveResolution(req.question.Name); len(answers) > 0 {
		s.cacheRecord(cacheKey, answers, ttl)
		return answers, false, status
	}

	answers, ttl, status := s.resolveCNAMEChain(req, make(map[string]bool))
	if len(answers) > 0 {
		s.cacheRecord(cacheKey, answers, ttl)
	}
	return answers, false, status
}

func (s *DNSServer) resolveResolution(domain string) ([]dns.RR, uint32, string) {
	var (
		records []dns.RR
		ttl     = uint32(s.Config.DNS.CacheTTL)
		status  = dns.RcodeToString[dns.RcodeSuccess]
	)

	ipFound, err := database.FetchResolution(s.DB, domain)
	if err != nil {
		log.Error("Database lookup error for domain (%s): %v", domain, err)
		return nil, 0, dns.RcodeToString[dns.RcodeServerFailure]
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
		status = dns.RcodeToString[dns.RcodeNameError]
	}

	return records, ttl, status
}

func (s *DNSServer) resolveCNAMEChain(req *Request, visited map[string]bool) ([]dns.RR, uint32, string) {
	if visited[req.question.Name] {
		return nil, 0, dns.RcodeToString[dns.RcodeServerFailure]
	}
	visited[req.question.Name] = true

	answers, ttl, status := s.queryUpstream(req)
	if len(answers) > 0 {
		for _, answer := range answers {
			if _, ok := answer.(*dns.CNAME); ok {
				targetAnswers, targetTTL, targetStatus := s.resolveCNAMEChain(req, visited)
				if len(targetAnswers) > 0 {
					minTTL := min(targetTTL, ttl)
					return append(answers, targetAnswers...), minTTL, targetStatus
				}
				return answers, ttl, status
			}
		}
	}

	return answers, ttl, status
}

func (s *DNSServer) queryUpstream(req *Request) ([]dns.RR, uint32, string) {
	resultCh := make(chan *dns.Msg, 1)
	errCh := make(chan error, 1)

	go func() {
		in, _, err := s.dnsClient.Exchange(req.msg, s.Config.DNS.PreferredUpstream)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- in
	}()

	select {
	case in := <-resultCh:
		status := dns.RcodeToString[dns.RcodeServerFailure]
		if statusStr, ok := dns.RcodeToString[in.Rcode]; ok {
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
		log.Warning("Resolution error for domain (%s): %v", req.question.Name, err)
		s.Notifications.CreateNotification(&notification.Notification{
			Severity: notification.SeverityWarning,
			Category: notification.CategoryDNS,
			Text:     fmt.Sprintf("Resolution error for domain (%s)", req.question.Name),
		})
		return nil, 0, dns.RcodeToString[dns.RcodeServerFailure]

	case <-time.After(5 * time.Second):
		log.Warning("DNS lookup for %s timed out", req.question.Name)
		return nil, 0, dns.RcodeToString[dns.RcodeServerFailure]
	}
}

func (s *DNSServer) handleBlacklisted(request *Request) model.RequestLogEntry {
	request.msg.Rcode = dns.RcodeSuccess
	rr4 := &dns.A{
		Hdr: dns.RR_Header{
			Name:   request.question.Name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    uint32(s.Config.DNS.CacheTTL),
		},
		A: net.ParseIP("0.0.0.0"),
	}

	rr6 := &dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   request.question.Name,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    uint32(s.Config.DNS.CacheTTL),
		},
		AAAA: net.ParseIP("::"),
	}

	request.msg.Answer = append(request.msg.Answer, rr4, rr6)
	_ = request.w.WriteMsg(request.msg)

	return model.RequestLogEntry{
		Domain:            request.question.Name,
		Status:            dns.RcodeToString[dns.RcodeSuccess],
		QueryType:         dns.TypeToString[request.question.Qtype],
		ResponseSizeBytes: request.msg.Len(),
		Timestamp:         request.sent,
		ResponseTime:      time.Since(request.sent),
		Blocked:           true,
		ClientInfo:        request.client,
	}
}
