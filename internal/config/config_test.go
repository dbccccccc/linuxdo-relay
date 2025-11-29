package config

import (
	"os"
	"testing"
)

func unsetEnv(t *testing.T, key string) {
	orig, existed := os.LookupEnv(key)
	if existed {
		t.Cleanup(func() {
			_ = os.Setenv(key, orig)
		})
	} else {
		t.Cleanup(func() {
			_ = os.Unsetenv(key)
		})
	}
	_ = os.Unsetenv(key)
}

func TestLoadConfigRequiresPostgresDSN(t *testing.T) {
	t.Setenv("APP_PG_DSN", "")
	t.Setenv("APP_REDIS_ADDR", "localhost:6379")
	t.Setenv("APP_JWT_SECRET", "secret")

	_, err := Load()
	if err == nil {
		t.Fatalf("expected error for missing APP_PG_DSN")
	}
}

func TestLoadConfigRequiresRedisAddr(t *testing.T) {
	t.Setenv("APP_PG_DSN", "postgres://user:pass@localhost:5432/db")
	t.Setenv("APP_REDIS_ADDR", "")
	t.Setenv("APP_JWT_SECRET", "secret")

	_, err := Load()
	if err == nil {
		t.Fatalf("expected error for missing APP_REDIS_ADDR")
	}
}

func TestLoadConfigRequiresJWTSecret(t *testing.T) {
	t.Setenv("APP_PG_DSN", "postgres://user:pass@localhost:5432/db")
	t.Setenv("APP_REDIS_ADDR", "localhost:6379")
	t.Setenv("APP_JWT_SECRET", "")

	_, err := Load()
	if err == nil {
		t.Fatalf("expected error for missing APP_JWT_SECRET")
	}
}

func TestLoadConfigAppliesDefaults(t *testing.T) {
	dsn := "postgres://user:pass@localhost:5432/db?sslmode=disable"
	secret := "test-secret"

	t.Setenv("APP_PG_DSN", dsn)
	t.Setenv("APP_REDIS_ADDR", "localhost:6379")
	t.Setenv("APP_JWT_SECRET", secret)
	unsetEnv(t, "APP_HTTP_LISTEN")
	unsetEnv(t, "APP_SIGNUP_CREDITS")
	unsetEnv(t, "APP_DEFAULT_MODEL_CREDIT_COST")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.PostgresDSN != dsn {
		t.Fatalf("unexpected dsn: %s", cfg.PostgresDSN)
	}
	if cfg.JWTSecret != secret {
		t.Fatalf("unexpected secret")
	}
	if cfg.HTTPListen != ":8080" {
		t.Fatalf("expected default listen, got %s", cfg.HTTPListen)
	}
	if cfg.SignupCredits != 100 {
		t.Fatalf("expected default signup credits, got %d", cfg.SignupCredits)
	}
	if cfg.DefaultModelCreditCost != 1 {
		t.Fatalf("expected default model cost, got %d", cfg.DefaultModelCreditCost)
	}
}

func TestLoadConfigParsesInts(t *testing.T) {
	t.Setenv("APP_PG_DSN", "dsn")
	t.Setenv("APP_REDIS_ADDR", "localhost:6379")
	t.Setenv("APP_JWT_SECRET", "secret")
	t.Setenv("APP_SIGNUP_CREDITS", "250")
	t.Setenv("APP_DEFAULT_MODEL_CREDIT_COST", "5")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SignupCredits != 250 {
		t.Fatalf("expected signup credits 250, got %d", cfg.SignupCredits)
	}
	if cfg.DefaultModelCreditCost != 5 {
		t.Fatalf("expected model cost 5, got %d", cfg.DefaultModelCreditCost)
	}
}
