package services

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/marksartdev/trading/internal/broker"
	"github.com/marksartdev/trading/internal/log"
)

const (
	mainAction      log.Action = "main"
	statAction      log.Action = "statistic"
	dealsAction     log.Action = "deals"
	statGrpcAction  log.Action = "statistic - gRPC"
	dealsGrpcAction log.Action = "deals - gRPC"
)

// Broker service.
type brokerService struct {
	logger     log.Logger
	clientRepo broker.ClientRepo
	dealRepo   broker.DealRepo
	posRepo    broker.PositionRepo
	statRepo   broker.StatisticRepo
	exchange   broker.ExchangeService
	cancel     context.CancelFunc
}

// NewBrokerService creates new broker.
func NewBrokerService(
	logger log.Logger,
	clientRepo broker.ClientRepo,
	dealRepo broker.DealRepo,
	posRepo broker.PositionRepo,
	statRepo broker.StatisticRepo,
	exchange broker.ExchangeService,
) broker.BrokerService {
	return &brokerService{
		logger:     logger,
		clientRepo: clientRepo,
		dealRepo:   dealRepo,
		posRepo:    posRepo,
		statRepo:   statRepo,
		exchange:   exchange,
	}
}

// Start runs service.
func (b *brokerService) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b.cancel = cancel

	g := &errgroup.Group{}

	g.Go(func() error {
		b.consumeStatistic(ctx)
		return nil
	})
	g.Go(func() error {
		b.consumeResults(ctx)
		return nil
	})

	b.logger.Info(mainAction, "started")
	if err := g.Wait(); err != nil {
		b.logger.Error(mainAction, err)
	}
	b.logger.Info(mainAction, "stopped")
}

// Stop stops service.
func (b *brokerService) Stop() {
	if b.cancel != nil {
		b.cancel()
		return
	}

	b.logger.Error(mainAction, fmt.Errorf("cancel func dose not initialized"))
}

// GetClient returns client. Create client if it is not exist.
func (b *brokerService) GetClient(login string) (broker.Client, error) {
	client, ok, err := b.clientRepo.Get(login)
	if err != nil {
		return broker.Client{}, err
	}

	if !ok {
		client.Login = login
		client.Balance = 100000000
		if err := b.clientRepo.Add(&client); err != nil {
			return broker.Client{}, err
		}
	}

	return client, nil
}

// GetProfile returns profile.
func (b *brokerService) GetProfile(login string) (broker.Profile, error) {
	client, err := b.GetClient(login)
	if err != nil {
		return broker.Profile{}, err
	}

	positions, err := b.posRepo.Get(client.ID)
	if err != nil {
		return broker.Profile{}, err
	}

	deals, err := b.dealRepo.GetOpened(client.ID)
	if err != nil {
		return broker.Profile{}, err
	}

	return broker.Profile{
		ClientID:  client.ID,
		Balance:   client.Balance,
		Positions: positions,
		OpenDeals: deals,
	}, nil
}

// Create creates deal and send it to exchange service.
func (b *brokerService) Create(deal broker.Deal) (broker.Deal, error) {
	dealID, err := b.exchange.Create(deal)
	if err != nil {
		return broker.Deal{}, err
	}

	deal.ID = dealID
	deal.Status = broker.DealStatusNew
	if err := b.dealRepo.Add(deal); err != nil {
		return broker.Deal{}, err
	}

	return deal, nil
}

// Cancel canceled deal.
func (b *brokerService) Cancel(dealID int64) (bool, error) {
	ok, err := b.exchange.Cancel(dealID)
	if err != nil {
		return false, err
	}

	if ok {
		if err := b.dealRepo.UpdateStatus(dealID, broker.DealStatusCanceled); err != nil {
			return false, err
		}
	}

	return ok, nil
}

// History returns ticker history.
func (b *brokerService) History(ticker string) ([]broker.OHLCV, error) {
	history, err := b.statRepo.Get(ticker)
	if err != nil {
		return nil, err
	}

	return history, nil
}

// Consumes statistic.
func (b *brokerService) consumeStatistic(ctx context.Context) {
	in := make(chan broker.OHLCV, 100)

	g := &errgroup.Group{}

	g.Go(func() error {
		b.logger.Info(statGrpcAction, "started")
		defer b.logger.Info(statGrpcAction, "stopped")

		if err := b.exchange.Statistic(ctx, in); err != nil {
			b.logger.Error(statGrpcAction, err)
		}
		return nil
	})

	b.logger.Info(statAction, "started")
	defer b.logger.Info(statAction, "stopped")

	for ohlcv := range in {
		if err := b.statRepo.Add(ohlcv); err != nil {
			b.logger.Error(statAction, err)
		}
	}

	if err := g.Wait(); err != nil {
		b.logger.Error(statAction, err)
	}
}

// Consumes results of deals.
func (b *brokerService) consumeResults(ctx context.Context) {
	in := make(chan broker.Deal, 100)

	g := &errgroup.Group{}

	g.Go(func() error {
		b.logger.Info(dealsGrpcAction, "started")
		defer b.logger.Info(dealsGrpcAction, "stopped")

		if err := b.exchange.Results(ctx, in); err != nil {
			b.logger.Error(dealsGrpcAction, err)
		}
		return nil
	})

	b.logger.Info(dealsAction, "started")
	defer b.logger.Info(dealsAction, "stopped")

	for deal := range in {
		if err := b.dealRepo.Update(deal); err != nil {
			b.logger.Error(dealsAction, err)
			continue
		}

		position := broker.Position{
			ClientID: deal.ClientID,
			Ticker:   deal.Ticker,
			Amount:   deal.Amount,
		}

		var err error
		if deal.Type == broker.Buy {
			err = b.posRepo.Add(position)
		} else {
			err = b.posRepo.Remove(position)
		}

		if err != nil {
			b.logger.Error(dealsAction, err)
			continue
		}

		if deal.Type == broker.Buy {
			err = b.clientRepo.SubBalance(deal.ClientID, deal.Price*float64(deal.Amount))
		} else {
			err = b.clientRepo.SumBalance(deal.ClientID, deal.Price*float64(deal.Amount))
		}

		if err != nil {
			b.logger.Error(dealsAction, err)
			continue
		}

	}

	if err := g.Wait(); err != nil {
		b.logger.Error(dealsAction, err)
	}
}
