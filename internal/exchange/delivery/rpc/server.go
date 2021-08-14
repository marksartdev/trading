package rpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/marksartdev/trading/internal/exchange"
	"github.com/marksartdev/trading/internal/log"
)

const errLimit = 10

const gRPC log.Action = "gRPC"

// gRPC server.
type exchangeServer struct {
	logger  log.Logger
	service exchange.ExchangeService
	UnimplementedExchangeServer
}

// NewExchangeServer creates new gRPC server.
func NewExchangeServer(logger log.Logger, service exchange.ExchangeService) ExchangeServer {
	return &exchangeServer{logger: logger, service: service}
}

// Statistic streams statistic.
func (e exchangeServer) Statistic(brokerID *BrokerID, stream Exchange_StatisticServer) error {
	var errCount int

	broker := exchange.Broker{
		ID:         brokerID.GetID(),
		InstanceID: time.Now().UnixNano(),
	}

	e.logger.Info(gRPC, fmt.Sprintf("start streaming statistic for brocker %d", brokerID.GetID()))
	defer e.logger.Info(gRPC, fmt.Sprintf("stop streaming statistic for brocker %d", brokerID.GetID()))
	defer e.service.StatisticUnsubscribe(broker)

	ch := make(chan exchange.OHLCV, 100)
	e.service.Statistic(broker, ch)

	for st := range ch {
		ohlcv := OHLCV{
			ID:       st.ID,
			Time:     st.Time.Unix(),
			Interval: int32(st.Interval.Seconds()),
			Open:     st.Open,
			High:     st.High,
			Low:      st.Low,
			Close:    st.Close,
			Volume:   st.Volume,
			Ticker:   st.Ticker,
		}

		err := stream.Send(&ohlcv)
		if err != nil {
			if s, ok := status.FromError(err); ok {
				if s.Code() == codes.Unavailable {
					return nil
				}
			}
			e.logger.Error(gRPC, err)

			errCount++
			if errCount > errLimit {
				return err
			}
		}
	}

	return nil
}

// Create adds a deal to exchange queue.
func (e exchangeServer) Create(_ context.Context, deal *Deal) (*DealID, error) {
	d := exchange.Deal{
		ID:       deal.GetID(),
		BrokerID: deal.GetBrokerID(),
		ClientID: deal.GetClientID(),
		Ticker:   deal.GetTicker(),
		Amount:   deal.GetAmount(),
		Partial:  deal.GetPartial(),
		Time:     time.Unix(deal.GetTime(), 0),
		Price:    deal.GetPrice(),
	}

	d = e.service.Create(d)

	e.logger.Info(gRPC, fmt.Sprintf("%q request from broker %d wath handled", "Create", deal.GetBrokerID()))

	return &DealID{ID: d.ID, BrokerID: d.BrokerID}, nil
}

// Cancel remove deal from exchange queue.
func (e exchangeServer) Cancel(_ context.Context, dealID *DealID) (*CancelResult, error) {
	ok := e.service.Cancel(dealID.GetID())

	e.logger.Info(gRPC, fmt.Sprintf("%q request from broker %d wath handled", "Cancel", dealID.GetBrokerID()))

	return &CancelResult{Success: ok}, nil
}

// Results streams results of deals.
func (e exchangeServer) Results(brokerID *BrokerID, stream Exchange_ResultsServer) error {
	var errCount int

	broker := exchange.Broker{
		ID:         brokerID.GetID(),
		InstanceID: time.Now().UnixNano(),
	}

	e.logger.Info(gRPC, fmt.Sprintf("start streaming results for brocker %d", brokerID.GetID()))
	defer e.logger.Info(gRPC, fmt.Sprintf("stop streaming results for brocker %d", brokerID.GetID()))
	defer e.service.ResultsUnsubscribe(broker)

	ch := make(chan exchange.Deal, 100)
	e.service.Results(broker, ch)

	for r := range ch {
		res := Deal{
			ID:       r.ID,
			BrokerID: r.BrokerID,
			ClientID: r.ClientID,
			Ticker:   r.Ticker,
			Amount:   r.Amount,
			Partial:  r.Partial,
			Time:     r.Time.Unix(),
			Price:    r.Price,
		}

		err := stream.Send(&res)
		if err != nil {
			if s, ok := status.FromError(err); ok {
				if s.Code() == codes.Unavailable {
					return nil
				}
			}
			e.logger.Error(gRPC, err)

			errCount++
			if errCount > errLimit {
				return err
			}
		}
	}

	return nil
}
