package broker

// Position user tickers.
type Position struct {
	ClientID int64
	Ticker   string
	Amount   int32
}

// PositionRepo position repository.
type PositionRepo interface {
	Add(position Position) error
	Get(clientID int64) ([]Position, error)
	Remove(position Position) error
}
