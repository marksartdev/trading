package app

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/marksartdev/trading/internal/config"
)

// Init setups logger and loads config.
func Init() (*zap.SugaredLogger, config.Config) {
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

	return sugar, cfg
}
