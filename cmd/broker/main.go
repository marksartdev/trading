package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/marksartdev/trading/internal/app"
	"github.com/marksartdev/trading/internal/broker"
	"github.com/marksartdev/trading/internal/broker/database"
	brokerRpc "github.com/marksartdev/trading/internal/broker/delivery/rpc"
	"github.com/marksartdev/trading/internal/broker/repository"
	"github.com/marksartdev/trading/internal/broker/services"
	exchangeRpc "github.com/marksartdev/trading/internal/exchange/delivery/rpc"
	"github.com/marksartdev/trading/internal/log"
)

const http log.Action = "http"

func main() {
	logger, cfg := app.Init()

	db, err := database.New(cfg.Broker.DB)
	if err != nil {
		logger.Fatal(err)
	}

	clientRepo := repository.NewClientRepo(db)
	dealRepo := repository.NewDealRepo(db)
	posRepo := repository.NewPositionRepo(db)
	statRepo := repository.NewStatisticRepo(db)

	conn, err := grpc.Dial(":8000", grpc.WithInsecure())
	if err != nil {
		logger.Fatal(err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			logger.Error(err)
		}
	}(conn)

	exchangeClient := exchangeRpc.NewExchangeClient(conn)
	exchangeService := brokerRpc.NewExchangeService(1, exchangeClient)
	serviceLogger := log.NewLogger(logger, "Broker", log.Blue())

	service := services.NewBrokerService(serviceLogger, clientRepo, dealRepo, posRepo, statRepo, exchangeService)
	_, err = service.Create(broker.Deal{
		ClientID: 1,
		Ticker:   "SPFB.RTS",
		Type:     broker.Sell,
		Amount:   400,
		Price:    1000,
	})
	if err != nil {
		logger.Fatal(err)
	}

	go func() {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

		<-done
		fmt.Println()
		service.Stop()
	}()

	srvLogger := log.NewLogger(logger, "Server", log.Green())

	srvLogger.Info(http, "started")
	service.Start()
	srvLogger.Info(http, "stopped")
}
