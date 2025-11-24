package server

import (
	"fmt"
	"time"

	"linuxdo-relay/internal/models"

	"github.com/gin-gonic/gin"
)

// nonBlockingSave ignores errors and ensures that logging failures never break
// main request handling paths.
func nonBlockingSave(err error) {
	if err != nil {
		fmt.Println("log insert failed:", err)
	}
}

func recordAPILog(app *AppContext, userID uint, model, status string, statusCode int, errorMessage, ip string) {
	if app == nil || app.DB == nil {
		return
	}
	log := &models.APILog{
		UserID:       userID,
		Model:        model,
		Status:       status,
		StatusCode:   statusCode,
		ErrorMessage: errorMessage,
		IPAddress:    ip,
		CreatedAt:    time.Now(),
	}
	nonBlockingSave(app.DB.Create(log).Error)
}

// recordAPILogFromContext is a convenience helper for relay routes to record
// a single API call using values from gin.Context.
func recordAPILogFromContext(app *AppContext, c *gin.Context, model string, statusCode int, status, errorMessage string) {
	if c == nil {
		return
	}
	uidVal, _ := c.Get("user_id")
	userID, _ := uidVal.(uint)
	ip := ""
	if c.Request != nil {
		ip = c.ClientIP()
	}
	recordAPILog(app, userID, model, status, statusCode, errorMessage, ip)
}

func recordOperationLog(app *AppContext, userID uint, opType, details string) {
	if app == nil || app.DB == nil {
		return
	}
	log := &models.OperationLog{
		UserID:        userID,
		OperationType: opType,
		Details:       details,
		CreatedAt:     time.Now(),
	}
	nonBlockingSave(app.DB.Create(log).Error)
}

func recordLoginLog(app *AppContext, userID uint, ip, userAgent string) {
	if app == nil || app.DB == nil {
		return
	}
	log := &models.LoginLog{
		UserID:    userID,
		IPAddress: ip,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	}
	nonBlockingSave(app.DB.Create(log).Error)
}
