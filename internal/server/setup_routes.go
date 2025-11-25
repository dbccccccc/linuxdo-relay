package server

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"linuxdo-relay/internal/config"
	"linuxdo-relay/internal/runtimeconfig"
	"linuxdo-relay/internal/storage"
	"linuxdo-relay/internal/storage/migrate"
)

type setupHandler struct {
	config *config.Config
}

func RegisterSetupWizardRoutes(r *gin.Engine, cfg *config.Config) {
	handler := &setupHandler{config: cfg}

	r.GET("/setup/status", handler.handleStatus)
	r.POST("/setup/database", handler.handleSaveDatabase)
	r.POST("/setup/migrate", handler.handleRunMigrations)

	r.Static("/assets", "./web/dist/assets")
	r.StaticFile("/favicon.ico", "./web/dist/favicon.ico")

	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.File("./web/dist/index.html")
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "server is in setup mode"})
	})
}

func (h *setupHandler) handleStatus(c *gin.Context) {
	store := runtimeconfig.NewStore(h.config.RuntimeConfigPath)
	data, err := store.Load()
	if err != nil {
		if err == runtimeconfig.ErrNotFound {
			c.JSON(http.StatusOK, gin.H{"mode": "unconfigured"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if data.Database.DSN == "" {
		c.JSON(http.StatusOK, gin.H{"mode": "unconfigured"})
		return
	}

	db, err := storage.OpenDB(data.Database.DSN)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"mode":    "invalid_credentials",
			"details": err.Error(),
		})
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("setup status: close db: %v", err)
		}
	}()

	runner := migrate.NewRunner(db.DB, h.config.MigrationsDir)
	res, err := runner.Check(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	runtimeSummary := gin.H{
		"database_configured": data.Database.DSN != "",
		"redis_configured":    data.Redis.Addr != "" || data.Redis.Password != "",
	}

	c.JSON(http.StatusOK, gin.H{
		"mode":    string(res.Status),
		"result":  res,
		"runtime": runtimeSummary,
	})
}

func (h *setupHandler) handleSaveDatabase(c *gin.Context) {
	var body struct {
		DSN           string `json:"dsn"`
		RedisAddr     string `json:"redis_addr"`
		RedisPassword string `json:"redis_password"`
		Force         bool   `json:"force"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if body.DSN == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dsn is required"})
		return
	}

	db, err := storage.OpenDB(body.DSN)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("setup save db: close db: %v", err)
		}
	}()

	store := runtimeconfig.NewStore(h.config.RuntimeConfigPath)
	data, err := store.Load()
	if err != nil && err != runtimeconfig.ErrNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if data == nil {
		data = &runtimeconfig.Data{}
	}

	data.Database.DSN = body.DSN
	if body.RedisAddr != "" {
		data.Redis.Addr = body.RedisAddr
	}
	if body.RedisPassword != "" {
		data.Redis.Password = body.RedisPassword
	}

	if err := store.Save(data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	runner := migrate.NewRunner(db.DB, h.config.MigrationsDir)
	res, err := runner.ApplyPending(c.Request.Context(), body.Force)
	if err != nil {
		mode := "unknown"
		if res != nil {
			mode = string(res.Status)
		}
		status := http.StatusBadRequest
		if err == migrate.ErrSchemaDrift {
			status = http.StatusPreconditionFailed
		}
		c.JSON(status, gin.H{"error": err.Error(), "mode": mode})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mode": string(res.Status), "result": res})
}

func (h *setupHandler) handleRunMigrations(c *gin.Context) {
	store := runtimeconfig.NewStore(h.config.RuntimeConfigPath)
	data, err := store.Load()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "runtime config missing"})
		return
	}
	if data.Database.DSN == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "database dsn not configured"})
		return
	}

	force := c.Query("force") == "true"

	db, err := storage.OpenDB(data.Database.DSN)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("setup migrate: close db: %v", err)
		}
	}()

	runner := migrate.NewRunner(db.DB, h.config.MigrationsDir)
	res, err := runner.ApplyPending(c.Request.Context(), force)
	if err != nil {
		mode := "unknown"
		if res != nil {
			mode = string(res.Status)
		}
		status := http.StatusBadRequest
		if err == migrate.ErrSchemaDrift {
			status = http.StatusPreconditionFailed
		}
		c.JSON(status, gin.H{"error": err.Error(), "mode": mode})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mode": string(res.Status), "result": res})
}
