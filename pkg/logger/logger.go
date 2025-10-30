package logger

import (
	"postchi/pkg/env"
	"time"

	"go.uber.org/zap"
)

type ElasticLog struct {
	ServiceLogCode string
	Channel        string
	UserAgent      string
	Timestamp      time.Time
}

type Logger struct {
	Log *zap.Logger
	env *env.Envs
}

type LoggerInterface interface {
	StdLog(level string, message string)
}

func Init(envs *env.Envs) (LoggerInterface, error) {
	logger := Logger{env: envs}
	var err error
	logger.Log, err = zap.NewProduction()
	if err != nil {
		return nil, err
	}

	return &logger, nil
}

func (logger *Logger) StdLog(level string, message string) {
	switch {
	case level == "info" && (logger.env.LOG_LEVEL == "info"):
		logger.Log.Info(message)
	case level == "warn":
		logger.Log.Warn(message)
	case level == "error":
		logger.Log.Info(message)
	case level == "panic":
		logger.Log.Panic(message)
	}
}
