package api

import (
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

		if apiKey := c.GetHeader("api-key"); apiKey != "" {
			if api.KeyManager.VerifyApiKey(apiKey) {
				c.Next()
				return
			}
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		cookie, err := c.Cookie("jwt")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization cookie"})
			return
		}

		claims, err := parseToken(cookie)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		username, ok := claims["username"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		now := time.Now().Unix()
		exp, ok := claims["exp"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token expiration"})
			return
		}
		expiration := int64(exp)

		if now > expiration {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			return
		}

		if now > expiration-int64(TokenDuration/2) {
			newToken, err := generateToken(username)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to renew token"})
				return
			}
			setAuthCookie(c.Writer, newToken)
		}

		c.Set("username", username)
		c.Next()
	}
}

func parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, err
	}

	return claims, nil
}

func generateToken(username string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"username": username,
		"exp":      now.Add(TokenDuration).Unix(),
		"iat":      now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(Secret))
}

func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(TokenDuration),
	})
}
