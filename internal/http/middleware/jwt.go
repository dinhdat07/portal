package middleware

import (
	"net/http"
	"portal-system/internal/platform/token"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuth(manager *token.Manager, authCookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""
		auth := c.GetHeader("Authorization")

		if auth != "" {
			parts := strings.Split(auth, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		if tokenString == "" {
			cookieToken, err := c.Cookie(authCookieName)
			if err == nil {
				tokenString = cookieToken
			}
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		// use manager parser to parse tokenstring for authorize
		claims, err := manager.Parse(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorize format",
			})
			return
		}

		// set necessary field in gin context
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Set("username", claims.Username)

		c.Next()
	}
}
