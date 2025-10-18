package api

import (
	"fmt"
	"goaway/backend/audit"
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

	err := api.ResolutionService.CreateResolution(newResolution.IP, newResolution.Domain)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	api.DNSServer.RemoveCachedDomain(newResolution.Domain)

	api.DNSServer.AuditService.CreateAudit(&audit.Entry{
		Topic:   audit.TopicResolution,
		Message: fmt.Sprintf("Added new resolution '%s'", newResolution.Domain),
	})
	c.Status(http.StatusOK)
}

func (api *API) getResolutions(c *gin.Context) {
	resolutions, err := api.ResolutionService.GetResolutions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resolutions)
}

func (api *API) deleteResolution(c *gin.Context) {
	domain := c.Query("domain")
	ip := c.Query("ip")

	rowsAffected, err := api.ResolutionService.DeleteResolution(ip, domain)
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

	api.DNSServer.AuditService.CreateAudit(&audit.Entry{
		Topic:   audit.TopicResolution,
		Message: fmt.Sprintf("Removed resolution '%s'", domain),
	})
	c.Status(http.StatusOK)
}
