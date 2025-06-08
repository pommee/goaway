package api

import (
	"database/sql"
	"encoding/json"
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

	err = api.Whitelist.AddDomain(newDomain.Domain)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (api *API) getWhitelistedDomains(c *gin.Context) {
	query := "SELECT domain FROM whitelist"
	rows, err := api.DBManager.Conn.Query(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var domains = make([]string, 0)
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		domains = append(domains, domain)
	}

	c.JSON(http.StatusOK, domains)
}

func (api *API) deleteWhitelistedDomain(c *gin.Context) {
	newDomain := c.Query("domain")

	err := api.Whitelist.RemoveDomain(newDomain)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
