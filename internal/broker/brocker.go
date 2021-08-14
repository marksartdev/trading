package broker

import "context"

// Profile user profile.
type Profile struct {
	ClientID  int64
	Balance   float64
	Positions []Position
	OpenDeals []Deal
}

// ExchangeService stock exchange service.
type ExchangeService interface {
	Statistic(ctx context.Context, out chan OHLCV) error
	Create(deal Deal) (int64, error)
	Cancel(dealID int64) (bool, error)
	Results(ctx context.Context, out chan Deal) error
}

// BrokerService broker service.
type BrokerService interface {
	Start()
	Stop()
	GetClient(login string) (Client, error)
	GetProfile(login string) (Profile, error)
	Create(deal Deal) (Deal, error)
	Cancel(dealID int64) (bool, error)
	History(ticker string) ([]OHLCV, error)
}
