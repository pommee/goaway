package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) registerNotificationRoutes() {
	api.routes.GET("/notifications", api.fetchNotifications)

	api.routes.DELETE("/notification", api.markNotificationAsRead)
}

func (api *API) fetchNotifications(c *gin.Context) {
	notifications, err := api.NotificationService.GetNotifications()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, notifications)
}

func (api *API) markNotificationAsRead(c *gin.Context) {
	type NotificationsRead struct {
		NotificationIDs []int `json:"notificationIds"`
	}

	var request NotificationsRead
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to parse request"})
		return
	}

	err = api.NotificationService.MarkNotificationsAsRead(request.NotificationIDs)
	if err != nil {
		log.Warning("Unable to mark notifications as read %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Unable to mark notifications as read %v", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
