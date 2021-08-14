package broker

import "time"

// OHLCV statistic.
type OHLCV struct {
	ID       int64
	Ticker   string
	Time     time.Time
	Interval time.Duration
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   int32
}

// StatisticRepo statistic repository.
type StatisticRepo interface {
	Add(ohlcv OHLCV) error
	Get(ticker string) ([]OHLCV, error)
}
