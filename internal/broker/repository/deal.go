package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/marksartdev/trading/internal/broker"
)

// Deal entity.
type Deal struct {
	ID        int64             `gorm:"primarykey"`
	ClientID  int64             `gorm:"not null"`
	Ticker    string            `gorm:"not null"`
	Vol       int32             `gorm:"not null"`
	Partial   bool              `gorm:"not null"`
	Price     float64           `gorm:"not null"`
	Type      broker.DealType   `gorm:"not null"`
	Status    broker.DealStatus `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Deal repository.
type dealRepo struct {
	db *gorm.DB
}

// NewDealRepo creates new deal repository.
func NewDealRepo(db *gorm.DB) broker.DealRepo {
	return dealRepo{db: db}
}

// Add adds deal to repository.
func (d dealRepo) Add(deal broker.Deal) error {
	entity := Deal{
		ID:       deal.ID,
		ClientID: deal.ClientID,
		Ticker:   deal.Ticker,
		Vol:      deal.Amount,
		Price:    deal.Price,
		Type:     deal.Type,
		Status:   deal.Status,
	}

	return d.db.Create(&entity).Error
}

// GetOpened returns opened deals.
func (d dealRepo) GetOpened(clientID int64) ([]broker.Deal, error) {
	var entities []Deal

	err := d.db.Where(Deal{ClientID: clientID, Status: broker.DealStatusNew}).Find(&entities).Error
	if err != nil {
		return nil, err
	}

	deals := make([]broker.Deal, len(entities))
	for i := range deals {
		deals[i] = broker.Deal{
			ID:       entities[i].ID,
			ClientID: entities[i].ClientID,
			Ticker:   entities[i].Ticker,
			Type:     entities[i].Type,
			Amount:   entities[i].Vol,
			Price:    entities[i].Price,
			Status:   entities[i].Status,
			Time:     entities[i].CreatedAt,
		}
	}

	return deals, nil
}

// Update updates deal.
func (d dealRepo) Update(deal broker.Deal) error {
	return d.db.Where(Deal{ID: deal.ID}).Updates(Deal{
		Vol:     deal.Amount,
		Partial: deal.Partial,
		Price:   deal.Price,
		Status:  deal.Status,
	}).Error
}

// UpdateStatus updates deal status.
func (d dealRepo) UpdateStatus(dealID int64, status broker.DealStatus) error {
	return d.db.Model(Deal{}).Where(Deal{ID: dealID}).Update("status", status).Error
}
