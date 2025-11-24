package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"linuxdo-relay/internal/auth"
	"linuxdo-relay/internal/config"
	"linuxdo-relay/internal/server"
	"linuxdo-relay/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	r := gin.Default()

	// init DB & Redis
	db := storage.NewDB(cfg.PostgresDSN)
	redisClient := storage.NewRedis(cfg.RedisAddr, cfg.RedisPassword)

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

	addr := cfg.HTTPListen
	if addr == "" {
		addr = ":8080"
	}

	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
