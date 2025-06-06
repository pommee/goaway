package api

import (
	"encoding/json"
	"fmt"
	"goaway/backend/dns/database"
	"goaway/backend/dns/server/prefetch"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (api *API) registerBlacklistRoutes() {
	api.routes.POST("/prefetch", api.createPrefetchedDomain)
	api.routes.GET("/prefetch", api.fetchPrefetchedDomains)
	api.routes.GET("/removeFromCustom", api.removeDomainFromCustom)
	api.routes.GET("/domains", api.getDomains)
	api.routes.GET("/topBlockedDomains", api.getTopBlockedDomains)
	api.routes.GET("/getDomainsForList", api.getDomainsForList)
	api.routes.DELETE("/prefetch", api.deletePrefetchedDomain)
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

func (api *API) getTopBlockedDomains(c *gin.Context) {
	_, blocked, _ := api.Blacklist.GetAllowedAndBlocked()
	topBlockedDomains, err := database.GetTopBlockedDomains(api.DBManager.Conn, blocked)
	if err != nil {
		log.Error("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"domains": topBlockedDomains})
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
