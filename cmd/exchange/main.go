package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/marksartdev/trading/internal/config"
	"github.com/marksartdev/trading/internal/exchange"
	"github.com/marksartdev/trading/internal/exchange/repository/memory"
	"github.com/marksartdev/trading/internal/exchange/services"
)

func main() {
	loggerConf := zap.NewDevelopmentConfig()
	loggerConf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	loggerConf.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("02.01.2006 15:04:05.000")
	loggerConf.DisableCaller = true
	loggerConf.DisableStacktrace = true

	logger, err := loggerConf.Build()
	if err != nil {
		log.Fatal(err)
	}

	sugar := logger.Sugar()

	cfg, err := config.Load("default")
	if err != nil {
		sugar.Fatal(err)
	}

	dealQueue := memory.NewDealQueue()
	tickService := services.NewTickService(sugar)
	exchangeService := services.NewExchangeService(sugar, dealQueue, tickService, cfg.Tickers)

	// todo remove it.
	broker := make(chan exchange.Deal, 10)

	go func() {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

		<-done
		fmt.Println()
		exchangeService.Stop()
		close(broker) // todo remove it.
	}()

	// todo remove it.
	deal := exchangeService.Create(exchange.Deal{
		BrokerID: 2,
		ClientID: 1,
		Ticker:   "SPFB.Si",
		Amount:   10,
		Partial:  false,
		Time:     time.Now(),
		Price:    -1000,
	})
	sugar.Info(deal)

	// todo remove it.
	exchangeService.Results(2, broker)

	go exchangeService.Start()

	// todo remove it.
	for deal := range broker {
		fmt.Println(deal)
	}
}
