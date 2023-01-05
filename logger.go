package main

import "go.uber.org/zap"

func initLogger() (*zap.SugaredLogger, error) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	logger := zapLogger.Sugar()
	return logger, nil
}
