package api

import (
	"encoding/json"
	"fmt"
	"goaway/backend/dns/lists"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (api *API) registerListsRoutes() {
	api.routes.POST("/custom", api.updateCustom)

	api.routes.GET("/lists", api.getLists)
	api.routes.GET("/addList", api.addList)
	api.routes.GET("/fetchUpdatedList", api.fetchUpdatedList)
	api.routes.GET("/runUpdateList", api.runUpdateList)
	api.routes.GET("/toggleBlocklist", api.toggleBlocklist)
	api.routes.GET("/updateBlockStatus", api.handleUpdateBlockStatus)

	api.routes.DELETE("/list", api.removeList)
}

func (api *API) updateCustom(c *gin.Context) {
	type UpdateListRequest struct {
		Domains []string `json:"domains"`
	}

	updatedList, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var request UpdateListRequest
	if err := json.Unmarshal(updatedList, &request); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	err = api.Blacklist.AddCustomDomains(request.Domains)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update custom blocklist."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"blockedLen": len(request.Domains)})
}

func (api *API) getLists(c *gin.Context) {
	lists, err := api.Blacklist.GetSourceStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"lists": lists})
}

func (api *API) addList(c *gin.Context) {
	name := c.Query("name")
	url := c.Query("url")

	if api.Blacklist.BlocklistURL[name] != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List already exists"})
		return
	}

	err := api.Blacklist.FetchAndLoadHosts(url, name)
	if err != nil {
		log.Error("%v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	blockedDomains, err := api.Blacklist.PopulateBlocklistCache()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	api.Blacklist.BlocklistURL[name] = url

	c.JSON(http.StatusOK, gin.H{"list": lists.SourceStats{
		URL:          url,
		BlockedCount: blockedDomains,
		LastUpdated:  time.Now().Unix(),
		Active:       false,
	}})
}

func (api *API) fetchUpdatedList(c *gin.Context) {
	name := c.Query("name")
	url := c.Query("url")

	if api.Blacklist.BlocklistURL[name] == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List does not exist"})
		return
	}

	if name == "" || url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'name' or 'url' query parameter"})
		return
	}

	remoteDomains, remoteChecksum, err := api.Blacklist.FetchRemoteHostsList(url)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dbDomains, dbChecksum, err := api.Blacklist.FetchDBHostsList(name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if remoteChecksum == dbChecksum {
		c.JSON(http.StatusOK, gin.H{"updateAvailable": false, "message": "No list updates available"})
		return
	}

	diff := func(a, b []string) []string {
		mb := make(map[string]struct{}, len(b))
		for _, x := range b {
			mb[x] = struct{}{}
		}
		diff := make([]string, 0)
		for _, x := range a {
			if _, found := mb[x]; !found {
				diff = append(diff, x)
			}
		}
		return diff
	}

	c.JSON(http.StatusOK, gin.H{
		"updateAvailable": true,
		"remoteChecksum":  remoteChecksum,
		"dbChecksum":      dbChecksum,
		"diffAdded":       diff(remoteDomains, dbDomains),
		"diffRemoved":     diff(dbDomains, remoteDomains),
	})
}

func (api *API) runUpdateList(c *gin.Context) {
	name := c.Query("name")
	url := c.Query("url")

	if api.Blacklist.BlocklistURL[name] == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List does not exist"})
		return
	}

	if name == "" || url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'name' or 'url' query parameter"})
		return
	}

	err := api.Blacklist.RemoveSourceAndDomains(name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = api.Blacklist.FetchAndLoadHosts(url, name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = api.Blacklist.PopulateBlocklistCache()
	if err != nil {
		message := fmt.Sprintf("Unable to re-populate the blocklist cache: %v", err)
		log.Warning("%s", message)
		c.JSON(http.StatusBadGateway, gin.H{"error": message})
	}

	c.Status(http.StatusOK)
}

func (api *API) toggleBlocklist(c *gin.Context) {
	blocklist := c.Query("blocklist")

	if blocklist == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid blocklist"})
		return
	}

	err := api.Blacklist.ToggleBlocklistStatus(blocklist)
	if err != nil {
		log.Error("Failed to toggle blocklist status: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to toggle status for %s", blocklist)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Toggled status for %s", blocklist)})
}

func (api *API) handleUpdateBlockStatus(c *gin.Context) {
	domain := c.Query("domain")
	blocked := c.Query("blocked")
	if domain == "" || blocked == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing query parameters"})
		return
	}

	action := map[string]func(string) error{
		"true":  api.Blacklist.AddBlacklistedDomain,
		"false": api.Blacklist.RemoveDomain,
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

func (api *API) removeList(c *gin.Context) {
	name := c.Query("name")

	if api.Blacklist.BlocklistURL[name] == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List does not exist"})
		return
	}

	err := api.Blacklist.RemoveSourceAndDomains(name)
	if err != nil {
		log.Error("%v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	delete(api.Blacklist.BlocklistURL, name)
	c.Status(http.StatusOK)
}
