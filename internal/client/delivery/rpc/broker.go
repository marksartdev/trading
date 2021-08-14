package rpc

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/marksartdev/trading/internal/broker/delivery/rpc"
	"github.com/marksartdev/trading/internal/client"
	"github.com/marksartdev/trading/internal/log"
)

const (
	timeout    = 5 * time.Second
	layout     = "Time: %s  Int: %ds  O: %.2f  H: %.2f  L: %.2f  C: %.2f  Val: %d"
	timeLayout = "15:04"
)

const (
	createAction  log.Action = "create"
	cancelAction  log.Action = "cancel"
	profileAction log.Action = "profile"
	statAction    log.Action = "statistic"
)

// BrokerService delivery service, which responses with strings.
type BrokerService interface {
	Create(login string, ticker, dealType string, amount int32, price float64) (string, error)
	Cancel(login string, dealID int64) (string, error)
	Profile(login string) (string, error)
	Statistic(login string, ticker string) (string, error)
}

// Broker service.
type brokerService struct {
	logger log.Logger
	client rpc.BrokerClient
}

// NewBrokerService creates new broker service.
func NewBrokerService(logger log.Logger, client rpc.BrokerClient) BrokerService {
	return &brokerService{logger: logger, client: client}
}

// Create sends deal to broker.
func (b brokerService) Create(login string, ticker, dealType string, amount int32, price float64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := rpc.CreateDeal{
		Client: &rpc.Client{Login: login},
		Ticker: ticker,
		Type:   dealType,
		Amount: amount,
		Price:  price,
	}

	resp, err := b.client.Create(ctx, &req)
	if err != nil {
		b.logger.Error(createAction, err)
		return "", err
	}

	return fmt.Sprintf("Сделка №%d зарегестрирована", resp.GetID()), nil
}

// Cancel sends request to cancel deal.
func (b brokerService) Cancel(login string, dealID int64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := rpc.CancelDeal{
		Client: &rpc.Client{Login: login},
		DealID: &rpc.DealID{ID: dealID},
	}

	resp, err := b.client.Cancel(ctx, &req)
	if err != nil {
		b.logger.Error(cancelAction, err)
		return "", err
	}

	if resp.GetOK() {
		return fmt.Sprintf("Заявка №%d успешно отменена", dealID), nil
	}

	return fmt.Sprintf("Не удалось отменить заявку №%d", dealID), nil
}

// Profile returns client's profile.
func (b brokerService) Profile(login string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := rpc.Client{Login: login}
	resp, err := b.client.GetProfile(ctx, &req)
	if err != nil {
		b.logger.Error(profileAction, err)
		return "", err
	}

	res := []string{
		"Ваш профиль",
		fmt.Sprintf("Баланс: %.2f", resp.Balance),
		"",
		"Активы:",
	}

	positions := resp.GetPositions()
	for i := range positions {
		res = append(res, fmt.Sprintf("    %s: %d", positions[i].Ticker, positions[i].Amount))
	}

	res = append(res, "", "Открытые сделки:")

	deals := resp.GetDeals()
	for i := range deals {
		res = append(res, fmt.Sprintf("    %d  %s  %s  %d  %.2f",
			deals[i].ID, deals[i].Ticker, deals[i].Type, deals[i].Amount, deals[i].Price))
	}

	return strings.Join(res, "\n"), nil
}

// Statistic returns ticker statistic.
func (b brokerService) Statistic(login string, ticker string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := rpc.Ticker{
		Client: &rpc.Client{Login: login},
		Name:   ticker,
	}

	resp, err := b.client.Statistic(ctx, &req)
	if err != nil {
		b.logger.Error(statAction, err)
		return "", err
	}

	data := make(map[string]client.OHLCV)

	prices := resp.GetPrices()
	for i := range prices {
		tm := time.Unix(prices[i].GetTime(), 0).Format(timeLayout)

		item := data[tm]

		if item.Open == 0 {
			item.Open = prices[i].GetOpen()
			item.Low = prices[i].GetLow()
		}

		if item.High < prices[i].GetHigh() {
			item.High = prices[i].GetHigh()
		}

		if item.Low > prices[i].GetLow() {
			item.Low = prices[i].GetLow()
		}

		item.Close = prices[i].GetClose()
		item.Volume += prices[i].GetVol()
		item.Interval += int32(time.Duration(prices[i].GetInterval()).Seconds())

		data[tm] = item
	}

	times := make([]string, len(data))
	var i int

	for tm := range data {
		times[i] = tm
		i++
	}

	sort.Strings(times)

	res := []string{fmt.Sprintf("Ticker: %s", ticker)}

	for _, tm := range times {
		str := fmt.Sprintf(
			layout,
			tm,
			data[tm].Interval,
			data[tm].Open,
			data[tm].High,
			data[tm].Low,
			data[tm].Close,
			data[tm].Volume,
		)
		res = append(res, str)
	}

	return strings.Join(res, "\n"), nil
}
