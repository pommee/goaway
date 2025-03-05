package api

import (
	"goaway/internal/user"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/server") {
			c.Next()
			return
		}

		cookie, err := c.Cookie("jwt")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization cookie required"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(cookie, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
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

			if time.Now().Unix() > expirationTime-int64(tokenDuration/2) {
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
		"exp":      time.Now().Add(tokenDuration).Unix(),
		"iat":      time.Now().Unix(),
	})
	return token.SignedString([]byte(jwtSecret))
}

func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(tokenDuration),
	})
}

func (api *API) validateCredentials(username, password string) bool {
	existingUser := &user.User{Username: username, Password: password}
	return existingUser.Authenticate(api.DnsServer.DB)
}
