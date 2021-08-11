package services

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/marksartdev/trading/internal/exchange"
)

const (
	timeIdx = 3
	lastIdx = 4
)

// Service for working with ticks.
type tickService struct {
	logger exchange.Logger
}

// NewTickService creates new tick service.
func NewTickService(logger exchange.Logger) exchange.TickService {
	return &tickService{logger: logger}
}

// StartReading starts reading ticks from a file and sending it to channel.
func (t *tickService) StartReading(ctx context.Context, ticker string, out chan exchange.Tick) {
	f, err := os.Open(filepath.Join("assets", fmt.Sprintf("%s.txt", ticker)))
	if err != nil {
		t.logger.Error(t.wrapMsg(ticker, err.Error()))
		return
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	t.start(ctx, ticker, scanner, out)
}

// Starts reading ticks from a file and sending it to channel.
func (t *tickService) start(ctx context.Context, ticker string, scanner *bufio.Scanner, out chan exchange.Tick) {
	timeTicker := time.NewTicker(time.Second)
	now := time.Now().Format("150405")

	var (
		buffer []string
		data   []string
	)

	t.logger.Info(t.wrapMsg(ticker, "started"))

	// Skip headers.
	if scanner.Scan() {
		scanner.Text()
	}

	for scanner.Scan() {
		if buffer != nil {
			data = buffer
			buffer = nil
		} else {
			data = strings.Split(scanner.Text(), ",")
		}

		// Skip all past ticks.
		if strings.TrimSpace(data[timeIdx]) < now {
			continue
		}

		select {
		case <-ctx.Done():
			t.logger.Info(t.wrapMsg(ticker, "stopped"))
			return
		case <-timeTicker.C:
			firstTickTime := strings.TrimSpace(data[timeIdx])
			if err := t.processTick(data, ticker, out); err != nil {
				t.logger.Error(t.wrapMsg(ticker, err.Error()))
				continue
			}

			for scanner.Scan() {
				data = strings.Split(scanner.Text(), ",")
				tickTime := strings.TrimSpace(data[timeIdx])

				if tickTime != firstTickTime {
					buffer = data
					break
				}

				if err := t.processTick(data, ticker, out); err != nil {
					t.logger.Error(t.wrapMsg(ticker, err.Error()))
					continue
				}
			}
		}
	}
}

// Processes tick.
func (t *tickService) processTick(data []string, ticker string, out chan exchange.Tick) error {
	price, err := strconv.ParseFloat(data[lastIdx], 64)
	if err != nil {
		return err
	}

	out <- exchange.Tick{Ticker: ticker, Price: price}

	return nil
}

// Wraps message.
func (t *tickService) wrapMsg(detail, msg string) string {
	return fmt.Sprintf("Tick service (%s): %s", detail, msg)
}