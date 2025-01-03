package website

import (
	"crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"goaway/internal/logger"
	"goaway/internal/server"
	"goaway/internal/settings"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ping/ping"
	"github.com/golang-jwt/jwt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

var log = logger.GetLogger()

const (
	tokenDuration = 5 * time.Minute
	jwtSecret     = "the-secret-key"
)

type Credentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type API struct {
	router                *gin.Engine
	dnsServer             *server.DNSServer
	port                  int
	DisableAuthentication bool
	adminPassword         string
}

func (api *API) Start(content embed.FS, dnsServer *server.DNSServer, port int) {
	password, exists := os.LookupEnv("GOAWAY_PASSWORD")
	if !exists {
		api.adminPassword = generateRandomPassword(14)
		log.Info("Randomly generated admin password: %s", api.adminPassword)
	} else {
		api.adminPassword = password
		log.Info("Using custom password: [hidden]")
	}

	gin.SetMode(gin.ReleaseMode)
	api.router = gin.New()
	dnsServer.WebServer = api.router
	api.dnsServer = dnsServer
	api.port = port

	api.setupRoutes()
	api.serveEmbeddedContent(content)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		log.Info("Starting server on port %d", port)
		api.router.Run(fmt.Sprintf(":%d", port))
	}()

	wg.Wait()
}

func generateRandomPassword(length int) string {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Error("Error generating random bytes: %v", err)
	}

	password := base64.StdEncoding.EncodeToString(randomBytes)
	return password
}

func (api *API) setupRoutes() {
	api.router.POST("/login", api.handleLogin)
	api.router.GET("/server", api.handleServer)
	api.router.GET("/authentication", api.getAuthentication)

	authorized := api.router.Group("/")
	if !api.DisableAuthentication {
		authorized.Use(authMiddleware())
	} else {
		log.Info("Authentication is disabled.")
	}

	authorized.GET("/metrics", api.handleMetrics)
	authorized.GET("/queriesData", api.handleQueriesData)
	authorized.GET("/updateBlockStatus", api.handleUpdateBlockStatus)
	authorized.GET("/domains", api.getDomains)
	authorized.GET("/settings", api.getSettings)
	authorized.POST("/settings", api.updateSettings)
	authorized.GET("/clients", api.getClients)
	authorized.GET("/upstreams", api.getUpstreams)
	authorized.POST("/upstreams", api.createUpstreams)
	authorized.DELETE("/upstreams", api.removeUpstreams)
	authorized.GET("/preferredUpstream", api.setPreferredUpstream)
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

func generateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(tokenDuration).Unix(),
		"iat":      time.Now().Unix(),
	})
	return token.SignedString([]byte(jwtSecret))
}

func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(tokenDuration),
	})
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/server") {
			c.Next()
			return
		}

		cookie, err := c.Cookie("jwt")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization cookie required"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(cookie, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			expirationTime := int64(claims["exp"].(float64))
			if time.Now().Unix() > expirationTime {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
				c.Abort()
				return
			}

			if time.Now().Unix() > expirationTime-int64(tokenDuration/2) {
				newToken, err := generateToken(claims["username"].(string))
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to renew token"})
					c.Abort()
					return
				}
				setAuthCookie(c.Writer, newToken)
			}

			c.Set("username", claims["username"])
		}

		c.Next()
	}
}

func (api *API) serveEmbeddedContent(content embed.FS) {
	mimeTypes := map[string]string{
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
	}

	ipAddress, err := getServerIP()
	if err != nil {
		fmt.Println("Error getting IP address:", err)
	}

	err = fs.WalkDir(content, "website", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		fileContent, err := content.ReadFile(path)
		if err != nil {
			return err
		}

		mimeType := mimeTypes[strings.ToLower(path[strings.LastIndex(path, "."):])]
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		route := strings.TrimPrefix(path, "website")
		api.router.GET(route, func(c *gin.Context) {
			c.Header("X-Server-IP", ipAddress)
			c.Data(http.StatusOK, mimeType, fileContent)
		})

		return nil
	})

	if err != nil {
		log.Error("Error serving embedded content: %v", err)
	}
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

	c.JSON(http.StatusOK, gin.H{
		"allowed":           allowedQueries,
		"blocked":           blockedQueries,
		"total":             totalQueries,
		"percentageBlocked": percentageBlocked,
		"domainBlockLen":    len(websiteServer.dnsServer.Blacklist.Domains),
	})
}

func (websiteServer *API) validateCredentials(username, password string) bool {
	return username == "admin" && password == websiteServer.adminPassword
}

func (websiteServer *API) handleQueriesData(c *gin.Context) {
	queriesData := websiteServer.dnsServer.RequestLog
	c.JSON(http.StatusOK, gin.H{
		"details": queriesData,
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
	domains := make([]string, 0, len(websiteServer.dnsServer.Blacklist.Domains))
	for domain := range websiteServer.dnsServer.Blacklist.Domains {
		domains = append(domains, domain)
	}
	c.JSON(http.StatusOK, gin.H{"domains": domains})
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

	for _, logEntry := range websiteServer.dnsServer.RequestLog {
		client := logEntry.ClientInfo
		if client == nil {
			continue
		}

		if existing, exists := uniqueClients[client.IP]; !exists || logEntry.Timestamp.After(existing.LastSeen) {
			uniqueClients[client.IP] = struct {
				Name     string
				LastSeen time.Time
			}{
				Name:     client.Name,
				LastSeen: logEntry.Timestamp,
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
