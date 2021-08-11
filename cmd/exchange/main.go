package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"

	"github.com/marksartdev/trading/internal/config"
	"github.com/marksartdev/trading/internal/exchange/delivery/rpc"
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

	exchangeService := services.NewExchangeService(sugar, dealQueue, tickService, cfg.Tickers, cfg.Interval)
	exchangeServer := rpc.NewExchangeServer(sugar, exchangeService)

	lis, err := net.Listen("tcp", ":8000")
	if err != nil {
		sugar.Fatal(err)
	}

	srv := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_zap.StreamServerInterceptor(logger))),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_zap.UnaryServerInterceptor(logger))),
	)
	rpc.RegisterExchangeServer(srv, exchangeServer)

	go func() {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

		<-done
		fmt.Println()
		exchangeService.Stop()
		srv.Stop()
	}()

	go exchangeService.Start()

	if err := srv.Serve(lis); err != nil {
		sugar.Fatal(err)
	}
}
