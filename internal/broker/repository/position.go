package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/marksartdev/trading/internal/broker"
)

// Position entity.
type Position struct {
	ID        int64  `gorm:"primarykey"`
	ClientID  int64  `gorm:"not null"`
	Ticker    string `gorm:"not null"`
	Vol       int32  `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Position repository.
type positionRepo struct {
	db *gorm.DB
}

// NewPositionRepo creates new position repository.
func NewPositionRepo(db *gorm.DB) broker.PositionRepo {
	return positionRepo{db: db}
}

// Add adds position to repository.
func (p positionRepo) Add(position broker.Position) error {
	var entity Position

	err := p.db.Where(Position{ClientID: position.ClientID, Ticker: position.Ticker}).FirstOrCreate(&entity).Error
	if err != nil {
		return err
	}

	entity.Vol += position.Amount

	return p.db.Save(&entity).Error
}

// Get returns positions from repository.
func (p positionRepo) Get(clientID int64) ([]broker.Position, error) {
	var entities []Position

	err := p.db.Where(Position{ClientID: clientID}).Find(&entities).Error
	if err != nil {
		return nil, err
	}

	positions := make([]broker.Position, len(entities))
	for i := range positions {
		positions[i] = broker.Position{
			ClientID: entities[i].ClientID,
			Ticker:   entities[i].Ticker,
			Amount:   entities[i].Vol,
		}
	}

	return positions, nil
}

// Remove deletes position from repository.
func (p positionRepo) Remove(position broker.Position) error {
	var entity Position

	err := p.db.Where(Position{ClientID: position.ClientID, Ticker: position.Ticker}).FirstOrCreate(&entity).Error
	if err != nil {
		return err
	}

	entity.Vol -= position.Amount

	return p.db.Save(&entity).Error
}
