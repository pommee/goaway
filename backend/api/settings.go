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

	// Temporary filename for export the database into
	tempExport := "export_temp.db"

	// remove in case it already exists, otherwise VACUUM INTO will fail
	_ = os.Remove(tempExport)

	// Create a new connection to a temp file and vacuum into it
	_, err := api.DBManager.Conn.Exec(fmt.Sprintf("VACUUM INTO '%s';", tempExport))
	if err != nil {
		log.Error("Failed to write WAL to temp export: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare database for export"})
		return
	}

	file, err := os.Open(tempExport)
	if err != nil {
		log.Error("Error opening database export file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	defer func(tx *os.File) {
		_ = file.Close()

		// remove the temporary export file after sending it
		_ = os.Remove(tempExport)
	}(file)

	fileInfo, err := file.Stat()
	if err != nil {
		log.Error("Error getting file info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename=database.db")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Header("Cache-Control", "no-cache")

	c.DataFromReader(http.StatusOK, fileInfo.Size(), "application/octet-stream", file, nil)
}
