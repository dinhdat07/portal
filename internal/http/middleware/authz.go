package middleware

import (
	"net/http"
	"portal-system/internal/domain/enum"

	"github.com/gin-gonic/gin"
)

func RequireRole(role enum.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists || userRole != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "forbidden",
			})
			return
		}
		c.Next()
	}
}
