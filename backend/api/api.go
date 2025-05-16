package api

import (
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/base64"
	"fmt"
	notification "goaway/backend"
	"goaway/backend/api/key"
	api "goaway/backend/api/key"
	"goaway/backend/api/user"
	"goaway/backend/dns/blacklist"
	"goaway/backend/dns/server"
	"goaway/backend/logging"
	"goaway/backend/settings"
	"net"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	_ "github.com/mattn/go-sqlite3"
)

var log = logging.GetLogger()

type Credentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type API struct {
	Authentication bool
	Config         *settings.Config
	router         *gin.Engine
	routes         *gin.RouterGroup
	DB             *sql.DB
	Blacklist      *blacklist.Blacklist
	KeyManager     *api.ApiKeyManager
	WS             *websocket.Conn
	DNSPort        int
	Version        string
	Commit         string
	Date           string
	Notifications  *notification.NotificationManager
}

func (api *API) Start(content embed.FS, dnsServer *server.DNSServer, errorChannel chan struct{}) {
	gin.SetMode(gin.ReleaseMode)
	api.router = gin.New()
	api.router.Use(gzip.Gzip(gzip.DefaultCompression))
	api.routes = api.router.Group("/api")
	var allowedOrigins []string

	if !api.Config.DevMode {
		allowedOrigins = append(allowedOrigins, "*")
	} else {
		log.Warning("No embedded content found, not serving")
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

	api.KeyManager = key.NewApiKeyManager(api.DB)

	api.setupRoutes()
	api.setupAuthorizedRoutes()
	api.setupWebsocket(dnsServer)

	if !api.Config.DevMode {
		api.ServeEmbeddedContent(content)
	}

	go func() {
		const maxRetries = 10
		const retryDelay = 10 * time.Second
		addr := fmt.Sprintf(":%d", api.Config.API.Port)

		for i := 1; i <= maxRetries; i++ {
			listener, err := net.Listen("tcp", addr)
			if err != nil {
				log.Error("Failed to start server (attempt %d/%d): %v", i, maxRetries, err)
				if i < maxRetries {
					time.Sleep(retryDelay)
					continue
				}
				errorChannel <- struct{}{}
				return
			}

			log.Info("Web server started on port :%d", api.Config.API.Port)

			if serverIP, err := getServerIP(); err == nil {
				log.Info("Web interface available at http://%s:%d", serverIP, api.Config.API.Port)
			} else {
				log.Error("Could not determine server IP: %v", err)
			}

			if err := api.router.RunListener(listener); err != nil {
				log.Error("Server error: %v", err)
				errorChannel <- struct{}{}
			}
			return
		}
	}()
}

func (api *API) SetupAuth() {
	newUser := &user.User{Username: "admin"}
	if newUser.Exists(api.DB) {
		return
	}

	log.Info("Creating a new admin user")

	if password, exists := os.LookupEnv("GOAWAY_PASSWORD"); exists {
		newUser.Password = password
		log.Info("Using custom password: [hidden]")
	} else {
		newUser.Password = generateRandomPassword()
		log.Info("Randomly generated admin password: %s", newUser.Password)
	}

	if err := newUser.Create(api.DB); err != nil {
		log.Error("Unable to create new user: %v", err)
	}
}

func (api *API) setupRoutes() {
	api.router.POST("/api/login", api.handleLogin)
	api.router.GET("/api/server", api.handleServer)
	api.router.GET("/api/authentication", api.getAuthentication)
	api.router.GET("/api/dnsMetrics", api.handleMetrics)
}

func (api *API) setupAuthorizedRoutes() {
	if api.Authentication {
		api.SetupAuth()
		api.routes.Use(api.authMiddleware())
	} else {
		log.Info("Authentication is disabled.")

		if api.Config.DevMode {
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

	api.routes.POST("/upstream", api.createUpstream)
	api.routes.POST("/settings", api.updateSettings)
	api.routes.POST("/custom", api.updateCustom)
	api.routes.POST("/resolution", api.createResolution)
	api.routes.POST("/pause", api.pauseBlocking)

	api.routes.POST("/apiKey", api.createAPIKey)
	api.routes.GET("/apiKey", api.getAPIKeys)
	api.routes.GET("/deleteApiKey", api.deleteAPIKey)

	api.routes.GET("/notifications", api.fetchNotifications)
	api.routes.GET("/removeFromCustom", api.removeDomainFromCustom)
	api.routes.GET("/queries", api.getQueries)
	api.routes.GET("/queryTimestamps", api.getQueryTimestamps)
	api.routes.GET("/queryTypes", api.getQueryTypes)
	api.routes.GET("/updateBlockStatus", api.handleUpdateBlockStatus)
	api.routes.GET("/domains", api.getDomains)
	api.routes.GET("/settings", api.getSettings)
	api.routes.GET("/clients", api.getClients)
	api.routes.GET("/clientDetails", api.getClientDetails)
	api.routes.GET("/resolutions", api.getResolutions)
	api.routes.GET("/upstreams", api.getUpstreams)
	api.routes.GET("/topBlockedDomains", api.getTopBlockedDomains)
	api.routes.GET("/topClients", api.getTopClients)
	api.routes.GET("/lists", api.getLists)
	api.routes.GET("/addList", api.addList)
	api.routes.GET("/fetchUpdatedList", api.fetchUpdatedList)
	api.routes.GET("/runUpdateList", api.runUpdateList)
	api.routes.GET("/getDomainsForList", api.getDomainsForList)
	api.routes.GET("/runUpdate", api.runUpdate)
	api.routes.GET("/pause", api.getBlocking)
	api.routes.GET("/toggleBlocklist", api.toggleBlocklist)
	api.routes.GET("/exportDatabase", api.exportDatabase)

	api.routes.PUT("/password", api.updatePassword)
	api.routes.PUT("/preferredUpstream", api.updatePreferredUpstream)

	api.routes.DELETE("/upstream", api.deleteUpstream)
	api.routes.DELETE("/queries", api.clearQueries)
	api.routes.DELETE("/list", api.removeList)
	api.routes.DELETE("/resolution", api.deleteResolution)
	api.routes.DELETE("/pause", api.clearBlocking)
	api.routes.DELETE("/notification", api.markNotificationAsRead)
}

func (api *API) setupWebsocket(dnsServer *server.DNSServer) {
	api.router.GET("/api/liveQueries", func(c *gin.Context) {
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		api.WS = conn

		if dnsServer != nil {
			dnsServer.WS = conn
		}
	})
}

func generateRandomPassword() string {
	randomBytes := make([]byte, 14)
	if _, err := rand.Read(randomBytes); err != nil {
		log.Error("Error generating random bytes: %v", err)
	}
	return base64.RawStdEncoding.EncodeToString(randomBytes)
}
