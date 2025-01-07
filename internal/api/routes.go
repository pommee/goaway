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
			api.Router.GET("/", func(c *gin.Context) {
				c.Header("X-Server-IP", ipAddress)
				c.Data(http.StatusOK, mimeType, fileContent)
			})
		}

		api.Router.GET("/"+route, func(c *gin.Context) {
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

func (websiteServer *API) handleServer(c *gin.Context) {
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
		"portDNS":           websiteServer.port,
		"portWebsite":       websiteServer.dnsServer.Config.Port,
		"totalMem":          float64(vMem.Total) / 1024 / 1024 / 1024,
		"usedMem":           float64(vMem.Used) / 1024 / 1024 / 1024,
		"usedMemPercentage": float64(vMem.Free) / 1024 / 1024 / 1024,
		"cpuUsage":          cpuUsage[0],
		"cpuTemp":           temp,
	})
}

func (websiteServer *API) getAuthentication(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"disabled": websiteServer.DisableAuthentication})
}

func (websiteServer *API) handleMetrics(c *gin.Context) {
	allowedQueries := websiteServer.dnsServer.Counters.AllowedRequests
	blockedQueries := websiteServer.dnsServer.Counters.BlockedRequests
	totalQueries := allowedQueries + blockedQueries

	var percentageBlocked float64
	if totalQueries > 0 {
		percentageBlocked = (float64(blockedQueries) / float64(totalQueries)) * 100
	}

	domainsLength, _ := websiteServer.dnsServer.Blacklist.CountDomains()
	c.JSON(http.StatusOK, gin.H{
		"allowed":           allowedQueries,
		"blocked":           blockedQueries,
		"total":             totalQueries,
		"percentageBlocked": percentageBlocked,
		"domainBlockLen":    domainsLength,
	})
}

func (websiteServer *API) getQueryTimestamps(c *gin.Context) {
	type QueryEntry struct {
		Timestamp time.Time `json:"timestamp"`
		Blocked   bool      `json:"blocked"`
	}

	rows, err := websiteServer.dnsServer.DB.Query("SELECT timestamp, blocked FROM request_log")
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

func (websiteServer *API) handleQueriesData(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.DefaultQuery("search", "")
	sortColumn := c.DefaultQuery("sortColumn", "timestamp")
	sortDirection := c.DefaultQuery("sortDirection", "desc")

	offset := (page - 1) * pageSize

	query := `
		SELECT timestamp, domain, blocked, cached, response_time_ns, client_ip, client_name
		FROM request_log
		WHERE domain LIKE ?
		ORDER BY ` + sortColumn + ` ` + sortDirection + `
		LIMIT ? OFFSET ?`

	rows, err := websiteServer.dnsServer.DB.Query(query, "%"+search+"%", pageSize, offset)
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	queries := []server.RequestLogEntry{}
	for rows.Next() {
		var query server.RequestLogEntry
		if query.ClientInfo == nil {
			query.ClientInfo = &server.Client{}
		}
		if err := rows.Scan(&query.Timestamp, &query.Domain, &query.Blocked, &query.Cached, &query.ResponseTimeNS, &query.ClientInfo.IP, &query.ClientInfo.Name); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		queries = append(queries, query)
	}

	var totalRecords int
	err = websiteServer.dnsServer.DB.QueryRow(`SELECT COUNT(*) FROM request_log WHERE domain LIKE ?`, "%"+search+"%").Scan(&totalRecords)
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

func (websiteServer *API) handleUpdateBlockStatus(c *gin.Context) {
	domain := c.Query("domain")
	blocked := c.Query("blocked")
	if domain == "" || blocked == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing query parameters"})
		return
	}

	action := map[string]func(string) error{
		"true":  websiteServer.dnsServer.Blacklist.AddDomain,
		"false": websiteServer.dnsServer.Blacklist.RemoveDomain,
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

func (websiteServer *API) getDomains(c *gin.Context) {
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

	domains, total, err := websiteServer.dnsServer.Blacklist.LoadPaginatedBlacklist(pageInt, pageSizeInt, search)
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

func (websiteServer *API) getSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"settings": websiteServer.dnsServer.Config,
	})
}

func (websiteServer *API) updateSettings(c *gin.Context) {
	var updatedSettings map[string]interface{}
	if err := c.BindJSON(&updatedSettings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid settings data",
		})
		return
	}

	settings.UpdateSettings(websiteServer.dnsServer, updatedSettings)
	settingsJson, _ := json.MarshalIndent(updatedSettings, "", "  ")
	log.Info("Updated settings!")
	log.Debug("%s", string(settingsJson))

	c.JSON(http.StatusOK, gin.H{
		"settings": websiteServer.dnsServer.Config,
	})
}

func (websiteServer *API) getClients(c *gin.Context) {
	uniqueClients := make(map[string]struct {
		Name     string
		LastSeen time.Time
	})

	rows, err := websiteServer.dnsServer.DB.Query("SELECT client_ip, client_name, timestamp FROM request_log")
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

func (websiteServer *API) getUpstreams(c *gin.Context) {
	upstreams := websiteServer.dnsServer.Config.UpstreamDNS
	results := make([]map[string]string, 0)

	preferredUpstream := websiteServer.dnsServer.Config.PreferredUpstream

	for _, upstream := range upstreams {
		host := strings.TrimSuffix(upstream, ":53")
		entry := make(map[string]string)
		entry["upstream"] = upstream

		if upstream == preferredUpstream {
			entry["preferred"] = "true"
		} else {
			entry["preferred"] = "false"
		}

		start := time.Now()
		names, addrErr := net.LookupAddr(host)
		duration := time.Since(start)

		if addrErr != nil {
			entry["name"] = "Error: " + addrErr.Error()
			entry["dnsPing"] = "Error: " + addrErr.Error()
		} else if len(names) > 0 {
			entry["name"] = strings.TrimSuffix(names[0], ".")
			entry["dnsPing"] = duration.String()
		} else {
			entry["name"] = "No name found"
			entry["dnsPing"] = duration.String()
		}

		pinger, err := ping.NewPinger(host)
		if err != nil {
			entry["icmpPing"] = "Error: " + err.Error()
		} else {
			pinger.Count = 1
			pinger.Timeout = 2 * time.Second
			pinger.OnRecv = func(pkt *ping.Packet) {
				entry["icmpPing"] = pkt.Rtt.String()
			}

			err := pinger.Run()
			if err != nil {
				entry["icmpPing"] = "Error: " + err.Error()
			}
		}

		results = append(results, entry)
	}

	c.JSON(http.StatusOK, gin.H{
		"upstreams":         results,
		"preferredUpstream": preferredUpstream,
	})
}

func (websiteServer *API) createUpstreams(c *gin.Context) {
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
		for _, existing := range websiteServer.dnsServer.Config.UpstreamDNS {
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
	websiteServer.dnsServer.Config.UpstreamDNS = append(
		websiteServer.dnsServer.Config.UpstreamDNS,
		filteredUpstreams...,
	)

	settings.SaveSettings(&websiteServer.dnsServer.Config)
	c.JSON(http.StatusOK, gin.H{"added_upstreams": filteredUpstreams})
}

func (websiteServer *API) removeUpstreams(c *gin.Context) {
	upstreamToDelete := c.Query("upstream")

	if upstreamToDelete == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'upstream' query parameter"})
		return
	}

	var updatedUpstreams []string
	for _, upstream := range websiteServer.dnsServer.Config.UpstreamDNS {
		if upstream != upstreamToDelete {
			updatedUpstreams = append(updatedUpstreams, upstream)
		}
	}

	websiteServer.dnsServer.Config.UpstreamDNS = updatedUpstreams

	c.JSON(http.StatusOK, gin.H{
		"message": "Upstream removed successfully",
	})
}

func (websiteServer *API) clearLogs(c *gin.Context) {
	result, err := websiteServer.dnsServer.DB.Exec("DELETE FROM request_log")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not clear logs", "reason": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()

	websiteServer.dnsServer.Counters = server.CounterDetails{}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Cleared %d logs", rowsAffected),
	})
}

func (websiteServer *API) setPreferredUpstream(c *gin.Context) {
	upstreamToSet := c.DefaultQuery("upstream", "")

	if upstreamToSet == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Upstream is required"})
		return
	}

	var found bool
	for _, upstream := range websiteServer.dnsServer.Config.UpstreamDNS {
		if upstream == upstreamToSet {
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Upstream not found"})
		return
	}

	websiteServer.dnsServer.Config.PreferredUpstream = upstreamToSet
	updatedMsg := fmt.Sprintf("Preferred upstream set to %s", websiteServer.dnsServer.Config.PreferredUpstream)
	log.Info("%s", updatedMsg)
	settings.SaveSettings(&websiteServer.dnsServer.Config)
	c.JSON(http.StatusOK, gin.H{"message": updatedMsg})
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
