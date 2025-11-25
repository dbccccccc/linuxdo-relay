package storage

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func NewDB(dsn string) *DB {
	gdb, err := openGorm(dsn)
	if err != nil {
		log.Fatalf("open postgres: %v", err)
	}
	return &DB{DB: gdb}
}

func OpenDB(dsn string) (*DB, error) {
	gdb, err := openGorm(dsn)
	if err != nil {
		return nil, err
	}
	return &DB{DB: gdb}, nil
}

func openGorm(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
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
