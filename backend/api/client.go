package api

import (
	model "goaway/backend/dns/server/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) registerClientRoutes() {
	api.routes.GET("/clients", api.getClients)
	api.routes.GET("/topClients", api.getTopClients)

	api.routes.GET("/client/:ip/details", api.getClientDetails)
	api.routes.GET("/client/:ip/history", api.getClientHistory)

	api.routes.PUT("/client/:ip/bypass/:bypass", api.updateClientBypass)
}

func (api *API) getClients(c *gin.Context) {
	uniqueClients, err := api.RequestService.FetchAllClients()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clients := make([]model.Client, 0, len(uniqueClients))
	for _, entry := range uniqueClients {
		clients = append(clients, entry)
	}

	c.JSON(http.StatusOK, clients)
}

func (api *API) getClientDetails(c *gin.Context) {
	ip := c.Param("ip")

	requestDetails, mostQueriedDomain, domainQueryCount, err := api.RequestService.GetClientDetailsWithDomains(ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	clientDetails, err := api.RequestService.FetchClient(ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]any{
		"totalRequests":     requestDetails.TotalRequests,
		"uniqueDomains":     requestDetails.UniqueDomains,
		"blockedRequests":   requestDetails.BlockedRequests,
		"cachedRequests":    requestDetails.CachedRequests,
		"avgResponseTimeMs": requestDetails.AvgResponseTimeMs,
		"mostQueriedDomain": mostQueriedDomain,
		"allDomains":        domainQueryCount,
		"clientInfo":        clientDetails,
	})
}

func (api *API) getClientHistory(c *gin.Context) {
	ip := c.Param("ip")
	history, err := api.RequestService.GetClientHistory(ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]any{
		"history": history,
	})
}

func (api *API) getTopClients(c *gin.Context) {
	topClients, err := api.RequestService.GetTopClients()
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, topClients)
}

func (api *API) updateClientBypass(c *gin.Context) {
	ip := c.Param("ip")
	bypass := c.Param("bypass")

	if bypass != "true" && bypass != "false" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bypass value must be true or false"})
		return
	}

	err := api.RequestService.UpdateClientBypass(ip, bypass == "true")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Refresh DNS server client caches to reflect the updated bypass status
	err = api.DNS.PopulateClientCaches()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh DNS server client caches"})
		return
	}

	c.Status(http.StatusOK)
}
