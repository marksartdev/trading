package client

// OHLCV statistic.
type OHLCV struct {
	Interval int32
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   int32
}
