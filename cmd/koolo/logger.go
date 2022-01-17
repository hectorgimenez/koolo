package main

import (
	"go.uber.org/zap"
)

func NewLogger(debug bool, logFilePath string) (*zap.Logger, error) {
	cfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:       debug,
		DisableCaller:     false,
		DisableStacktrace: false,
		Encoding:          "console",
		EncoderConfig:     zap.NewDevelopmentEncoderConfig(),
		OutputPaths:       []string{"stdout", logFilePath},
		ErrorOutputPaths:  []string{"stderr", logFilePath},
	}

	return cfg.Build()
}
