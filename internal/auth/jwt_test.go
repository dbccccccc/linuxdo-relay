package auth

import (
	"testing"
	"time"
)

func TestGenerateAndParseToken(t *testing.T) {
	secret := "unit-test-secret"
	token, err := GenerateToken(secret, 10, "admin", 2, time.Minute)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	claims, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("parse token failed: %v", err)
	}
	if claims.UserID != 10 || claims.Role != "admin" || claims.Level != 2 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestParseTokenRejectsWrongSecret(t *testing.T) {
	secret := "unit-test-secret"
	token, err := GenerateToken(secret, 1, "user", 1, time.Minute)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	if _, err := ParseToken("other-secret", token); err == nil {
		t.Fatalf("expected parse error with wrong secret")
	}
}
