package main

import (
	"google.golang.org/grpc"

	"github.com/marksartdev/trading/internal/app"
	brokerRpc "github.com/marksartdev/trading/internal/broker/delivery/rpc"
	clientRpc "github.com/marksartdev/trading/internal/client/delivery/rpc"
	"github.com/marksartdev/trading/internal/client/services"
	"github.com/marksartdev/trading/internal/log"
)

func main() {
	logger, cfg := app.Init()

	conn, err := grpc.Dial(":8001", grpc.WithInsecure())
	if err != nil {
		logger.Fatal(err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			logger.Error(err)
		}
	}(conn)

	brokerLogger := log.NewLogger(logger, "Broker", log.Purple())
	brokerClient := brokerRpc.NewBrokerClient(conn)
	brokerService := clientRpc.NewBrokerService(brokerLogger, brokerClient)

	botLogger := log.NewLogger(logger, "Client", log.Green())

	bot := services.NewTelegramBot(botLogger, cfg.Client, brokerService)
	if err := bot.Start(); err != nil {
		logger.Fatal(err)
	}
}
