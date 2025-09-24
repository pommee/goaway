package api

import (
	"context"
	"goaway/backend/dns/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	SeveritySuccess = "success"
	SeverityInfo    = "info"
	SeverityWarning = "warning"
	SeverityError   = "error"
)

type DiscordSettings struct {
	Enabled bool   `json:"enabled"`
	Name    string `json:"name"`
	Webhook string `json:"webhook"`
}

type AlertSettings struct {
	Discord DiscordSettings `json:"discord"`
}

func (api *API) registerAlertRoutes() {
	api.routes.POST("/alert", api.setAlert)
	api.routes.POST("/alert/test", api.testAlert)

	api.routes.GET("/alert", api.getAlert)
}

func (api *API) setAlert(c *gin.Context) {
	var request AlertSettings
	err := c.Bind(&request)
	if err != nil {
		log.Error("Failed to parse alert settings: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	err = api.DNSServer.Alerts.SaveAlert(database.Alert{
		Type:    "discord",
		Enabled: request.Discord.Enabled,
		Name:    request.Discord.Name,
		Webhook: request.Discord.Webhook,
	})
	if err != nil {
		log.Error("Failed to save alert settings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save alert settings"})
		return
	}

	c.Status(http.StatusOK)
}

func (api *API) getAlert(c *gin.Context) {
	alerts, err := api.DNSServer.Alerts.GetAllAlerts()
	if err != nil {
		log.Error("Failed to retrieve alert settings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve alert settings"})
		return
	}

	var response AlertSettings
	for _, alert := range alerts {
		if alert.Type == "discord" {
			response.Discord = DiscordSettings{
				Enabled: alert.Enabled,
				Name:    alert.Name,
				Webhook: alert.Webhook,
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

func (api *API) testAlert(c *gin.Context) {
	var request AlertSettings
	err := c.Bind(&request)
	if err != nil {
		log.Error("Failed to parse alert settings: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	err = api.DNSServer.Alerts.SendTest(
		context.Background(),
		"discord",
		request.Discord.Name,
		request.Discord.Webhook,
	)
	if err != nil {
		log.Error("Failed to save alert settings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save alert settings"})
		return
	}

	c.Status(http.StatusOK)
}
