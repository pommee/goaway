package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"goaway/backend/api/user"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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
	query := "SELECT password FROM user WHERE username = ?"

	var hashedPassword string
	err := api.DBManager.Conn.QueryRow(query, username).Scan(&hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info("Authentication attempt for non-existent or invalid credentials")
		} else {
			log.Warning("Database error during authentication: %v", err)
		}
		return false
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
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
	type PasswordChange struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}

	updatedList, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var request PasswordChange
	if err := json.Unmarshal(updatedList, &request); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if !api.validateCredentials("admin", request.CurrentPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is not valid"})
		return
	}

	existingUser := user.User{Username: "admin", Password: request.NewPassword}
	if err = existingUser.UpdatePassword(api.DBManager.Conn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to update password"})
		return
	}

	log.Info("Password has been changed!")
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

	apiKey, err := api.KeyManager.CreateApiKey(request.Name)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"key": apiKey})
}

func (api *API) getAPIKeys(c *gin.Context) {
	apiKeys, err := api.KeyManager.GetAllApiKeys()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"keys": apiKeys})
}

func (api *API) deleteAPIKey(c *gin.Context) {
	keyToDelete := c.Query("key")

	err := api.KeyManager.DeleteApiKey(keyToDelete)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted api key!"})
}
