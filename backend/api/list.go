package api

import (
	"context"
	"encoding/json"
	"fmt"
	"goaway/backend/alert"
	"goaway/backend/audit"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func (api *API) registerListsRoutes() {
	api.routes.POST("/custom", api.updateCustom)
	api.routes.POST("/addList", api.addList)
	api.routes.POST("/addLists", api.addLists)

	api.routes.GET("/lists", api.getLists)
	api.routes.GET("/fetchUpdatedList", api.fetchUpdatedList)
	api.routes.GET("/runUpdateList", api.runUpdateList)
	api.routes.GET("/toggleBlocklist", api.toggleBlocklist)
	api.routes.GET("/updateBlockStatus", api.handleUpdateBlockStatus)

	api.routes.PATCH("/listName", api.updateListName)

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

	api.DNSServer.Audits.CreateAudit(&audit.Entry{
		Topic:   audit.TopicList,
		Message: fmt.Sprintf("Added %d domains to custom blacklist", len(request.Domains)),
	})
	c.Status(http.StatusOK)
}

func (api *API) getLists(c *gin.Context) {
	lists, err := api.Blacklist.GetAllListStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lists)
}

func (api *API) addList(c *gin.Context) {
	type NewListRequest struct {
		Name   string `json:"name"`
		URL    string `json:"url"`
		Active bool   `json:"active"`
	}

	var newList NewListRequest
	err := c.Bind(&newList)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	err = api.ValidateURLAndName(newList.URL, newList.Name, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err = api.Blacklist.FetchAndLoadHosts(newList.URL, newList.Name); err != nil {
		log.Error("Failed to fetch and load hosts: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := api.Blacklist.PopulateBlocklistCache(); err != nil {
		log.Error("Failed to populate blocklist cache: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	if err := api.Blacklist.AddSource(newList.Name, newList.URL); err != nil {
		log.Error("Failed to add source: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !newList.Active {
		if err := api.Blacklist.ToggleBlocklistStatus(newList.Name); err != nil {
			log.Error("Failed to toggle blocklist status: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to toggle status for " + newList.Name})
			return
		}
	}

	_, addedList, err := api.Blacklist.GetListStatistics(newList.Name)
	if err != nil {
		log.Error("Failed to get list statistics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get list statistics"})
		return
	}

	api.DNSServer.Audits.CreateAudit(&audit.Entry{
		Topic:   audit.TopicList,
		Message: fmt.Sprintf("New blacklist with name '%s' was added", addedList.Name),
	})

	c.JSON(http.StatusOK, addedList)
}

func (api *API) addLists(c *gin.Context) {
	type NewList struct {
		Name   string `json:"name" binding:"required"`
		URL    string `json:"url" binding:"required,url"`
		Active bool   `json:"active"`
	}
	var payload struct {
		Lists []NewList `json:"lists" binding:"required,dive"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var addedList []NewList
	var ignoredList []NewList
	for _, list := range payload.Lists {
		if api.Blacklist.URLExists(list.URL) {
			ignoredList = append(ignoredList, list)
			continue
		}

		if err := api.Blacklist.FetchAndLoadHosts(list.URL, list.Name); err != nil {
			log.Error("Failed to fetch and load hosts: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if _, err := api.Blacklist.PopulateBlocklistCache(); err != nil {
			log.Error("Failed to populate blocklist cache: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		if err := api.Blacklist.AddSource(list.Name, list.URL); err != nil {
			log.Error("Failed to add source: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !list.Active {
			if err := api.Blacklist.ToggleBlocklistStatus(list.Name); err != nil {
				log.Error("Failed to toggle blocklist status: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to toggle status for " + list.Name})
				return
			}
		}

		addedList = append(addedList, list)
	}

	if len(addedList) > 0 {
		api.DNSServer.Audits.CreateAudit(&audit.Entry{
			Topic:   audit.TopicList,
			Message: fmt.Sprintf("Added %d new blacklists in bulk", len(addedList)),
		})
	}

	c.JSON(http.StatusOK, gin.H{"ignored": ignoredList})
}

func (api *API) updateListName(c *gin.Context) {
	oldName := c.Query("old")
	newName := c.Query("new")
	url := c.Query("url")

	if oldName == "" || newName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New and old names are required"})
		return
	}

	if !api.Blacklist.NameExists(oldName, url) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List with that name and url combination does not exist"})
		return
	}

	err := api.Blacklist.UpdateSourceName(oldName, newName, url)
	if err != nil {
		log.Warning("%s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (api *API) fetchUpdatedList(c *gin.Context) {
	name := c.Query("name")
	url := c.Query("url")

	if !api.Blacklist.NameExists(name, url) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List with that name and url combination does not exist"})
		return
	}

	if name == "" || url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'name' or 'url' query parameter"})
		return
	}

	availableUpdate, err := api.Blacklist.CheckIfUpdateAvailable(url, name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if availableUpdate.RemoteChecksum == availableUpdate.DBChecksum {
		c.JSON(http.StatusOK, gin.H{"updateAvailable": false, "message": "No list updates available"})
		return
	}

	c.JSON(http.StatusOK, availableUpdate)
}

func (api *API) runUpdateList(c *gin.Context) {
	name := c.Query("name")
	url := c.Query("url")

	if !api.Blacklist.NameExists(name, url) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List does not exist"})
		return
	}

	if name == "" || url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'name' or 'url' query parameter"})
		return
	}

	err := api.Blacklist.RemoveSourceAndDomains(name, url)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = api.Blacklist.FetchAndLoadHosts(url, name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	go func() {
		_ = api.DNSServer.Alerts.SendToAll(context.Background(), alert.Message{
			Title:    "System",
			Content:  fmt.Sprintf("List '%s' with url '%s' was updated! ", name, url),
			Severity: SeveritySuccess,
		})
	}()

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
	url := c.Query("url")

	if !api.Blacklist.NameExists(name, url) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List does not exist"})
		return
	}

	err := api.Blacklist.RemoveSourceAndDomains(name, url)
	if err != nil {
		log.Error("%v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	if removed := api.Blacklist.RemoveSourceByNameAndURL(name, url); !removed {
		log.Error("Failed to remove source with name '%s' and url '%s'", name, url)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to remove the list"})
		return
	}

	api.DNSServer.Audits.CreateAudit(&audit.Entry{
		Topic:   audit.TopicList,
		Message: fmt.Sprintf("Blacklist with name '%s' was deleted", name),
	})
	c.Status(http.StatusOK)
}

func (api *API) ValidateURLAndName(URL, name string, c *gin.Context) error {
	if name == "" || URL == "" {
		return fmt.Errorf("name and URL are required")
	}

	if _, err := url.ParseRequestURI(URL); err != nil {
		return fmt.Errorf("invalid URL format")
	}

	if api.Blacklist.URLExists(URL) {
		return fmt.Errorf("list with the same URL already exists")
	}

	return nil
}
