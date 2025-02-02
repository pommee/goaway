package api

import (
	"embed"
	"encoding/json"
	"fmt"
	"goaway/internal/server"
	"goaway/internal/settings"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ping/ping"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func (api *API) ServeEmbeddedContent(content embed.FS) {
	mimeTypes := map[string]string{
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
	}

	ipAddress, err := getServerIP()
	if err != nil {
		log.Error("Error getting IP address: %v", err)
		return
	}

	err = fs.WalkDir(content, "website", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking through path %s: %w", path, err)
		}
		if d.IsDir() {
			return nil
		}

		fileContent, err := content.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", path, err)
		}

		ext := strings.ToLower(filepath.Ext(path))
		mimeType := mimeTypes[ext]
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		route := strings.TrimPrefix(path, "website/")
		if route == "index.html" {
			api.router.GET("/", func(c *gin.Context) {
				c.Header("X-Server-IP", ipAddress)
				c.Data(http.StatusOK, mimeType, fileContent)
			})
		}

		api.router.GET("/"+route, func(c *gin.Context) {
			c.Header("X-Server-IP", ipAddress)
			c.Data(http.StatusOK, mimeType, fileContent)
		})

		return nil
	})

	if err != nil {
		log.Error("Error serving embedded content: %v", err)
	}
}

func (api *API) handleLogin(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if !api.validateCredentials(creds.Username, creds.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := generateToken(creds.Username)
	if err != nil {
		log.Error("Failed to create token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	setAuthCookie(c.Writer, token)
	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}

func (apiServer *API) handleServer(c *gin.Context) {
	cpuUsage, err := cpu.Percent(0, false)
	if err != nil {
		log.Error("%s", err)
	}

	temp, err := getCPUTemperature()
	if err != nil {
		log.Error("%s", err)
	}

	vMem, err := mem.VirtualMemory()
	if err != nil {
		log.Error("%s", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"portDNS":           apiServer.Config.Port,
		"portWebsite":       apiServer.DnsServer.Config.Port,
		"totalMem":          float64(vMem.Total) / 1024 / 1024 / 1024,
		"usedMem":           float64(vMem.Used) / 1024 / 1024 / 1024,
		"usedMemPercentage": float64(vMem.Free) / 1024 / 1024 / 1024,
		"cpuUsage":          cpuUsage[0],
		"cpuTemp":           temp,
	})
}

func (apiServer *API) getAuthentication(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"enabled": apiServer.Config.Authentication})
}

func (apiServer *API) handleMetrics(c *gin.Context) {
	allowedQueries := apiServer.DnsServer.Counters.AllowedRequests
	blockedQueries := apiServer.DnsServer.Counters.BlockedRequests
	totalQueries := allowedQueries + blockedQueries

	var percentageBlocked float64
	if totalQueries > 0 {
		percentageBlocked = (float64(blockedQueries) / float64(totalQueries)) * 100
	}

	domainsLength, _ := apiServer.DnsServer.Blacklist.CountDomains()
	c.JSON(http.StatusOK, gin.H{
		"allowed":           allowedQueries,
		"blocked":           blockedQueries,
		"total":             totalQueries,
		"percentageBlocked": percentageBlocked,
		"domainBlockLen":    domainsLength,
	})
}

func (apiServer *API) getQueryTimestamps(c *gin.Context) {
	type QueryEntry struct {
		Timestamp time.Time `json:"timestamp"`
		Blocked   bool      `json:"blocked"`
	}

	rows, err := apiServer.DnsServer.DB.Query("SELECT timestamp, blocked FROM request_log")
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	queries := []QueryEntry{}
	for rows.Next() {
		var query QueryEntry
		if err := rows.Scan(&query.Timestamp, &query.Blocked); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		queries = append(queries, query)
	}

	c.JSON(http.StatusOK, gin.H{
		"queries": queries,
	})
}

func (apiServer *API) handleQueriesData(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.DefaultQuery("search", "")
	sortColumn := c.DefaultQuery("sortColumn", "timestamp")
	sortDirection := c.DefaultQuery("sortDirection", "desc")

	offset := (page - 1) * pageSize

	sortColumn = map[string]string{
		"timestamp": "timestamp",
		"domain":    "domain",
		"client":    "client_ip",
	}[sortColumn]

	query := `
		SELECT timestamp, domain, blocked, cached, response_time_ns, client_ip, client_name
		FROM request_log
		WHERE domain LIKE ?
		ORDER BY ` + sortColumn + ` ` + sortDirection + `
		LIMIT ? OFFSET ?`

	rows, err := apiServer.DnsServer.DB.Query(query, "%"+search+"%", pageSize, offset)
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	queries := []server.RequestLogEntry{}
	for rows.Next() {
		var query server.RequestLogEntry
		query.ClientInfo = &server.Client{}
		if err := rows.Scan(&query.Timestamp, &query.Domain, &query.Blocked, &query.Cached, &query.ResponseTimeNS, &query.ClientInfo.IP, &query.ClientInfo.Name); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		queries = append(queries, query)
	}

	var totalRecords int
	err = apiServer.DnsServer.DB.QueryRow(`SELECT COUNT(*) FROM request_log WHERE domain LIKE ?`, "%"+search+"%").Scan(&totalRecords)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"draw":            c.DefaultQuery("draw", "1"),
		"recordsTotal":    totalRecords,
		"recordsFiltered": totalRecords,
		"details":         queries,
	})
}

func (apiServer *API) handleUpdateBlockStatus(c *gin.Context) {
	domain := c.Query("domain")
	blocked := c.Query("blocked")
	if domain == "" || blocked == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing query parameters"})
		return
	}

	action := map[string]func(string) error{
		"true":  apiServer.DnsServer.Blacklist.AddDomain,
		"false": apiServer.DnsServer.Blacklist.RemoveDomain,
	}[blocked]

	if action == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid value for blocked"})
		return
	}

	if err := action(domain); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}

	if blocked == "true" {
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s has been blacklisted.", domain)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s has been whitelisted.", domain)})
}

func (apiServer *API) getDomains(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")
	search := c.DefaultQuery("search", "")
	draw := c.DefaultQuery("draw", "1")

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeInt < 1 {
		pageSizeInt = 10
	}

	domains, total, err := apiServer.DnsServer.Blacklist.LoadPaginatedBlacklist(pageInt, pageSizeInt, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"draw":            draw,
		"domains":         domains,
		"recordsTotal":    total,
		"recordsFiltered": total,
	})
}

func (apiServer *API) getSettings(c *gin.Context) {
	dnsSettings := struct {
		Port                int      `json:"Port"`
		LoggingDisabled     bool     `json:"LoggingDisabled"`
		UpstreamDNS         []string `json:"UpstreamDNS"`
		PreferredUpstream   string   `json:"PreferredUpstream"`
		CacheTTL            int      `json:"CacheTTL"`
		StatisticsRetention int      `json:"StatisticsRetention"`
	}{
		Port:                apiServer.DnsServer.Config.Port,
		LoggingDisabled:     apiServer.DnsServer.Config.LoggingDisabled,
		UpstreamDNS:         apiServer.DnsServer.Config.UpstreamDNS,
		PreferredUpstream:   apiServer.DnsServer.Config.PreferredUpstream,
		CacheTTL:            int(apiServer.DnsServer.Config.CacheTTL.Seconds()),
		StatisticsRetention: apiServer.DnsServer.Config.StatisticsRetention,
	}

	c.JSON(http.StatusOK, gin.H{
		"api": apiServer.Config,
		"dns": dnsSettings,
	})
}

func (apiServer *API) updateSettings(c *gin.Context) {
	var updatedSettings map[string]interface{}
	if err := c.BindJSON(&updatedSettings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid settings data",
		})
		return
	}

	config := settings.Config{DNSServer: &apiServer.DnsServer.Config, APIServer: apiServer.Config}
	config.UpdateDNSSettings(updatedSettings)
	settingsJson, _ := json.MarshalIndent(updatedSettings, "", "  ")
	log.Info("Updated settings!")
	log.Debug("%s", string(settingsJson))

	apiServer.DnsServer.Config = *config.DNSServer
	apiServer.Config = config.APIServer

	c.JSON(http.StatusOK, gin.H{
		"api": apiServer.Config,
		"dns": apiServer.DnsServer.Config,
	})
}

func (apiServer *API) getClients(c *gin.Context) {
	uniqueClients := make(map[string]struct {
		Name     string
		LastSeen time.Time
	})

	rows, err := apiServer.DnsServer.DB.Query("SELECT client_ip, client_name, timestamp FROM request_log")
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var ip, name string
		var timestamp time.Time

		if err := rows.Scan(&ip, &name, &timestamp); err != nil {
			log.Error("%v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		if existing, exists := uniqueClients[ip]; !exists || timestamp.After(existing.LastSeen) {
			uniqueClients[ip] = struct {
				Name     string
				LastSeen time.Time
			}{
				Name:     name,
				LastSeen: timestamp,
			}
		}
	}

	var clients []map[string]interface{}
	for ip, entry := range uniqueClients {
		clients = append(clients, map[string]interface{}{
			"IP":       ip,
			"Name":     entry.Name,
			"lastSeen": entry.LastSeen,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"clients": clients,
	})
}

func (apiServer *API) getUpstreams(c *gin.Context) {
	upstreams := apiServer.DnsServer.Config.UpstreamDNS
	results := make([]map[string]string, len(upstreams))

	preferredUpstream := apiServer.DnsServer.Config.PreferredUpstream

	var wg sync.WaitGroup
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

func getUpstreamDetails(upstream, preferredUpstream string) map[string]string {
	host := strings.TrimSuffix(upstream, ":53")
	entry := map[string]string{
		"upstream":  upstream,
		"preferred": strconv.FormatBool(upstream == preferredUpstream),
	}

	entry["name"], entry["dnsPing"] = getDNSDetails(host)
	entry["icmpPing"] = getICMPPing(host)

	return entry
}

func getDNSDetails(host string) (string, string) {
	start := time.Now()
	names, err := net.LookupAddr(host)
	duration := time.Since(start)

	if err != nil {
		return "Error: " + err.Error(), "Error: " + err.Error()
	}
	if len(names) > 0 {
		return strings.TrimSuffix(names[0], "."), duration.String()
	}
	return "No name found", duration.String()
}

func getICMPPing(host string) string {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return "Error: " + err.Error()
	}
	pinger.Count = 1
	pinger.Timeout = 2 * time.Second

	var icmpPing string
	pinger.OnRecv = func(pkt *ping.Packet) {
		icmpPing = pkt.Rtt.String()
	}

	if err := pinger.Run(); err != nil {
		return "Error: " + err.Error()
	}
	return icmpPing
}

func (apiServer *API) createUpstreams(c *gin.Context) {
	type UpstreamsRequest struct {
		Upstreams []string `json:"upstreams"`
	}

	newUpstreams, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var request UpstreamsRequest
	if err := json.Unmarshal(newUpstreams, &request); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	var filteredUpstreams []string
	for _, upstream := range request.Upstreams {
		if !strings.Contains(upstream, ":") {
			upstream += ":53"
		}

		exists := false
		for _, existing := range apiServer.DnsServer.Config.UpstreamDNS {
			if existing == upstream {
				exists = true
				break
			}
		}

		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Upstream already exists"})
			return
		}
	}

	if len(filteredUpstreams) == 0 {
		log.Info("No new unique upstreams to add")
		c.JSON(http.StatusOK, gin.H{"message": "No new unique upstreams to add"})
		return
	}

	log.Info("Adding unique upstreams: %v", filteredUpstreams)
	apiServer.DnsServer.Config.UpstreamDNS = append(
		apiServer.DnsServer.Config.UpstreamDNS,
		filteredUpstreams...,
	)

	config := settings.Config{DNSServer: &apiServer.DnsServer.Config, APIServer: apiServer.Config}
	config.Save()
	c.JSON(http.StatusOK, gin.H{"added_upstreams": filteredUpstreams})
}

func (apiServer *API) removeUpstreams(c *gin.Context) {
	upstreamToDelete := c.Query("upstream")

	if upstreamToDelete == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'upstream' query parameter"})
		return
	}

	var updatedUpstreams []string
	for _, upstream := range apiServer.DnsServer.Config.UpstreamDNS {
		if upstream != upstreamToDelete {
			updatedUpstreams = append(updatedUpstreams, upstream)
		}
	}

	apiServer.DnsServer.Config.UpstreamDNS = updatedUpstreams

	c.JSON(http.StatusOK, gin.H{
		"message": "Upstream removed successfully",
	})
}

func (apiServer *API) clearLogs(c *gin.Context) {
	result, err := apiServer.DnsServer.DB.Exec("DELETE FROM request_log")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not clear logs", "reason": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()

	apiServer.DnsServer.Counters = server.CounterDetails{}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Cleared %d logs", rowsAffected),
	})
}

func (apiServer *API) setPreferredUpstream(c *gin.Context) {
	upstreamToSet := c.DefaultQuery("upstream", "")

	if upstreamToSet == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Upstream is required"})
		return
	}

	var found bool
	for _, upstream := range apiServer.DnsServer.Config.UpstreamDNS {
		if upstream == upstreamToSet {
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Upstream not found"})
		return
	}

	apiServer.DnsServer.Config.PreferredUpstream = upstreamToSet
	updatedMsg := fmt.Sprintf("Preferred upstream set to %s", apiServer.DnsServer.Config.PreferredUpstream)
	log.Info("%s", updatedMsg)

	config := settings.Config{DNSServer: &apiServer.DnsServer.Config, APIServer: apiServer.Config}
	config.Save()
	c.JSON(http.StatusOK, gin.H{"message": updatedMsg})
}

func (apiServer *API) getTopBlockedDomains(c *gin.Context) {
	rows, err := apiServer.DnsServer.DB.Query(`
		SELECT domain, COUNT(*) as hits
		FROM request_log
		WHERE blocked = 1
		GROUP BY domain
		ORDER BY hits DESC
		LIMIT 5
	`)
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var domains []map[string]interface{}
	for rows.Next() {
		var domain string
		var hits int
		if err := rows.Scan(&domain, &hits); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		domains = append(domains, map[string]interface{}{
			"name":      domain,
			"hits":      hits,
			"frequency": (hits * 100) / apiServer.DnsServer.Counters.BlockedRequests,
		})
	}

	c.JSON(http.StatusOK, gin.H{"domains": domains})
}

func (apiServer *API) getLists(c *gin.Context) {
	lists, err := apiServer.DnsServer.Blacklist.GetSourceStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"lists": lists})
}

func (apiServer *API) updateLists(c *gin.Context) {
	type UpdateListRequest struct {
		List    string   `json:"list"`
		Domains []string `json:"domains"`
	}

	updatedList, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var request UpdateListRequest
	if err := json.Unmarshal(updatedList, &request); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if request.List == "Custom" {
		apiServer.DnsServer.Blacklist.AddCustomDomains(request.Domains)
	} else {
		apiServer.DnsServer.Blacklist.AddDomains(request.Domains, request.List)
	}

	c.JSON(http.StatusOK, gin.H{"blockedLen": len(request.Domains)})
}

func (apiServer *API) getDomainsForList(c *gin.Context) {
	list := c.Query("list")
	if list == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'list' query parameter"})
		return
	}

	domains, err := apiServer.DnsServer.Blacklist.GetDomainsForList(list)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"domains": domains})
}

func (apiServer *API) updatePassword(c *gin.Context) {
	type PasswordChange struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}

	updatedList, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var request PasswordChange
	if err := json.Unmarshal(updatedList, &request); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if !apiServer.validateCredentials("admin", request.CurrentPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid current password"})
		return
	}

	apiServer.adminPassword = request.NewPassword
	log.Info("Password has been changed!")

	c.JSON(http.StatusOK, nil)
}

func getCPUTemperature() (float64, error) {
	tempFile := "/sys/class/thermal/thermal_zone0/temp"
	data, err := os.ReadFile(tempFile)
	if err != nil {
		return 0, err
	}

	tempStr := strings.TrimSpace(string(data))
	temp, err := strconv.ParseFloat(tempStr, 64)
	if err != nil {
		return 0, err
	}

	return temp / 1000, nil
}

func getServerIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("server IP not found")
}
