package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"goaway/backend/audit"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
	probing "github.com/prometheus-community/pro-bing"
)

type pingResult struct {
	Error      error
	Method     string
	Duration   time.Duration
	Successful bool
}

func (pr pingResult) String() string {
	if !pr.Successful {
		return fmt.Sprintf("Failed (%s)", pr.Method)
	}
	return pr.Duration.String()
}

func (api *API) registerUpstreamRoutes() {
	api.routes.POST("/upstream", api.createUpstream)
	api.routes.GET("/upstreams", api.getUpstreams)
	api.routes.PUT("/preferredUpstream", api.updatePreferredUpstream)
	api.routes.DELETE("/upstream", api.deleteUpstream)
}

func (api *API) createUpstream(c *gin.Context) {
	type UpstreamRequest struct {
		Upstream string `json:"upstream"`
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var request UpstreamRequest
	if err := json.Unmarshal(body, &request); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	upstream := request.Upstream
	if upstream == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Upstream is required"})
		return
	}

	lower := strings.ToLower(upstream)
	isDoH := strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
	isProvider := lower == "cloudflare" || lower == "google" || lower == "quad9" || lower == "dnspod" || lower == "auto" || lower == "fastest"
	if !isDoH && !isProvider && !strings.Contains(upstream, ":") {
		upstream += ":53"
	}

	if slices.Contains(api.Config.DNS.Upstream.Fallback, upstream) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Upstream already exists"})
		return
	}

	api.Config.DNS.Upstream.Fallback = append(api.Config.DNS.Upstream.Fallback, upstream)
	api.Config.Save()

	log.Info("Added %s as a new upstream", upstream)
	api.DNSServer.AuditService.CreateAudit(&audit.Entry{
		Topic:   audit.TopicUpstream,
		Message: fmt.Sprintf("Added a new upstream '%s'", request.Upstream),
	})

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Added %s as a new upstream", upstream)})
}

func (api *API) getUpstreams(c *gin.Context) {
	var (
		fallbackUpstreams = api.Config.DNS.Upstream.Fallback
		preferredUpstream = api.Config.DNS.Upstream.Preferred
		upstreamsToCheck  = make([]string, 0, len(fallbackUpstreams)+1)
		wg                sync.WaitGroup
	)

	upstreamsToCheck = append(upstreamsToCheck, fallbackUpstreams...)
	if preferredUpstream != "" && !slices.Contains(fallbackUpstreams, preferredUpstream) {
		upstreamsToCheck = append(upstreamsToCheck, preferredUpstream)
	}

	results := make([]map[string]any, len(upstreamsToCheck))
	wg.Add(len(upstreamsToCheck))

	for i, upstream := range upstreamsToCheck {
		go func(i int, upstream string) {
			defer wg.Done()
			results[i] = getUpstreamDetails(upstream, preferredUpstream)
		}(i, upstream)
	}

	wg.Wait()

	c.JSON(http.StatusOK, gin.H{
		"upstreams":         results,
		"preferredUpstream": preferredUpstream,
	})
}

func getUpstreamDetails(upstream, preferredUpstream string) map[string]any {
	var host, port string
	lower := strings.ToLower(upstream)
	isProvider := lower == "cloudflare" || lower == "google" || lower == "quad9" || lower == "dnspod" || lower == "auto" || lower == "fastest"
	isDoH := strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || isProvider
	if isDoH {
		var dohURL string
		if isProvider {
			dohURL = providerKeywordToURL(lower)
		} else {
			dohURL = upstream
		}
		if u, err := url.Parse(dohURL); err == nil {
			h, p, _ := net.SplitHostPort(u.Host)
			if h == "" {
				host = u.Host
			} else {
				host = h
				port = p
			}
			if port == "" {
				port = "443"
			}
			// keep original 'upstream' value for UI actions like setPreferred
		} else {
			host = upstream
			port = "443"
		}
	} else {
		h, p, err := net.SplitHostPort(upstream)
		if err != nil {
			host = upstream
			port = "53"
			upstream = net.JoinHostPort(host, port)
		} else {
			host = h
			port = p
		}
	}

	entry := map[string]any{
		"upstream":  upstream,
		"preferred": upstream == preferredUpstream,
		"host":      host,
		"port":      port,
	}

	entry["resolvedIP"] = resolveHostname(host)
	entry["upstreamName"] = getUpstreamName(host)

	if isDoH {
		dohURL := upstream
		if isProvider {
			dohURL = providerKeywordToURL(lower)
		}
		dohPing := measureDoHPing(dohURL)
		entry["dnsPing"] = dohPing.String()
		entry["dnsPingSuccess"] = dohPing.Successful
		entry["protocol"] = "doh"
	} else {
		dnsPingResult := measureDNSPing(upstream)
		entry["dnsPing"] = dnsPingResult.String()
		entry["dnsPingSuccess"] = dnsPingResult.Successful
		entry["protocol"] = "udp/dot"
	}

	icmpPingResult := measureICMPPing(host)
	entry["icmpPing"] = icmpPingResult.String()
	entry["icmpPingSuccess"] = icmpPingResult.Successful

	return entry
}

func providerKeywordToURL(keyword string) string {
	switch keyword {
	case "cloudflare":
		return "https://cloudflare-dns.com/dns-query"
	case "google":
		return "https://dns.google/dns-query"
	case "quad9":
		return "https://dns.quad9.net/dns-query"
	case "dnspod":
		return "https://doh.dnspod.cn/dns-query"
	case "auto", "fastest":
		// default for details view; actual fastest chosen at query time
		return "https://cloudflare-dns.com/dns-query"
	default:
		return keyword
	}
}

func measureDoHPing(upstreamURL string) pingResult {
	testDomains := []string{"google.com", "cloudflare.com", "quad9.net"}
	client := &http.Client{Timeout: 5 * time.Second}
	var totalDuration time.Duration
	successCount := 0
	var lastError error

	for _, domain := range testDomains {
		msg := &dns.Msg{}
		msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)
		msg.RecursionDesired = true
		packed, err := msg.Pack()
		if err != nil {
			lastError = err
			continue
		}

		// Try POST first
		req, err := http.NewRequest(http.MethodPost, upstreamURL, io.NopCloser(bytes.NewReader(packed)))
		if err != nil {
			lastError = err
			continue
		}
		req.Header.Set("Content-Type", "application/dns-message")
		req.Header.Set("Accept", "application/dns-message")

		start := time.Now()
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			body, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if readErr != nil {
				lastError = readErr
				continue
			}
			parsed := new(dns.Msg)
			if unpackErr := parsed.Unpack(body); unpackErr != nil {
				lastError = unpackErr
				continue
			}
			duration := time.Since(start)
			totalDuration += duration
			successCount++
			continue
		}
		if resp != nil {
			_ = resp.Body.Close()
		}

		// Fallback to GET (Base64URL encoded)
		b64 := base64.RawURLEncoding.EncodeToString(packed)
		getURL := upstreamURL
		if strings.Contains(upstreamURL, "?") {
			getURL += "&dns=" + b64
		} else {
			getURL += "?dns=" + b64
		}
		reqGet, err := http.NewRequest(http.MethodGet, getURL, nil)
		if err != nil {
			lastError = err
			continue
		}
		reqGet.Header.Set("Accept", "application/dns-message")
		start = time.Now()
		resp, err = client.Do(reqGet)
		if err != nil {
			lastError = err
			continue
		}
		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil || resp.StatusCode != http.StatusOK {
			if readErr != nil {
				lastError = readErr
			} else {
				lastError = fmt.Errorf("status %s", resp.Status)
			}
			continue
		}
		parsed := new(dns.Msg)
		if unpackErr := parsed.Unpack(body); unpackErr != nil {
			lastError = unpackErr
			continue
		}
		duration := time.Since(start)
		totalDuration += duration
		successCount++
	}

	if successCount == 0 {
		return pingResult{Duration: 0, Error: lastError, Method: "doh", Successful: false}
	}

	avg := totalDuration / time.Duration(successCount)
	return pingResult{Duration: avg, Error: nil, Method: "doh", Successful: true}
}

func resolveHostname(host string) string {
	if ip := net.ParseIP(host); ip != nil {
		return host
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Sprintf("Resolution failed: %v", err)
	}

	if len(ips) > 0 {
		return ips[0].IP.String()
	}

	return "No IP found"
}

func measureDNSPing(upstream string) pingResult {
	var (
		testDomains = []string{"google.com", "cloudflare.com", "quad9.net"}
		client      = &dns.Client{
			Timeout: 2 * time.Second,
		}
		totalDuration time.Duration
		successCount  int
		lastError     error
	)

	for _, domain := range testDomains {
		msg := &dns.Msg{}
		msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)
		msg.RecursionDesired = true

		start := time.Now()
		response, _, err := client.Exchange(msg, upstream)
		duration := time.Since(start)

		if err != nil {
			lastError = err
			continue
		}

		if response.Rcode != dns.RcodeSuccess {
			lastError = fmt.Errorf("DNS query failed with rcode: %d", response.Rcode)
			continue
		}

		totalDuration += duration
		successCount++
	}

	if successCount == 0 {
		return pingResult{
			Duration:   0,
			Error:      lastError,
			Method:     "dns",
			Successful: false,
		}
	}

	avgDuration := totalDuration / time.Duration(successCount)
	return pingResult{
		Duration:   avgDuration,
		Error:      nil,
		Method:     "dns",
		Successful: true,
	}
}

func measureICMPPing(host string) pingResult {
	icmpResult := tryICMPPing(host)
	if icmpResult.Successful {
		return icmpResult
	}

	tcpResult := tryTCPPing(host)
	return tcpResult
}

func tryICMPPing(host string) pingResult {
	pinger, err := probing.NewPinger(host)
	if err != nil {
		return pingResult{
			Duration:   0,
			Error:      err,
			Method:     "icmp",
			Successful: false,
		}
	}

	pinger.Count = 3
	pinger.Timeout = 3 * time.Second
	pinger.SetPrivileged(false)

	err = pinger.Run()
	if err != nil {
		return pingResult{
			Duration:   0,
			Error:      err,
			Method:     "icmp",
			Successful: false,
		}
	}

	stats := pinger.Statistics()
	if stats.PacketsRecv == 0 {
		return pingResult{
			Duration:   0,
			Error:      fmt.Errorf("no packets received"),
			Method:     "icmp",
			Successful: false,
		}
	}

	return pingResult{
		Duration:   stats.AvgRtt,
		Error:      nil,
		Method:     "icmp",
		Successful: true,
	}
}

func tryTCPPing(host string) pingResult {
	start := time.Now()

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, "53"), 2*time.Second)
	if err != nil {
		return pingResult{
			Duration:   0,
			Error:      err,
			Method:     "tcp",
			Successful: false,
		}
	}

	duration := time.Since(start)
	defer func() {
		_ = conn.Close()
	}()

	return pingResult{
		Duration:   duration,
		Error:      nil,
		Method:     "tcp",
		Successful: true,
	}
}

func getUpstreamName(host string) string {
	if ip := net.ParseIP(host); ip == nil {
		return host
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	names, err := net.DefaultResolver.LookupAddr(ctx, host)
	if err != nil {
		log.Info("Reverse lookup failed for %s: %v", host, err)
		return "unknown"
	}

	if len(names) > 0 {
		return strings.TrimSuffix(names[0], ".")
	}

	return "unknown"
}

func (api *API) updatePreferredUpstream(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var request struct {
		Upstream string `json:"upstream"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if request.Upstream == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Upstream is required"})
		return
	}

	if !slices.Contains(api.Config.DNS.Upstream.Fallback, request.Upstream) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Upstream not found"})
		return
	}

	if api.Config.DNS.Upstream.Preferred == request.Upstream {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Preferred upstream already set to %s", request.Upstream)})
		return
	}

	api.Config.DNS.Upstream.Preferred = request.Upstream
	message := fmt.Sprintf("Preferred upstream set to %s", request.Upstream)
	log.Info("%s", message)

	api.Config.Save()
	api.DNSServer.AuditService.CreateAudit(&audit.Entry{
		Topic:   audit.TopicUpstream,
		Message: fmt.Sprintf("New preferred upstream '%s'", request.Upstream),
	})

	c.JSON(http.StatusOK, gin.H{"message": message})
}

func (api *API) deleteUpstream(c *gin.Context) {
	upstreamToDelete := c.Query("upstream")

	if upstreamToDelete == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'upstream' query parameter"})
		return
	}

	if !slices.Contains(api.Config.DNS.Upstream.Fallback, upstreamToDelete) {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Upstream %s not found", upstreamToDelete)})
		return
	}

	var updatedUpstreams []string
	for _, upstream := range api.Config.DNS.Upstream.Fallback {
		if upstream != upstreamToDelete {
			updatedUpstreams = append(updatedUpstreams, upstream)
		}
	}

	api.Config.DNS.Upstream.Fallback = updatedUpstreams
	api.Config.Save()
	log.Info("Removed upstream: %s", upstreamToDelete)

	api.DNSServer.AuditService.CreateAudit(&audit.Entry{
		Topic:   audit.TopicUpstream,
		Message: fmt.Sprintf("Removed upstream '%s'", upstreamToDelete),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Upstream removed successfully",
	})
}
