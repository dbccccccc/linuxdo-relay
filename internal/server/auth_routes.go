package server

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	authpkg "linuxdo-relay/internal/auth"
	"linuxdo-relay/internal/models"
)

// linuxDoUserInfo is a minimal view of LinuxDo's user info response.
type linuxDoUserInfo struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	AvatarTemplate string `json:"avatar_template"`
	Active         bool   `json:"active"`
	TrustLevel     int    `json:"trust_level"`
	Silenced       bool   `json:"silenced"`
}

// RegisterAuthRoutes registers LinuxDo OAuth-based auth endpoints.
func RegisterAuthRoutes(r *gin.Engine, app *AppContext) {
	r.GET("/auth/linuxdo/web_login", func(c *gin.Context) {
		// mark this browser session as popup login mode, then redirect to LinuxDo OAuth
		state := uuid.NewString()
		c.SetCookie("oauth_state", state, 300, "/", "", false, true)
		c.SetCookie("oauth_mode", "popup", 300, "/", "", false, true)

		url := app.OAuth.AuthCodeURL(state)
		c.Redirect(http.StatusFound, url)
	})

	r.GET("/auth/linuxdo/login", func(c *gin.Context) {
		state := uuid.NewString()
		// Save state in short-lived cookie for CSRF protection.
		c.SetCookie("oauth_state", state, 300, "/", "", false, true)

		url := app.OAuth.AuthCodeURL(state)
		c.Redirect(http.StatusFound, url)
	})

	r.GET("/auth/linuxdo/callback", func(c *gin.Context) {
		state := c.Query("state")
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
			return
		}

		cookieState, err := c.Cookie("oauth_state")
		if err != nil || cookieState == "" || cookieState != state {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
			return
		}

		// Exchange code for access token.
		token, err := authpkg.ExchangeCode(c.Request.Context(), app.OAuth, code)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to exchange code"})
			return
		}

		// Fetch user info from LinuxDo.
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, app.Config.LinuxDoUserInfoURL, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build userinfo request"})
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch userinfo"})
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			c.JSON(http.StatusBadGateway, gin.H{"error": "unexpected userinfo status", "status": resp.StatusCode, "body": string(body)})
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read userinfo body"})
			return
		}

		var info linuxDoUserInfo
		if err := json.Unmarshal(body, &info); err != nil {
			// Include response body for debugging
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to parse userinfo",
				"details": err.Error(),
				"body":    string(body),
			})
			return
		}

		linuxID := info.ID
		if linuxID == "" {
			linuxID = info.Username
		}
		username := info.Username
		if username == "" {
			username = info.Name
		}
		if linuxID == "" || username == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "userinfo missing id or username",
				"info":  info,
			})
			return
		} // Find or create local user.
		signupCredits := app.Config.SignupCredits
		if signupCredits < 0 {
			signupCredits = 0
		}

		var user models.User
		if err := app.DB.Where("linuxdo_user_id = ?", linuxID).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Determine if this is the first user.
				var count int64
				if err := app.DB.Model(&models.User{}).Count(&count).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count users"})
					return
				}

				role := models.UserRoleUser
				if count == 0 {
					role = models.UserRoleAdmin
				}

				user = models.User{
					LinuxDoUserID:   linuxID,
					LinuxDoUsername: username,
					Role:            role,
					Level:           1,
					Status:          models.UserStatusNormal,
					Credits:         signupCredits,
				}
				if err := app.DB.Create(&user).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
					return
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
				return
			}
		} else {
			// Update username in case it changed on LinuxDo.
			if user.LinuxDoUsername != username {
				user.LinuxDoUsername = username
				if err := app.DB.Save(&user).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
					return
				}
			}
		}

		// Issue JWT for this user.
		tokenStr, err := authpkg.GenerateToken(app.JWTSecret, user.ID, user.Role, user.Level, 24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		// Record login log for audit
		ip := c.ClientIP()
		ua := c.Request.UserAgent()
		recordLoginLog(app, user.ID, ip, ua)

		// Read optional web login mode from cookie.
		mode, _ := c.Cookie("oauth_mode")

		// Clear state and mode cookies.
		c.SetCookie("oauth_state", "", -1, "/", "", false, true)
		if mode != "" {
			c.SetCookie("oauth_mode", "", -1, "/", "", false, true)
		}

		userPayload := gin.H{
			"id":               user.ID,
			"linuxdo_user_id":  user.LinuxDoUserID,
			"linuxdo_username": user.LinuxDoUsername,
			"role":             user.Role,
			"level":            user.Level,
			"status":           user.Status,
		}

		// If called from web popup mode, return a small HTML page that posts a message back.
		if mode == "popup" {
			payload := gin.H{
				"type":  "linuxdo-login-success",
				"token": tokenStr,
				"user":  userPayload,
			}
			payloadJSON, err := json.Marshal(payload)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build web response"})
				return
			}

			html := "<!doctype html><html><body><script>(function(){var data="
			html += string(payloadJSON)
			html += ";if(window.opener&&!window.opener.closed){window.opener.postMessage(data,'*');}window.close();})();</script></body></html>"

			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
			return
		}

		// Default: JSON response (for CLI / programmatic usage).
		c.JSON(http.StatusOK, gin.H{
			"token": tokenStr,
			"user":  userPayload,
		})
	})
}
