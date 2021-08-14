package broker

// Client broker client.
type Client struct {
	ID      int64
	Login   string
	Balance float64
}

// ClientRepo client repository.
type ClientRepo interface {
	Add(client *Client) error
	Get(login string) (Client, bool, error)
	SumBalance(clientID int64, amount float64) error
	SubBalance(clientID int64, amount float64) error
}
