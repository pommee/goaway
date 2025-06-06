package api

import (
	"database/sql"
	"errors"
	"goaway/backend/dns/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) registerClientRoutes() {
	api.routes.GET("/clients", api.getClients)
	api.routes.GET("/clientDetails", api.getClientDetails)
	api.routes.GET("/topClients", api.getTopClients)
}

func (api *API) getClients(c *gin.Context) {
	uniqueClients, err := database.FetchAllClients(api.DBManager.Conn)
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
	clientRequestDetails, err := database.GetClientRequestDetails(api.DBManager.Conn, clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mostQueriedDomain, err := database.GetMostQueriedDomainByIP(api.DBManager.Conn, clientIP)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	queriedDomains, err := database.GetAllQueriedDomainsByIP(api.DBManager.Conn, clientIP)
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

func (api *API) getTopClients(c *gin.Context) {
	topClients, err := database.GetTopClients(api.DBManager.Conn)
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"clients": topClients})
}
