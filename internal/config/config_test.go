package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"linuxdo-relay/internal/runtimeconfig"
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

func TestLoadConfigRequiresRequiredEnv(t *testing.T) {
	useTempRuntimePath(t)
	t.Setenv("APP_PG_DSN", "")
	t.Setenv("APP_JWT_SECRET", "secret")

	cfg, err := Load()
	if !errors.Is(err, ErrSetupRequired) {
		t.Fatalf("expected ErrSetupRequired, got %v", err)
	}
	if cfg == nil {
		t.Fatalf("expected cfg even when setup required")
	}
}

func TestLoadConfigAppliesDefaults(t *testing.T) {
	useTempRuntimePath(t)
	dsn := "postgres://user:pass@localhost:5432/db?sslmode=disable"
	secret := "test-secret"

	t.Setenv("APP_PG_DSN", dsn)
	t.Setenv("APP_JWT_SECRET", secret)
	unsetEnv(t, "APP_HTTP_LISTEN")
	unsetEnv(t, "APP_REDIS_ADDR")
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
	if cfg.RedisAddr != "localhost:6379" {
		t.Fatalf("expected default redis addr, got %s", cfg.RedisAddr)
	}
	if cfg.SignupCredits != 100 {
		t.Fatalf("expected default signup credits, got %d", cfg.SignupCredits)
	}
	if cfg.DefaultModelCreditCost != 1 {
		t.Fatalf("expected default model cost, got %d", cfg.DefaultModelCreditCost)
	}
}

func TestLoadConfigParsesInts(t *testing.T) {
	useTempRuntimePath(t)
	t.Setenv("APP_PG_DSN", "dsn")
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

func TestLoadConfigFallsBackToRuntimeConfig(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")
	t.Setenv("APP_RUNTIME_CONFIG_PATH", path)
	t.Setenv("APP_JWT_SECRET", "secret")
	unsetEnv(t, "APP_PG_DSN")

	store := runtimeconfig.NewStore(path)
	runtimeCfg := &runtimeconfig.Data{
		Database: runtimeconfig.DatabaseConfig{DSN: "postgres://runtime"},
		Redis:    runtimeconfig.RedisConfig{Addr: "runtime-redis:6379"},
	}
	if err := store.Save(runtimeCfg); err != nil {
		t.Fatalf("save runtime config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PostgresDSN != "postgres://runtime" {
		t.Fatalf("expected runtime DSN, got %s", cfg.PostgresDSN)
	}
	if cfg.RedisAddr != "runtime-redis:6379" {
		t.Fatalf("expected runtime redis, got %s", cfg.RedisAddr)
	}
}

func useTempRuntimePath(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("APP_RUNTIME_CONFIG_PATH", path)
}
