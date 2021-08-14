package broker

import "time"

// DealType deal type.
type DealType string

const (
	// Buy deal of buying.
	Buy DealType = "BUY"
	// Sell deal of selling.
	Sell DealType = "SELL"
)

// DealStatus deal status.
type DealStatus string

const (
	// DealStatusNew new deal status.
	DealStatusNew DealStatus = "NEW"
	// DealStatusCompleted completed deal status.
	DealStatusCompleted DealStatus = "COMPLETED"
	// DealStatusCanceled canceled deal status.
	DealStatusCanceled DealStatus = "CANCELED"
)

// Deal user deal.
type Deal struct {
	ID       int64
	ClientID int64
	Ticker   string
	Type     DealType
	Amount   int32
	Partial  bool
	Price    float64
	Status   DealStatus
	Time     time.Time
}

// DealRepo deal repository.
type DealRepo interface {
	Add(deal Deal) error
	GetOpened(clientID int64) ([]Deal, error)
	Update(deal Deal) error
	UpdateStatus(dealID int64, status DealStatus) error
}
