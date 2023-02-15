package log

import (
	"log"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	encoderConfig := zap.NewProductionConfig()
	encoderConfig.Level.SetLevel(zapcore.InfoLevel)
	if os.Getenv("env") == "local" {
		encoderConfig.Level.SetLevel(zapcore.DebugLevel)
	}
	encoderConfig.EncoderConfig.TimeKey = "timestamp"
	encoderConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339Nano)

	zapLogger, err := encoderConfig.Build()
	if err != nil {
		log.Fatalf("fail to build log. err: %s", err)
	}

	logger = zapLogger.With(zap.String("app", "backend-go-service"))
}

func Logger() *zap.Logger {
	return logger
}
