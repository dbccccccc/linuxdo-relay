package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	authpkg "linuxdo-relay/internal/auth"
	"linuxdo-relay/internal/config"
	"linuxdo-relay/internal/models"
	"linuxdo-relay/internal/storage"
)

type AppContext struct {
	Config    *config.Config
	DB        *storage.DB
	Redis     *storage.Redis
	OAuth     *oauth2.Config
	JWTSecret string
}

func SetupRoutes(r *gin.Engine, app *AppContext) {
	// health check
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// auth routes (LinuxDo OAuth)
	RegisterAuthRoutes(r, app)

	// authenticated routes
	authGroup := r.Group("/")
	authGroup.Use(AuthMiddleware(app))
	authGroup.Use(QuotaMiddleware(app))
	authGroup.Use(CreditMiddleware(app))

	// user info
	authGroup.GET("/me", func(c *gin.Context) {
		uidVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no user in context"})
			return
		}

		userID, ok := uidVal.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
			return
		}

		var user models.User
		if err := app.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":               user.ID,
			"linuxdo_user_id":  user.LinuxDoUserID,
			"linuxdo_username": user.LinuxDoUsername,
			"role":             user.Role,
			"level":            user.Level,
			"status":           user.Status,
			"credits":          user.Credits,
		})
	})

	authGroup.GET("/me/credit_transactions", func(c *gin.Context) {
		uidVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no user in context"})
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
			return
		}

		pageStr := c.DefaultQuery("page", "1")
		pageSizeStr := c.DefaultQuery("page_size", "20")
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(pageSizeStr)
		if pageSize <= 0 || pageSize > 100 {
			pageSize = 20
		}

		db := app.DB.Model(&models.CreditTransaction{}).Where("user_id = ?", userID)

		var total int64
		if err := db.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count credit transactions"})
			return
		}

		var txns []models.CreditTransaction
		if err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&txns).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load credit transactions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"total": total, "items": txns})
	})

	authGroup.GET("/me/check_in/status", func(c *gin.Context) {
		uidVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no user in context"})
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
			return
		}

		var user models.User
		if err := app.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
			return
		}

		status, err := loadTodayCheckInStatus(app, user.ID, user.Level, user.Credits)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load check-in status"})
			return
		}

		recent, err := fetchRecentCheckInLogs(app, user.ID, 7)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load check-in history"})
			return
		}

		resp := gin.H{
			"checked_in_today": status.CheckedToday,
			"today_reward":     status.Reward,
			"streak":           status.Streak,
			"credits":          status.Credits,
			"recent_logs":      recent,
		}
		if status.Config != nil {
			resp["config"] = gin.H{
				"level":                  status.Config.Level,
				"base_reward":            status.Config.BaseReward,
				"decay_threshold":        status.Config.DecayThreshold,
				"min_multiplier_percent": status.Config.MinMultiplierPercent,
			}
		}

		c.JSON(http.StatusOK, resp)
	})

	authGroup.POST("/me/check_in", func(c *gin.Context) {
		uidVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no user in context"})
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
			return
		}

		res, err := performDailyCheckIn(app, userID)
		if err != nil {
			if errors.Is(err, errAlreadyCheckedIn) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "already_checked_in"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check in"})
			return
		}

		recent, err := fetchRecentCheckInLogs(app, userID, 7)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load check-in history"})
			return
		}

		resp := gin.H{
			"reward":      res.Reward,
			"streak":      res.Streak,
			"credits":     res.Credits,
			"recent_logs": recent,
		}
		if res.Config != nil {
			resp["config"] = gin.H{
				"level":                  res.Config.Level,
				"base_reward":            res.Config.BaseReward,
				"decay_threshold":        res.Config.DecayThreshold,
				"min_multiplier_percent": res.Config.MinMultiplierPercent,
			}
		}
		c.JSON(http.StatusOK, resp)
	})

	// regenerate per-user API key (JWT-only, for web UI).
	authGroup.POST("/me/api_key/regenerate", func(c *gin.Context) {
		methodVal, _ := c.Get("auth_method")
		if m, _ := methodVal.(string); m != "jwt" {
			c.JSON(http.StatusForbidden, gin.H{"error": "api key auth not allowed for this endpoint"})
			return
		}

		uidVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no user in context"})
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
			return
		}

		var user models.User
		if err := app.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
			return
		}

		plain, hash, err := authpkg.GenerateUserAPIKey()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate api key"})
			return
		}

		now := time.Now()
		user.APIKeyHash = hash
		user.APIKeyCreatedAt = &now
		if err := app.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist api key"})
			return
		}

		// record operation log (do not store actual key)
		recordOperationLog(app, user.ID, "regenerate_api_key", "user regenerated API key")

		c.JSON(http.StatusOK, gin.H{
			"api_key":     plain,
			"created_at":  user.APIKeyCreatedAt,
			"description": "store this key securely; it will not be shown again",
		})
	})

	// User dashboard: quota usage and logs
	authGroup.GET("/me/quota_usage", func(c *gin.Context) {
		uidVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no user in context"})
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
			return
		}
		levelVal, _ := c.Get("level")
		level, _ := levelVal.(int)

		var rules []models.QuotaRule
		if err := app.DB.Where("level = ?", level).Order("model_pattern").Find(&rules).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load quota rules"})
			return
		}

		ctx := context.Background()
		res := make([]gin.H, 0, len(rules))
		for _, r := range rules {
			bucket := time.Now().Unix() / int64(r.WindowSeconds)
			key := fmt.Sprintf("quota:%d:%d:%s:%d", userID, level, r.ModelPattern, bucket)
			var used int64
			if val, err := app.Redis.Get(ctx, key).Result(); err == nil {
				if n, err2 := strconv.ParseInt(val, 10, 64); err2 == nil {
					used = n
				}
			}
			if used < 0 {
				used = 0
			}
			if used > int64(r.MaxRequests) {
				used = int64(r.MaxRequests)
			}
			res = append(res, gin.H{
				"model_pattern":  r.ModelPattern,
				"max_requests":   r.MaxRequests,
				"window_seconds": r.WindowSeconds,
				"used":           used,
				"remaining":      int64(r.MaxRequests) - used,
			})
		}

		c.JSON(http.StatusOK, gin.H{"items": res})
	})

	authGroup.GET("/me/api_logs", func(c *gin.Context) {
		uidVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no user in context"})
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
			return
		}

		pageStr := c.DefaultQuery("page", "1")
		pageSizeStr := c.DefaultQuery("page_size", "20")
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(pageSizeStr)
		if pageSize <= 0 || pageSize > 100 {
			pageSize = 20
		}

		db := app.DB.Model(&models.APILog{}).Where("user_id = ?", userID)

		if start := c.Query("start"); start != "" {
			db = db.Where("created_at >= ?", start)
		}
		if end := c.Query("end"); end != "" {
			db = db.Where("created_at <= ?", end)
		}

		var total int64
		if err := db.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count logs"})
			return
		}

		var logs []models.APILog
		if err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load logs"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"total": total, "items": logs})
	})

	authGroup.GET("/me/operation_logs", func(c *gin.Context) {
		uidVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no user in context"})
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
			return
		}

		pageStr := c.DefaultQuery("page", "1")
		pageSizeStr := c.DefaultQuery("page_size", "20")
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(pageSizeStr)
		if pageSize <= 0 || pageSize > 100 {
			pageSize = 20
		}

		db := app.DB.Model(&models.OperationLog{}).Where("user_id = ?", userID)
		if start := c.Query("start"); start != "" {
			db = db.Where("created_at >= ?", start)
		}
		if end := c.Query("end"); end != "" {
			db = db.Where("created_at <= ?", end)
		}

		var total int64
		if err := db.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count logs"})
			return
		}

		var logs []models.OperationLog
		if err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load logs"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"total": total, "items": logs})
	})

	// admin & relay routes (protected)
	RegisterAdminRoutes(authGroup, app)
	RegisterRelayRoutes(authGroup, app)
}

/*

	// User dashboard: quota usage and logs
	authGroup.GET("/me/quota_usage", func(c *gin.Context) {
		uidVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no user in context"})
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
			return
		}
		levelVal, _ := c.Get("level")
		level, _ := levelVal.(int)

		var rules []models.QuotaRule
		if err := app.DB.Where("level = ?", level).Order("model_pattern").Find(&rules).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load quota rules"})
			return
		}

		ctx := context.Background()
		res := make([]gin.H, 0, len(rules))
		for _, r := range rules {
			bucket := time.Now().Unix() / int64(r.WindowSeconds)
			key := fmt.Sprintf("quota:%d:%d:%s:%d", userID, level, r.ModelPattern, bucket)
			var used int64
			if val, err := app.Redis.Get(ctx, key).Result(); err == nil {
				if n, err2 := strconv.ParseInt(val, 10, 64); err2 == nil {
					used = n
				}
			}
			if used < 0 {
				used = 0
			}
			if used > int64(r.MaxRequests) {
				used = int64(r.MaxRequests)
			}
			res = append(res, gin.H{
				"model_pattern":  r.ModelPattern,
				"max_requests":  r.MaxRequests,
				"window_seconds": r.WindowSeconds,
				"used":          used,
				"remaining":     int64(r.MaxRequests) - used,
			})
		}

		c.JSON(http.StatusOK, gin.H{"items": res})
	})

	authGroup.GET("/me/api_logs", func(c *gin.Context) {
		uidVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no user in context"})
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
			return
		}

		pageStr := c.DefaultQuery("page", "1")
		pageSizeStr := c.DefaultQuery("page_size", "20")
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(pageSizeStr)
		if pageSize <= 0 || pageSize > 100 {
			pageSize = 20
		}

		db := app.DB.Model(&models.APILog{}).Where("user_id = ?", userID)

		if start := c.Query("start"); start != "" {
			db = db.Where("created_at >= ?", start)
		}
		if end := c.Query("end"); end != "" {
			db = db.Where("created_at <= ?", end)
		}

		var total int64
		if err := db.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count logs"})
			return
		}

		var logs []models.APILog
		if err := db.Order("created_at DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load logs"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"total": total, "items": logs})
	})

	authGroup.GET("/me/operation_logs", func(c *gin.Context) {
		uidVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no user in context"})
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
			return
		}

		pageStr := c.DefaultQuery("page", "1")
		pageSizeStr := c.DefaultQuery("page_size", "20")
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(pageSizeStr)
		if pageSize <= 0 || pageSize > 100 {
			pageSize = 20
		}

		db := app.DB.Model(&models.OperationLog{}).Where("user_id = ?", userID)
		if start := c.Query("start"); start != "" {
			db = db.Where("created_at >= ?", start)
		}
		if end := c.Query("end"); end != "" {
			db = db.Where("created_at <= ?", end)
		}

		var total int64
		if err := db.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count logs"})
			return
		}

		var logs []models.OperationLog
		if err := db.Order("created_at DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load logs"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"total": total, "items": logs})
	})


*/
