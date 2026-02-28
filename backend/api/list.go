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

	err = api.BlacklistService.AddCustomDomains(context.Background(), request.Domains)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	api.DNSServer.AuditService.CreateAudit(&audit.Entry{
		Topic:   audit.TopicList,
		Message: fmt.Sprintf("Added %d domains to custom blacklist", len(request.Domains)),
	})
	c.Status(http.StatusOK)
}

func (api *API) getLists(c *gin.Context) {
	lists, err := api.BlacklistService.GetAllListStatistics(context.Background())
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

	err = api.validateURLAndName(newList.URL, newList.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err = api.BlacklistService.FetchAndLoadHosts(context.Background(), newList.URL, newList.Name); err != nil {
		log.Error("Failed to fetch and load hosts: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := api.BlacklistService.PopulateCache(context.Background()); err != nil {
		log.Error("Failed to populate blocklist cache: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	if err := api.BlacklistService.AddSource(context.Background(), newList.Name, newList.URL); err != nil {
		log.Error("Failed to add source: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !newList.Active {
		if err := api.BlacklistService.ToggleBlocklistStatus(context.Background(), newList.Name); err != nil {
			log.Error("Failed to toggle blocklist status: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to toggle status for " + newList.Name})
			return
		}
	}

	_, addedList, err := api.BlacklistService.GetListStatistics(context.Background(), newList.Name)
	if err != nil {
		log.Error("Failed to get list statistics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get list statistics"})
		return
	}

	api.DNSServer.AuditService.CreateAudit(&audit.Entry{
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
		if api.BlacklistService.URLExists(list.URL) {
			ignoredList = append(ignoredList, list)
			continue
		}

		if err := api.BlacklistService.FetchAndLoadHosts(context.Background(), list.URL, list.Name); err != nil {
			log.Error("Failed to fetch and load hosts: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := api.BlacklistService.PopulateCache(context.Background()); err != nil {
			log.Error("Failed to populate blocklist cache: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		if err := api.BlacklistService.AddSource(context.Background(), list.Name, list.URL); err != nil {
			log.Error("Failed to add source: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !list.Active {
			if err := api.BlacklistService.ToggleBlocklistStatus(context.Background(), list.Name); err != nil {
				log.Error("Failed to toggle blocklist status: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to toggle status for " + list.Name})
				return
			}
		}

		addedList = append(addedList, list)
	}

	if len(addedList) > 0 {
		api.DNSServer.AuditService.CreateAudit(&audit.Entry{
			Topic:   audit.TopicList,
			Message: fmt.Sprintf("Added %d new blacklists in bulk", len(addedList)),
		})
	}

	c.JSON(http.StatusOK, gin.H{"ignored": ignoredList})
}

func (api *API) updateListName(c *gin.Context) {
	oldName := c.Query("old")
	newName := c.Query("new")
	listURL := c.Query("url")

	if oldName == "" || newName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New and old names are required"})
		return
	}

	if !api.BlacklistService.NameExists(oldName, listURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List with that name and url combination does not exist"})
		return
	}

	err := api.BlacklistService.UpdateSourceName(context.Background(), oldName, newName, listURL)
	if err != nil {
		log.Warning("%s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (api *API) fetchUpdatedList(c *gin.Context) {
	name := c.Query("name")
	listURL := c.Query("url")

	if !api.BlacklistService.NameExists(name, listURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List with that name and url combination does not exist"})
		return
	}

	if name == "" || listURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'name' or 'url' query parameter"})
		return
	}

	availableUpdate, err := api.BlacklistService.CheckIfUpdateAvailable(context.Background(), listURL, name)
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
	listURL := c.Query("url")

	if !api.BlacklistService.NameExists(name, listURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List does not exist"})
		return
	}

	if name == "" || listURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'name' or 'url' query parameter"})
		return
	}

	err := api.BlacklistService.RemoveSourceAndDomains(context.Background(), name, listURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = api.BlacklistService.FetchAndLoadHosts(context.Background(), listURL, name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	go func() {
		_ = api.DNSServer.AlertService.SendToAll(context.Background(), alert.Message{
			Title:    "System",
			Content:  fmt.Sprintf("List '%s' with url '%s' was updated! ", name, listURL),
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

	err := api.BlacklistService.ToggleBlocklistStatus(context.Background(), blocklist)
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

	action := map[string]func(context.Context, string) error{
		"true":  api.BlacklistService.AddBlacklistedDomain,
		"false": api.BlacklistService.RemoveDomain,
	}[blocked]

	if action == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid value for blocked"})
		return
	}

	if err := action(context.Background(), domain); err != nil {
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
	listURL := c.Query("url")

	if !api.BlacklistService.NameExists(name, listURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List does not exist"})
		return
	}

	err := api.BlacklistService.RemoveSourceAndDomains(context.Background(), name, listURL)
	if err != nil {
		log.Error("%v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	if removed := api.BlacklistService.RemoveSourceByNameAndURL(name, listURL); !removed {
		log.Error("Failed to remove source with name '%s' and url '%s'", name, listURL)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to remove the list"})
		return
	}

	api.DNSServer.AuditService.CreateAudit(&audit.Entry{
		Topic:   audit.TopicList,
		Message: fmt.Sprintf("Blacklist with name '%s' was deleted", name),
	})
	c.Status(http.StatusOK)
}

func (api *API) validateURLAndName(listURL, name string) error {
	if name == "" || listURL == "" {
		return fmt.Errorf("name and URL are required")
	}

	if _, err := url.ParseRequestURI(listURL); err != nil {
		return fmt.Errorf("invalid URL format")
	}

	if api.BlacklistService.URLExists(listURL) {
		return fmt.Errorf("list with the same URL already exists")
	}

	return nil
}
