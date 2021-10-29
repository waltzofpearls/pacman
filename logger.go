package main

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newLogger() (*zap.Logger, error) {
	conf := zap.NewProductionConfig()
	conf.Encoding = "console"
	conf.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	return conf.Build()
}
