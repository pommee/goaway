package api

import (
	"fmt"
	"goaway/backend/dns/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) registerResolutionRoutes() {
	api.routes.POST("/resolution", api.createResolution)

	api.routes.GET("/resolutions", api.getResolutions)

	api.routes.DELETE("/resolution", api.deleteResolution)
}

func (api *API) createResolution(c *gin.Context) {
	type NewResolution struct {
		IP     string
		Domain string
	}

	var newResolution NewResolution
	if err := c.BindJSON(&newResolution); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid resolution data",
		})
		return
	}

	err := database.CreateNewResolution(api.DBManager.Conn, newResolution.IP, newResolution.Domain)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	api.DNSServer.RemoveCachedDomain(newResolution.Domain)
	c.Status(http.StatusOK)
}

func (api *API) getResolutions(c *gin.Context) {
	resolutions, err := database.FetchResolutions(api.DBManager.Conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"resolutions": resolutions})
}

func (api *API) deleteResolution(c *gin.Context) {
	domain := c.Query("domain")
	ip := c.Query("ip")

	rowsAffected, err := database.DeleteResolution(api.DBManager.Conn, ip, domain)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s does not exist", domain)})
		return
	}

	api.DNSServer.RemoveCachedDomain(domain)
	c.JSON(http.StatusOK, gin.H{"deleted": rowsAffected})
}
