package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	authpkg "linuxdo-relay/internal/auth"
	"linuxdo-relay/internal/models"
)

// AuthMiddleware validates Authorization header and injects user info into
// context. It supports both JWT (for web login) and per-user API keys for
// programmatic access.
func AuthMiddleware(app *AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

		// First, try to parse as JWT (web login flow).
		if claims, err := authpkg.ParseToken(app.JWTSecret, tokenStr); err == nil {
			var user models.User
			if err := app.DB.First(&user, claims.UserID).Error; err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
				return
			}
			if user.Status != models.UserStatusNormal {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user disabled"})
				return
			}
			c.Set("user_id", user.ID)
			c.Set("role", user.Role)
			c.Set("level", user.Level)
			c.Set("auth_method", "jwt")
			c.Next()
			return
		}

		// If not a valid JWT, treat as per-user API key. We only accept keys
		// that start with the expected prefix to avoid confusing random tokens
		// with API keys.
		if !strings.HasPrefix(tokenStr, "sk-") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Look up user by hashed API key.
		hash := authpkg.HashAPIKey(tokenStr)
		var user models.User
		if err := app.DB.Where("api_key_hash = ?", hash).First(&user).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		if user.Status != models.UserStatusNormal {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user disabled"})
			return
		}

		// inject into context
		c.Set("user_id", user.ID)
		c.Set("role", user.Role)
		c.Set("level", user.Level)
		c.Set("auth_method", "api_key")

		c.Next()
	}
}
