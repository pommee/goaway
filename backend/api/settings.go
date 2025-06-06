package api

import (
	"encoding/json"
	"fmt"
	"goaway/backend/settings"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func (api *API) registerSettingsRoutes() {
	api.routes.POST("/settings", api.updateSettings)

	api.routes.GET("/settings", api.getSettings)
	api.routes.GET("/exportDatabase", api.exportDatabase)
}

func (api *API) updateSettings(c *gin.Context) {
	var updatedSettings settings.Config
	if err := c.BindJSON(&updatedSettings); err != nil {
		log.Warning("Could not save new settings, reason: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid settings data",
		})
		return
	}

	api.Config.UpdateSettings(updatedSettings)
	log.Info("Updated settings!")
	settingsJson, _ := json.MarshalIndent(updatedSettings, "", "  ")
	log.Debug("%s", string(settingsJson))

	c.JSON(http.StatusOK, gin.H{
		"config": api.Config,
	})
}

func (api *API) getSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"settings": api.Config,
	})
}

func (api *API) exportDatabase(c *gin.Context) {
	log.Debug("Starting export of database")

	const databaseName = "database.db"
	if _, err := os.Stat(databaseName); err != nil {
		if os.IsNotExist(err) {
			log.Error("Database file not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "Database file not found"})
		} else {
			log.Error("Error accessing database file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	file, err := os.Open(databaseName)
	if err != nil {
		log.Error("Error opening database file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer func(tx *os.File) {
		_ = file.Close()
	}(file)

	fileInfo, err := file.Stat()
	if err != nil {
		log.Error("Error getting file info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+databaseName)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Header("Cache-Control", "no-cache")

	c.DataFromReader(http.StatusOK, fileInfo.Size(), "application/octet-stream", file, nil)
}
