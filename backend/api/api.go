package api

import (
	"crypto/rand"
	"embed"
	"encoding/base64"
	"fmt"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"goaway/backend/api/key"
	"goaway/backend/api/ratelimit"
	"goaway/backend/api/user"
	"goaway/backend/dns/database"
	"goaway/backend/dns/lists"
	"goaway/backend/dns/server"
	"goaway/backend/dns/server/prefetch"
	"goaway/backend/logging"
	notification "goaway/backend/notifications"
	"goaway/backend/settings"
)

var log = logging.GetLogger()

const (
	maxRetries = 10
	retryDelay = 10 * time.Second
)

type API struct {
	Authentication bool
	Config         *settings.Config
	DNSPort        int

	Version string
	Commit  string
	Date    string

	router *gin.Engine
	routes *gin.RouterGroup

	DNSServer                *server.DNSServer
	DBManager                *database.DatabaseManager
	Blacklist                *lists.Blacklist
	Whitelist                *lists.Whitelist
	KeyManager               *key.ApiKeyManager
	PrefetchedDomainsManager *prefetch.Manager
	Notifications            *notification.Manager

	WSQueries       *websocket.Conn
	WSCommunication *websocket.Conn

	RateLimiter *ratelimit.RateLimiter
}

func (api *API) Start(content embed.FS, errorChannel chan struct{}) {
	api.initializeRouter()
	api.configureCORS()
	api.KeyManager = key.NewApiKeyManager(api.DBManager)
	api.RateLimiter = ratelimit.NewRateLimiter(
		api.Config.API.RateLimiterConfig.Enabled,
		api.Config.API.RateLimiterConfig.MaxTries,
		api.Config.API.RateLimiterConfig.Window,
	)
	api.setupRoutes()

	if api.Config.Dashboard {
		api.ServeEmbeddedContent(content)
	}

	api.startServer(errorChannel)
}

func (api *API) initializeRouter() {
	gin.SetMode(gin.ReleaseMode)
	api.router = gin.New()

	// Ignore compression on this route as otherwise it has problems with exposing the Content-Length header
	ignoreCompression := gzip.WithExcludedPaths([]string{"/api/exportDatabase"})
	api.router.Use(gzip.Gzip(gzip.DefaultCompression, ignoreCompression))
	api.routes = api.router.Group("/api")
}

func (api *API) configureCORS() {
	var (
		corsConfig = cors.Config{
			AllowOrigins:     []string{},
			AllowMethods:     []string{"POST", "GET", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Content-Type", "Authorization", "Cookie"},
			ExposeHeaders:    []string{"Set-Cookie"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}
	)

	if api.Config.Dashboard {
		corsConfig.AllowOrigins = append(corsConfig.AllowOrigins, "*")
	} else {
		log.Warning("Dashboard UI is disabled")
		corsConfig.AllowOrigins = append(corsConfig.AllowOrigins, "http://localhost:8081")
		api.routes.Use(cors.New(corsConfig))
	}

	api.router.Use(cors.New(corsConfig))
	api.setupAuthAndMiddleware()
}

func (api *API) setupRoutes() {
	api.registerServerRoutes()
	api.registerAuthRoutes()
	api.registerBlacklistRoutes()
	api.registerWhitelistRoutes()
	api.registerClientRoutes()
	api.registerAuditRoutes()
	api.registerDNSRoutes()
	api.registerUpstreamRoutes()
	api.registerListsRoutes()
	api.registerResolutionRoutes()
	api.registerSettingsRoutes()
	api.registerNotificationRoutes()
	api.registerAlertRoutes()
}

func (api *API) setupAuthAndMiddleware() {
	if api.Authentication {
		api.SetupAuth()
		api.routes.Use(api.authMiddleware())
	} else {
		log.Warning("Authentication is disabled.")
	}
}

func (api *API) SetupAuth() {
	newUser := &user.User{Username: "admin"}
	if newUser.Exists(api.DBManager.Conn) {
		return
	}

	password := api.getOrGeneratePassword()
	newUser.Password = password

	if err := newUser.Create(api.DBManager.Conn); err != nil {
		log.Error("Unable to create new user: %v", err)
	}
}

func (api *API) getOrGeneratePassword() string {
	if password, exists := os.LookupEnv("GOAWAY_PASSWORD"); exists {
		log.Info("Using custom password: [hidden]")
		return password
	}

	password := generateRandomPassword()
	log.Info("Randomly generated admin password: %s", password)
	return password
}

func (api *API) startServer(errorChannel chan struct{}) {
	go func() {
		addr := fmt.Sprintf(":%d", api.Config.API.Port)

		for attempt := 1; attempt <= maxRetries; attempt++ {
			if api.attemptServerStart(addr, attempt, errorChannel) {
				return
			}

			if attempt < maxRetries {
				time.Sleep(retryDelay)
			}
		}

		log.Error("Failed to start server after %d attempts", maxRetries)
		errorChannel <- struct{}{}
	}()
}

func (api *API) attemptServerStart(addr string, attempt int, errorChannel chan struct{}) bool {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("Failed to start server (attempt %d/%d): %v", attempt, maxRetries, err)
		return false
	}

	if serverIP, err := GetServerIP(); err == nil {
		log.Info("Web interface available at http://%s:%d", serverIP, api.Config.API.Port)
	} else {
		log.Info("Web server started on port :%d", api.Config.API.Port)
	}

	if err := api.router.RunListener(listener); err != nil {
		log.Error("Server error: %v", err)
		errorChannel <- struct{}{}
	}

	return true
}

func (api *API) ServeEmbeddedContent(content embed.FS) {
	ipAddress, err := GetServerIP()
	if err != nil {
		log.Error("Error getting IP address: %v", err)
		return
	}

	if err := api.serveStaticFiles(content); err != nil {
		log.Error("Error serving embedded content: %v", err)
		return
	}

	api.serveIndexHTML(content, ipAddress)
}

func (api *API) serveStaticFiles(content embed.FS) error {
	return fs.WalkDir(content, "client/dist", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking through path %s: %w", path, err)
		}

		if d.IsDir() || path == "client/dist/index.html" {
			return nil
		}

		return api.registerStaticFile(content, path)
	})
}

func (api *API) registerStaticFile(content embed.FS, path string) error {
	fileContent, err := content.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", path, err)
	}

	mimeType := api.getMimeType(path)
	route := strings.TrimPrefix(path, "client/dist/")

	api.router.GET("/"+route, func(c *gin.Context) {
		c.Data(http.StatusOK, mimeType, fileContent)
	})

	return nil
}

func (api *API) getMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return mimeType
}

func (api *API) serveIndexHTML(content embed.FS, ipAddress string) {
	indexContent, err := content.ReadFile("client/dist/index.html")
	if err != nil {
		log.Error("Error reading index.html: %v", err)
		return
	}

	indexWithConfig := injectServerConfig(string(indexContent), ipAddress, api.Config.API.Port)
	handleIndexHTML := func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.Data(http.StatusOK, "text/html", []byte(indexWithConfig))
	}

	api.router.GET("/", handleIndexHTML)
	api.router.NoRoute(handleIndexHTML)
}

func injectServerConfig(htmlContent, serverIP string, port int) string {
	serverConfigScript := fmt.Sprintf(`<script>
	window.SERVER_CONFIG = {
		ip: "%s",
		port: "%d"
	};
	</script>`, serverIP, port)

	return strings.Replace(
		htmlContent,
		"<head>",
		"<head>\n  "+serverConfigScript,
		1,
	)
}

func GetServerIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String(), nil
		}
	}

	return "", fmt.Errorf("server IP not found")
}

func generateRandomPassword() string {
	randomBytes := make([]byte, 14)
	if _, err := rand.Read(randomBytes); err != nil {
		log.Error("Error generating random bytes: %v", err)
	}
	return base64.RawStdEncoding.EncodeToString(randomBytes)
}
