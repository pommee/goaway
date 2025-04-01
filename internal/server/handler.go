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

func (s *DNSServer) processQuery(request *Request) model.RequestLogEntry {
	if request.question.Qtype == dns.TypePTR || strings.HasSuffix(request.question.Name, ".in-addr.arpa.") {
		return s.handlePTRQuery(request)
	}

	domainName := strings.TrimSuffix(request.question.Name, ".")
	isBlacklisted, err := s.Blacklist.IsBlacklisted(domainName)
	if err != nil {
		log.Error("%v", err)
	}

	if isBlacklisted {
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

func (s *DNSServer) handlePTRQuery(request *Request) model.RequestLogEntry {
	ptrName := request.question.Name
	ipParts := strings.TrimSuffix(ptrName, ".in-addr.arpa.")
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

	responseSizeBytes := request.msg.Len()
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
		ResponseSizeBytes: responseSizeBytes,
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

	responseSizeBytes := request.msg.Len()
	request.msg.Answer = append(request.msg.Answer, ptr)
	_ = request.w.WriteMsg(request.msg)

	return model.RequestLogEntry{
		Domain:            request.question.Name,
		Status:            "NoError",
		QueryType:         "PTR",
		IP:                []string{hostname},
		ResponseSizeBytes: responseSizeBytes,
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

	if status == "NXDomain" {
		request.msg.Rcode = dns.RcodeNameError
	} else if status == "ServFail" {
		request.msg.Rcode = dns.RcodeServerFailure
	}

	var resolvedHostnames []string
	for _, answer := range answers {
		if ptr, ok := answer.(*dns.PTR); ok {
			resolvedHostnames = append(resolvedHostnames, ptr.Ptr)
		}
	}

	responseSizeBytes := request.msg.Len()
	_ = request.w.WriteMsg(request.msg)
	s.Counters.AllowedRequests++

	return model.RequestLogEntry{
		Domain:            request.question.Name,
		Status:            status,
		QueryType:         "PTR",
		IP:                resolvedHostnames,
		ResponseSizeBytes: responseSizeBytes,
		Timestamp:         request.sent,
		ResponseTime:      time.Since(request.sent),
		Blocked:           false,
		Cached:            false,
		ClientInfo:        request.client,
	}
}

func (s *DNSServer) handleStandardQuery(request *Request) model.RequestLogEntry {
	answers, cached, status := s.resolve(request.question.Name, request.question.Qtype)
	request.msg.Answer = append(request.msg.Answer, answers...)

	if status == "NXDomain" {
		request.msg.Rcode = dns.RcodeNameError
	} else if status == "ServFail" {
		request.msg.Rcode = dns.RcodeServerFailure
	}

	var resolvedAddresses []string
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

	responseSizeBytes := request.msg.Len()
	_ = request.w.WriteMsg(request.msg)
	s.Counters.AllowedRequests++

	return model.RequestLogEntry{
		Domain:            request.question.Name,
		Status:            status,
		QueryType:         dns.TypeToString[request.question.Qtype],
		IP:                resolvedAddresses,
		ResponseSizeBytes: responseSizeBytes,
		Timestamp:         request.sent,
		ResponseTime:      time.Since(request.sent),
		Blocked:           false,
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
		ttl     uint32 = uint32(s.Config.CacheTTL.Seconds())
		status  string = "NOERROR"
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

	return answers, ttl, status
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

	responseSizeBytes := request.msg.Len()
	request.msg.Answer = append(request.msg.Answer, rr4, rr6)
	_ = request.w.WriteMsg(request.msg)

	s.Counters.BlockedRequests++

	return model.RequestLogEntry{
		Domain:            request.question.Name,
		Status:            status,
		QueryType:         dns.TypeToString[request.question.Qtype],
		IP:                []string{""},
		ResponseSizeBytes: responseSizeBytes,
		Timestamp:         request.sent,
		ResponseTime:      time.Since(request.sent),
		Blocked:           true,
		Cached:            false,
		ClientInfo:        request.client,
	}
}
