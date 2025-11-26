package server

import (
	crand "crypto/rand"
	"errors"
	"math/big"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"linuxdo-relay/internal/models"
)

var (
	errAlreadyCheckedIn = errors.New("already checked in today")
	cstLocation         = time.FixedZone("UTC+8", 8*3600)
)

type checkInStatus struct {
	CheckedToday bool
	Reward       int
	Streak       int
	Credits      int
}

type checkInSpinResult struct {
	RewardOption      models.CheckInRewardOption
	MultiplierPercent int
	FinalReward       int
	Streak            int
	Credits           int
	CheckInDate       time.Time
	CheckedToday      bool
	WheelIndex        int
}

func determineCheckInDate(now time.Time) time.Time {
	local := now.In(cstLocation)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

func fetchCheckInLog(tx *gorm.DB, userID uint, date time.Time) (*models.CheckInLog, error) {
	var log models.CheckInLog
	if err := tx.Where("user_id = ? AND check_in_date = ?", userID, date).First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func fetchRecentCheckInLogs(app *AppContext, userID uint, limit int) ([]models.CheckInLog, error) {
	if limit <= 0 {
		limit = 7
	}
	var logs []models.CheckInLog
	if err := app.DB.Where("user_id = ?", userID).
		Order("check_in_date DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func loadRewardOptions(tx *gorm.DB) ([]models.CheckInRewardOption, error) {
	var options []models.CheckInRewardOption
	if err := tx.Order("sort_order ASC, id ASC").Find(&options).Error; err != nil {
		return nil, err
	}
	if len(options) == 0 {
		return nil, errors.New("no check-in reward options configured")
	}
	return options, nil
}

func loadDecayRules(tx *gorm.DB) ([]models.CheckInDecayRule, error) {
	var rules []models.CheckInDecayRule
	if err := tx.Order("sort_order ASC, threshold ASC").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func randomInt(max int) int {
	if max <= 0 {
		return 0
	}
	n, err := crand.Int(crand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0
	}
	return int(n.Int64())
}

func selectRewardOption(options []models.CheckInRewardOption) (models.CheckInRewardOption, int) {
	total := 0
	for _, opt := range options {
		if opt.Probability > 0 {
			total += opt.Probability
		}
	}
	if total <= 0 {
		total = len(options)
	}
	randVal := randomInt(total)
	cumulative := 0
	for idx, opt := range options {
		weight := opt.Probability
		if weight <= 0 {
			weight = 1
		}
		cumulative += weight
		if randVal < cumulative {
			return opt, idx
		}
	}
	return options[len(options)-1], len(options) - 1
}

func calculateMultiplier(rules []models.CheckInDecayRule, currentCredits int) int {
	multiplier := 100
	for _, rule := range rules {
		if currentCredits > rule.Threshold {
			multiplier = rule.MultiplierPercent
		} else {
			break
		}
	}
	if multiplier <= 0 {
		return 1
	}
	if multiplier > 100 {
		return 100
	}
	return multiplier
}

func performDailyCheckIn(app *AppContext, userID uint) (*checkInSpinResult, error) {
	if app == nil || app.DB == nil {
		return nil, errors.New("app context missing")
	}

	var result checkInSpinResult
	err := app.DB.Transaction(func(tx *gorm.DB) error {
		today := determineCheckInDate(time.Now())

		if _, err := fetchCheckInLog(tx, userID, today); err == nil {
			return errAlreadyCheckedIn
		} else if err != gorm.ErrRecordNotFound {
			return err
		}

		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}

		rewardOptions, err := loadRewardOptions(tx)
		if err != nil {
			return err
		}
		decayRules, err := loadDecayRules(tx)
		if err != nil {
			return err
		}

		selected, idx := selectRewardOption(rewardOptions)
		multiplier := calculateMultiplier(decayRules, user.Credits)
		finalReward := selected.Credits * multiplier / 100
		if finalReward < 1 {
			finalReward = 1
		}

		yesterday := today.AddDate(0, 0, -1)
		streak := 1
		if log, err := fetchCheckInLog(tx, userID, yesterday); err == nil {
			streak = log.Streak + 1
		}

		credits, err := adjustUserCreditsTx(tx, userID, finalReward, creditReasonCheckIn, nil)
		if err != nil {
			return err
		}

		entry := models.CheckInLog{
			UserID:        userID,
			CheckInDate:   today,
			EarnedCredits: finalReward,
			Streak:        streak,
		}
		if err := tx.Create(&entry).Error; err != nil {
			return err
		}

		result = checkInSpinResult{
			RewardOption:      selected,
			MultiplierPercent: multiplier,
			FinalReward:       finalReward,
			Streak:            streak,
			Credits:           credits,
			CheckInDate:       today,
			CheckedToday:      true,
			WheelIndex:        idx,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func loadTodayCheckInStatus(app *AppContext, userID uint, credits int) (*checkInStatus, error) {
	if app == nil || app.DB == nil {
		return nil, errors.New("app context missing")
	}
	today := determineCheckInDate(time.Now())

	var log models.CheckInLog
	err := app.DB.Where("user_id = ? AND check_in_date = ?", userID, today).First(&log).Error
	if err == nil {
		return &checkInStatus{
			CheckedToday: true,
			Reward:       log.EarnedCredits,
			Streak:       log.Streak,
			Credits:      credits,
		}, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	yesterday := today.AddDate(0, 0, -1)
	streak := 0
	if yLog, err := fetchCheckInLog(app.DB.DB, userID, yesterday); err == nil {
		streak = yLog.Streak
	}

	return &checkInStatus{
		CheckedToday: false,
		Reward:       0,
		Streak:       streak,
		Credits:      credits,
	}, nil
}
