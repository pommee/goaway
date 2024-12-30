package website

import (
	"embed"
	"fmt"
	"goaway/internal/server"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type API struct {
	router    *gin.Engine
	dnsServer *server.DNSServer
	port      int
}

func (websiteServer *API) create() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	websiteServer.router = router
}

func (websiteServer *API) Start(content embed.FS, dnsServer *server.DNSServer, port int) {
	websiteServer.create()
	websiteServer.dnsServer = dnsServer
	websiteServer.port = port
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		websiteServer.router.Run(fmt.Sprintf(":%d", port))
	}()

	websiteServer.serve()
	websiteServer.serveWebsite(content)

	wg.Wait()
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

func (websiteServer *API) serve() {
	websiteServer.router.GET("/server", websiteServer.handleServer)
	websiteServer.router.GET("/metrics", websiteServer.handleMetrics)
	websiteServer.router.GET("/queriesData", websiteServer.handleQueriesData)
	websiteServer.router.GET("/updateBlockStatus", websiteServer.handleUpdateBlockStatus)
	websiteServer.router.GET("/domains", websiteServer.getDomains)
}

func (websiteServer *API) handleServer(c *gin.Context) {
	cpuUsage, err := cpu.Percent(0, false)
	if err != nil {
		log.Fatal(err)
	}

	temp, err := getCPUTemperature()
	if err != nil {
		log.Fatal(err)
	}

	vMem, err := mem.VirtualMemory()
	if err != nil {
		log.Fatal(err)
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
		log.Println(err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"domain": domain, "blocked": blocked})
}

func (websiteServer *API) getDomains(c *gin.Context) {
	domains := make([]string, 0, len(websiteServer.dnsServer.Blacklist.Domains))
	for domain := range websiteServer.dnsServer.Blacklist.Domains {
		domains = append(domains, domain)
	}
	c.JSON(http.StatusOK, gin.H{"domains": domains})
}

func (websiteServer *API) serveWebsite(content embed.FS) {
	// Create a map to associate file extensions with their MIME types
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

		ext := strings.ToLower(path[strings.LastIndex(path, "."):])
		mimeType, ok := mimeTypes[ext]
		if !ok {
			// Default MIME type if not found
			mimeType = "application/octet-stream"
		}

		// Generate the route based on the file path
		route := strings.TrimPrefix(path, "website")
		websiteServer.router.GET(route, func(c *gin.Context) {
			// Set the server IP address header
			c.Header("X-Server-IP", ipAddress)

			// If the file is an HTML file, add a script to store the IP address in localStorage
			if ext == ".html" {
				script := `
					<script>
						window.onload = function() {
							// Check if the IP is not already stored
							if (!localStorage.getItem("serverIP")) {
								var serverIP = document.location.origin;
								localStorage.setItem("serverIP", serverIP);
							}
						}
					</script>
				`
				fileContent = append(fileContent, []byte(script)...)
			}

			c.Data(200, mimeType, fileContent)
		})

		return nil
	})
	if err != nil {
		fmt.Println("Error embedding files:", err)
	}
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
