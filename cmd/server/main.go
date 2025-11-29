package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"linuxdo-relay/internal/auth"
	"linuxdo-relay/internal/config"
	"linuxdo-relay/internal/logger"
	"linuxdo-relay/internal/server"
	"linuxdo-relay/internal/storage"
)

// Version is set at build time via -ldflags
var Version = "dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// init DB with connection pooling
	db, err := storage.OpenDB(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	logger.Info("database connected")

	// auto migrate database schema
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}
	logger.Info("database migrated")

	// init Redis with connection verification
	redisClient, err := storage.NewRedisWithPing(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	logger.Info("redis connected", "addr", cfg.RedisAddr)

	// init OAuth config
	oauthCfg := auth.NewLinuxDoOAuthConfig(cfg)

	app := &server.AppContext{
		Config:    cfg,
		DB:        db,
		Redis:     redisClient,
		OAuth:     oauthCfg,
		JWTSecret: cfg.JWTSecret,
		Version:   Version,
	}

	r := gin.Default()
	server.SetupRoutes(r, app)

	addr := cfg.HTTPListen
	if addr == "" {
		addr = ":8080"
	}

	logger.Info("starting server", "addr", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
