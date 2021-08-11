package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/marksartdev/trading/internal/exchange"
)

// DealQueue queue of deals.
type DealQueue interface {
	Add(deal exchange.Deal)
	Get(ticker string, price float64) []exchange.Deal
	Delete(dealID int64)
}

// Service for exchanging.
type exchangeService struct {
	mu          *sync.Mutex
	logger      exchange.Logger
	dealQueue   DealQueue
	tickService exchange.TickService
	tickers     []string
	tickerAmt   map[string]int32
	statObs     map[int32]chan exchange.OHLCV
	dealsObs    map[int32]chan exchange.Deal
	cancel      context.CancelFunc
}

// NewExchangeService creates new exchange service.
func NewExchangeService(logger exchange.Logger, dealQueue DealQueue, tickService exchange.TickService, tickers []string,
) exchange.ExchangeService {
	return &exchangeService{
		mu:          &sync.Mutex{},
		logger:      logger,
		dealQueue:   dealQueue,
		tickService: tickService,
		tickers:     tickers,
		tickerAmt:   make(map[string]int32),
		statObs:     make(map[int32]chan exchange.OHLCV),
		dealsObs:    make(map[int32]chan exchange.Deal),
	}
}

// Start starts working.
func (e *exchangeService) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e.cancel = cancel

	ch := make(chan exchange.Tick, 100)
	g := &errgroup.Group{}

	g.Go(func() error {
		e.completeDeals(ctx, ch)
		return nil
	})

	for _, ticker := range e.tickers {
		ticker := ticker
		g.Go(func() error {
			e.tickService.StartReading(ctx, ticker, ch)
			return nil
		})
	}

	e.logger.Info(e.wrapMsg("main", "started"))

	if err := g.Wait(); err != nil {
		e.logger.Info(e.wrapMsg("main", err.Error()))
	}

	e.logger.Info(e.wrapMsg("main", "stopped"))
}

// Stop stops working.
func (e *exchangeService) Stop() {
	if e.cancel != nil {
		e.cancel()
		return
	}

	e.logger.Error(e.wrapMsg("main", "cancel func dose not initialized"))
}

// Statistic adds observer for a statistic.
func (e *exchangeService) Statistic(brokerID int32, ch chan exchange.OHLCV) {
	e.statObs[brokerID] = ch
}

// Create adds a deal to queue.
func (e *exchangeService) Create(deal exchange.Deal) exchange.Deal {
	deal.ID = time.Now().UnixNano()
	e.dealQueue.Add(deal)

	return deal
}

// Cancel removes deal from queue.
func (e *exchangeService) Cancel(dealID int64) {
	e.dealQueue.Delete(dealID)
}

// Results adds observer for deals.
func (e *exchangeService) Results(brokerID int32, ch chan exchange.Deal) {
	e.dealsObs[brokerID] = ch
}

// Completes deals.
func (e *exchangeService) completeDeals(ctx context.Context, in chan exchange.Tick) {
	e.logger.Info(e.wrapMsg("deals", "started"))

	for {
		select {
		case <-ctx.Done():
			e.logger.Info(e.wrapMsg("deals", "stopped"))
			return
		case tick := <-in:
			deals := e.dealQueue.Get(tick.Ticker, tick.Price)
			for _, deal := range deals {
				var completed bool

				e.mu.Lock()
				if deal.Price > 0 && e.tickerAmt[deal.Ticker] > 0 {
					e.completePurchase(&deal)
					completed = true
				}

				if deal.Price < 0 {
					e.completeSale(&deal)
					completed = true
				}
				e.mu.Unlock()

				if completed {
					e.dealsObs[deal.BrokerID] <- deal
				}
			}
		}
	}
}

// Completes purchase.
func (e *exchangeService) completePurchase(deal *exchange.Deal) {
	if e.tickerAmt[deal.Ticker] < deal.Amount {
		deal.Amount = e.tickerAmt[deal.Ticker]
		deal.Partial = true
		e.tickerAmt[deal.Ticker] = 0
	} else {
		e.tickerAmt[deal.Ticker] -= deal.Amount
	}

	e.dealQueue.Delete(deal.ID)
}

// Completes sale.
func (e *exchangeService) completeSale(deal *exchange.Deal) {
	e.tickerAmt[deal.Ticker] += deal.Amount
	e.dealQueue.Delete(deal.ID)
}

// Wraps message.
func (e *exchangeService) wrapMsg(detail, msg string) string {
	return fmt.Sprintf("Exchange service (%s): %s", detail, msg)
}
