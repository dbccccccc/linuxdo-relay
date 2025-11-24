package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"linuxdo-relay/internal/models"
)

// QuotaMiddleware enforces per-user, per-level, per-model request limits
// using Redis as a simple counter with fixed time windows.
//
// Limits are defined in the quota_rules table and are purely count-based
// (no token-based quota). If no matching rule exists, the request passes.
func QuotaMiddleware(app *AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Only guard relay endpoints; admin and auth routes are not limited here.
		isRelayPath := strings.HasPrefix(path, "/v1/chat/completions") ||
			strings.HasPrefix(path, "/v1/messages") ||
			strings.HasPrefix(path, "/v1beta/models/")
		if !isRelayPath {
			c.Next()
			return
		}

		uidVal, ok := c.Get("user_id")
		if !ok {
			c.Next()
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
			return
		}

		levelVal, ok := c.Get("level")
		if !ok {
			c.Next()
			return
		}
		level, ok := levelVal.(int)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid user level type"})
			return
		}

		model := extractModelForQuota(c, path)
		// If we cannot determine model, fall back to no limit.
		if model == "" {
			c.Next()
			return
		}

		rule, err := findQuotaRuleForRequest(app, level, model)
		if err != nil {
			// Fail-open on DB errors.
			fmt.Println("quota: failed to load rules:", err)
			c.Next()
			return
		}
		if rule == nil {
			c.Next()
			return
		}

		if app.Redis == nil || app.Redis.Client == nil {
			c.Next()
			return
		}

		// Fixed-window counter using Redis INCR + EXPIRE.
		if rule.WindowSeconds <= 0 || rule.MaxRequests <= 0 {
			c.Next()
			return
		}

		now := time.Now()
		window := int64(rule.WindowSeconds)
		if window <= 0 {
			c.Next()
			return
		}
		aligned := now
		if window >= 86400 {
			aligned = aligned.Add(8 * time.Hour)
		}
		bucket := aligned.Unix() / window
		key := fmt.Sprintf("quota:%d:%d:%s:%d", userID, level, rule.ModelPattern, bucket)

		ctx := context.Background()
		cnt, err := app.Redis.Incr(ctx, key).Result()
		if err != nil {
			fmt.Println("quota: redis error:", err)
			c.Next()
			return
		}
		if cnt == 1 {
			// First hit in this window: set expiry slightly longer than window.
			_ = app.Redis.Expire(ctx, key, time.Duration(rule.WindowSeconds+5)*time.Second).Err()
		}

		if cnt > int64(rule.MaxRequests) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "quota_exceeded",
				"message": fmt.Sprintf("request limit exceeded for model %s", model),
			})
			return
		}

		c.Next()
	}
}

// findQuotaRuleForRequest selects the most specific quota rule for a given
// user level and model name. ModelPattern is treated as a simple prefix.
func findQuotaRuleForRequest(app *AppContext, level int, model string) (*models.QuotaRule, error) {
	var rules []models.QuotaRule
	if err := app.DB.Where("level = ?", level).Find(&rules).Error; err != nil {
		return nil, err
	}

	var best *models.QuotaRule
	bestLen := -1
	for i := range rules {
		p := rules[i].ModelPattern
		if p == "" || strings.HasPrefix(model, p) {
			if len(p) > bestLen {
				bestLen = len(p)
				best = &rules[i]
			}
		}
	}
	return best, nil
}

// extractModelForQuota extracts the requested model name from different
// provider-style endpoints without consuming the request body for handlers.
func extractModelForQuota(c *gin.Context, path string) string {
	// OpenAI /v1/chat/completions
	if strings.HasPrefix(path, "/v1/chat/completions") {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return ""
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		var tmp struct {
			Model string `json:"model"`
		}
		if err := json.Unmarshal(body, &tmp); err != nil {
			return ""
		}
		return tmp.Model
	}

	// Claude /v1/messages
	if strings.HasPrefix(path, "/v1/messages") {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return ""
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		var tmp struct {
			Model string `json:"model"`
		}
		if err := json.Unmarshal(body, &tmp); err != nil {
			return ""
		}
		return tmp.Model
	}

	// Gemini /v1beta/models/*: model is encoded in URL path.
	if strings.HasPrefix(path, "/v1beta/models/") {
		param := c.Param("path")
		return extractGeminiModelName(param)
	}

	return ""
}
