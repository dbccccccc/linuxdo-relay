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
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("open postgres: %v", err)
	}
	return &DB{DB: gdb}
}
