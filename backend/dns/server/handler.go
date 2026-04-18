package server

import (
	"bufio"
	"context"
	"fmt"
	arp "goaway/backend/dns"
	model "goaway/backend/dns/server/models"
	"goaway/backend/notification"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"codeberg.org/miekg/dns/rdata"
)

var (
	blackholeIPv4 = netip.MustParseAddr("0.0.0.0")
	blackholeIPv6 = netip.MustParseAddr("::")
	IPv4Loopback  = netip.MustParseAddr("127.0.0.1")
)

const (
	unknownHostname = "unknown"
)

func trimDomainDot(name string) string {
	if name != "" && name[len(name)-1] == '.' {
		return name[:len(name)-1]
	}
	return name
}

func isPTRQuery(request *Request, domainName string) bool {
	return request.Question.Header().Class == dns.TypePTR || strings.HasSuffix(domainName, "in-addr.arpa.")
}

func (s *DNSServer) checkAndUpdatePauseStatus() {
	if s.Config.DNS.Status.Paused &&
		s.Config.DNS.Status.PausedAt.After(s.Config.DNS.Status.PauseTime) {
		s.Config.DNS.Status.Paused = false
	}
}

func (s *DNSServer) shouldBlockQuery(client *model.Client, domainName, fullName string) bool {
	if client.Bypass {
		log.Debug("Allowing client '%s' to bypass %s", client.IP, fullName)
		return false
	}

	return !s.Config.DNS.Status.Paused &&
		s.BlacklistService.IsBlacklisted(domainName) &&
		!s.WhitelistService.IsWhitelisted(fullName)
}

func (s *DNSServer) processQuery(request *Request) model.RequestLogEntry {
	domainName := trimDomainDot(request.Question.Header().Name)

	if isPTRQuery(request, domainName) {
		return s.handlePTRQuery(request)
	}

	if ip, found := s.reverseHostnameLookup(domainName); found {
		return s.respondWithHostnameA(request, ip)
	}

	s.checkAndUpdatePauseStatus()

	if s.shouldBlockQuery(request.Client, domainName, domainName) {
		return s.handleBlacklisted(request)
	}

	if isLocalLookup(domainName) {
		val, err := s.LocalForwardLookup(request)
		if err != nil {
			log.Debug("Reverse lookup failed for %s: %v", domainName, err)
		} else {
			return val
		}
	}

	return s.handleStandardQuery(request)
}

func (s *DNSServer) reverseHostnameLookup(requestedHostname string) (netip.Addr, bool) {
	trimmed := strings.TrimSuffix(requestedHostname, ".")
	if value, ok := s.clientHostnameCache.Load(trimmed); ok {
		if client, ok := value.(*model.Client); ok {
			return client.IP, true
		}
	}

	return netip.Addr{}, false
}

func (s *DNSServer) getClientInfo(clientIP netip.Addr) *model.Client {
	var isLoopback = clientIP.IsLoopback()
	if isLoopback {
		if localIP, err := getLocalIP(); err == nil {
			clientIP = localIP
		} else {
			log.Warning("Failed to get local IP: %v", err)
			clientIP = IPv4Loopback
		}
	}

	if loaded, ok := s.clientIPCache.Load(clientIP); ok {
		if client, ok := loaded.(*model.Client); ok {
			return client
		}
	}

	macAddress := arp.GetMacAddress(clientIP)
	hostname := s.resolveHostname(clientIP)

	if isLoopback {
		if h, err := os.Hostname(); err == nil {
			hostname = h
		} else {
			hostname = "localhost"
		}
	}

	vendor := s.lookupVendor(clientIP.String(), macAddress)
	client := &model.Client{
		IP:       clientIP,
		LastSeen: time.Now(),
		Name:     hostname,
		Mac:      macAddress,
		Vendor:   vendor,
		Bypass:   false,
	}

	log.Debug("Saving new client: %s", client.IP)
	_ = s.PopulateClientCaches()

	return client
}

func (s *DNSServer) lookupVendor(clientIP, macAddress string) string {
	if macAddress == unknownHostname {
		return ""
	}

	vendor, err := s.MACService.FindVendor(macAddress)
	if err == nil && vendor != "" {
		return vendor
	}

	log.Debug("Lookup vendor for mac %s", macAddress)
	vendor, err = arp.GetMacVendor(macAddress)
	if err != nil {
		log.Warning(
			"Was not able to find vendor for addr '%s' with MAC '%s'. %v",
			clientIP, macAddress, err,
		)
		return ""
	}

	s.MACService.SaveMac(clientIP, macAddress, vendor)
	return vendor
}

func (s *DNSServer) resolveHostname(clientIP netip.Addr) string {
	if clientIP.IsLoopback() {
		hostname, err := os.Hostname()
		if err == nil {
			return hostname
		}
	}

	if hostname := s.reverseDNSLookup(clientIP); hostname != unknownHostname {
		return hostname
	}

	if hostname := s.avahiLookup(clientIP); hostname != unknownHostname {
		return hostname
	}

	if hostname := s.sshBannerLookup(clientIP); hostname != unknownHostname {
		return hostname
	}

	return unknownHostname
}

func (s *DNSServer) avahiLookup(clientIP netip.Addr) string {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "avahi-resolve-address", clientIP.String())
	output, err := cmd.Output()
	if err == nil {
		lines := strings.SplitSeq(string(output), "\n")
		for line := range lines {
			if strings.Contains(line, clientIP.String()) {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					hostname := strings.TrimSuffix(parts[1], ".local")
					if hostname != "" && hostname != clientIP.String() {
						log.Debug("Found hostname via avahi-resolve: %s -> %s", clientIP, hostname)
						return hostname
					}
				}
			}
		}
	}

	return unknownHostname
}

func (s *DNSServer) reverseDNSLookup(clientIP netip.Addr) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 2 * time.Second,
			}
			gateway := s.Config.DNS.Gateway
			if _, _, err := net.SplitHostPort(gateway); err != nil {
				gateway = net.JoinHostPort(gateway, "53")
			}
			return d.DialContext(ctx, "udp", gateway)
		},
	}

	if hostnames, err := resolver.LookupAddr(ctx, clientIP.String()); err == nil && len(hostnames) > 0 {
		hostname := strings.TrimSuffix(hostnames[0], ".")
		if hostname != clientIP.String() &&
			!strings.Contains(hostname, "in-addr.arpa") && !strings.HasPrefix(hostname, clientIP.String()) {
			log.Debug("Found hostname via reverse DNS: %s -> %s", clientIP, hostname)
			return hostname
		}
	}
	return unknownHostname
}

func (s *DNSServer) sshBannerLookup(clientIP netip.Addr) string {
	conn, err := net.DialTimeout("tcp", clientIP.String()+":22", 1*time.Second)
	if err != nil {
		return unknownHostname
	}
	defer func() {
		_ = conn.Close()
	}()

	err = conn.SetDeadline(time.Now().Add(2 * time.Second))
	if err != nil {
		log.Warning("Failed to set deadline for SSH banner lookup: %v", err)
		_ = conn.Close()
		return unknownHostname
	}

	reader := bufio.NewReader(conn)
	banner, err := reader.ReadString('\n')
	if err != nil {
		return unknownHostname
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
			if hostname != clientIP.String() && len(hostname) > 1 && hostname != "SSH" {
				log.Debug("Found hostname via SSH banner: %s -> %s", clientIP, hostname)
				return hostname
			}
		}
	}

	return unknownHostname
}

func getLocalIP() (netip.Addr, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return netip.Addr{}, err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipv4 := ipnet.IP.To4(); ipv4 != nil {
				ip, _ := netip.AddrFromSlice(ipv4)
				return ip, nil
			}
		}
	}

	return IPv4Loopback, fmt.Errorf("no non-loopback IPv4 address found")
}

func (s *DNSServer) handlePTRQuery(request *Request) model.RequestLogEntry {
	ipParts := strings.TrimSuffix(request.Question.Header().Name, ".in-addr.arpa.")
	parts := strings.Split(ipParts, ".")

	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	ipStr := strings.Join(parts, ".")

	if ipStr == IPv4Loopback.String() {
		return s.respondWithLocalhost(request)
	}

	if !isPrivateIP(ipStr) {
		return s.forwardPTRQueryUpstream(request)
	}

	hostname := s.RequestService.GetClientNameFromIP(ipStr)
	if hostname == unknownHostname {
		if ip, err := netip.ParseAddr(ipStr); err == nil {
			hostname = s.resolveHostname(ip)
		} else {
			log.Warning("Failed to parse IP for hostname lookup: %v", err)
			hostname = unknownHostname
		}
	}

	if hostname != unknownHostname {
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
		Hdr: dns.Header{
			Name:  request.Question.Header().Name,
			TTL:   3600,
			Class: dns.ClassINET,
		},
		PTR: rdata.PTR{
			Ptr: "localhost.lan.",
		},
	}

	request.Msg.Answer = []dns.RR{ptr}

	request.Respond(s.NotificationService)
	return model.RequestLogEntry{
		Timestamp: request.Sent,
		Domain:    request.Question.Header().Name,
		Status:    dnsutil.CodeToString(dns.RcodeSuccess),
		IP: []model.ResolvedIP{
			{
				IP:    IPv4Loopback,
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

func (s *DNSServer) respondWithHostnameA(request *Request, hostIP netip.Addr) model.RequestLogEntry {
	request.Msg.Response = true
	request.Msg.Authoritative = false
	request.Msg.RecursionAvailable = true
	request.Msg.Rcode = dns.RcodeSuccess

	response := &dns.A{
		Hdr: dns.Header{
			Name:  request.Question.Header().Name,
			TTL:   60,
			Class: dns.ClassINET,
		},
		A: rdata.A{
			Addr: hostIP,
		},
	}

	request.Msg.Answer = []dns.RR{response}
	request.Respond(s.NotificationService)
	return s.respondWithType(request, dns.TypeA, hostIP)
}

func (s *DNSServer) respondWithHostnamePTR(request *Request, hostname string) model.RequestLogEntry {
	request.Msg.Response = true
	request.Msg.Authoritative = false
	request.Msg.RecursionAvailable = true
	request.Msg.Rcode = dns.RcodeSuccess

	ptr := &dns.PTR{
		Hdr: dns.Header{
			Name:  request.Question.Header().Name,
			TTL:   3600,
			Class: dns.ClassINET,
		},
		PTR: rdata.PTR{
			Ptr: hostname + ".",
		},
	}

	request.Msg.Answer = []dns.RR{ptr}
	request.Respond(s.NotificationService)
	ip, err := netip.ParseAddr(hostname)
	if err != nil {
		log.Warning("Not able to parse ip for hostname %s", hostname)
		return model.RequestLogEntry{
			Timestamp:         request.Sent,
			Domain:            request.Question.Header().Name,
			Status:            dnsutil.CodeToString(dns.RcodeSuccess),
			IP:                []model.ResolvedIP{},
			Blocked:           false,
			Cached:            false,
			ResponseTime:      time.Since(request.Sent),
			ClientInfo:        request.Client,
			QueryType:         "PTR",
			ResponseSizeBytes: request.Msg.Len(),
			Protocol:          request.Protocol,
		}
	}
	return s.respondWithType(request, dns.TypePTR, ip)
}

func (s *DNSServer) respondWithType(request *Request, rType uint16, ip netip.Addr) model.RequestLogEntry {
	return model.RequestLogEntry{
		Domain:    request.Question.Header().Name,
		Status:    dnsutil.CodeToString(dns.RcodeSuccess),
		QueryType: dnsutil.TypeToString(request.Question.Header().Class),
		IP: []model.ResolvedIP{
			{
				IP:    ip,
				RType: dnsutil.TypeToString(rType),
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
			if ip, err := netip.ParseAddr(ptr.Ptr); err == nil {
				resolvedHostnames = append(resolvedHostnames, model.ResolvedIP{
					IP:    ip,
					RType: "PTR",
				})
			} else {
				log.Warning("Failed to parse PTR data: %v", err)
			}
		}
	}

	request.Respond(s.NotificationService)
	return model.RequestLogEntry{
		Domain:            request.Question.Header().Name,
		Status:            status,
		QueryType:         dnsutil.TypeToString(request.Question.Header().Class),
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
	if rcode, ok := dns.StringToRcode[status]; ok {
		request.Msg.Rcode = rcode
	} else {
		request.Msg.Rcode = dns.RcodeServerFailure
	}

	for _, a := range answers {
		switch rr := a.(type) {
		case *dns.A:
			if ip, err := netip.ParseAddr(rr.A.String()); err == nil {
				resolved = append(resolved, model.ResolvedIP{
					IP:    ip,
					RType: "A",
				})
			} else {
				log.Warning("Failed to parse A record: %v", err)
			}
		case *dns.AAAA:
			if ip, err := netip.ParseAddr(rr.AAAA.String()); err == nil {
				resolved = append(resolved, model.ResolvedIP{
					IP:    ip,
					RType: "AAAA",
				})
			} else {
				log.Warning("Failed to parse AAAA record: %v", err)
			}
		case *dns.PTR:
			if ip, err := netip.ParseAddr(rr.Ptr); err == nil {
				resolved = append(resolved, model.ResolvedIP{
					IP:    ip,
					RType: "PTR",
				})
			} else {
				log.Warning("Failed to parse PTR record: %v", err)
			}
		case *dns.CNAME:
			if ip, err := netip.ParseAddr(rr.Target); err == nil {
				resolved = append(resolved, model.ResolvedIP{
					IP:    ip,
					RType: "CNAME",
				})
			} else {
				log.Warning("Failed to parse CNAME record: %v", err)
			}
		case *dns.SVCB:
			if ip, err := netip.ParseAddr(rr.Target); err == nil {
				resolved = append(resolved, model.ResolvedIP{
					IP:    ip,
					RType: "SVCB",
				})
			} else {
				log.Warning("Failed to parse SVCB record: %v", err)
			}
		case *dns.MX:
			if ip, err := netip.ParseAddr(rr.Mx); err == nil {
				resolved = append(resolved, model.ResolvedIP{
					IP:    ip,
					RType: "MX",
				})
			} else {
				log.Warning("Failed to parse MX record: %v", err)
			}
		case *dns.TXT:
			if ip, err := netip.ParseAddr(rr.Txt[0]); err == nil {
				resolved = append(resolved, model.ResolvedIP{
					IP:    ip,
					RType: "TXT",
				})
			} else {
				log.Warning("Failed to parse TXT record: %v", err)
			}
		case *dns.NS:
			if ip, err := netip.ParseAddr(rr.Ns); err == nil {
				resolved = append(resolved, model.ResolvedIP{
					IP:    ip,
					RType: "NS",
				})
			} else {
				log.Warning("Failed to parse NS record: %v", err)
			}
		case *dns.SOA:
			if ip, err := netip.ParseAddr(rr.Ns); err == nil {
				resolved = append(resolved, model.ResolvedIP{
					IP:    ip,
					RType: "SOA",
				})
			} else {
				log.Warning("Failed to parse SOA record: %v", err)
			}
		case *dns.SRV:
			if ip, err := netip.ParseAddr(fmt.Sprintf("%s:%d", rr.Target, rr.Port)); err == nil {
				resolved = append(resolved, model.ResolvedIP{
					IP:    ip,
					RType: "SRV",
				})
			} else {
				log.Warning("Failed to parse SRV record: %v", err)
			}
		case *dns.HTTPS:
			if ip, err := netip.ParseAddr(rr.Target); err == nil {
				resolved = append(resolved, model.ResolvedIP{
					IP:    ip,
					RType: "HTTPS",
				})
			} else {
				log.Warning("Failed to parse HTTPS record: %v", err)
			}
		case *dns.CAA:
			if ip, err := netip.ParseAddr(fmt.Sprintf("%s %d %s", rr.Tag, rr.Flag, rr.Value)); err == nil {
				resolved = append(resolved, model.ResolvedIP{
					IP:    ip,
					RType: "CAA",
				})
			} else {
				log.Warning("Failed to parse CAA record: %v", err)
			}
		case *dns.DNSKEY:
			if ip, err := netip.ParseAddr(fmt.Sprintf("flags:%d protocol:%d algorithm:%d", rr.Flags, rr.Protocol, rr.Algorithm)); err == nil {
				resolved = append(resolved, model.ResolvedIP{
					IP:    ip,
					RType: "DNSKEY",
				})
			} else {
				log.Warning("Failed to parse DNSKEY record: %v", err)
			}
		default:
			log.Warning("Unhandled record type '%s' while requesting '%s'", dnsutil.TypeToString(rr.Header().Class), request.Question.Header().Name)
		}
	}

	request.Respond(s.NotificationService)
	return model.RequestLogEntry{
		Domain:            request.Question.Header().Name,
		Status:            status,
		QueryType:         dnsutil.TypeToString(request.Question.Header().Class),
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
	cacheKey := req.Question.Header().Name + ":" + strconv.Itoa(int(req.Question.Header().Class))
	if cached, found := s.DomainCache.Load(cacheKey); found {
		if ipAddresses, valid := s.getCachedRecord(cached); valid {
			return ipAddresses, true, dnsutil.CodeToString(dns.RcodeSuccess)
		}
	}

	if answers, ttl, status := s.resolveResolution(req.Question.Header().Name); len(answers) > 0 {
		s.CacheRecord(cacheKey, req.Question.Header().Name, answers, ttl)
		return answers, false, status
	}

	answers, ttl, status := s.resolveCNAMEChain(req, make(map[string]bool))
	if len(answers) > 0 {
		s.CacheRecord(cacheKey, req.Question.Header().Name, answers, ttl)
	}
	return answers, false, status
}

func (s *DNSServer) resolveResolution(domain string) ([]dns.RR, uint32, string) {
	var (
		records []dns.RR
		rr      dns.RR
		ttl     = uint32(s.Config.DNS.CacheTTL)
		status  = dnsutil.CodeToString(dns.RcodeSuccess)
	)

	ipFound, err := s.ResolutionService.GetResolution(domain)
	if err != nil {
		log.Error("Database lookup error for domain (%s): %v", domain, err)
		return nil, 0, dnsutil.CodeToString(dns.RcodeServerFailure)
	}

	if ipFound == "" {
		status = dnsutil.CodeToString(dns.RcodeNameError)
	} else {
		ip, err := netip.ParseAddr(ipFound)
		if err != nil {
			log.Warning("Failed to parse IP from database for domain '%s': %v", domain, err)
			return nil, 0, dnsutil.CodeToString(dns.RcodeServerFailure)
		}
		if ip.Is4() {
			rr = &dns.A{
				Hdr: dns.Header{
					Name:  dnsutil.Fqdn(domain),
					TTL:   ttl,
					Class: dns.ClassINET,
				},
				A: rdata.A{Addr: ip}}
		} else {
			rr = &dns.AAAA{
				Hdr: dns.Header{
					Name:  dnsutil.Fqdn(domain),
					TTL:   ttl,
					Class: dns.ClassINET,
				},
				AAAA: rdata.AAAA{Addr: ip},
			}
		}
		records = append(records, rr)
	}
	return records, ttl, status
}

func (s *DNSServer) resolveCNAMEChain(req *Request, visited map[string]bool) ([]dns.RR, uint32, string) {
	if visited[req.Question.Header().Name] {
		return nil, 0, dnsutil.CodeToString(dns.RcodeServerFailure)
	}
	visited[req.Question.Header().Name] = true

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
		go s.WSCom(communicationMessage{IP: "", Client: false, Upstream: true, DNS: false})

		q := req.Question.Header()
		upstreamMsg := dns.NewMsg(q.Name, q.Class)
		upstreamMsg.RecursionDesired = true
		upstreamMsg.ID = dns.ID()

		upstream := s.Config.DNS.Upstream.Preferred
		proto := "udp"

		// Use TCP for DoT and TCP clients, UDP for others
		if req.Protocol == model.DoT || req.Protocol == model.TCP {
			proto = "tcp"
		}

		log.Debug("Sending query using '%s' as upstream", upstream)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		in, _, err := s.dnsClient.Exchange(ctx, upstreamMsg, proto, upstream)
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
		go s.WSCom(communicationMessage{IP: "", Client: false, Upstream: false, DNS: true})

		status := dnsutil.CodeToString(dns.RcodeServerFailure)
		if statusStr, ok := dns.RcodeToString[in.Rcode]; ok {
			status = statusStr
		}

		var ttl uint32 = 3600
		if len(in.Answer) > 0 {
			ttl = in.Answer[0].Header().TTL
			for _, a := range in.Answer {
				if a.Header().TTL < ttl {
					ttl = a.Header().TTL
				}
			}
		} else if len(in.Ns) > 0 {
			ttl = in.Ns[0].Header().TTL
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
		log.Warning("Upstream resolution error for domain (%s): %v", req.Question.Header().Name, err)
		s.NotificationService.SendNotification(
			notification.SeverityWarning,
			notification.CategoryDNS,
			fmt.Sprintf("Upstream resolution error for domain (%s)", req.Question.Header().Name),
		)
		return nil, 0, dnsutil.CodeToString(dns.RcodeServerFailure)

	case <-time.After(5 * time.Second):
		log.Warning("Upstream lookup for %s timed out", req.Question.Header().Name)
		return nil, 0, dnsutil.CodeToString(dns.RcodeServerFailure)
	}
}

func (s *DNSServer) LocalForwardLookup(req *Request) (model.RequestLogEntry, error) {
	hostname := strings.ReplaceAll(req.Question.Header().Name, ".in-addr.arpa.", "")
	hostname = strings.ReplaceAll(hostname, ".ip6.arpa.", "")
	if !strings.HasSuffix(hostname, ".") {
		hostname += "."
	}

	queryType := req.Question.Header().Class
	if queryType == 0 {
		queryType = dns.TypeA
	}

	dnsMsg := dns.NewMsg(hostname, queryType)
	client := &dns.Client{}
	start := time.Now()
	log.Debug("Performing local forward lookup for %s", hostname)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	in, _, err := client.Exchange(ctx, dnsMsg, "udp", s.Config.DNS.Gateway)
	responseTime := time.Since(start)

	if err != nil {
		log.Error("DNS exchange error for %s: %v", hostname, err)
		return model.RequestLogEntry{}, fmt.Errorf("forward DNS query failed: %w", err)
	}

	if in.Rcode != dns.RcodeSuccess {
		status := dnsutil.CodeToString(in.Rcode)
		log.Info("DNS query for %s returned status %s", hostname, status)
		return model.RequestLogEntry{}, fmt.Errorf("forward lookup failed with status: %s", status)
	}

	var ips []model.ResolvedIP
	for _, answer := range in.Answer {
		if a, ok := answer.(*dns.A); ok {
			ips = append(ips, model.ResolvedIP{IP: a.Addr})
		}
	}

	if len(ips) == 0 && queryType == dns.TypeA {
		return model.RequestLogEntry{}, fmt.Errorf("no A records found for hostname: %s", hostname)
	}

	req.Msg.Rcode = in.Rcode
	req.Msg.Answer = in.Answer

	req.Respond(s.NotificationService)
	entry := model.RequestLogEntry{
		Domain:            req.Question.Header().Name,
		Status:            dnsutil.CodeToString(in.Rcode),
		QueryType:         dnsutil.TypeToString(queryType),
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

	switch request.Question.Header().Class {
	case dns.TypeA:
		request.Msg.Answer = []dns.RR{&dns.A{
			Hdr: dns.Header{
				Name: request.Question.Header().Name,

				Class: dns.ClassINET,
				TTL:   cacheTTL,
			},
			A: rdata.A{Addr: blackholeIPv4},
		}}
		resolved = []model.ResolvedIP{{IP: blackholeIPv4, RType: "A"}}
	case dns.TypeAAAA:
		request.Msg.Answer = []dns.RR{&dns.AAAA{
			Hdr: dns.Header{
				Name:  request.Question.Header().Name,
				TTL:   cacheTTL,
				Class: dns.ClassINET,
			},
			AAAA: rdata.AAAA{Addr: blackholeIPv6},
		}}
		resolved = []model.ResolvedIP{{IP: blackholeIPv6, RType: "AAAA"}}
	default:
		request.Msg.Rcode = dns.RcodeNameError
		request.Msg.Answer = nil
		resolved = nil
	}

	if len(request.Msg.Question) == 0 {
		return model.RequestLogEntry{
			Domain: "unknown",
		}
	}

	request.Respond(s.NotificationService)
	return model.RequestLogEntry{
		Domain:            request.Question.Header().Name,
		Status:            dnsutil.CodeToString(request.Msg.Rcode),
		QueryType:         dnsutil.TypeToString(request.Question.Header().Class),
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
