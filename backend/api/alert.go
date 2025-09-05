package api

import (
	"encoding/json"
	"goaway/backend/alert"
	"io"
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

	api.routes.GET("/alert", api.getAlert)
}

func (api *API) setAlert(c *gin.Context) {

	alerts, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var request AlertSettings
	if err := json.Unmarshal(alerts, &request); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	err = api.DNSServer.Alerts.SaveAlert(alert.Alert{
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
