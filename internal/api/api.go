package api

import (
	"crypto/rand"
	"embed"
	"encoding/base64"
	"fmt"
	"goaway/internal/logging"
	"goaway/internal/server"
	"goaway/internal/settings"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var log = logging.GetLogger()

const (
	tokenDuration = 5 * time.Minute
	jwtSecret     = "the-secret-key"
)

type Credentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type API struct {
	router        *gin.Engine
	DnsServer     *server.DNSServer
	Config        *settings.APIServerConfig
	adminPassword string
}

func (api *API) Start(content embed.FS, dnsServer *server.DNSServer, errorChannel chan struct{}) {
	gin.SetMode(gin.ReleaseMode)
	api.router = gin.New()
	api.DnsServer = dnsServer
	api.DnsServer.WebServer = api.router

	api.setupRoutes()
	api.setupAuthorizedRoutes()
	api.ServeEmbeddedContent(content)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		log.Info("Starting server on port %d", api.Config.Port)
		if err := api.router.Run(fmt.Sprintf(":%d", api.Config.Port)); err != nil {
			log.Error("%v", err)
			errorChannel <- struct{}{}
		}
	}()

	wg.Wait()
}

func (api *API) SetupAuth() {
	if password, exists := os.LookupEnv("GOAWAY_PASSWORD"); exists {
		api.adminPassword = password
		log.Info("Using custom password: [hidden]")
	} else {
		api.adminPassword = generateRandomPassword()
		log.Info("Randomly generated admin password: %s", api.adminPassword)
	}
}

func (api *API) setupRoutes() {
	api.router.POST("/login", api.handleLogin)
	api.router.GET("/server", api.handleServer)
	api.router.GET("/authentication", api.getAuthentication)
	api.router.GET("/metrics", api.handleMetrics)
}

func (api *API) setupAuthorizedRoutes() {
	authorized := api.router.Group("/")
	if api.Config.Authentication {
		api.SetupAuth()
		authorized.Use(authMiddleware())
	} else {
		log.Info("Authentication is disabled.")
	}

	authorized.POST("/upstreams", api.createUpstreams)
	authorized.POST("/settings", api.updateSettings)
	authorized.POST("/lists", api.updateLists)

	authorized.GET("/queriesData", api.handleQueriesData)
	authorized.GET("/queryTimestamps", api.getQueryTimestamps)
	authorized.GET("/updateBlockStatus", api.handleUpdateBlockStatus)
	authorized.GET("/domains", api.getDomains)
	authorized.GET("/settings", api.getSettings)
	authorized.GET("/clients", api.getClients)
	authorized.GET("/upstreams", api.getUpstreams)
	authorized.GET("/preferredUpstream", api.setPreferredUpstream)
	authorized.GET("/topBlockedDomains", api.getTopBlockedDomains)
	authorized.GET("/lists", api.getLists)
	authorized.GET("/getDomainsForList", api.getDomainsForList)

	authorized.DELETE("/upstreams", api.removeUpstreams)
	authorized.DELETE("/logs", api.clearLogs)
}

func generateRandomPassword() string {
	randomBytes := make([]byte, 14)
	if _, err := rand.Read(randomBytes); err != nil {
		log.Error("Error generating random bytes: %v", err)
	}

	return base64.RawStdEncoding.EncodeToString(randomBytes)
}
