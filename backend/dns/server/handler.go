package server

import (
	"bufio"
	"context"
	"fmt"
	arp "goaway/backend/dns"
	"goaway/backend/dns/database"
	model "goaway/backend/dns/server/models"
	notification "goaway/backend/notifications"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

var (
	blackholeIPv4 = net.ParseIP("0.0.0.0")
	blackholeIPv6 = net.ParseIP("::")
)

func trimDomainDot(name string) string {
	if len(name) > 0 && name[len(name)-1] == '.' {
		return name[:len(name)-1]
	}
	return name
}

func isPTRQuery(request *Request, domainName string) bool {
	return request.Question.Qtype == dns.TypePTR || strings.HasSuffix(domainName, "in-addr.arpa.")
}

func (s *DNSServer) checkAndUpdatePauseStatus() {
	if s.Config.DNS.Status.Paused &&
		time.Since(s.Config.DNS.Status.PausedAt).Seconds() >= float64(s.Config.DNS.Status.PauseTime) {
		s.Config.DNS.Status.Paused = false
	}
}

func (s *DNSServer) shouldBlockQuery(domainName, fullName string) bool {
	return !s.Config.DNS.Status.Paused &&
		s.Blacklist.IsBlacklisted(domainName) &&
		!s.Whitelist.IsWhitelisted(fullName)
}

func (s *DNSServer) processQuery(request *Request) model.RequestLogEntry {
	domainName := trimDomainDot(request.Question.Name)

	if isPTRQuery(request, domainName) {
		return s.handlePTRQuery(request)
	}

	if ip, found := s.reverseHostnameLookup(request.Question.Name); found {
		return s.respondWithHostnameA(request, ip)
	}

	s.checkAndUpdatePauseStatus()

	if s.shouldBlockQuery(domainName, request.Question.Name) {
		return s.handleBlacklisted(request)
	}

	if isLocalLookup(request.Question.Name) {
		val, err := s.LocalForwardLookup(request)
		if err != nil {
			log.Debug("Reverse lookup failed for %s: %v", request.Question.Name, err)
		} else {
			return val
		}
	}

	return s.handleStandardQuery(request)
}

func (s *DNSServer) GetVendor(mac string) (string, error) {
	s.DBManager.Mutex.Lock()
	defer s.DBManager.Mutex.Unlock()
	return database.FindVendor(s.DBManager.Conn, mac)
}

func (s *DNSServer) SaveMacVendor(clientIP, mac, vendor string) {
	s.DBManager.Mutex.Lock()
	defer s.DBManager.Mutex.Unlock()

	log.Debug("Saving new MAC address: %s %s", mac, vendor)
	database.SaveMacEntry(s.DBManager.Conn, clientIP, mac, vendor)
}

func (s *DNSServer) reverseHostnameLookup(requestedHostname string) (string, bool) {
	trimmed := strings.TrimSuffix(requestedHostname, ".")

	if value, ok := s.hostnameCache.Load(trimmed); ok {
		if ip, ok := value.(string); ok {
			return ip, true
		}
	}

	return "", false
}

func (s *DNSServer) getClientInfo(remoteAddr string) *model.Client {
	clientIP, _, _ := net.SplitHostPort(remoteAddr)

	if cachedClient, ok := s.clientCache.Load(clientIP); ok {
		return cachedClient.(*model.Client)
	}

	macAddress := arp.GetMacAddress(clientIP)
	hostname := s.resolveHostname(clientIP)
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
	}

	client := model.Client{IP: resultIP, Name: hostname, MAC: macAddress}
	s.clientCache.Store(clientIP, &client)

	if client.Name != "unknown" {
		s.hostnameCache.Store(client.Name, client.IP)
	}

	return &client
}

func (s *DNSServer) resolveHostname(clientIP string) string {
	if hostname := s.reverseDNSLookup(clientIP); hostname != "unknown" {
		return hostname
	}

	if hostname := s.avahiLookup(clientIP); hostname != "unknown" {
		return hostname
	}

	if hostname := s.sshBannerLookup(clientIP); hostname != "unknown" {
		return hostname
	}

	return "unknown"
}

func (s *DNSServer) avahiLookup(clientIP string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "avahi-resolve-address", clientIP)
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, clientIP) {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					hostname := strings.TrimSuffix(parts[1], ".local")
					if hostname != "" && hostname != clientIP {
						log.Debug("Found hostname via avahi-resolve: %s -> %s", clientIP, hostname)
						return hostname
					}
				}
			}
		}
	}

	return "unknown"
}

func (s *DNSServer) reverseDNSLookup(clientIP string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resolver := &net.Resolver{}
	if hostnames, err := resolver.LookupAddr(ctx, clientIP); err == nil && len(hostnames) > 0 {
		hostname := strings.TrimSuffix(hostnames[0], ".")
		if hostname != clientIP &&
			!strings.Contains(hostname, "in-addr.arpa") && !strings.HasPrefix(hostname, clientIP) {
			log.Debug("Found hostname via reverse DNS: %s -> %s", clientIP, hostname)
			return hostname
		}
	}
	return "unknown"
}

func (s *DNSServer) sshBannerLookup(clientIP string) string {
	conn, err := net.DialTimeout("tcp", clientIP+":22", 1*time.Second)
	if err != nil {
		return "unknown"
	}
	defer func() {
		_ = conn.Close()
	}()

	err = conn.SetDeadline(time.Now().Add(2 * time.Second))
	if err != nil {
		log.Warning("Failed to set deadline for SSH banner lookup: %v", err)
		_ = conn.Close()
		return "unknown"
	}

	reader := bufio.NewReader(conn)
	banner, err := reader.ReadString('\n')
	if err != nil {
		return "unknown"
	}

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`SSH-2\.0-OpenSSH_[0-9.]+.*?(\w+)`),
		regexp.MustCompile(`SSH.*?(\w+)\.local`),
		regexp.MustCompile(`(\w+)@(\w+)`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(banner)
		if len(matches) > 1 {
			hostname := matches[1]
			if hostname != clientIP && len(hostname) > 1 && hostname != "SSH" {
				log.Debug("Found hostname via SSH banner: %s -> %s", clientIP, hostname)
				return hostname
			}
		}
	}

	return "unknown"
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
	ipParts := strings.TrimSuffix(request.Question.Name, ".in-addr.arpa.")
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

	hostname := database.GetClientNameFromRequestLog(s.DBManager.Conn, ipStr)
	if hostname == "unknown" {
		hostname = s.resolveHostname(ipStr)
	}

	if hostname != "unknown" {
		return s.respondWithHostnamePTR(request, hostname)
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
	request.Msg.Response = true
	request.Msg.Authoritative = false
	request.Msg.RecursionAvailable = true
	request.Msg.Rcode = dns.RcodeSuccess

	ptr := &dns.PTR{
		Hdr: dns.RR_Header{
			Name:   request.Question.Name,
			Rrtype: dns.TypePTR,
			Class:  dns.ClassINET,
			Ttl:    3600,
		},
		Ptr: "localhost.lan.",
	}

	request.Msg.Answer = []dns.RR{ptr}
	_ = request.ResponseWriter.WriteMsg(request.Msg)

	return model.RequestLogEntry{
		Timestamp: request.Sent,
		Domain:    request.Question.Name,
		Status:    dns.RcodeToString[dns.RcodeSuccess],
		IP: []model.ResolvedIP{
			{
				IP:    "localhost.lan",
				RType: "PTR",
			},
		},
		Blocked:           false,
		Cached:            false,
		ResponseTime:      time.Since(request.Sent),
		ClientInfo:        request.Client,
		QueryType:         "PTR",
		ResponseSizeBytes: request.Msg.Len(),
		Protocol:          request.Protocol,
	}
}

func (s *DNSServer) respondWithHostnameA(request *Request, hostIP string) model.RequestLogEntry {
	request.Msg.Response = true
	request.Msg.Authoritative = false
	request.Msg.RecursionAvailable = true
	request.Msg.Rcode = dns.RcodeSuccess

	response := &dns.A{
		Hdr: dns.RR_Header{
			Name:   request.Question.Name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    60,
		},
		A: net.ParseIP(hostIP),
	}

	request.Msg.Answer = []dns.RR{response}
	_ = request.ResponseWriter.WriteMsg(request.Msg)

	return s.respondWithType(request, dns.TypeA, hostIP)
}

func (s *DNSServer) respondWithHostnamePTR(request *Request, hostname string) model.RequestLogEntry {
	request.Msg.Response = true
	request.Msg.Authoritative = false
	request.Msg.RecursionAvailable = true
	request.Msg.Rcode = dns.RcodeSuccess

	ptr := &dns.PTR{
		Hdr: dns.RR_Header{
			Name:   request.Question.Name,
			Rrtype: dns.TypePTR,
			Class:  dns.ClassINET,
			Ttl:    3600,
		},
		Ptr: hostname + ".",
	}

	request.Msg.Answer = []dns.RR{ptr}
	_ = request.ResponseWriter.WriteMsg(request.Msg)

	return s.respondWithType(request, dns.TypePTR, hostname)
}

func (s *DNSServer) respondWithType(request *Request, rType uint16, ip string) model.RequestLogEntry {
	return model.RequestLogEntry{
		Domain:    request.Question.Name,
		Status:    dns.RcodeToString[dns.RcodeSuccess],
		QueryType: dns.TypeToString[request.Question.Qtype],
		IP: []model.ResolvedIP{
			{
				IP:    ip,
				RType: dns.TypeToString[rType],
			},
		},
		ResponseSizeBytes: request.Msg.Len(),
		Timestamp:         request.Sent,
		ResponseTime:      time.Since(request.Sent),
		Blocked:           false,
		Cached:            false,
		ClientInfo:        request.Client,
		Protocol:          request.Protocol,
	}
}

func (s *DNSServer) forwardPTRQueryUpstream(request *Request) model.RequestLogEntry {
	answers, _, status := s.QueryUpstream(request)
	request.Msg.Answer = append(request.Msg.Answer, answers...)

	if rcode, ok := dns.StringToRcode[status]; ok {
		request.Msg.Rcode = rcode
	} else {
		request.Msg.Rcode = dns.RcodeServerFailure
	}

	request.Msg.Response = true
	request.Msg.Authoritative = false
	request.Msg.RecursionAvailable = true

	var resolvedHostnames []model.ResolvedIP
	for _, answer := range answers {
		if ptr, ok := answer.(*dns.PTR); ok {
			resolvedHostnames = append(resolvedHostnames, model.ResolvedIP{
				IP:    ptr.Ptr,
				RType: "PTR",
			})
		}
	}

	_ = request.ResponseWriter.WriteMsg(request.Msg)

	return model.RequestLogEntry{
		Domain:            request.Question.Name,
		Status:            status,
		QueryType:         dns.TypeToString[request.Question.Qtype],
		IP:                resolvedHostnames,
		ResponseSizeBytes: request.Msg.Len(),
		Timestamp:         request.Sent,
		ResponseTime:      time.Since(request.Sent),
		ClientInfo:        request.Client,
		Protocol:          request.Protocol,
	}
}

func (s *DNSServer) handleStandardQuery(request *Request) model.RequestLogEntry {
	answers, cached, status := s.Resolve(request)
	resolved := make([]model.ResolvedIP, 0, len(answers))

	request.Msg.Answer = answers
	request.Msg.Response = true
	request.Msg.Authoritative = false
	if request.Msg.RecursionDesired {
		request.Msg.RecursionAvailable = true
	}
	if len(answers) == 0 {
		request.Msg.Rcode = dns.RcodeServerFailure
	}

	for _, a := range answers {
		switch rr := a.(type) {
		case *dns.A:
			resolved = append(resolved, model.ResolvedIP{
				IP:    rr.A.String(),
				RType: "A",
			})
		case *dns.AAAA:
			resolved = append(resolved, model.ResolvedIP{
				IP:    rr.AAAA.String(),
				RType: "AAAA",
			})
		case *dns.PTR:
			resolved = append(resolved, model.ResolvedIP{
				IP:    rr.Ptr,
				RType: "PTR",
			})
		case *dns.CNAME:
			resolved = append(resolved, model.ResolvedIP{
				IP:    rr.Target,
				RType: "CNAME",
			})
		case *dns.SVCB:
			resolved = append(resolved, model.ResolvedIP{
				IP:    rr.Target,
				RType: "SVCB",
			})
		case *dns.MX:
			resolved = append(resolved, model.ResolvedIP{
				IP:    rr.Mx,
				RType: "MX",
			})
		case *dns.TXT:
			resolved = append(resolved, model.ResolvedIP{
				IP:    rr.Txt[0],
				RType: "TXT",
			})
		case *dns.NS:
			resolved = append(resolved, model.ResolvedIP{
				IP:    rr.Ns,
				RType: "NS",
			})
		case *dns.SOA:
			resolved = append(resolved, model.ResolvedIP{
				IP:    rr.Ns,
				RType: "SOA",
			})
		case *dns.SRV:
			resolved = append(resolved, model.ResolvedIP{
				IP:    fmt.Sprintf("%s:%d", rr.Target, rr.Port),
				RType: "SRV",
			})
		case *dns.HTTPS:
			resolved = append(resolved, model.ResolvedIP{
				IP:    rr.Target,
				RType: "HTTPS",
			})
		case *dns.CAA:
			resolved = append(resolved, model.ResolvedIP{
				IP:    fmt.Sprintf("%s: %s", rr.Tag, rr.Value),
				RType: "CAA",
			})
		case *dns.DNSKEY:
			resolved = append(resolved, model.ResolvedIP{
				IP:    fmt.Sprintf("flags:%d protocol:%d algorithm:%d", rr.Flags, rr.Protocol, rr.Algorithm),
				RType: "DNSKEY",
			})
		default:
			log.Warning("Unhandled record type '%s' while requesting '%s'", dns.TypeToString[rr.Header().Rrtype], request.Question.Name)
		}
	}

	err := request.ResponseWriter.WriteMsg(request.Msg)
	if err != nil {
		log.Warning("Could not write query response. client: [%s] with query [%v], err: %v", request.Client.IP, request.Msg.Answer, err.Error())
		s.Notifications.CreateNotification(&notification.Notification{
			Severity: notification.SeverityWarning,
			Category: notification.CategoryDNS,
			Text:     fmt.Sprintf("Could not write query response. Client: %s, err: %v", request.Client.IP, err.Error()),
		})
	}

	return model.RequestLogEntry{
		Domain:            request.Question.Name,
		Status:            status,
		QueryType:         dns.TypeToString[request.Question.Qtype],
		IP:                resolved,
		ResponseSizeBytes: request.Msg.Len(),
		Timestamp:         request.Sent,
		ResponseTime:      time.Since(request.Sent),
		Cached:            cached,
		ClientInfo:        request.Client,
		Protocol:          request.Protocol,
	}
}

func (s *DNSServer) Resolve(req *Request) ([]dns.RR, bool, string) {
	cacheKey := req.Question.Name + ":" + strconv.Itoa(int(req.Question.Qtype))
	if cached, found := s.Cache.Load(cacheKey); found {
		if ipAddresses, valid := s.getCachedRecord(cached); valid {
			return ipAddresses, true, dns.RcodeToString[dns.RcodeSuccess]
		}
	}

	if answers, ttl, status := s.resolveResolution(req.Question.Name); len(answers) > 0 {
		s.CacheRecord(cacheKey, req.Question.Name, answers, ttl)
		return answers, false, status
	}

	if answers, ttl, status := s.QueryUpstream(req); status != dns.RcodeToString[dns.RcodeServerFailure] {
		s.CacheRecord(cacheKey, req.Question.Name, answers, ttl)
		return answers, false, status
	}

	answers, ttl, status := s.resolveCNAMEChain(req, make(map[string]bool))
	if len(answers) > 0 {
		s.CacheRecord(cacheKey, req.Question.Name, answers, ttl)
	}
	return answers, false, status
}

func (s *DNSServer) resolveResolution(domain string) ([]dns.RR, uint32, string) {
	var (
		records []dns.RR
		ttl     = uint32(s.Config.DNS.CacheTTL)
		status  = dns.RcodeToString[dns.RcodeSuccess]
	)

	ipFound, err := database.FetchResolution(s.DBManager.Conn, domain)
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
	if visited[req.Question.Name] {
		return nil, 0, dns.RcodeToString[dns.RcodeServerFailure]
	}
	visited[req.Question.Name] = true

	answers, ttl, status := s.QueryUpstream(req)
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

func (s *DNSServer) QueryUpstream(req *Request) ([]dns.RR, uint32, string) {
	resultCh := make(chan *dns.Msg, 1)
	errCh := make(chan error, 1)

	go func() {
		go s.WSCom(communicationMessage{false, true, false, ""})

		upstreamMsg := &dns.Msg{}
		upstreamMsg.SetQuestion(req.Question.Name, req.Question.Qtype)
		upstreamMsg.RecursionDesired = true
		upstreamMsg.Id = dns.Id()

		upstream := s.Config.DNS.PreferredUpstream
		if s.dnsClient.Net == "tcp-tls" && !strings.HasSuffix(upstream, ":853") {
			host, _, err := net.SplitHostPort(upstream)
			if err == nil {
				upstream = net.JoinHostPort(host, "853")
			}
		}
		in, _, err := s.dnsClient.Exchange(upstreamMsg, upstream)
		if err != nil {
			errCh <- err
			return
		}

		if in == nil {
			errCh <- fmt.Errorf("nil response from upstream")
			return
		}

		resultCh <- in
	}()

	select {
	case in := <-resultCh:
		go s.WSCom(communicationMessage{false, false, true, ""})

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
		} else if len(in.Ns) > 0 {
			ttl = in.Ns[0].Header().Ttl
		}

		if len(in.Ns) > 0 {
			req.Msg.Ns = make([]dns.RR, len(in.Ns))
			copy(req.Msg.Ns, in.Ns)
		}
		if len(in.Extra) > 0 {
			req.Msg.Extra = make([]dns.RR, len(in.Extra))
			copy(req.Msg.Extra, in.Extra)
		}

		return in.Answer, ttl, status

	case err := <-errCh:
		log.Warning("Resolution error for domain (%s): %v", req.Question.Name, err)
		s.Notifications.CreateNotification(&notification.Notification{
			Severity: notification.SeverityWarning,
			Category: notification.CategoryDNS,
			Text:     fmt.Sprintf("Resolution error for domain (%s)", req.Question.Name),
		})
		return nil, 0, dns.RcodeToString[dns.RcodeServerFailure]

	case <-time.After(5 * time.Second):
		log.Warning("DNS lookup for %s timed out", req.Question.Name)
		return nil, 0, dns.RcodeToString[dns.RcodeServerFailure]
	}
}

func (s *DNSServer) LocalForwardLookup(req *Request) (model.RequestLogEntry, error) {
	hostname := strings.ReplaceAll(req.Question.Name, ".in-addr.arpa.", "")
	hostname = strings.ReplaceAll(hostname, ".ip6.arpa.", "")
	if !strings.HasSuffix(hostname, ".") {
		hostname += "."
	}

	dnsMsg := new(dns.Msg)
	dnsMsg.SetQuestion(hostname, dns.TypeA)

	client := &dns.Client{Net: "udp"}
	start := time.Now()
	in, _, err := client.Exchange(dnsMsg, s.Config.DNS.Gateway)
	responseTime := time.Since(start)

	if err != nil {
		log.Error("DNS exchange error for %s: %v", hostname, err)
		return model.RequestLogEntry{}, fmt.Errorf("forward DNS query failed: %w", err)
	}

	if in.Rcode != dns.RcodeSuccess {
		status := dns.RcodeToString[in.Rcode]
		log.Info("DNS query for %s returned status %s", hostname, status)
		return model.RequestLogEntry{}, fmt.Errorf("forward lookup failed with status: %s", status)
	}

	var ips []model.ResolvedIP
	for _, answer := range in.Answer {
		if a, ok := answer.(*dns.A); ok {
			ips = append(ips, model.ResolvedIP{IP: a.A.String()})
		}
	}

	if len(ips) == 0 {
		return model.RequestLogEntry{}, fmt.Errorf("no A records found for hostname: %s", hostname)
	}

	req.Msg.Rcode = in.Rcode
	req.Msg.Answer = in.Answer
	if writeErr := req.ResponseWriter.WriteMsg(req.Msg); writeErr != nil {
		log.Error("failed to write DNS response: %v", writeErr)
	}

	entry := model.RequestLogEntry{
		Domain:            req.Question.Name,
		Status:            dns.RcodeToString[in.Rcode],
		QueryType:         dns.TypeToString[dns.TypeA],
		IP:                ips,
		ResponseSizeBytes: in.Len(),
		Timestamp:         start,
		ResponseTime:      responseTime,
		Blocked:           false,
		Cached:            false,
		ClientInfo:        req.Client,
		Protocol:          model.UDP,
	}

	return entry, nil
}

func isLocalLookup(qname string) bool {
	return strings.HasSuffix(qname, ".in-addr.arpa.") || strings.HasSuffix(qname, ".ip6.arpa.")
}

func (s *DNSServer) handleBlacklisted(request *Request) model.RequestLogEntry {
	request.Msg.Response = true
	request.Msg.Authoritative = false
	request.Msg.RecursionAvailable = true
	request.Msg.Rcode = dns.RcodeSuccess

	var resolved []model.ResolvedIP
	cacheTTL := uint32(s.Config.DNS.CacheTTL)

	switch request.Question.Qtype {
	case dns.TypeA:
		request.Msg.Answer = []dns.RR{&dns.A{
			Hdr: dns.RR_Header{
				Name:   request.Question.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    cacheTTL,
			},
			A: blackholeIPv4,
		}}
		resolved = []model.ResolvedIP{{IP: blackholeIPv4.String(), RType: "A"}}
	case dns.TypeAAAA:
		request.Msg.Answer = []dns.RR{&dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   request.Question.Name,
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    cacheTTL,
			},
			AAAA: blackholeIPv6,
		}}
		resolved = []model.ResolvedIP{{IP: blackholeIPv6.String(), RType: "AAAA"}}
	default:
		request.Msg.Rcode = dns.RcodeNameError
		request.Msg.Answer = nil
		resolved = nil
	}

	if len(request.Msg.Question) == 0 {
		request.Msg.Question = []dns.Question{request.Question}
	}

	_ = request.ResponseWriter.WriteMsg(request.Msg)

	return model.RequestLogEntry{
		Domain:            request.Question.Name,
		Status:            dns.RcodeToString[request.Msg.Rcode],
		QueryType:         dns.TypeToString[request.Question.Qtype],
		IP:                resolved,
		ResponseSizeBytes: request.Msg.Len(),
		Timestamp:         request.Sent,
		ResponseTime:      time.Since(request.Sent),
		Blocked:           true,
		Cached:            false,
		ClientInfo:        request.Client,
		Protocol:          request.Protocol,
	}
}
