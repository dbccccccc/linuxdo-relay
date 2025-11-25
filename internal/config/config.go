package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"linuxdo-relay/internal/runtimeconfig"
)

type Config struct {
	HTTPListen        string
	PostgresDSN       string
	RedisAddr         string
	RedisPassword     string
	RuntimeConfigPath string
	MigrationsDir     string

	LinuxDoClientID     string
	LinuxDoClientSecret string
	LinuxDoAuthURL      string
	LinuxDoTokenURL     string
	LinuxDoUserInfoURL  string
	LinuxDoRedirectURL  string

	JWTSecret string

	SignupCredits          int
	DefaultModelCreditCost int
}

var ErrSetupRequired = errors.New("runtime setup required")

func Load() (*Config, error) {
	runtimeStore := runtimeconfig.NewStore(runtimeconfig.DefaultPath())
	runtimeData, err := runtimeStore.Load()
	if err != nil && !errors.Is(err, runtimeconfig.ErrNotFound) {
		return nil, err
	}

	cfg := &Config{
		HTTPListen:             getEnv("APP_HTTP_LISTEN", ":8080"),
		PostgresDSN:            os.Getenv("APP_PG_DSN"),
		RedisAddr:              getEnv("APP_REDIS_ADDR", ""),
		RedisPassword:          os.Getenv("APP_REDIS_PASSWORD"),
		RuntimeConfigPath:      runtimeStore.Path,
		MigrationsDir:          getEnv("APP_MIGRATIONS_DIR", "migrations"),
		LinuxDoClientID:        os.Getenv("APP_LINUXDO_CLIENT_ID"),
		LinuxDoClientSecret:    os.Getenv("APP_LINUXDO_CLIENT_SECRET"),
		LinuxDoAuthURL:         getEnv("APP_LINUXDO_AUTH_URL", "https://connect.linux.do/oauth2/authorize"),
		LinuxDoTokenURL:        getEnv("APP_LINUXDO_TOKEN_URL", "https://connect.linux.do/oauth2/token"),
		LinuxDoUserInfoURL:     getEnv("APP_LINUXDO_USERINFO_URL", "https://connect.linux.do/api/user"),
		LinuxDoRedirectURL:     os.Getenv("APP_LINUXDO_REDIRECT_URL"),
		JWTSecret:              os.Getenv("APP_JWT_SECRET"),
		SignupCredits:          getEnvInt("APP_SIGNUP_CREDITS", 100),
		DefaultModelCreditCost: getEnvInt("APP_DEFAULT_MODEL_CREDIT_COST", 1),
	}

	if runtimeData != nil {
		if cfg.PostgresDSN == "" {
			cfg.PostgresDSN = runtimeData.Database.DSN
		}
		if cfg.RedisAddr == "" {
			cfg.RedisAddr = runtimeData.Redis.Addr
		}
		if cfg.RedisPassword == "" {
			cfg.RedisPassword = runtimeData.Redis.Password
		}
	}

	if cfg.RedisAddr == "" {
		cfg.RedisAddr = "localhost:6379"
	}

	if cfg.SignupCredits < 0 {
		cfg.SignupCredits = 0
	}
	if cfg.DefaultModelCreditCost < 0 {
		cfg.DefaultModelCreditCost = 0
	}

	// If database is not configured, enter setup mode
	// JWTSecret check is skipped in setup mode
	if cfg.PostgresDSN == "" {
		return cfg, ErrSetupRequired
	}

	// Only require JWTSecret when database is configured (normal mode)
	if cfg.JWTSecret == "" {
		return cfg, fmt.Errorf("APP_JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
