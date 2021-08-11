package rpc

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/marksartdev/trading/internal/exchange"
)

const errLimit = 10

// gRPC server.
type exchangeServer struct {
	logger  *zap.SugaredLogger
	service exchange.ExchangeService
	UnimplementedExchangeServer
}

// NewExchangeServer creates new gRPC server.
func NewExchangeServer(logger *zap.SugaredLogger, service exchange.ExchangeService) ExchangeServer {
	return &exchangeServer{logger: logger, service: service}
}

// Statistic streams statistic.
func (e exchangeServer) Statistic(brokerID *BrokerID, stream Exchange_StatisticServer) error {
	var errCount int

	e.logger.Info(e.wrapMsg(fmt.Sprintf("start streaming statistic for brocker %d", brokerID.GetID())))
	defer e.logger.Info(e.wrapMsg(fmt.Sprintf("stop streaming statistic for brocker %d", brokerID.GetID())))

	ch := make(chan exchange.OHLCV, 100)
	e.service.Statistic(brokerID.GetID(), ch)

	for st := range ch {
		ohlcv := OHLCV{
			ID:       st.ID,
			Time:     int32(st.Time.Unix()),
			Interval: int32(st.Interval.Seconds()),
			Open:     float32(st.Open),
			High:     float32(st.High),
			Low:      float32(st.Low),
			Close:    float32(st.Close),
			Volume:   st.Volume,
			Ticker:   st.Ticker,
		}

		err := stream.Send(&ohlcv)
		if err != nil {
			e.logger.Error(e.wrapErr(err))

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
		Time:     time.Unix(int64(deal.GetTime()), 0),
		Price:    float64(deal.GetPrice()),
	}

	d = e.service.Create(d)

	return &DealID{ID: d.ID, BrokerID: d.BrokerID}, nil
}

// Cancel remove deal from exchange queue.
func (e exchangeServer) Cancel(_ context.Context, dealID *DealID) (*CancelResult, error) {
	ok := e.service.Cancel(dealID.GetID())

	return &CancelResult{Success: ok}, nil
}

// Results streams results of deals.
func (e exchangeServer) Results(brokerID *BrokerID, stream Exchange_ResultsServer) error {
	var errCount int

	e.logger.Info(e.wrapMsg(fmt.Sprintf("start streaming results for brocker %d", brokerID.GetID())))
	defer e.logger.Info(e.wrapMsg(fmt.Sprintf("stop streaming results for brocker %d", brokerID.GetID())))

	ch := make(chan exchange.Deal, 100)
	e.service.Results(brokerID.GetID(), ch)

	for r := range ch {
		res := Deal{
			ID:       r.ID,
			BrokerID: r.BrokerID,
			ClientID: r.ClientID,
			Ticker:   r.Ticker,
			Amount:   r.Amount,
			Partial:  r.Partial,
			Time:     int32(r.Time.Unix()),
			Price:    float32(r.Price),
		}

		err := stream.Send(&res)
		if err != nil {
			e.logger.Error(e.wrapErr(err))

			errCount++
			if errCount > errLimit {
				return err
			}
		}
	}

	return nil
}

func (e exchangeServer) wrapErr(err error) string {
	return e.wrapMsg(err.Error())
}

func (e exchangeServer) wrapMsg(msg string) string {
	return fmt.Sprintf("\033[0;32mgRPC [exchanger]\033[0m %s", msg)
}
