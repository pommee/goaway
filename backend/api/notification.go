package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) registerNotificationRoutes() {
	api.routes.GET("/notifications", api.fetchNotifications)

	api.routes.DELETE("/notification", api.markNotificationAsRead)
}

func (api *API) fetchNotifications(c *gin.Context) {
	notifications, err := api.Notifications.ReadNotifications()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, notifications)
}

func (api *API) markNotificationAsRead(c *gin.Context) {
	type NotificationsRead struct {
		NotificationIDs []int `json:"notificationIds"`
	}

	notificationsRead, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var request NotificationsRead
	if err := json.Unmarshal(notificationsRead, &request); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	err = api.Notifications.MarkNotificationsAsRead(request.NotificationIDs)
	if err != nil {
		log.Warning("Unable to mark notifications as read %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Unable to mark notifications as read %v", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
