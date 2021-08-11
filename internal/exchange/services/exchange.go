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
	Delete(dealID int64) bool
}

// Service for exchanging.
type exchangeService struct {
	mu          *sync.Mutex
	logger      exchange.Logger
	dealQueue   DealQueue
	tickService exchange.TickService
	tickers     []string
	interval    time.Duration
	tickerAmt   map[string]int32
	statObs     map[int32]chan exchange.OHLCV
	dealsObs    map[int32]chan exchange.Deal
	cancel      context.CancelFunc
}

// NewExchangeService creates new exchange service.
func NewExchangeService(
	logger exchange.Logger,
	dealQueue DealQueue,
	tickService exchange.TickService,
	tickers []string,
	interval time.Duration,
) exchange.ExchangeService {
	return &exchangeService{
		mu:          &sync.Mutex{},
		logger:      logger,
		dealQueue:   dealQueue,
		tickService: tickService,
		tickers:     tickers,
		interval:    interval,
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

	in := make(chan exchange.Tick, 100)
	out1 := make(chan exchange.Tick, 100)
	out2 := make(chan exchange.Tick, 100)

	g := &errgroup.Group{}

	g.Go(func() error {
		e.retransmitTick(ctx, in, out1, out2)
		return nil
	})

	g.Go(func() error {
		e.sendStatistic(ctx, out1)
		return nil
	})

	g.Go(func() error {
		e.completeDeals(ctx, out2)
		return nil
	})

	for _, ticker := range e.tickers {
		ticker := ticker
		g.Go(func() error {
			e.tickService.StartReading(ctx, ticker, in)
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
func (e *exchangeService) Cancel(dealID int64) bool {
	return e.dealQueue.Delete(dealID)
}

// Results adds observer for deals.
func (e *exchangeService) Results(brokerID int32, ch chan exchange.Deal) {
	e.dealsObs[brokerID] = ch
}

// Retransmits ticks to other channels.
func (e *exchangeService) retransmitTick(ctx context.Context, in chan exchange.Tick, out ...chan exchange.Tick) {
	for {
		select {
		case <-ctx.Done():
			return
		case tick := <-in:
			for _, ch := range out {
				ch <- tick
			}
		}
	}
}

// Sends a statistic to all subscribers.
func (e *exchangeService) sendStatistic(ctx context.Context, in chan exchange.Tick) {
	var (
		ohlcv      exchange.OHLCV
		closePrice float64
	)

	t := time.NewTicker(e.interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case tick := <-in:
			if ohlcv.ID == 0 {
				ohlcv = exchange.OHLCV{
					ID:       time.Now().UnixNano(),
					Time:     time.Now(),
					Interval: e.interval,
					Open:     tick.Price,
					High:     tick.Price,
					Low:      tick.Price,
					Ticker:   tick.Ticker,
				}
			}

			if tick.Price > ohlcv.High {
				ohlcv.High = tick.Price
			}

			if tick.Price < ohlcv.Low {
				ohlcv.Low = tick.Price
			}

			ohlcv.Volume += tick.Vol
			closePrice = tick.Price
		case <-t.C:
			if ohlcv.ID == 0 {
				continue
			}

			ohlcv.Close = closePrice

			for _, ch := range e.statObs {
				ch <- ohlcv
			}

			ohlcv = exchange.OHLCV{}
		}
	}
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
