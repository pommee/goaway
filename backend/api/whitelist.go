package api

import (
	"encoding/json"
	"goaway/backend/database"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) registerWhitelistRoutes() {
	api.routes.POST("/whitelist", api.addWhitelisted)
	api.routes.GET("/whitelist", api.getWhitelistedDomains)
	api.routes.DELETE("/whitelist", api.deleteWhitelistedDomain)
}

func (api *API) addWhitelisted(c *gin.Context) {
	type NewWhitelistedDomain struct {
		Domain string `json:"domain"`
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var newDomain NewWhitelistedDomain
	if err := json.Unmarshal(body, &newDomain); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	err = api.WhitelistService.AddDomain(newDomain.Domain)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (api *API) getWhitelistedDomains(c *gin.Context) {
	var domains []string

	err := api.DBConn.Model(&database.Whitelist{}).Pluck("domain", &domains).Error
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to retrieve whitelisted domains"})
		return
	}

	c.JSON(http.StatusOK, domains)
}

func (api *API) deleteWhitelistedDomain(c *gin.Context) {
	newDomain := c.Query("domain")

	err := api.WhitelistService.RemoveDomain(newDomain)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
