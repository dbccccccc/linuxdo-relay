package storage

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"linuxdo-relay/internal/models"
)

// DBConfig holds database connection pool settings.
type DBConfig struct {
	MaxOpenConns    int           // Maximum number of open connections to the database
	MaxIdleConns    int           // Maximum number of idle connections in the pool
	ConnMaxLifetime time.Duration // Maximum amount of time a connection may be reused
	ConnMaxIdleTime time.Duration // Maximum amount of time a connection may be idle
}

// DefaultDBConfig returns sensible defaults for connection pooling.
func DefaultDBConfig() DBConfig {
	return DBConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

type DB struct {
	*gorm.DB
}

// OpenDB creates a new DB connection with default pool settings.
func OpenDB(dsn string) (*DB, error) {
	return OpenDBWithConfig(dsn, DefaultDBConfig())
}

// OpenDBWithConfig creates a new DB connection with custom pool settings.
func OpenDBWithConfig(dsn string, cfg DBConfig) (*DB, error) {
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return &DB{DB: gdb}, nil
}

// AutoMigrate runs GORM auto migration for all models.
func (db *DB) AutoMigrate() error {
	return db.DB.AutoMigrate(
		&models.User{},
		&models.Channel{},
		&models.QuotaRule{},
		&models.ModelCreditRule{},
		&models.CreditTransaction{},
		&models.APILog{},
		&models.OperationLog{},
		&models.LoginLog{},
		&models.CheckInLog{},
		&models.CheckInRewardOption{},
		&models.CheckInDecayRule{},
	)
}

func (db *DB) Close() error {
	if db == nil || db.DB == nil {
		return nil
	}
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
