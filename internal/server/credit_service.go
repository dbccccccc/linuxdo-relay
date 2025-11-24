package server

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"linuxdo-relay/internal/models"
)

type creditAdjustmentOptions struct {
	ModelName string
	RequestID string
}

func adjustUserCredits(app *AppContext, userID uint, delta int, reason string) (int, error) {
	if app == nil || app.DB == nil {
		return 0, fmt.Errorf("app context missing")
	}
	var balance int
	err := app.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		balance, err = adjustUserCreditsTx(tx, userID, delta, reason, nil)
		return err
	})
	return balance, err
}

func adjustUserCreditsTx(tx *gorm.DB, userID uint, delta int, reason string, opts *creditAdjustmentOptions) (int, error) {
	if tx == nil {
		return 0, fmt.Errorf("tx is required")
	}
	if delta == 0 {
		return 0, fmt.Errorf("delta must not be zero")
	}
	if reason == "" {
		reason = creditReasonManualAdjust
	}

	query := tx.Model(&models.User{}).Where("id = ?", userID)
	if delta < 0 {
		query = query.Where("credits >= ?", -delta)
	}
	res := query.UpdateColumn("credits", gorm.Expr("credits + ?", delta))
	if res.Error != nil {
		return 0, res.Error
	}
	if res.RowsAffected == 0 {
		return 0, errInsufficientCredits
	}

	var user models.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
		return 0, err
	}

	txn := models.CreditTransaction{
		UserID:    userID,
		Delta:     delta,
		Reason:    reason,
		Status:    creditStatusCommitted,
		ModelName: optsValue(opts, func(o *creditAdjustmentOptions) string { return o.ModelName }),
		RequestID: optsValue(opts, func(o *creditAdjustmentOptions) string { return o.RequestID }),
	}
	if err := tx.Create(&txn).Error; err != nil {
		return 0, err
	}

	return user.Credits, nil
}

func optsValue[T any](opts *creditAdjustmentOptions, getter func(*creditAdjustmentOptions) T) T {
	var zero T
	if opts == nil {
		return zero
	}
	return getter(opts)
}
