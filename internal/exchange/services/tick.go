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
	"github.com/marksartdev/trading/internal/log"
)

const (
	timeIdx = 3
	lastIdx = 4
	volIdx  = 5
)

// Service for working with ticks.
type tickService struct {
	logger log.Logger
}

// NewTickService creates new tick service.
func NewTickService(logger log.Logger) exchange.TickService {
	return &tickService{logger: logger}
}

// StartReading starts reading ticks from a file and sending it to channel.
func (t *tickService) StartReading(ctx context.Context, ticker string, out chan exchange.Tick) {
	f, err := os.Open(filepath.Join("assets", fmt.Sprintf("%s.txt", ticker)))
	if err != nil {
		t.logger.Error(mainAction, err)
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
	defer timeTicker.Stop()

	now := time.Now().Format("150405")

	var (
		buffer []string
		data   []string
	)

	t.logger.Info(log.Action(ticker), "started")
	defer t.logger.Info(log.Action(ticker), "stopped")

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
			return
		case <-timeTicker.C:
			firstTickTime := strings.TrimSpace(data[timeIdx])
			if err := t.processTick(data, ticker, out); err != nil {
				t.logger.Error(log.Action(ticker), err)
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
					t.logger.Error(log.Action(ticker), err)
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

	vol, err := strconv.ParseInt(data[volIdx], 10, 32)
	if err != nil {
		return err
	}

	out <- exchange.Tick{Ticker: ticker, Price: price, Vol: int32(vol)}

	return nil
}

// Wraps message.
func (t *tickService) wrapMsg(detail, msg string) string {
	return fmt.Sprintf("Tick service (%s): %s", detail, msg)
}
