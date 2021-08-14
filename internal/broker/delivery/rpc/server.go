package rpc

import (
	"context"
	"fmt"
	"time"

	"github.com/marksartdev/trading/internal/broker"
	"github.com/marksartdev/trading/internal/log"
)

type brokerServer struct {
	logger  log.Logger
	service broker.BrokerService
	UnimplementedBrokerServer
}

const gRPC = "gRPC"

// NewBrokerServer creates new broker server.
func NewBrokerServer(logger log.Logger, service broker.BrokerService) BrokerServer {
	return &brokerServer{logger: logger, service: service}
}

// GetProfile returns client profile.
func (b brokerServer) GetProfile(_ context.Context, client *Client) (*Profile, error) {
	profile, err := b.service.GetProfile(client.GetLogin())
	if err != nil {
		b.logger.Error(gRPC, err)
		return nil, err
	}

	positions := make([]*Position, len(profile.Positions))
	for i := range positions {
		positions[i] = &Position{
			Ticker: profile.Positions[i].Ticker,
			Amount: profile.Positions[i].Amount,
		}
	}

	deals := make([]*Deal, len(profile.OpenDeals))
	for i := range deals {
		deals[i] = &Deal{
			ID:     profile.OpenDeals[i].ID,
			Ticker: profile.OpenDeals[i].Ticker,
			Type:   string(profile.OpenDeals[i].Type),
			Amount: profile.OpenDeals[i].Amount,
			Price:  profile.OpenDeals[i].Price,
			Time:   profile.OpenDeals[i].Time.Unix(),
		}
	}

	resp := Profile{
		Balance:   profile.Balance,
		Positions: positions,
		Deals:     deals,
	}

	b.logRequest(client.GetLogin(), "GetProfile")
	return &resp, nil
}

// Create creates deal.
func (b brokerServer) Create(_ context.Context, deal *CreateDeal) (*DealID, error) {
	client, err := b.service.GetClient(deal.GetClient().GetLogin())
	if err != nil {
		b.logger.Error(gRPC, err)
		return nil, err
	}

	d := broker.Deal{
		ClientID: client.ID,
		Ticker:   deal.GetTicker(),
		Type:     broker.DealType(deal.GetType()),
		Amount:   deal.GetAmount(),
		Price:    deal.GetPrice(),
		Time:     time.Now(),
	}

	d, err = b.service.Create(d)
	if err != nil {
		b.logger.Error(gRPC, err)
		return nil, err
	}

	b.logRequest(deal.GetClient().GetLogin(), "Create")
	return &DealID{ID: d.ID}, nil
}

// Cancel cancels deal.
func (b brokerServer) Cancel(_ context.Context, deal *CancelDeal) (*Success, error) {
	ok, err := b.service.Cancel(deal.GetDealID().GetID())
	if err != nil {
		b.logger.Error(gRPC, err)
		return nil, err
	}

	b.logRequest(deal.GetClient().GetLogin(), "Cancel")
	return &Success{OK: ok}, nil
}

// Statistic returns ticker statistics.
func (b brokerServer) Statistic(_ context.Context, ticker *Ticker) (*OHLCV, error) {
	stats, err := b.service.History(ticker.GetName())
	if err != nil {
		b.logger.Error(gRPC, err)
		return nil, err
	}

	prices := make([]*Price, len(stats))
	for i := range prices {
		prices[i] = &Price{
			Time:     stats[i].Time.Unix(),
			Interval: int32(stats[i].Interval.Nanoseconds()),
			Open:     stats[i].Open,
			High:     stats[i].High,
			Low:      stats[i].Low,
			Close:    stats[i].Close,
			Vol:      stats[i].Volume,
		}
	}

	resp := OHLCV{
		Prices: prices,
	}

	b.logRequest(ticker.GetClient().GetLogin(), "Statistic")
	return &resp, nil
}

func (b brokerServer) logRequest(login string, request string) {
	b.logger.Info(gRPC, fmt.Sprintf("%q request from client %s wath handled", request, login))
}
