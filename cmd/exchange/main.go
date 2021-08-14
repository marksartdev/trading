package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/marksartdev/trading/internal/app"
	"github.com/marksartdev/trading/internal/exchange/delivery/rpc"
	"github.com/marksartdev/trading/internal/exchange/repository/memory"
	"github.com/marksartdev/trading/internal/exchange/services"
	"github.com/marksartdev/trading/internal/log"
)

const http log.Action = "http"

func main() {
	logger, cfg := app.Init()

	dealQueue := memory.NewDealQueue()
	tickLogger := log.NewLogger(logger, "Ticker", log.Blue())
	ticks := services.NewTickService(tickLogger)

	exchangeLogger := log.NewLogger(logger, "Exchanger", log.Purple())
	service := services.NewExchangeService(exchangeLogger, dealQueue, ticks, cfg.Exchange.Tickers, cfg.Exchange.Interval)

	srvLogger := log.NewLogger(logger, "Server", log.Green())
	grpcServer := rpc.NewExchangeServer(srvLogger, service)

	lis, err := net.Listen("tcp", ":8000")
	if err != nil {
		logger.Fatal(err)
	}

	srv := grpc.NewServer()
	rpc.RegisterExchangeServer(srv, grpcServer)

	go func() {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

		<-done
		fmt.Println()
		service.Stop()
	}()

	go func() {
		service.Start()
		srv.Stop()
	}()

	srvLogger.Info(http, "started")
	if err := srv.Serve(lis); err != nil {
		logger.Fatal(err)
	}
	srvLogger.Info(http, "stopped")
}
