package server

import (
	"errors"
	"math"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"linuxdo-relay/internal/models"
)

var (
	errAlreadyCheckedIn = errors.New("already checked in today")
	cstLocation         = time.FixedZone("UTC+8", 8*3600)
)

type checkInResult struct {
	Reward       int
	Streak       int
	Credits      int
	CheckInDate  time.Time
	CheckedToday bool
	Config       *models.CheckInConfig
}

func determineCheckInDate(now time.Time) time.Time {
	local := now.In(cstLocation)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

func loadCheckInConfig(tx *gorm.DB, level int) (*models.CheckInConfig, error) {
	var cfg models.CheckInConfig
	if err := tx.Where("level <= ?", level).Order("level DESC").First(&cfg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err2 := tx.Order("level ASC").First(&cfg).Error; err2 != nil {
				return nil, err2
			}
			return &cfg, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func calculateCheckInReward(cfg *models.CheckInConfig, currentCredits int) int {
	if cfg == nil || cfg.BaseReward <= 0 {
		return 0
	}

	// 如果余额低于阈值，给予全额奖励
	if currentCredits < cfg.DecayThreshold {
		return cfg.BaseReward
	}

	// 计算超出阈值的部分
	excess := currentCredits - cfg.DecayThreshold

	// 每超出 100 积分，衰减 5%（可调整此比例）
	decayRate := 5.0 // 每 100 积分衰减 5%
	decayPercent := (float64(excess) / 100.0) * decayRate

	// 计算当前倍数：100% - 衰减百分比
	multiplierPercent := 100.0 - decayPercent

	// 限制在最低倍数和 100% 之间
	minMultiplier := float64(cfg.MinMultiplierPercent)
	if multiplierPercent < minMultiplier {
		multiplierPercent = minMultiplier
	}
	if multiplierPercent > 100 {
		multiplierPercent = 100
	}

	// 计算最终奖励
	reward := int(math.Round(float64(cfg.BaseReward) * multiplierPercent / 100.0))
	if reward < 1 {
		reward = 1
	}
	return reward
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

func performDailyCheckIn(app *AppContext, userID uint) (*checkInResult, error) {
	if app == nil || app.DB == nil {
		return nil, errors.New("app context missing")
	}
	var result checkInResult
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

		cfg, err := loadCheckInConfig(tx, user.Level)
		if err != nil {
			return err
		}

		reward := calculateCheckInReward(cfg, user.Credits)
		if reward <= 0 {
			return errors.New("calculated reward is zero")
		}

		yesterday := today.AddDate(0, 0, -1)
		streak := 1
		if log, err := fetchCheckInLog(tx, userID, yesterday); err == nil {
			streak = log.Streak + 1
		}

		credits, err := adjustUserCreditsTx(tx, userID, reward, creditReasonCheckIn, nil)
		if err != nil {
			return err
		}

		entry := models.CheckInLog{
			UserID:        userID,
			CheckInDate:   today,
			EarnedCredits: reward,
			Streak:        streak,
		}
		if err := tx.Create(&entry).Error; err != nil {
			return err
		}

		result = checkInResult{
			Reward:       reward,
			Streak:       streak,
			Credits:      credits,
			CheckInDate:  today,
			CheckedToday: true,
			Config:       cfg,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func loadTodayCheckInStatus(app *AppContext, userID uint, level int, credits int) (*checkInResult, error) {
	if app == nil || app.DB == nil {
		return nil, errors.New("app context missing")
	}
	today := determineCheckInDate(time.Now())

	var log models.CheckInLog
	err := app.DB.Where("user_id = ? AND check_in_date = ?", userID, today).First(&log).Error
	if err == nil {
		cfg, cfgErr := loadCheckInConfig(app.DB.DB, level)
		if cfgErr != nil {
			return nil, cfgErr
		}
		return &checkInResult{
			Reward:       log.EarnedCredits,
			Streak:       log.Streak,
			Credits:      credits,
			CheckInDate:  today,
			CheckedToday: true,
			Config:       cfg,
		}, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	cfg, err := loadCheckInConfig(app.DB.DB, level)
	if err != nil {
		return nil, err
	}

	reward := calculateCheckInReward(cfg, credits)
	yesterday := today.AddDate(0, 0, -1)
	streak := 0
	if yLog, err := fetchCheckInLog(app.DB.DB, userID, yesterday); err == nil {
		streak = yLog.Streak
	}

	return &checkInResult{
		Reward:       reward,
		Streak:       streak,
		Credits:      credits,
		CheckInDate:  today,
		CheckedToday: false,
		Config:       cfg,
	}, nil
}
