package api

import (
	"fmt"
	"goaway/backend/dns/database"
	"goaway/backend/dns/server"
	"goaway/backend/updater"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func (api *API) registerServerRoutes() {
	api.setupWSLiveCommunication(api.PrefetchedDomainsManager.DNS)

	api.router.GET("/api/server", api.handleServer)
	api.router.GET("/api/dnsMetrics", api.handleMetrics)
	api.routes.GET("/runUpdate", api.runUpdate)
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

	dbSize, err := getDBSizeMB()
	if err != nil {
		log.Error("%s", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"portDNS":           api.Config.DNS.Port,
		"portWebsite":       api.DNSPort,
		"totalMem":          float64(vMem.Total) / 1024 / 1024 / 1024,
		"usedMem":           float64(vMem.Used) / 1024 / 1024 / 1024,
		"usedMemPercentage": float64(vMem.Used) / 1024 / 1024 / 1024,
		"cpuUsage":          cpuUsage[0],
		"cpuTemp":           temp,
		"dbSize":            dbSize,
		"version":           api.Version,
		"inAppUpdate":       api.Config.InAppUpdate,
		"commit":            api.Commit,
		"date":              api.Date,
	})
}

func getCPUTemperature() (float64, error) {
	tempFile := "/sys/class/thermal/thermal_zone0/temp"

	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		return 0, nil // Temperature file does not exist, return 0
	}

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

func getDBSizeMB() (float64, error) {
	var totalSize int64

	basePath := "data"
	files := []string{
		filepath.Join(basePath, "database.db"),
		filepath.Join(basePath, "database.db-wal"),
		filepath.Join(basePath, "database.db-shm"),
	}

	for _, filename := range files {
		info, err := os.Stat(filename)
		if err != nil {
			// Only return error if the main DB file is missing.
			if filename == "database.db" {
				return 0, err
			}
			// WAL/SHM files may not exist temporarily — that's fine.
			continue
		}
		totalSize += info.Size()
	}

	return float64(totalSize) / (1024 * 1024), nil
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
		"clients":           database.GetDistinctRequestIP(api.DBManager.Conn),
	})
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

	sendSSE("[info] Starting update process...")
	err := updater.SelfUpdate(sendSSE, api.Config.BinaryPath)
	if err != nil {
		sendSSE(fmt.Sprintf("[error] %s", err.Error()))
		c.Status(http.StatusBadRequest)
	} else {
		sendSSE("[info] Update successful!")
		c.Status(http.StatusOK)
	}
}

func (api *API) setupWSLiveCommunication(dnsServer *server.DNSServer) {
	api.router.GET("/api/liveCommunication", func(c *gin.Context) {
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		api.WSCommunication = conn

		if dnsServer != nil {
			dnsServer.WSCommunication = conn
		}
	})
}
