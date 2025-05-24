package api

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"goaway/backend/api/models"
	"goaway/backend/api/user"
	"goaway/backend/dns/database"
	"goaway/backend/dns/server/prefetch"
	"goaway/backend/settings"
	"goaway/backend/updater"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	probing "github.com/prometheus-community/pro-bing"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func (api *API) ServeEmbeddedContent(content embed.FS) {
	ipAddress, err := getServerIP()
	if err != nil {
		log.Error("Error getting IP address: %v", err)
		return
	}

	err = fs.WalkDir(content, "client/dist", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking through path %s: %w", path, err)
		}
		if d.IsDir() || path == "client/dist/index.html" {
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

		route := strings.TrimPrefix(path, "client/dist/")
		api.router.GET("/"+route, func(c *gin.Context) {
			c.Data(http.StatusOK, mimeType, fileContent)
		})

		return nil
	})
	if err != nil {
		log.Error("Error serving embedded content: %v", err)
		return
	}

	indexContent, err := content.ReadFile("client/dist/index.html")
	if err != nil {
		log.Error("Error reading index.html: %v", err)
		return
	}

	indexWithConfig := injectServerConfig(string(indexContent), ipAddress, api.Config.API.Port)
	handleIndexHTML := func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.Data(http.StatusOK, "text/html", []byte(indexWithConfig))
	}

	api.router.GET("/", handleIndexHTML)
	api.router.NoRoute(handleIndexHTML)
}

func injectServerConfig(htmlContent, serverIP string, port int) string {
	serverConfigScript := `<script>
	window.SERVER_CONFIG = {
		ip: "` + serverIP + `",
		port: "` + strconv.Itoa(port) + `"
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

		c.Header("Access-Control-Allow-Origin", "*")
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
		"portDNS":           api.Config.DNS.Port,
		"portWebsite":       api.DNSPort,
		"totalMem":          float64(vMem.Total) / 1024 / 1024 / 1024,
		"usedMem":           float64(vMem.Used) / 1024 / 1024 / 1024,
		"usedMemPercentage": float64(vMem.Free) / 1024 / 1024 / 1024,
		"cpuUsage":          cpuUsage[0],
		"cpuTemp":           temp,
		"dbSize":            dbSize,
		"version":           api.Version,
		"commit":            api.Commit,
		"date":              api.Date,
	})
}

func (api *API) getAuthentication(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"enabled": api.Authentication})
}

func (api *API) handleMetrics(c *gin.Context) {
	allowed, blocked, err := api.Blacklist.GetAllowedAndBlocked()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	total := allowed + blocked

	var percentageBlocked float64
	if total > 0 {
		percentageBlocked = (float64(blocked) / float64(total)) * 100
	}

	domainsLength, _ := api.Blacklist.CountDomains()
	c.JSON(http.StatusOK, gin.H{
		"allowed":           allowed,
		"blocked":           blocked,
		"total":             total,
		"percentageBlocked": percentageBlocked,
		"domainBlockLen":    domainsLength,
		"clients":           database.GetDistinctRequestIP(api.DB),
	})
}

func (api *API) getQueryTimestamps(c *gin.Context) {
	timestamps, err := database.GetRequestSummaryByInterval(api.DB)
	if err != nil {
		log.Error("%v", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"queries": timestamps,
	})
}

func (api *API) getQueryTypes(c *gin.Context) {
	queries, err := database.GetUniqueQueryTypes(api.DB)
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
	queries, err := database.FetchQueries(api.DB, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	total, err := database.CountQueries(api.DB, query.Search)
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

func (api *API) handleUpdateBlockStatus(c *gin.Context) {
	domain := c.Query("domain")
	blocked := c.Query("blocked")
	if domain == "" || blocked == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing query parameters"})
		return
	}

	action := map[string]func(string) error{
		"true":  api.Blacklist.AddDomain,
		"false": api.Blacklist.RemoveDomain,
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

	domains, total, err := api.Blacklist.LoadPaginatedBlacklist(pageInt, pageSizeInt, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	c.JSON(http.StatusOK, gin.H{
		"settings": api.Config,
	})
}

func (api *API) updateSettings(c *gin.Context) {
	var updatedSettings settings.Config
	if err := c.BindJSON(&updatedSettings); err != nil {
		log.Warning("Could not save new settings, reason: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid settings data",
		})
		return
	}

	api.Config.UpdateSettings(updatedSettings)
	log.Info("Updated settings!")
	settingsJson, _ := json.MarshalIndent(updatedSettings, "", "  ")
	log.Debug("%s", string(settingsJson))

	c.JSON(http.StatusOK, gin.H{
		"config": api.Config,
	})
}

func (api *API) getClients(c *gin.Context) {
	uniqueClients, err := database.FetchAllClients(api.DB)
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
	clientRequestDetails, err := database.GetClientRequestDetails(api.DB, clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mostQueriedDomain, err := database.GetMostQueriedDomainByIP(api.DB, clientIP)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	queriedDomains, err := database.GetAllQueriedDomainsByIP(api.DB, clientIP)
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
	resolutions, err := database.FetchResolutions(api.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"resolutions": resolutions})
}

func (api *API) createResolution(c *gin.Context) {
	type NewResolution struct {
		IP     string
		Domain string
	}

	var newResolution NewResolution
	if err := c.BindJSON(&newResolution); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid resolution data",
		})
		return
	}

	err := database.CreateNewResolution(api.DB, newResolution.IP, newResolution.Domain)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	c.Status(http.StatusOK)
}

func (api *API) pauseBlocking(c *gin.Context) {
	type BlockTime struct {
		Time int `json:"time"`
	}

	var blockTime BlockTime
	if err := c.BindJSON(&blockTime); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid time data",
		})
		return
	}

	api.Config.DNS.Status = settings.Status{
		Paused:    true,
		PausedAt:  time.Now(),
		PauseTime: blockTime.Time,
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (api *API) deleteResolution(c *gin.Context) {
	domain := c.Query("domain")
	ip := c.Query("ip")

	rowsAffected, err := database.DeleteResolution(api.DB, ip, domain)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s does not exist", domain)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": rowsAffected})
}

func (api *API) clearBlocking(c *gin.Context) {
	api.Config.DNS.Status = settings.Status{}
	c.JSON(http.StatusOK, gin.H{})
}

func (api *API) markNotificationAsRead(c *gin.Context) {
	type NotificationsRead struct {
		NotificationIDs []int `json:"notificationIds"`
	}

	notificationsRead, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var request NotificationsRead
	if err := json.Unmarshal(notificationsRead, &request); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	err = api.Notifications.MarkNotificationsAsRead(request.NotificationIDs)
	if err != nil {
		log.Warning("Unable to mark notifications as read %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Unable to mark notifications as read %v", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (api *API) getUpstreams(c *gin.Context) {
	upstreams := api.Config.DNS.UpstreamDNS
	results := make([]map[string]any, len(upstreams))

	preferredUpstream := api.Config.DNS.PreferredUpstream

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

func getUpstreamDetails(upstream, preferredUpstream string) map[string]any {
	host := strings.TrimSuffix(upstream, ":53")
	entry := map[string]any{
		"upstream":  upstream,
		"preferred": upstream == preferredUpstream,
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
	pinger, err := probing.NewPinger(host)
	if err != nil {
		return "Error: " + err.Error()
	}
	pinger.Count = 3
	pinger.Timeout = 2 * time.Second

	err = pinger.Run()
	if err != nil {
		log.Warning("Could not get ICMP ping from host: %s", host)
		panic(err)
	}
	stats := pinger.Statistics()
	return stats.AvgRtt.String()
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

func (api *API) deleteUpstream(c *gin.Context) {
	upstreamToDelete := c.Query("upstream")

	if upstreamToDelete == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'upstream' query parameter"})
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

	c.JSON(http.StatusOK, gin.H{
		"message": "Upstream removed successfully",
	})
}

func (api *API) clearQueries(c *gin.Context) {
	result, err := api.DB.Exec("DELETE FROM request_log")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not clear logs", "reason": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()

	api.Blacklist.Vacuum()

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Cleared %d logs", rowsAffected),
	})
}

func (api *API) getTopBlockedDomains(c *gin.Context) {
	_, blocked, _ := api.Blacklist.GetAllowedAndBlocked()
	topBlockedDomains, err := database.GetTopBlockedDomains(api.DB, blocked)
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"domains": topBlockedDomains})
}

func (api *API) getTopClients(c *gin.Context) {
	topClients, err := database.GetTopClients(api.DB)
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"clients": topClients})
}

func (api *API) getLists(c *gin.Context) {
	lists, err := api.Blacklist.GetSourceStatistics()
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

	err = api.Blacklist.AddCustomDomains(request.Domains)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update custom blocklist."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"blockedLen": len(request.Domains)})
}

func (api *API) createPrefetchedDomain(c *gin.Context) {
	type NewPrefetch struct {
		Domain  string `json:"domain"`
		Refresh int    `json:"refresh"`
		QType   int    `json:"qtype"`
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var prefetchedDomain NewPrefetch
	if err := json.Unmarshal(body, &prefetchedDomain); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	err = api.PrefetchedDomainsManager.AddPrefetchedDomain(prefetchedDomain.Domain, prefetchedDomain.Refresh, prefetchedDomain.QType)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (api *API) fetchPrefetchedDomains(c *gin.Context) {
	prefetchedDomains := make([]prefetch.PrefetchedDomain, 0)
	for _, b := range api.PrefetchedDomainsManager.Domains {
		prefetchedDomains = append(prefetchedDomains, b)
	}
	c.JSON(http.StatusOK, gin.H{"domains": prefetchedDomains})
}

func (api *API) deletePrefetchedDomain(c *gin.Context) {
	domainPrefetchToDelete := c.Query("domain")

	domain := api.PrefetchedDomainsManager.Domains[domainPrefetchToDelete]
	if (domain == prefetch.PrefetchedDomain{}) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s does not exist", domainPrefetchToDelete)})
		return
	}

	err := api.PrefetchedDomainsManager.RemovePrefetchedDomain(domainPrefetchToDelete)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (api *API) fetchNotifications(c *gin.Context) {
	notifications, err := api.Notifications.ReadNotifications()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

func (api *API) removeDomainFromCustom(c *gin.Context) {
	domain := c.Query("domain")

	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty domain name"})
	}

	err := api.Blacklist.RemoveCustomDomain(domain)
	if err != nil {
		log.Debug("Error occured while removing domain from custom list: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update custom blocklist."})
		return
	}

	c.Status(http.StatusOK)
}

func (api *API) addList(c *gin.Context) {
	name := c.Query("name")
	url := c.Query("url")

	if api.Blacklist.BlocklistURL[name] != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List already exists"})
		return
	}

	err := api.Blacklist.FetchAndLoadHosts(url, name)
	if err != nil {
		log.Error("%v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	api.Blacklist.BlocklistURL[name] = url

	c.JSON(http.StatusOK, nil)
}

func (api *API) fetchUpdatedList(c *gin.Context) {
	name := c.Query("name")
	url := c.Query("url")

	if api.Blacklist.BlocklistURL[name] == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List does not exist"})
		return
	}

	if name == "" || url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'name' or 'url' query parameter"})
		return
	}

	remoteDomains, remoteChecksum, err := api.Blacklist.FetchRemoteHostsList(url)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dbDomains, dbChecksum, err := api.Blacklist.FetchDBHostsList(name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if remoteChecksum == dbChecksum {
		c.JSON(http.StatusOK, gin.H{"updateAvailable": false, "message": "No list updates available"})
		return
	}

	diff := func(a, b []string) []string {
		mb := make(map[string]struct{}, len(b))
		for _, x := range b {
			mb[x] = struct{}{}
		}
		diff := make([]string, 0)
		for _, x := range a {
			if _, found := mb[x]; !found {
				diff = append(diff, x)
			}
		}
		return diff
	}

	c.JSON(http.StatusOK, gin.H{
		"updateAvailable": true,
		"remoteChecksum":  remoteChecksum,
		"dbChecksum":      dbChecksum,
		"diffAdded":       diff(remoteDomains, dbDomains),
		"diffRemoved":     diff(dbDomains, remoteDomains),
	})
}

func (api *API) runUpdateList(c *gin.Context) {
	name := c.Query("name")
	url := c.Query("url")

	if api.Blacklist.BlocklistURL[name] == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List does not exist"})
		return
	}

	if name == "" || url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'name' or 'url' query parameter"})
		return
	}

	err := api.Blacklist.RemoveSourceAndDomains(name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = api.Blacklist.FetchAndLoadHosts(url, name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = api.Blacklist.PopulateBlocklistCache()
	if err != nil {
		message := fmt.Sprintf("Unable to re-populate the blocklist cache: %v", err)
		log.Warning("%s", message)
		c.JSON(http.StatusBadGateway, gin.H{"error": message})
	}

	c.Status(http.StatusOK)
}

func (api *API) removeList(c *gin.Context) {
	name := c.Query("name")

	if api.Blacklist.BlocklistURL[name] == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List does not exist"})
		return
	}

	err := api.Blacklist.RemoveSourceAndDomains(name)
	if err != nil {
		log.Error("%v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	delete(api.Blacklist.BlocklistURL, name)
	c.JSON(http.StatusOK, nil)
}

func (api *API) getDomainsForList(c *gin.Context) {
	list := c.Query("list")
	if list == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'list' query parameter"})
		return
	}

	domains, err := api.Blacklist.GetDomainsForList(list)
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

	sendSSE("[i] Starting update process...")
	err := updater.SelfUpdate(sendSSE)
	if err != nil {
		sendSSE(fmt.Sprintf("[ERROR] Update failed: %s", err.Error()))
		c.Status(http.StatusBadRequest)
	} else {
		sendSSE("[i] Update successful!")
		c.Status(http.StatusOK)
	}
}

func (api *API) getBlocking(c *gin.Context) {
	if api.Config.DNS.Status.Paused {
		elapsed := time.Since(api.Config.DNS.Status.PausedAt).Seconds()
		remainingTime := api.Config.DNS.Status.PauseTime - int(elapsed)

		if remainingTime <= 0 {
			c.JSON(http.StatusOK, gin.H{"paused": false})
		} else {
			c.JSON(http.StatusOK, gin.H{"paused": true, "timeLeft": remainingTime})
		}
	}

	if !api.Config.DNS.Status.Paused {
		c.JSON(http.StatusOK, gin.H{"paused": false})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is not valid"})
		return
	}

	existingUser := user.User{Username: "admin", Password: request.NewPassword}
	if err = existingUser.UpdatePassword(api.DB); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to update password"})
		return
	}

	log.Info("Password has been changed!")
	c.Status(http.StatusOK)
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

func (api *API) toggleBlocklist(c *gin.Context) {
	blocklist := c.Query("blocklist")

	if blocklist == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid blocklist"})
		return
	}

	err := api.Blacklist.ToggleBlocklistStatus(blocklist)
	if err != nil {
		log.Error("Failed to toggle blocklist status: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to toggle status for %s", blocklist)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Toggled status for %s", blocklist)})
}

func (api *API) exportDatabase(c *gin.Context) {
	log.Debug("Starting export of database")

	const databaseName = "database.db"
	if _, err := os.Stat(databaseName); err != nil {
		if os.IsNotExist(err) {
			log.Error("Database file not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "Database file not found"})
		} else {
			log.Error("Error accessing database file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	file, err := os.Open(databaseName)
	if err != nil {
		log.Error("Error opening database file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer func(tx *os.File) {
		_ = file.Close()
	}(file)

	fileInfo, err := file.Stat()
	if err != nil {
		log.Error("Error getting file info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+databaseName)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Header("Cache-Control", "no-cache")

	c.DataFromReader(http.StatusOK, fileInfo.Size(), "application/octet-stream", file, nil)
}

func (api *API) createAPIKey(c *gin.Context) {
	type NewApiKeyName struct {
		Name string `json:"name"`
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var request NewApiKeyName
	if err := json.Unmarshal(body, &request); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	apiKey, err := api.KeyManager.CreateApiKey(request.Name)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"key": apiKey})
}

func (api *API) getAPIKeys(c *gin.Context) {
	apiKeys, err := api.KeyManager.GetAllApiKeys()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"keys": apiKeys})
}

func (api *API) deleteAPIKey(c *gin.Context) {
	keyToDelete := c.Query("key")

	err := api.KeyManager.DeleteApiKey(keyToDelete)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted api key!"})
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
	return float64(file.Size()) / (1024 * 1024), nil
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
