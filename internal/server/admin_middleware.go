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

// JWTOnlyMiddleware ensures that only JWT authentication is allowed.
// This is used for user management endpoints that should not be accessible via API key.
func JWTOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method, _ := c.Get("auth_method")
		if m, _ := method.(string); m != "jwt" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "this endpoint requires web login, API key not allowed",
			})
			return
		}
		c.Next()
	}
}

// APIKeyOnlyMiddleware ensures that only API key authentication is allowed.
// This is used for LLM relay endpoints that should not be accessible via JWT.
func APIKeyOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method, _ := c.Get("auth_method")
		if m, _ := method.(string); m != "api_key" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "this endpoint requires API key, web login token not allowed",
			})
			return
		}
		c.Next()
	}
}
