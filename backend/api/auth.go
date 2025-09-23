package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"goaway/backend/alert"
	"goaway/backend/api/user"
	"goaway/backend/audit"
	"goaway/backend/dns/database"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (api *API) registerAuthRoutes() {
	api.router.POST("/api/login", api.handleLogin)
	api.router.GET("/api/authentication", api.getAuthentication)
	api.routes.PUT("/password", api.updatePassword)

	api.routes.POST("/apiKey", api.createAPIKey)
	api.routes.GET("/apiKey", api.getAPIKeys)
	api.routes.GET("/deleteApiKey", api.deleteAPIKey)
}

func (api *API) validateCredentials(username, password string) bool {
	existingUser := &user.User{Username: username, Password: password}
	return existingUser.Authenticate(api.DBManager.Conn)
}

func (api *API) handleLogin(c *gin.Context) {
	allowed, timeUntilReset := api.RateLimiter.CheckLimit(c.ClientIP())
	if !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":             "Too many login attempts. Please try again later.",
			"retryAfterSeconds": timeUntilReset,
		})
		return
	}

	var creds user.Credentials
	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if err := creds.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if api.authenticateUser(creds.Username, creds.Password) {
		token, err := generateToken(creds.Username)
		if err != nil {
			log.Info("Token generation failed for user %s: %v", creds.Username, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Authentication service temporarily unavailable",
			})
			return
		}

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")

		setAuthCookie(c.Writer, token)
		c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid username or password",
		})
	}
}

func (api *API) authenticateUser(username, password string) bool {
	var user database.User

	if err := api.DBManager.Conn.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info("Authentication attempt for non-existent or invalid credentials")
		} else {
			log.Warning("Database error during authentication: %v", err)
		}
		return false
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		log.Info("Password comparison failed for user: %s", username)
		return false
	}

	log.Info("Successful authentication for user: %s", username)
	return true
}

func (api *API) getAuthentication(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"enabled": api.Authentication})
}

func (api *API) updatePassword(c *gin.Context) {
	type passwordChange struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}

	var newCredentials passwordChange
	if err := c.BindJSON(&newCredentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if !api.validateCredentials("admin", newCredentials.CurrentPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is not valid"})
		return
	}

	existingUser := user.User{Username: "admin", Password: newCredentials.NewPassword}
	if err := existingUser.UpdatePassword(api.DBManager.Conn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to update password"})
		return
	}

	logMsg := fmt.Sprintf("Password changed for user '%s'", existingUser.Username)
	api.DNSServer.Audits.CreateAudit(&audit.Entry{
		Topic:   audit.TopicUser,
		Message: logMsg,
	})
	go api.DNSServer.Alerts.SendToAll(context.Background(), alert.Message{
		Title:    "System",
		Content:  logMsg,
		Severity: SeverityWarning,
	})

	log.Warning("%s", logMsg)
	c.Status(http.StatusOK)
}

func (api *API) createAPIKey(c *gin.Context) {
	type NewApiKeyName struct {
		Name string `json:"name"`
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var request NewApiKeyName
	if err := json.Unmarshal(body, &request); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	apiKey, err := api.KeyManager.CreateKey(request.Name)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	go api.DNSServer.Alerts.SendToAll(context.Background(), alert.Message{
		Title:    "System",
		Content:  fmt.Sprintf("New API key created with the name '%s'", request.Name),
		Severity: SeverityWarning,
	})

	c.JSON(http.StatusOK, apiKey)
}

func (api *API) getAPIKeys(c *gin.Context) {
	apiKeys, err := api.KeyManager.GetAllKeys()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, apiKeys)
}

func (api *API) deleteAPIKey(c *gin.Context) {
	keyName := c.Query("name")

	err := api.KeyManager.DeleteKey(keyName)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted api key!"})
}
