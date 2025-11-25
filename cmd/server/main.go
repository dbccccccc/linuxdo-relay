package main

import (
	"context"
	"errors"
	"log"

	"github.com/gin-gonic/gin"

	"linuxdo-relay/internal/auth"
	"linuxdo-relay/internal/config"
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
		// init DB & Redis
		db := storage.NewDB(cfg.PostgresDSN)
		redisClient := storage.NewRedis(cfg.RedisAddr, cfg.RedisPassword)

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

	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
