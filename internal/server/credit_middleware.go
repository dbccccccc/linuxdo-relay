package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"linuxdo-relay/internal/models"
)

const (
	creditReasonModelRequest = "model_request"
	creditReasonManualAdjust = "manual_adjust"
	creditReasonRefund       = "model_refund"
	creditReasonCheckIn      = "daily_check_in"

	creditStatusReserved  = "reserved"
	creditStatusCommitted = "committed"
	creditStatusReverted  = "reverted"
)

var errInsufficientCredits = errors.New("insufficient credits")

// CreditMiddleware reserves per-request credits before proxying upstream. On
// failure responses the reservation is refunded.
func CreditMiddleware(app *AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isRelayPath(c) {
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

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		model := extractModelForQuota(c, path)
		if model == "" {
			c.Next()
			return
		}

		cost, err := determineCreditCost(app, model)
		if err != nil {
			fmt.Println("credit: failed to determine cost:", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "credit_cost_lookup_failed"})
			return
		}
		if cost <= 0 {
			c.Next()
			return
		}

		requestID := uuid.NewString()
		txnID, err := reserveCreditsForRequest(app, userID, model, cost, requestID)
		if err != nil {
			if errors.Is(err, errInsufficientCredits) {
				c.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{
					"error":   "credit_insufficient",
					"message": "not enough credits for this model",
				})
				return
			}
			fmt.Println("credit: reserve failed:", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "credit_reserve_failed"})
			return
		}

		c.Set("credit_cost", cost)
		c.Set("credit_request_id", requestID)
		c.Set("credit_txn_id", txnID)
		c.Set("credit_model", model)

		c.Next()

		statusCode := c.Writer.Status()
		if statusCode >= 200 && statusCode < 300 {
			commitReservedCredits(app, txnID)
			return
		}

		refundReservedCredits(app, txnID, userID, cost)
	}
}

func isRelayPath(c *gin.Context) bool {
	path := c.FullPath()
	if path == "" && c.Request != nil {
		path = c.Request.URL.Path
	}
	return strings.HasPrefix(path, "/v1/chat/completions") ||
		strings.HasPrefix(path, "/v1/messages") ||
		strings.HasPrefix(path, "/v1beta/models/")
}

func determineCreditCost(app *AppContext, model string) (int, error) {
	if app == nil || app.DB == nil || model == "" {
		return 0, nil
	}
	var rules []models.ModelCreditRule
	if err := app.DB.Order("model_pattern ASC").Find(&rules).Error; err != nil {
		return 0, err
	}
	defaultCost := 0
	if app.Config != nil {
		defaultCost = app.Config.DefaultModelCreditCost
	}
	return selectCreditCost(model, rules, defaultCost), nil
}

func selectCreditCost(model string, rules []models.ModelCreditRule, defaultCost int) int {
	bestCost := defaultCost
	bestLen := -1
	for _, rule := range rules {
		pattern := rule.ModelPattern
		if pattern == "" {
			if bestLen < 0 {
				bestCost = rule.CreditCost
				bestLen = 0
			}
			continue
		}
		if strings.HasPrefix(model, pattern) && len(pattern) > bestLen {
			bestLen = len(pattern)
			bestCost = rule.CreditCost
		}
	}
	if bestCost < 0 {
		bestCost = 0
	}
	return bestCost
}

func reserveCreditsForRequest(app *AppContext, userID uint, model string, cost int, requestID string) (uint, error) {
	if app == nil || app.DB == nil {
		return 0, fmt.Errorf("app context missing")
	}
	if cost <= 0 {
		return 0, nil
	}
	var txnID uint
	err := app.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.User{}).
			Where("id = ? AND credits >= ?", userID, cost).
			UpdateColumn("credits", gorm.Expr("credits - ?", cost))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errInsufficientCredits
		}
		txn := models.CreditTransaction{
			UserID:    userID,
			Delta:     -cost,
			Reason:    creditReasonModelRequest,
			Status:    creditStatusReserved,
			ModelName: model,
			RequestID: requestID,
		}
		if err := tx.Create(&txn).Error; err != nil {
			return err
		}
		txnID = txn.ID
		return nil
	})
	return txnID, err
}

func commitReservedCredits(app *AppContext, txnID uint) {
	if app == nil || app.DB == nil || txnID == 0 {
		return
	}
	if err := app.DB.Model(&models.CreditTransaction{}).
		Where("id = ? AND status = ?", txnID, creditStatusReserved).
		Updates(map[string]interface{}{
			"status":     creditStatusCommitted,
			"updated_at": time.Now(),
		}).Error; err != nil {
		fmt.Println("credit: commit failed:", err)
	}
}

func refundReservedCredits(app *AppContext, txnID uint, userID uint, cost int) {
	if app == nil || app.DB == nil || txnID == 0 || cost <= 0 {
		return
	}
	err := app.DB.Transaction(func(tx *gorm.DB) error {
		var txn models.CreditTransaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&txn, txnID).Error; err != nil {
			return err
		}
		if txn.Status != creditStatusReserved {
			return nil
		}
		if err := tx.Model(&models.User{}).
			Where("id = ?", userID).
			UpdateColumn("credits", gorm.Expr("credits + ?", cost)).Error; err != nil {
			return err
		}
		txn.Status = creditStatusReverted
		txn.UpdatedAt = time.Now()
		if err := tx.Save(&txn).Error; err != nil {
			return err
		}
		refundTxn := models.CreditTransaction{
			UserID:    userID,
			Delta:     cost,
			Reason:    creditReasonRefund,
			Status:    creditStatusCommitted,
			ModelName: txn.ModelName,
		}
		return tx.Create(&refundTxn).Error
	})
	if err != nil {
		fmt.Println("credit: refund failed:", err)
	}
}
