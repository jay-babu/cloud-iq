package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger  *zap.Logger
	SLogger *zap.SugaredLogger
)

func Stage() (stage string) {
	stage = "DEV"
	if stage, ok := os.LookupEnv("STAGE"); ok {
		return stage
	}
	return
}

func init() {
	var config zap.Config
	if Stage() == "DEV" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	Logger, _ = config.Build()
	SLogger = Logger.Sugar()
}
