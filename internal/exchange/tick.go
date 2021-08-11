package exchange

import "context"

// Tick stock exchange tick.
type Tick struct {
	Ticker string
	Price  float64
}

// TickService service for working with ticks.
type TickService interface {
	StartReading(ctx context.Context, ticker string, chOut chan Tick)
}
