package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPListen    string
	PostgresDSN   string
	RedisAddr     string
	RedisPassword string

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

func Load() (*Config, error) {
	cfg := &Config{
		HTTPListen:             getEnv("APP_HTTP_LISTEN", ":8080"),
		PostgresDSN:            os.Getenv("APP_PG_DSN"),
		RedisAddr:              os.Getenv("APP_REDIS_ADDR"),
		RedisPassword:          os.Getenv("APP_REDIS_PASSWORD"),
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

	// Validate required environment variables
	if cfg.PostgresDSN == "" {
		return nil, fmt.Errorf("APP_PG_DSN is required")
	}
	if cfg.RedisAddr == "" {
		return nil, fmt.Errorf("APP_REDIS_ADDR is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("APP_JWT_SECRET is required")
	}

	if cfg.SignupCredits < 0 {
		cfg.SignupCredits = 0
	}
	if cfg.DefaultModelCreditCost < 0 {
		cfg.DefaultModelCreditCost = 0
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
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
