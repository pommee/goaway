package api

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"goaway/internal/api/models"
	"goaway/internal/database"
	"goaway/internal/server"
	"goaway/internal/settings"
	"goaway/internal/updater"
	"goaway/internal/user"
	"io"
	"io/fs"
	"mime"
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
	ipAddress, err := getServerIP()
	if err != nil {
		log.Error("Error getting IP address: %v", err)
		return
	}

	err = fs.WalkDir(content, "website/dist", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking through path %s: %w", path, err)
		}
		if d.IsDir() || path == "website/dist/index.html" {
			return nil
		}

		fileContent, err := content.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", path, err)
		}

		ext := strings.ToLower(filepath.Ext(path))
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		route := strings.TrimPrefix(path, "website/dist/")
		api.router.GET("/"+route, func(c *gin.Context) {
			c.Data(http.StatusOK, mimeType, fileContent)
		})

		return nil
	})
	if err != nil {
		log.Error("Error serving embedded content: %v", err)
		return
	}

	indexContent, err := content.ReadFile("website/dist/index.html")
	if err != nil {
		log.Error("Error reading index.html: %v", err)
		return
	}

	indexWithConfig := injectServerConfig(string(indexContent), ipAddress)
	handleIndexHTML := func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.Data(http.StatusOK, "text/html", []byte(indexWithConfig))
	}

	api.router.GET("/", handleIndexHTML)
	api.router.NoRoute(handleIndexHTML)
}

func injectServerConfig(htmlContent, serverIP string) string {
	serverConfigScript := `<script>
	window.SERVER_CONFIG = {
		serverIP: "` + serverIP + `",
		apiBaseURL: "http:\//` + serverIP + `:8080\/api"
	};
	</script>`

	return strings.Replace(
		htmlContent,
		"<head>",
		"<head>\n  "+serverConfigScript,
		1,
	)
}

func (api *API) handleLogin(c *gin.Context) {
	var creds Credentials
	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if api.validateCredentials(creds.Username, creds.Password) {
		token, err := generateToken(creds.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
			return
		}

		serverIP, _ := getServerIP()
		c.Header("Access-Control-Allow-Origin", fmt.Sprintf("http://%s:8080", serverIP))
		c.Header("Access-Control-Allow-Credentials", "true")

		setAuthCookie(c.Writer, token)
		c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}
}

func (api *API) handleServer(c *gin.Context) {
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
		"portDNS":           api.Config.Port,
		"portWebsite":       api.DnsServer.Config.Port,
		"totalMem":          float64(vMem.Total) / 1024 / 1024 / 1024,
		"usedMem":           float64(vMem.Used) / 1024 / 1024 / 1024,
		"usedMemPercentage": float64(vMem.Free) / 1024 / 1024 / 1024,
		"cpuUsage":          cpuUsage[0],
		"cpuTemp":           temp,
		"dbSize":            dbSize,
		"version":           api.Version,
	})
}

func (api *API) getAuthentication(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"enabled": api.Config.Authentication})
}

func (api *API) handleMetrics(c *gin.Context) {
	allowedQueries := api.DnsServer.Counters.AllowedRequests
	blockedQueries := api.DnsServer.Counters.BlockedRequests
	totalQueries := allowedQueries + blockedQueries

	var percentageBlocked float64
	if totalQueries > 0 {
		percentageBlocked = (float64(blockedQueries) / float64(totalQueries)) * 100
	}

	domainsLength, _ := api.DnsServer.Blacklist.CountDomains()
	c.JSON(http.StatusOK, gin.H{
		"allowed":           allowedQueries,
		"blocked":           blockedQueries,
		"total":             totalQueries,
		"percentageBlocked": percentageBlocked,
		"domainBlockLen":    domainsLength,
		"clients":           database.GetDistinctRequestIP(api.DnsServer.DB),
	})
}

func (api *API) getQueryTimestamps(c *gin.Context) {
	timestamps, err := database.GetRequestTimestampAndBlocked(api.DnsServer.DB)
	if err != nil {
		log.Error("%v", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"queries": timestamps,
	})
}

func (api *API) getQueryTypes(c *gin.Context) {
	queries, err := database.GetUniqueQueryTypes(api.DnsServer.DB)
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"queries": queries,
	})
}

func (api *API) getQueries(c *gin.Context) {
	query := parseQueryParams(c)
	queries, err := database.FetchQueries(api.DnsServer.DB, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	total, err := database.CountQueries(api.DnsServer.DB, query.Search)
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

func (api *API) liveQueries(c *gin.Context) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	api.DnsServer.WS = conn
}

func (api *API) handleUpdateBlockStatus(c *gin.Context) {
	domain := c.Query("domain")
	blocked := c.Query("blocked")
	if domain == "" || blocked == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing query parameters"})
		return
	}

	action := map[string]func(string) error{
		"true":  api.DnsServer.Blacklist.AddDomain,
		"false": api.DnsServer.Blacklist.RemoveDomain,
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

func (api *API) getDomains(c *gin.Context) {
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

	domains, total, err := api.DnsServer.Blacklist.LoadPaginatedBlacklist(pageInt, pageSizeInt, search)
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

func (api *API) getSettings(c *gin.Context) {
	dnsSettings := struct {
		Port                int      `json:"Port"`
		LoggingDisabled     bool     `json:"LoggingDisabled"`
		UpstreamDNS         []string `json:"UpstreamDNS"`
		PreferredUpstream   string   `json:"PreferredUpstream"`
		CacheTTL            int      `json:"CacheTTL"`
		StatisticsRetention int      `json:"StatisticsRetention"`
	}{
		Port:                api.DnsServer.Config.Port,
		LoggingDisabled:     api.DnsServer.Config.LoggingDisabled,
		UpstreamDNS:         api.DnsServer.Config.UpstreamDNS,
		PreferredUpstream:   api.DnsServer.Config.PreferredUpstream,
		CacheTTL:            int(api.DnsServer.Config.CacheTTL.Seconds()),
		StatisticsRetention: api.DnsServer.Config.StatisticsRetention,
	}

	c.JSON(http.StatusOK, gin.H{
		"api": api.Config,
		"dns": dnsSettings,
	})
}

func (api *API) updateSettings(c *gin.Context) {
	var updatedSettings map[string]interface{}
	if err := c.BindJSON(&updatedSettings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid settings data",
		})
		return
	}

	config := settings.Config{DNSServer: &api.DnsServer.Config, APIServer: api.Config}
	config.UpdateDNSSettings(updatedSettings)
	log.Info("Updated settings!")
	settingsJson, _ := json.MarshalIndent(updatedSettings, "", "  ")
	log.Debug("%s", string(settingsJson))

	api.DnsServer.Config = *config.DNSServer
	api.Config = config.APIServer

	c.JSON(http.StatusOK, gin.H{
		"api": api.Config,
		"dns": api.DnsServer.Config,
	})
}

func (api *API) getClients(c *gin.Context) {
	uniqueClients, err := database.FetchAllClients(api.DnsServer.DB)
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

func (api *API) getClientDetails(c *gin.Context) {
	clientIP := c.DefaultQuery("clientIP", "")
	clientRequestDetails, err := database.GetClientRequestDetails(api.DnsServer.DB, clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mostQueriedDomain, err := database.GetMostQueriedDomainByIP(api.DnsServer.DB, clientIP)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	queriedDomains, err := database.GetAllQueriedDomainsByIP(api.DnsServer.DB, clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"details": map[string]interface{}{
			"ip":                clientIP,
			"totalRequests":     clientRequestDetails.TotalRequests,
			"uniqueDomains":     clientRequestDetails.UniqueDomains,
			"blockedRequests":   clientRequestDetails.BlockedRequests,
			"cachedRequests":    clientRequestDetails.CachedRequests,
			"avgResponseTimeMs": clientRequestDetails.AvgResponseTimeMs,
			"mostQueriedDomain": mostQueriedDomain,
			"lastSeen":          clientRequestDetails.LastSeen,
			"allDomains":        queriedDomains,
		},
	})
}

func (api *API) getResolutions(c *gin.Context) {
	resolutions, err := database.FetchResolutions(api.DnsServer.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"resolutions": resolutions})
}

func (api *API) createResolution(c *gin.Context) {
	var newResolution models.NewResolution
	if err := c.BindJSON(&newResolution); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid resolution data",
		})
		return
	}

	database.CreateNewResolution(api.DnsServer.DB, newResolution.IP, newResolution.Domain)

	c.Status(http.StatusOK)
}

func (api *API) deleteResolution(c *gin.Context) {
	domain := c.Query("domain")
	ip := c.Query("ip")

	rowsAffected, err := database.DeleteResolution(api.DnsServer.DB, ip, domain)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": rowsAffected})
}

func (api *API) getUpstreams(c *gin.Context) {
	upstreams := api.DnsServer.Config.UpstreamDNS
	results := make([]map[string]string, len(upstreams))

	preferredUpstream := api.DnsServer.Config.PreferredUpstream

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

func (api *API) createUpstreams(c *gin.Context) {
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
		for _, existing := range api.DnsServer.Config.UpstreamDNS {
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
	api.DnsServer.Config.UpstreamDNS = append(
		api.DnsServer.Config.UpstreamDNS,
		filteredUpstreams...,
	)

	config := settings.Config{DNSServer: &api.DnsServer.Config, APIServer: api.Config}
	config.Save()
	c.JSON(http.StatusOK, gin.H{"added_upstreams": filteredUpstreams})
}

func (api *API) removeUpstreams(c *gin.Context) {
	upstreamToDelete := c.Query("upstream")

	if upstreamToDelete == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'upstream' query parameter"})
		return
	}

	var updatedUpstreams []string
	for _, upstream := range api.DnsServer.Config.UpstreamDNS {
		if upstream != upstreamToDelete {
			updatedUpstreams = append(updatedUpstreams, upstream)
		}
	}

	api.DnsServer.Config.UpstreamDNS = updatedUpstreams

	c.JSON(http.StatusOK, gin.H{
		"message": "Upstream removed successfully",
	})
}

func (api *API) clearQueries(c *gin.Context) {
	result, err := api.DnsServer.DB.Exec("DELETE FROM request_log")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not clear logs", "reason": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()

	api.DnsServer.Counters = server.CounterDetails{}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Cleared %d logs", rowsAffected),
	})
}

func (api *API) setPreferredUpstream(c *gin.Context) {
	upstreamToSet := c.DefaultQuery("upstream", "")

	if upstreamToSet == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Upstream is required"})
		return
	}

	var found bool
	for _, upstream := range api.DnsServer.Config.UpstreamDNS {
		if upstream == upstreamToSet {
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Upstream not found"})
		return
	}

	api.DnsServer.Config.PreferredUpstream = upstreamToSet
	updatedMsg := fmt.Sprintf("Preferred upstream set to %s", api.DnsServer.Config.PreferredUpstream)
	log.Info("%s", updatedMsg)

	config := settings.Config{DNSServer: &api.DnsServer.Config, APIServer: api.Config}
	config.Save()
	c.JSON(http.StatusOK, gin.H{"message": updatedMsg})
}

func (api *API) getTopBlockedDomains(c *gin.Context) {
	topBlockedDomains, err := database.GetTopBlockedDomains(api.DnsServer.DB, api.DnsServer.Counters.BlockedRequests)
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"domains": topBlockedDomains})
}

func (api *API) getTopClients(c *gin.Context) {
	topClients, err := database.GetTopClients(api.DnsServer.DB)
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"clients": topClients})
}

func (api *API) getLists(c *gin.Context) {
	lists, err := api.DnsServer.Blacklist.GetSourceStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"lists": lists})
}

func (api *API) updateCustom(c *gin.Context) {
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

	err = api.DnsServer.Blacklist.AddCustomDomains(request.Domains)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update custom blocklist."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"blockedLen": len(request.Domains)})
}

func (api *API) toggleBlocklist(c *gin.Context) {
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

	err = api.DnsServer.Blacklist.ToggleBlocklistStatus(request.Name)
	if err != nil {
		log.Error("Failed to toggle blocklist status: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Toggled status for %s", request.Name)})
}

func (api *API) addList(c *gin.Context) {
	name := c.Query("name")
	url := c.Query("url")

	if api.DnsServer.Blacklist.BlocklistURL[name] != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List already exists"})
		return
	}

	err := api.DnsServer.Blacklist.FetchAndLoadHosts(url, name)
	if err != nil {
		log.Error("%v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	api.DnsServer.Blacklist.BlocklistURL[name] = url

	c.JSON(http.StatusOK, nil)
}

func (api *API) removeList(c *gin.Context) {
	name := c.Query("name")

	if api.DnsServer.Blacklist.BlocklistURL[name] == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List does not exist"})
		return
	}

	err := api.DnsServer.Blacklist.RemoveSourceAndDomains(name)
	if err != nil {
		log.Error("%v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	delete(api.DnsServer.Blacklist.BlocklistURL, name)
	c.JSON(http.StatusOK, nil)
}

func (api *API) getDomainsForList(c *gin.Context) {
	list := c.Query("list")
	if list == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'list' query parameter"})
		return
	}

	domains, err := api.DnsServer.Blacklist.GetDomainsForList(list)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"domains": domains})
}

func (api *API) runUpdate(c *gin.Context) {
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

func (api *API) updatePassword(c *gin.Context) {
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

	if !api.validateCredentials("admin", request.CurrentPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid current password"})
		return
	}

	existingUser := user.User{Username: "admin", Password: request.NewPassword}
	if err = existingUser.UpdatePassword(api.DnsServer.DB); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to update password"})
		return
	}

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
