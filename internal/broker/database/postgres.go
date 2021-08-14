package database

import (
	"gorm.io/gorm"

	"github.com/marksartdev/trading/internal/broker/repository"
	"github.com/marksartdev/trading/internal/config"
	"github.com/marksartdev/trading/internal/database"
)

// New creates new database client.
func New(cfg config.DB) (*gorm.DB, error) {
	db, err := database.New(cfg)
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		repository.Client{},
		repository.Deal{},
		repository.Position{},
		repository.OHLCV{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
