package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/marksartdev/trading/internal/broker"
)

// Client entity.
type Client struct {
	ID        int64   `gorm:"primarykey"`
	Login     string  `gorm:"not null"`
	Balance   float64 `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Client repository.
type clientRepo struct {
	db *gorm.DB
}

// NewClientRepo creates new client repository.
func NewClientRepo(db *gorm.DB) broker.ClientRepo {
	return clientRepo{db: db}
}

// Add adds client ot repository.
func (c clientRepo) Add(client broker.Client) error {
	entity := Client{
		Login:   client.Login,
		Balance: client.Balance,
	}

	return c.db.Create(&entity).Error
}

// Get returns client from repository.
func (c clientRepo) Get(clientID int64) (broker.Client, error) {
	var entity Client

	err := c.db.Where(Client{ID: clientID}).First(&entity).Error
	if err != nil {
		return broker.Client{}, err
	}

	return broker.Client{
		ClientID: entity.ID,
		Login:    entity.Login,
		Balance:  entity.Balance,
	}, nil
}

// SumBalance adds new sum to client balance.
func (c clientRepo) SumBalance(clientID int64, amount float64) error {
	return c.db.
		Model(&Client{}).
		Where(Client{ID: clientID}).
		Update("balance", gorm.Expr("balance + ?", amount)).
		Error
}

// SubBalance removes sum from client balance.
func (c clientRepo) SubBalance(clientID int64, amount float64) error {
	return c.db.
		Model(&Client{}).
		Where(Client{ID: clientID}).
		Update("balance", gorm.Expr("balance - ?", amount)).
		Error
}
