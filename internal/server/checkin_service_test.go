package server

import (
	"testing"

	"linuxdo-relay/internal/models"
)

func TestCalculateCheckInReward(t *testing.T) {
	cfg := &models.CheckInConfig{
		BaseReward:           10,
		DecayThreshold:       1000, // 积分余额阈值
		MinMultiplierPercent: 50,   // 最低 50%
	}

	tests := []struct {
		name           string
		currentCredits int
		expectedReward int
		description    string
	}{
		{"余额为0", 0, 10, "低于阈值，全额奖励"},
		{"余额500", 500, 10, "低于阈值，全额奖励"},
		{"余额999", 999, 10, "低于阈值，全额奖励"},
		{"余额1000", 1000, 10, "刚好阈值，全额奖励"},
		{"余额1100", 1100, 10, "超出100，衰减5%，奖励9.5→10"},
		{"余额1200", 1200, 9, "超出200，衰减10%，奖励9"},
		{"余额1500", 1500, 8, "超出500，衰减25%，奖励7.5→8"},
		{"余额2000", 2000, 5, "超出1000，衰减50%，触及最低50%"},
		{"余额5000", 5000, 5, "超出很多，但不低于最低50%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reward := calculateCheckInReward(cfg, tt.currentCredits)
			if reward != tt.expectedReward {
				t.Errorf("%s: 期望 %d，实际 %d", tt.description, tt.expectedReward, reward)
			}
		})
	}
}
