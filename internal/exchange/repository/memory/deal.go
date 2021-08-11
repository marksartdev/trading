package memory

import (
	"sync"

	"github.com/marksartdev/trading/internal/exchange"
	"github.com/marksartdev/trading/internal/exchange/services"
)

// Queue of deals.
type dealQueue struct {
	mu    *sync.RWMutex
	deals []exchange.Deal
}

// NewDealQueue creates new queue of deals.
func NewDealQueue() services.DealQueue {
	return &dealQueue{mu: &sync.RWMutex{}}
}

// Add adds a deal to queue.
func (d *dealQueue) Add(deal exchange.Deal) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.deals = append(d.deals, deal)
}

// Get returns deals by a ticker and price.
func (d *dealQueue) Get(ticker string, price float64) []exchange.Deal {
	var res []exchange.Deal

	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, deal := range d.deals {
		if deal.Ticker != ticker {
			continue
		}

		if deal.Price < 0 && -deal.Price <= price {
			res = append(res, deal)
		}

		if deal.Price > 0 && deal.Price >= price {
			res = append(res, deal)
		}
	}

	return res
}

// Delete removes deal from queue.
func (d *dealQueue) Delete(dealID int64) bool {
	idx := -1

	d.mu.Lock()
	defer d.mu.Unlock()

	for i := range d.deals {
		if d.deals[i].ID == dealID {
			idx = i
			break
		}
	}

	if idx == -1 {
		return false
	}

	d.deals = append(d.deals[:idx], d.deals[idx+1:]...)

	return true
}
