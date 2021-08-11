package exchange

import "time"

// OHLCV statistic.
type OHLCV struct {
	ID       int64
	Time     time.Time
	Interval time.Duration
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Ticker   string
}

// Deal purchase/sale of ticker.
type Deal struct {
	ID       int64
	BrokerID int32
	ClientID int32
	Ticker   string
	Amount   int32
	Partial  bool
	Time     time.Time
	Price    float64
}

// ExchangeService service for exchanging.
type ExchangeService interface {
	Start()
	Stop()
	Statistic(brokerID int32, ch chan OHLCV)
	Create(deal Deal) Deal
	Cancel(dealID int64)
	Results(brokerID int32, ch chan Deal)
}
