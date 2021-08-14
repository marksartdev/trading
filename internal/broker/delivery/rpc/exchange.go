package rpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/marksartdev/trading/internal/broker"
	"github.com/marksartdev/trading/internal/exchange/delivery/rpc"
)

const timeout = 10 * time.Second

// ExchangeService client of exchange service.
type ExchangeService struct {
	brokerID int64
	client   rpc.ExchangeClient
}

// NewExchangeService creates new exchange service.
func NewExchangeService(brokerID int64, client rpc.ExchangeClient) *ExchangeService {
	return &ExchangeService{brokerID: brokerID, client: client}
}

// Statistic subscribes to statistic.
func (e ExchangeService) Statistic(ctx context.Context, out chan broker.OHLCV) error {
	in := rpc.BrokerID{ID: e.brokerID}

	stream, err := e.client.Statistic(ctx, &in)
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			if s, ok := status.FromError(err); ok {
				if s.Code() == codes.Canceled || s.Code() == codes.Unavailable {
					close(out)
					return nil
				}
			}
			return err
		}

		ohlcv := broker.OHLCV{
			ID:       resp.GetID(),
			Ticker:   resp.GetTicker(),
			Time:     time.Unix(resp.GetTime(), 0),
			Interval: time.Duration(resp.GetInterval()),
			Open:     resp.GetOpen(),
			High:     resp.GetHigh(),
			Low:      resp.GetLow(),
			Close:    resp.GetClose(),
			Volume:   resp.GetVolume(),
		}

		out <- ohlcv
	}
}

// Create sends deal to exchange service.
func (e ExchangeService) Create(deal broker.Deal) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	price := deal.Price
	if deal.Type == broker.Sell {
		price = -price
	}

	in := rpc.Deal{
		BrokerID: e.brokerID,
		ClientID: deal.ClientID,
		Ticker:   deal.Ticker,
		Amount:   deal.Amount,
		Time:     deal.Time.Unix(),
		Price:    price,
	}

	resp, err := e.client.Create(ctx, &in)
	if err != nil {
		return 0, err
	}

	return resp.GetID(), nil
}

// Cancel send deal cancel to exchange service.
func (e ExchangeService) Cancel(dealID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	in := rpc.DealID{
		ID:       dealID,
		BrokerID: e.brokerID,
	}

	resp, err := e.client.Cancel(ctx, &in)
	if err != nil {
		return false, err
	}

	return resp.GetSuccess(), nil
}

// Results subscribes to result of deals.
func (e ExchangeService) Results(ctx context.Context, out chan broker.Deal) error {
	in := rpc.BrokerID{ID: e.brokerID}

	stream, err := e.client.Results(ctx, &in)
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			if s, ok := status.FromError(err); ok {
				if s.Code() == codes.Canceled || s.Code() == codes.Unavailable {
					close(out)
					return nil
				}
			}
			return err
		}

		dealType := broker.Buy
		price := resp.GetPrice()
		if price < 0 {
			dealType = broker.Sell
			price = -price
		}

		deal := broker.Deal{
			ID:       resp.GetID(),
			ClientID: resp.GetClientID(),
			Ticker:   resp.GetTicker(),
			Type:     dealType,
			Amount:   resp.GetAmount(),
			Partial:  resp.GetPartial(),
			Price:    price,
			Status:   broker.DealStatusCompleted,
			Time:     time.Unix(resp.GetTime(), 0),
		}

		out <- deal
	}
}
