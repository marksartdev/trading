package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/marksartdev/trading/internal/config"
)

const postgresDSN = "host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable TimeZone=%s"

// New creates new Postgres client.
func New(cfg config.DB) (*gorm.DB, error) {
	dsn := fmt.Sprintf(postgresDSN, cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.TimeZone)

	client, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return nil, err
	}

	if err := client.AutoMigrate(); err != nil {
		return nil, err
	}

	return client, nil
}
