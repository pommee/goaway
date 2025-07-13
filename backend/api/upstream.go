package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
	probing "github.com/prometheus-community/pro-bing"
)

type PingResult struct {
	Duration   time.Duration
	Error      error
	Method     string
	Successful bool
}

func (pr PingResult) String() string {
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

	if !strings.Contains(upstream, ":") {
		upstream += ":53"
	}

	if slices.Contains(api.Config.DNS.UpstreamDNS, upstream) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Upstream already exists"})
		return
	}

	api.Config.DNS.UpstreamDNS = append(api.Config.DNS.UpstreamDNS, upstream)
	api.Config.Save()

	log.Info("Added %s as a new upstream", upstream)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Added %s as a new upstream", upstream)})
}

func (api *API) getUpstreams(c *gin.Context) {
	var (
		upstreams         = api.Config.DNS.UpstreamDNS
		results           = make([]map[string]any, len(upstreams))
		preferredUpstream = api.Config.DNS.PreferredUpstream
		wg                sync.WaitGroup
	)
	wg.Add(len(upstreams))

	for i, upstream := range upstreams {
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
	host, port, err := net.SplitHostPort(upstream)
	if err != nil {
		host = upstream
		port = "53"
		upstream = net.JoinHostPort(host, port)
	}

	entry := map[string]any{
		"upstream":  upstream,
		"preferred": upstream == preferredUpstream,
		"host":      host,
		"port":      port,
	}

	entry["resolvedIP"] = resolveHostname(host)
	entry["upstreamName"] = getUpstreamName(host)

	dnsPingResult := measureDNSPing(upstream)
	entry["dnsPing"] = dnsPingResult.String()
	entry["dnsPingSuccess"] = dnsPingResult.Successful

	icmpPingResult := measureICMPPing(host)
	entry["icmpPing"] = icmpPingResult.String()
	entry["icmpPingSuccess"] = icmpPingResult.Successful

	return entry
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

func measureDNSPing(upstream string) PingResult {
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
		return PingResult{
			Duration:   0,
			Error:      lastError,
			Method:     "dns",
			Successful: false,
		}
	}

	avgDuration := totalDuration / time.Duration(successCount)
	return PingResult{
		Duration:   avgDuration,
		Error:      nil,
		Method:     "dns",
		Successful: true,
	}
}

func measureICMPPing(host string) PingResult {
	icmpResult := tryICMPPing(host)
	if icmpResult.Successful {
		return icmpResult
	}

	tcpResult := tryTCPPing(host)
	return tcpResult
}

func tryICMPPing(host string) PingResult {
	pinger, err := probing.NewPinger(host)
	if err != nil {
		return PingResult{
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
		return PingResult{
			Duration:   0,
			Error:      err,
			Method:     "icmp",
			Successful: false,
		}
	}

	stats := pinger.Statistics()
	if stats.PacketsRecv == 0 {
		return PingResult{
			Duration:   0,
			Error:      fmt.Errorf("no packets received"),
			Method:     "icmp",
			Successful: false,
		}
	}

	return PingResult{
		Duration:   stats.AvgRtt,
		Error:      nil,
		Method:     "icmp",
		Successful: true,
	}
}

func tryTCPPing(host string) PingResult {
	start := time.Now()

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, "53"), 2*time.Second)
	if err != nil {
		return PingResult{
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

	return PingResult{
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

	if !slices.Contains(api.Config.DNS.UpstreamDNS, request.Upstream) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Upstream not found"})
		return
	}

	if api.Config.DNS.PreferredUpstream == request.Upstream {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Preferred upstream already set to %s", request.Upstream)})
		return
	}

	api.Config.DNS.PreferredUpstream = request.Upstream
	message := fmt.Sprintf("Preferred upstream set to %s", request.Upstream)
	log.Info("%s", message)

	api.Config.Save()
	c.JSON(http.StatusOK, gin.H{"message": message})
}

func (api *API) deleteUpstream(c *gin.Context) {
	upstreamToDelete := c.Query("upstream")

	if upstreamToDelete == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'upstream' query parameter"})
		return
	}

	if !slices.Contains(api.Config.DNS.UpstreamDNS, upstreamToDelete) {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Upstream %s not found", upstreamToDelete)})
		return
	}

	var updatedUpstreams []string
	for _, upstream := range api.Config.DNS.UpstreamDNS {
		if upstream != upstreamToDelete {
			updatedUpstreams = append(updatedUpstreams, upstream)
		}
	}

	api.Config.DNS.UpstreamDNS = updatedUpstreams
	api.Config.Save()
	log.Info("Removed upstream: %s", upstreamToDelete)

	c.JSON(http.StatusOK, gin.H{
		"message": "Upstream removed successfully",
	})
}
