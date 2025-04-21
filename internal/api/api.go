package api

import (
	"crypto/rand"
	"embed"
	"encoding/base64"
	"fmt"
	"goaway/internal/logging"
	"goaway/internal/server"
	"goaway/internal/settings"
	"goaway/internal/user"
	"net"

	"os"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var log = logging.GetLogger()

const (
	tokenDuration = 5 * time.Minute
	jwtSecret     = "kMNSRwKip7Yet4rb2z8"
)

type Credentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type API struct {
	router    *gin.Engine
	routes    *gin.RouterGroup
	DnsServer *server.DNSServer
	Config    *settings.APIServerConfig
	Version   string
	Commit    string
	Date      string
}

func (api *API) Start(serveContent bool, content embed.FS, dnsServer *server.DNSServer, errorChannel chan struct{}) {
	gin.SetMode(gin.ReleaseMode)
	api.router = gin.New()
	api.routes = api.router.Group("/api")
	api.DnsServer = dnsServer
	api.DnsServer.WebServer = api.router
	var allowedOrigins []string

	if serveContent {
		allowedOrigins = append(allowedOrigins, "*")
	} else {
		allowedOrigins = append(allowedOrigins, "http://localhost:8081")
	}

	api.router.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Cookie"},
		ExposeHeaders:    []string{"Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api.setupRoutes()
	api.setupAuthorizedRoutes(!serveContent)

	if serveContent {
		api.ServeEmbeddedContent(content)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		addr := fmt.Sprintf(":%d", api.Config.Port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Error("Failed to start server: %v", err)
			errorChannel <- struct{}{}
			return
		}
		log.Info("Web server started on port :%d", api.Config.Port)

		serverIP, err := getServerIP()
		if err != nil {
			log.Error("Could not start web server, reason: %v", err)
		}
		log.Info("Web interface available at http://%s:%d", serverIP, api.Config.Port)

		if err := api.router.RunListener(listener); err != nil {
			log.Error("Server error: %v", err)
			errorChannel <- struct{}{}
		}
	}()

	wg.Wait()
}

func (api *API) SetupAuth() {
	newUser := &user.User{Username: "admin"}
	if newUser.Exists(api.DnsServer.DB) {
		return
	}

	log.Info("Creating a new admin user")
	if password, exists := os.LookupEnv("GOAWAY_PASSWORD"); exists {
		newUser.Password = password
		err := newUser.Create(api.DnsServer.DB)
		if err != nil {
			log.Error("Unable to create new user: %v", err)
		}
		log.Info("Using custom password: [hidden]")
	} else {
		newUser.Password = generateRandomPassword()
		err := newUser.Create(api.DnsServer.DB)
		if err != nil {
			log.Error("Unable to create new user: %v", err)
		}
		log.Info("Randomly generated admin password: %s", newUser.Password)
	}
}

func (api *API) setupRoutes() {
	api.router.POST("/api/login", api.handleLogin)
	api.router.GET("/api/server", api.handleServer)
	api.router.GET("/api/authentication", api.getAuthentication)
	api.router.GET("/api/metrics", api.handleMetrics)
}

func (api *API) setupAuthorizedRoutes(devmode bool) {
	if api.Config.Authentication {
		api.SetupAuth()
		api.routes.Use(authMiddleware())
	} else {
		log.Info("Authentication is disabled.")

		if devmode {
			api.routes.Use(cors.New(cors.Config{
				AllowOrigins:     []string{"http://localhost:8081"},
				AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowHeaders:     []string{"Content-Type", "Authorization", "Cookie"},
				ExposeHeaders:    []string{"Set-Cookie"},
				AllowCredentials: true,
				MaxAge:           12 * time.Hour,
			}))
		}
	}

	api.routes.POST("/upstreams", api.createUpstreams)
	api.routes.POST("/settings", api.updateSettings)
	api.routes.POST("/custom", api.updateCustom)
	api.routes.POST("/toggleBlocklist", api.toggleBlocklist)
	api.routes.POST("/resolution", api.createResolution)
	api.routes.POST("/pause", api.pauseBlocking)

	api.routes.GET("/queries", api.getQueries)
	api.routes.GET("/liveQueries", api.liveQueries)
	api.routes.GET("/queryTimestamps", api.getQueryTimestamps)
	api.routes.GET("/queryTypes", api.getQueryTypes)
	api.routes.GET("/updateBlockStatus", api.handleUpdateBlockStatus)
	api.routes.GET("/domains", api.getDomains)
	api.routes.GET("/settings", api.getSettings)
	api.routes.GET("/clients", api.getClients)
	api.routes.GET("/clientDetails", api.getClientDetails)
	api.routes.GET("/resolutions", api.getResolutions)
	api.routes.GET("/upstreams", api.getUpstreams)
	api.routes.GET("/preferredUpstream", api.setPreferredUpstream)
	api.routes.GET("/topBlockedDomains", api.getTopBlockedDomains)
	api.routes.GET("/topClients", api.getTopClients)
	api.routes.GET("/lists", api.getLists)
	api.routes.GET("/addList", api.addList)
	api.routes.GET("/getDomainsForList", api.getDomainsForList)
	api.routes.GET("/runUpdate", api.runUpdate)
	api.routes.GET("/pause", api.getBlocking)

	api.routes.PUT("/password", api.updatePassword)

	api.routes.DELETE("/upstreams", api.removeUpstreams)
	api.routes.DELETE("/queries", api.clearQueries)
	api.routes.DELETE("/list", api.removeList)
	api.routes.DELETE("/resolution", api.deleteResolution)
	api.routes.DELETE("/pause", api.clearBlocking)
}

func generateRandomPassword() string {
	randomBytes := make([]byte, 14)
	if _, err := rand.Read(randomBytes); err != nil {
		log.Error("Error generating random bytes: %v", err)
	}

	return base64.RawStdEncoding.EncodeToString(randomBytes)
}
