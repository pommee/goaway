package api

import (
	"goaway/backend/api/user"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenDuration = 5 * time.Minute
	Secret        = "kMNSRwKip7Yet4rb2z8"
)

func (api *API) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/server") {
			c.Next()
			return
		}

		apiKey := c.GetHeader("api-key")
		if apiKey != "" {
			api.KeyManager.VerifyApiKey(apiKey)
			return
		}

		cookie, err := c.Cookie("jwt")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization cookie required"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(cookie, func(t *jwt.Token) (any, error) {
			return []byte(Secret), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			expirationTime := int64(claims["exp"].(float64))
			if time.Now().Unix() > expirationTime {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
				c.Abort()
				return
			}

			if time.Now().Unix() > expirationTime-int64(TokenDuration/2) {
				newToken, err := generateToken(claims["username"].(string))
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to renew token"})
					c.Abort()
					return
				}
				setAuthCookie(c.Writer, newToken)
			}

			c.Set("username", claims["username"])
		}

		c.Next()
	}
}

func generateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(TokenDuration).Unix(),
		"iat":      time.Now().Unix(),
	})
	return token.SignedString([]byte(Secret))
}

func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  time.Now().Add(TokenDuration),
	})
}

func (api *API) validateCredentials(username, password string) bool {
	existingUser := &user.User{Username: username, Password: password}
	return existingUser.Authenticate(api.DB)
}
