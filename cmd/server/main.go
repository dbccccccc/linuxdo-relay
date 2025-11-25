package main

import (
	"context"
	"errors"
	"log"

	"github.com/gin-gonic/gin"

	"linuxdo-relay/internal/auth"
	"linuxdo-relay/internal/config"
	"linuxdo-relay/internal/logger"
	"linuxdo-relay/internal/server"
	"linuxdo-relay/internal/storage"
	"linuxdo-relay/internal/storage/migrate"
)

func main() {
	cfg, err := config.Load()
	setupMode := false
	if err != nil {
		if errors.Is(err, config.ErrSetupRequired) {
			setupMode = true
		} else {
			log.Fatalf("load config: %v", err)
		}
	}

	r := gin.Default()

	if setupMode {
		server.RegisterSetupWizardRoutes(r, cfg)
	} else {
		// init DB with connection pooling
		db, err := storage.OpenDB(cfg.PostgresDSN)
		if err != nil {
			log.Fatalf("open database: %v", err)
		}
		logger.Info("database connected", "dsn", "***")

		// init Redis with connection verification
		redisClient, err := storage.NewRedisWithPing(cfg.RedisAddr, cfg.RedisPassword)
		if err != nil {
			log.Fatalf("connect redis: %v", err)
		}
		logger.Info("redis connected", "addr", cfg.RedisAddr)

		runner := migrate.NewRunner(db.DB, cfg.MigrationsDir)
		if _, err := runner.ApplyPending(context.Background(), false); err != nil {
			log.Fatalf("apply migrations: %v", err)
		}

		// init OAuth config
		oauthCfg := auth.NewLinuxDoOAuthConfig(cfg)

		app := &server.AppContext{
			Config:    cfg,
			DB:        db,
			Redis:     redisClient,
			OAuth:     oauthCfg,
			JWTSecret: cfg.JWTSecret,
		}

		server.SetupRoutes(r, app)
	}

	addr := cfg.HTTPListen
	if addr == "" {
		addr = ":8080"
	}

	logger.Info("starting server", "addr", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
