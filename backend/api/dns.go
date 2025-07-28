package api

import (
	"goaway/backend/api/models"
	"goaway/backend/audit"
	"goaway/backend/dns/database"
	"goaway/backend/dns/server"
	model "goaway/backend/dns/server/models"
	"goaway/backend/settings"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func (api *API) registerDNSRoutes() {
	api.setupWSLiveQueries(api.PrefetchedDomainsManager.DNS)

	api.routes.POST("/pause", api.pauseBlocking)
	api.routes.GET("/pause", api.getBlocking)
	api.routes.GET("/queries", api.getQueries)
	api.routes.GET("/queryTimestamps", api.getQueryTimestamps)
	api.routes.GET("/responseSizeTimestamps", api.getResponseSizeTimestamps)
	api.routes.GET("/queryTypes", api.getQueryTypes)

	api.routes.DELETE("/queries", api.clearQueries)
	api.routes.DELETE("/pause", api.clearBlocking)
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

	log.Info("DNS blocking paused for %d seconds", blockTime.Time)
	c.Status(http.StatusOK)
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

func (api *API) getQueries(c *gin.Context) {
	query := parseQueryParams(c)

	type result struct {
		queries []model.RequestLogEntry
		total   int
		err     error
	}

	queryCh := make(chan result, 1)
	countCh := make(chan result, 1)

	go func() {
		queries, err := database.FetchQueries(api.DBManager.Conn, query)
		queryCh <- result{queries: queries, err: err}
	}()

	go func() {
		total, err := database.CountQueries(api.DBManager.Conn, query.Search)
		countCh <- result{total: total, err: err}
	}()

	queryResult := <-queryCh
	countResult := <-countCh

	if queryResult.err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": queryResult.err.Error()})
		return
	}

	if countResult.err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": countResult.err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"draw":            c.DefaultQuery("draw", "1"),
		"recordsTotal":    countResult.total,
		"recordsFiltered": countResult.total,
		"queries":         queryResult.queries,
	})
}

func parseQueryParams(c *gin.Context) models.QueryParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.DefaultQuery("search", "")
	sortColumn := c.DefaultQuery("sortColumn", "timestamp")
	sortDirection := c.DefaultQuery("sortDirection", "desc")

	validColumns := map[string]string{
		"timestamp":         "timestamp",
		"domain":            "domain",
		"client":            "client_ip",
		"ip":                "ip",
		"status":            "status",
		"responseTimeNS":    "response_time_ns",
		"queryType":         "query_type",
		"blocked":           "blocked",
		"cached":            "cached",
		"responseSizeBytes": "response_size_bytes",
	}

	column, ok := validColumns[sortColumn]
	if !ok {
		column = "timestamp"
	}

	sortDirection = strings.ToLower(sortDirection)
	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "desc"
	}

	return models.QueryParams{
		Page:      page,
		PageSize:  pageSize,
		Search:    search,
		Column:    column,
		Direction: strings.ToUpper(sortDirection),
		Offset:    (page - 1) * pageSize,
	}
}

func (api *API) getQueryTimestamps(c *gin.Context) {
	intervalParam := c.Query("interval")
	if intervalParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interval parameter is required"})
		return
	}

	interval, err := strconv.Atoi(intervalParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid interval parameter"})
		return
	}

	timestamps, err := database.GetRequestSummaryByInterval(interval, api.DBManager.Conn)
	if err != nil {
		log.Error("Failed to get request summary: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, timestamps)
}

func (api *API) getResponseSizeTimestamps(c *gin.Context) {
	intervalParam := c.Query("interval")
	interval, err := strconv.Atoi(intervalParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interval parameter"})
		return
	}

	timestamps, err := database.GetResponseSizeSummaryByInterval(interval, api.DBManager.Conn)
	if err != nil {
		log.Error("Error fetching response size timestamps: %v", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, timestamps)
}

func (api *API) getQueryTypes(c *gin.Context) {
	queries, err := database.GetUniqueQueryTypes(api.DBManager.Conn)
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, queries)
}

func (api *API) clearQueries(c *gin.Context) {
	_, err := api.DBManager.Conn.Exec("DELETE FROM request_log")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not clear logs", "reason": err.Error()})
		return
	}

	_, err = api.DBManager.Conn.Exec("DELETE FROM request_log_ips")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not clear logs", "reason": err.Error()})
		return
	}

	api.Blacklist.Vacuum()

	api.DNSServer.Audits.CreateAudit(&audit.Entry{
		Topic:   audit.TopicLogs,
		Message: "Logs were cleared",
	})

	c.JSON(http.StatusOK, gin.H{"message": "Cleared all logs"})
}

func (api *API) clearBlocking(c *gin.Context) {
	api.Config.DNS.Status = settings.Status{}
	c.JSON(http.StatusOK, gin.H{})
}

func (api *API) setupWSLiveQueries(dnsServer *server.DNSServer) {
	api.router.GET("/api/liveQueries", func(c *gin.Context) {
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		api.WSQueries = conn

		if dnsServer != nil {
			dnsServer.WSQueries = conn
		}
	})
}
