package server

import (
	"testing"

	"linuxdo-relay/internal/models"
)

func TestSelectCreditCostPrefersLongestPrefix(t *testing.T) {
	rules := []models.ModelCreditRule{
		{ModelPattern: "gpt-", CreditCost: 3},
		{ModelPattern: "gpt-4", CreditCost: 6},
		{ModelPattern: "claude", CreditCost: 4},
	}

	cost := selectCreditCost("gpt-4o-mini", rules, 1)
	if cost != 6 {
		t.Fatalf("expected longest prefix cost 6, got %d", cost)
	}
}

func TestSelectCreditCostFallsBackToDefault(t *testing.T) {
	rules := []models.ModelCreditRule{{ModelPattern: "claude", CreditCost: 5}}
	cost := selectCreditCost("gpt-4o", rules, 2)
	if cost != 2 {
		t.Fatalf("expected default cost 2, got %d", cost)
	}
}

func TestSelectCreditCostHandlesEmptyPattern(t *testing.T) {
	rules := []models.ModelCreditRule{{ModelPattern: "", CreditCost: 7}}
	cost := selectCreditCost("any", rules, 1)
	if cost != 7 {
		t.Fatalf("expected empty-pattern override to 7, got %d", cost)
	}
}
