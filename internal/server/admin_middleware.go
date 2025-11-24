package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminOnlyMiddleware ensures that only admin users can access the group.
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, ok := c.Get("role")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no role in context"})
			return
		}
		role, _ := roleVal.(string)
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin only"})
			return
		}
		c.Next()
	}
}
