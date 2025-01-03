package api

import (
	"crypto/rand"
	"embed"
	"encoding/base64"
	"fmt"
	"goaway/internal/logging"
	"goaway/internal/server"
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
	Router                *gin.Engine
	dnsServer             *server.DNSServer
	port                  int
	DisableAuthentication bool
	adminPassword         string
}

func (api *API) Start(content embed.FS, dnsServer *server.DNSServer, port int, errorChannel chan struct{}) {
	password, exists := os.LookupEnv("GOAWAY_PASSWORD")
	if !exists {
		api.adminPassword = generateRandomPassword(14)
		log.Info("Randomly generated admin password: %s", api.adminPassword)
	} else {
		api.adminPassword = password
		log.Info("Using custom password: [hidden]")
	}

	gin.SetMode(gin.ReleaseMode)
	api.Router = gin.New()
	dnsServer.WebServer = api.Router
	api.dnsServer = dnsServer
	api.port = port

	api.setupRoutes()
	api.ServeEmbeddedContent(content)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		log.Info("Starting server on port %d", port)
		err := api.Router.Run(fmt.Sprintf(":%d", port))
		if err != nil {
			log.Error("%v", err)
			errorChannel <- struct{}{}
		}
	}()

	wg.Wait()
}

func (api *API) setupRoutes() {
	api.Router.POST("/login", api.handleLogin)
	api.Router.GET("/server", api.handleServer)
	api.Router.GET("/authentication", api.getAuthentication)

	authorized := api.Router.Group("/")
	if !api.DisableAuthentication {
		authorized.Use(authMiddleware())
	} else {
		log.Info("Authentication is disabled.")
	}

	authorized.GET("/metrics", api.handleMetrics)
	authorized.GET("/queriesData", api.handleQueriesData)
	authorized.GET("/updateBlockStatus", api.handleUpdateBlockStatus)
	authorized.GET("/domains", api.getDomains)
	authorized.GET("/settings", api.getSettings)
	authorized.POST("/settings", api.updateSettings)
	authorized.GET("/clients", api.getClients)
	authorized.GET("/upstreams", api.getUpstreams)
	authorized.POST("/upstreams", api.createUpstreams)
	authorized.DELETE("/upstreams", api.removeUpstreams)
	authorized.GET("/preferredUpstream", api.setPreferredUpstream)
}

func generateRandomPassword(length int) string {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Error("Error generating random bytes: %v", err)
	}

	password := base64.StdEncoding.EncodeToString(randomBytes)
	return password
}
