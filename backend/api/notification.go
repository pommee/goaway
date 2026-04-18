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
	type PaginationParams struct {
		Page  int `form:"page,default=1"`
		Limit int `form:"limit,default=50"`
	}

	var params PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 50
	}

	result, err := api.NotificationService.GetNotifications(params.Page, params.Limit)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
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
