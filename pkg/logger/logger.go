package logger

import (
	"time"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

type ElasticLog struct {
	ServiceLogCode string
	Type           string
	Status         string
	StorageStatus  string
	MongoStatus    string
	Channel        string
	Variant        string
	FfmpegStatus   string
	BucketName     string
	FileSize       int
	Timestamp      time.Time
}

type Logger struct {
	Log     *zap.Logger
	elastic *elastic.Elastic[ElasticLog]
	env     *env.Envs
}

type LoggerInterface interface {
	StdLog(level string, message string) LoggerInterface
}

func Init(envs *env.Envs, elasticChannelSize int, elasticWorkers int) (LoggerInterface, error) {
	logger := Logger{env: envs}
	var err error
	if envs.PRODUCTION_MODE == "true" {
		logger.Log, err = zap.NewProduction()
		if err != nil {
			return nil, err
		}
	} else {
		logger.Log, err = zap.NewDevelopment()
		if err != nil {
			return nil, err
		}
	}
	logger.elastic = elastic.Initial[ElasticLog](envs.ELASTIC_URL, envs.ELASTIC_USERNAME, envs.ELASTIC_PASSWORD, envs.ELASTIC_INDEXNAME, elasticChannelSize, elasticWorkers, logger.Log)
	return &logger, nil
}

func (logger *Logger) StdLog(level string, message string) LoggerInterface {
	switch  {
	case level == "info" && (logger.env.LOG_LEVEL == "info"):
		logger.Log.Info(message)
	case level == "warn":
		logger.Log.Warn(message)
	case level == "error":
		logger.Log.Error(message)
	case level == "panic":
		logger.Log.Panic(message)
	}
	return logger
}


