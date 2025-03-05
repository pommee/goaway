package api

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"goaway/internal/api/models"
	"goaway/internal/database"
	"goaway/internal/server"
	"goaway/internal/settings"
	"goaway/internal/updater"
	"goaway/internal/user"
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
	"github.com/gorilla/websocket"
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

	dbSize, err := getDBSize()
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
		"dbSize":            dbSize,
		"version":           apiServer.Version,
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
		"clients":           database.GetDistinctRequestIP(apiServer.DnsServer.DB),
	})
}

func (apiServer *API) getQueryTimestamps(c *gin.Context) {
	timestamps, err := database.GetRequestTimestampAndBlocked(apiServer.DnsServer.DB)
	if err != nil {
		log.Error("%v", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"queries": timestamps,
	})
}

func (apiServer *API) getQueryTypes(c *gin.Context) {
	queries, err := database.GetUniqueQueryTypes(apiServer.DnsServer.DB)
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"queries": queries,
	})
}

func (apiServer *API) getQueries(c *gin.Context) {
	query := parseQueryParams(c)
	queries, err := database.FetchQueries(apiServer.DnsServer.DB, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	total, err := database.CountQueries(apiServer.DnsServer.DB, query.Search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"draw":            c.DefaultQuery("draw", "1"),
		"recordsTotal":    total,
		"recordsFiltered": total,
		"details":         queries,
	})
}

func parseQueryParams(c *gin.Context) models.QueryParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.DefaultQuery("search", "")
	sortColumn := c.DefaultQuery("sortColumn", "timestamp")
	sortDirection := c.DefaultQuery("sortDirection", "desc")

	validColumns := map[string]string{
		"timestamp": "timestamp",
		"domain":    "domain",
		"client":    "client_ip",
		"ip":        "ip",
	}

	column, ok := validColumns[sortColumn]
	if !ok {
		column = "timestamp"
	}

	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "desc"
	}

	return models.QueryParams{
		Page:      page,
		PageSize:  pageSize,
		Search:    search,
		Column:    column,
		Direction: sortDirection,
		Offset:    (page - 1) * pageSize,
	}
}

func (apiServer *API) liveQueries(c *gin.Context) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	apiServer.DnsServer.WS = conn
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
	uniqueClients, err := database.FetchAllClients(apiServer.DnsServer.DB)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clients := make([]map[string]interface{}, 0, len(uniqueClients))
	for ip, entry := range uniqueClients {
		clients = append(clients, map[string]interface{}{
			"ip":       ip,
			"name":     entry.Name,
			"lastSeen": entry.LastSeen,
			"mac":      entry.Mac,
			"vendor":   entry.Vendor,
		})
	}

	c.JSON(http.StatusOK, gin.H{"clients": clients})
}

func (apiServer *API) getClientDetails(c *gin.Context) {
	clientIP := c.DefaultQuery("clientIP", "")
	clientRequestDetails, err := database.GetClientRequestDetails(apiServer.DnsServer.DB, clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mostQueriedDomain, err := database.GetMostQueriedDomainByIP(apiServer.DnsServer.DB, clientIP)
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	queriedDomains, err := database.GetAllQueriedDomainsByIP(apiServer.DnsServer.DB, clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"details": map[string]interface{}{
			"IP":                clientIP,
			"TotalRequests":     clientRequestDetails.TotalRequests,
			"UniqueDomains":     clientRequestDetails.UniqueDomains,
			"BlockedRequests":   clientRequestDetails.BlockedRequests,
			"CachedRequests":    clientRequestDetails.CachedRequests,
			"AvgResponseTimeMs": clientRequestDetails.AvgResponseTimeMs,
			"MostQueriedDomain": mostQueriedDomain,
			"LastSeen":          clientRequestDetails.LastSeen,
			"AllDomains":        queriedDomains,
		},
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
	ips, err := net.LookupIP(host)
	duration := time.Since(start)

	if err != nil {
		return "Error: " + err.Error(), "Error: " + err.Error()
	}
	if len(ips) > 0 {
		return ips[0].String(), duration.String()
	}
	return "No IP found", duration.String()
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

		if exists {
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

func (apiServer *API) clearQueries(c *gin.Context) {
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
	topBlockedDomains, err := database.GetTopBlockedDomains(apiServer.DnsServer.DB, apiServer.DnsServer.Counters.BlockedRequests)
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"domains": topBlockedDomains})
}

func (apiServer *API) getTopClients(c *gin.Context) {
	topClients, err := database.GetTopClients(apiServer.DnsServer.DB)
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"clients": topClients})
}

func (apiServer *API) getLists(c *gin.Context) {
	lists, err := apiServer.DnsServer.Blacklist.GetSourceStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"lists": lists})
}

func (apiServer *API) updateCustom(c *gin.Context) {
	type UpdateListRequest struct {
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

	err = apiServer.DnsServer.Blacklist.AddCustomDomains(request.Domains)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update custom blocklist."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"blockedLen": len(request.Domains)})
}

func (apiServer *API) toggleBlocklist(c *gin.Context) {
	type ToggledBlocklistRequest struct {
		Name string `json:"name"`
	}

	updatedBlocklistName, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	var request ToggledBlocklistRequest
	if err := json.Unmarshal(updatedBlocklistName, &request); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON format"})
		return
	}

	err = apiServer.DnsServer.Blacklist.ToggleBlocklistStatus(request.Name)
	if err != nil {
		log.Error("Failed to toggle blocklist status: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Toggled status for %s", request.Name)})
}

func (apiServer *API) addList(c *gin.Context) {
	name := c.Query("name")
	url := c.Query("url")

	if apiServer.DnsServer.Blacklist.BlocklistURL[name] != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List already exists"})
		return
	}

	err := apiServer.DnsServer.Blacklist.FetchAndLoadHosts(url, name)
	if err != nil {
		log.Error("%v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	apiServer.DnsServer.Blacklist.BlocklistURL[name] = url

	c.JSON(http.StatusOK, nil)
}

func (apiServer *API) removeList(c *gin.Context) {
	name := c.Query("name")

	if apiServer.DnsServer.Blacklist.BlocklistURL[name] == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List does not exist"})
		return
	}

	err := apiServer.DnsServer.Blacklist.RemoveSourceAndDomains(name)
	if err != nil {
		log.Error("%v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	delete(apiServer.DnsServer.Blacklist.BlocklistURL, name)
	c.JSON(http.StatusOK, nil)
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

func (apiServer *API) runUpdate(c *gin.Context) {
	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming unsupported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	sendSSE := func(message string) {
		_, err := fmt.Fprintf(w, "data: %s\n\n", message)
		if err != nil {
			return
		}
		flusher.Flush()
	}

	sendSSE("[INFO] Starting update process...")
	err := updater.SelfUpdate(sendSSE)
	if err != nil {
		sendSSE(fmt.Sprintf("[ERROR] Update failed: %s", err.Error()))
	} else {
		sendSSE("[INFO] Update successful!")
	}
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

	user := user.User{Username: "admin", Password: request.NewPassword}
	user.UpdatePassword(apiServer.DnsServer.DB)
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

func getDBSize() (float64, error) {
	file, err := os.Stat("database.db")
	if err != nil {
		return 0, err
	}

	sizeInBytes := file.Size()
	sizeInMB := float64(sizeInBytes) / (1024 * 1024)

	return sizeInMB, nil
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
