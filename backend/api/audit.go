package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) registerAuditRoutes() {
	api.routes.GET("/audit", api.getAudits)
}

func (api *API) getAudits(c *gin.Context) {
	audits, err := api.DNSServer.AuditService.ReadAudits()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, audits)
}
