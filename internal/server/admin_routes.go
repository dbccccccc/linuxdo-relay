package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"linuxdo-relay/internal/models"
)

// validateModelUniqueness checks that the given models don't conflict with
// other channels. excludeChannelID is used when updating an existing channel.
func validateModelUniqueness(app *AppContext, excludeChannelID uint, newModels []string) error {
	if len(newModels) == 0 {
		return nil
	}

	var allChannels []models.Channel
	query := app.DB.DB
	if excludeChannelID > 0 {
		query = query.Where("id != ?", excludeChannelID)
	}
	if err := query.Find(&allChannels).Error; err != nil {
		return fmt.Errorf("failed to query channels")
	}

	// Build a map of model -> channel name for conflict detection
	modelToChannel := make(map[string]string)
	for _, ch := range allChannels {
		var existingModels []string
		if err := json.Unmarshal([]byte(ch.Models), &existingModels); err != nil {
			continue
		}
		for _, m := range existingModels {
			modelToChannel[m] = ch.Name
		}
	}

	// Check for conflicts
	for _, newModel := range newModels {
		if existingChannel, exists := modelToChannel[newModel]; exists {
			return fmt.Errorf("model '%s' already exists in channel '%s'", newModel, existingChannel)
		}
	}

	return nil
}

func validCheckInConfig(cfg *models.CheckInConfig) bool {
	if cfg == nil {
		return false
	}
	if cfg.Level <= 0 {
		return false
	}
	if cfg.BaseReward <= 0 || cfg.DecayThreshold <= 0 {
		return false
	}
	if cfg.MinMultiplierPercent <= 0 {
		cfg.MinMultiplierPercent = 10
	}
	if cfg.MinMultiplierPercent > 100 {
		cfg.MinMultiplierPercent = 100
	}
	return true
}

// RegisterAdminRoutes registers admin-only management APIs.
func RegisterAdminRoutes(r *gin.RouterGroup, app *AppContext) {
	admin := r.Group("/admin")
	admin.Use(AdminOnlyMiddleware())

	admin.GET("/channels", func(c *gin.Context) {
		var channels []models.Channel
		if err := app.DB.Order("id ASC").Find(&channels).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list channels"})
			return
		}
		c.JSON(http.StatusOK, channels)
	})

	admin.POST("/channels", func(c *gin.Context) {
		var in models.Channel
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if in.Status == "" {
			in.Status = models.ChannelStatusEn
		}

		// Validate model uniqueness: each model can only belong to one channel
		var newModels []string
		if err := json.Unmarshal([]byte(in.Models), &newModels); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid models JSON format"})
			return
		}

		if err := validateModelUniqueness(app, 0, newModels); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := app.DB.Create(&in).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create channel"})
			return
		}
		c.JSON(http.StatusOK, in)
	})

	admin.PUT("/channels/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var ch models.Channel
		if err := app.DB.First(&ch, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}
		if err := c.ShouldBindJSON(&ch); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}

		// Validate model uniqueness when updating
		var newModels []string
		if err := json.Unmarshal([]byte(ch.Models), &newModels); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid models JSON format"})
			return
		}

		if err := validateModelUniqueness(app, uint(id), newModels); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := app.DB.Save(&ch).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update channel"})
			return
		}
		c.JSON(http.StatusOK, ch)
	})

	admin.DELETE("/channels/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		if err := app.DB.Delete(&models.Channel{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete channel"})
			return
		}
		c.Status(http.StatusNoContent)
	})

	// quota rules management
	admin.GET("/quota_rules", func(c *gin.Context) {
		var rules []models.QuotaRule
		if err := app.DB.Order("level ASC, model_pattern ASC").Find(&rules).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list quota rules"})
			return
		}
		c.JSON(http.StatusOK, rules)
	})

	// model credit rules management
	admin.GET("/model_credit_rules", func(c *gin.Context) {
		var rules []models.ModelCreditRule
		if err := app.DB.Order("model_pattern ASC").Find(&rules).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list model credit rules"})
			return
		}
		c.JSON(http.StatusOK, rules)
	})

	admin.POST("/model_credit_rules", func(c *gin.Context) {
		var in models.ModelCreditRule
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		in.ModelPattern = strings.TrimSpace(in.ModelPattern)
		if in.ModelPattern == "" || in.CreditCost <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "model_pattern and credit_cost are required"})
			return
		}
		if err := app.DB.Create(&in).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create model credit rule"})
			return
		}
		c.JSON(http.StatusOK, in)
	})

	admin.PUT("/model_credit_rules/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var rule models.ModelCreditRule
		if err := app.DB.First(&rule, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "model credit rule not found"})
			return
		}
		if err := c.ShouldBindJSON(&rule); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		rule.ModelPattern = strings.TrimSpace(rule.ModelPattern)
		if rule.ModelPattern == "" || rule.CreditCost <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model credit rule fields"})
			return
		}
		if err := app.DB.Save(&rule).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update model credit rule"})
			return
		}
		c.JSON(http.StatusOK, rule)
	})

	admin.DELETE("/model_credit_rules/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		if err := app.DB.Delete(&models.ModelCreditRule{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete model credit rule"})
			return
		}
		c.Status(http.StatusNoContent)
	})

	admin.GET("/check_in_configs", func(c *gin.Context) {
		var configs []models.CheckInConfig
		if err := app.DB.Order("level ASC").Find(&configs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list check-in configs"})
			return
		}
		c.JSON(http.StatusOK, configs)
	})

	admin.POST("/check_in_configs", func(c *gin.Context) {
		var in models.CheckInConfig
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if !validCheckInConfig(&in) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid check-in config fields"})
			return
		}
		if err := app.DB.Create(&in).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create check-in config"})
			return
		}
		c.JSON(http.StatusOK, in)
	})

	admin.PUT("/check_in_configs/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var cfg models.CheckInConfig
		if err := app.DB.First(&cfg, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "check-in config not found"})
			return
		}
		if err := c.ShouldBindJSON(&cfg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if !validCheckInConfig(&cfg) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid check-in config fields"})
			return
		}
		if err := app.DB.Save(&cfg).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update check-in config"})
			return
		}
		c.JSON(http.StatusOK, cfg)
	})

	admin.DELETE("/check_in_configs/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		if err := app.DB.Delete(&models.CheckInConfig{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete check-in config"})
			return
		}
		c.Status(http.StatusNoContent)
	})

	admin.POST("/quota_rules", func(c *gin.Context) {
		var in models.QuotaRule
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if in.Level <= 0 || in.MaxRequests <= 0 || in.WindowSeconds <= 0 || in.ModelPattern == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quota rule fields"})
			return
		}
		if err := app.DB.Create(&in).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create quota rule"})
			return
		}
		c.JSON(http.StatusOK, in)
	})

	admin.PUT("/quota_rules/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var rule models.QuotaRule
		if err := app.DB.First(&rule, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "quota rule not found"})
			return
		}
		if err := c.ShouldBindJSON(&rule); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if rule.Level <= 0 || rule.MaxRequests <= 0 || rule.WindowSeconds <= 0 || rule.ModelPattern == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quota rule fields"})
			return
		}
		if err := app.DB.Save(&rule).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update quota rule"})
			return
		}
		c.JSON(http.StatusOK, rule)
	})

	admin.DELETE("/quota_rules/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		if err := app.DB.Delete(&models.QuotaRule{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete quota rule"})
			return
		}
		c.Status(http.StatusNoContent)
	})

	// user management
	admin.GET("/users", func(c *gin.Context) {
		var users []models.User
		if err := app.DB.Order("id ASC").Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
			return
		}

		// Return user list without sensitive fields
		result := make([]gin.H, len(users))
		for i, u := range users {
			result[i] = gin.H{
				"id":               u.ID,
				"linuxdo_user_id":  u.LinuxDoUserID,
				"linuxdo_username": u.LinuxDoUsername,
				"role":             u.Role,
				"level":            u.Level,
				"status":           u.Status,
				"credits":          u.Credits,
				"created_at":       u.CreatedAt,
				"updated_at":       u.UpdatedAt,
			}
		}
		c.JSON(http.StatusOK, result)
	})

	admin.PUT("/users/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var user models.User
		if err := app.DB.First(&user, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		// Only allow updating specific fields
		var input struct {
			Role   *string `json:"role"`
			Level  *int    `json:"level"`
			Status *string `json:"status"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}

		// Validate and update role
		if input.Role != nil {
			if *input.Role != models.UserRoleAdmin && *input.Role != models.UserRoleUser {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role, must be 'admin' or 'user'"})
				return
			}
			user.Role = *input.Role
		}

		// Validate and update level
		if input.Level != nil {
			if *input.Level < 1 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "level must be >= 1"})
				return
			}
			user.Level = *input.Level
		}

		// Validate and update status
		if input.Status != nil {
			if *input.Status != models.UserStatusNormal && *input.Status != models.UserStatusDisabled {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status, must be 'normal' or 'disabled'"})
				return
			}
			user.Status = *input.Status
		}

		if err := app.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":               user.ID,
			"linuxdo_user_id":  user.LinuxDoUserID,
			"linuxdo_username": user.LinuxDoUsername,
			"role":             user.Role,
			"level":            user.Level,
			"status":           user.Status,
			"updated_at":       user.UpdatedAt,
		})
	})

	admin.POST("/users/:id/credits", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var user models.User
		if err := app.DB.First(&user, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		var input struct {
			Delta  int    `json:"delta"`
			Reason string `json:"reason"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if input.Delta == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "delta must not be zero"})
			return
		}

		reason := strings.TrimSpace(input.Reason)
		if reason == "" {
			reason = creditReasonManualAdjust
		}

		balance, err := adjustUserCredits(app, uint(id), input.Delta, reason)
		if err != nil {
			if errors.Is(err, errInsufficientCredits) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient credits for deduction"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to adjust credits"})
			return
		}

		recordOperationLog(app, uint(id), "credit_adjust", fmt.Sprintf("delta=%d reason=%s", input.Delta, reason))
		c.JSON(http.StatusOK, gin.H{"user_id": id, "credits": balance})
	})

	// global stats & logs for admin dashboard
	admin.GET("/stats", func(c *gin.Context) {
		var userCount int64
		if err := app.DB.Model(&models.User{}).Count(&userCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count users"})
			return
		}

		var totalRequests int64
		if err := app.DB.Model(&models.APILog{}).Count(&totalRequests).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count requests"})
			return
		}

		var activeUsers int64
		if err := app.DB.Model(&models.APILog{}).
			Select("count(distinct user_id)").
			Where("created_at > now() - interval '24 hours'").
			Count(&activeUsers).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count active users"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total_users":      userCount,
			"total_requests":   totalRequests,
			"active_users_24h": activeUsers,
		})
	})

	admin.GET("/api_logs", func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		pageSizeStr := c.DefaultQuery("page_size", "20")
		userIDStr := c.Query("user_id")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			page = 1
		}
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize <= 0 || pageSize > 100 {
			pageSize = 20
		}

		db := app.DB.Model(&models.APILog{})
		if userIDStr != "" {
			if uid, err := strconv.Atoi(userIDStr); err == nil {
				db = db.Where("user_id = ?", uid)
			}
		}

		var total int64
		if err := db.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count api logs"})
			return
		}

		var logs []models.APILog
		if err := db.Order("created_at DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&logs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list api logs"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"total": total, "items": logs})
	})

	admin.GET("/credit_transactions", func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		pageSizeStr := c.DefaultQuery("page_size", "20")
		userIDStr := c.Query("user_id")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			page = 1
		}
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize <= 0 || pageSize > 100 {
			pageSize = 20
		}

		db := app.DB.Model(&models.CreditTransaction{})
		if userIDStr != "" {
			if uid, err := strconv.Atoi(userIDStr); err == nil {
				db = db.Where("user_id = ?", uid)
			}
		}

		var total int64
		if err := db.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count credit transactions"})
			return
		}

		var txns []models.CreditTransaction
		if err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&txns).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list credit transactions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"total": total, "items": txns})
	})

	admin.GET("/login_logs", func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		pageSizeStr := c.DefaultQuery("page_size", "20")
		userIDStr := c.Query("user_id")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			page = 1
		}
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize <= 0 || pageSize > 100 {
			pageSize = 20
		}

		db := app.DB.Model(&models.LoginLog{})
		if userIDStr != "" {
			if uid, err := strconv.Atoi(userIDStr); err == nil {
				db = db.Where("user_id = ?", uid)
			}
		}

		var total int64
		if err := db.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count login logs"})
			return
		}

		var logs []models.LoginLog
		if err := db.Order("created_at DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&logs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list login logs"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"total": total, "items": logs})
	})
}
