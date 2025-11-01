package main

import (
	"fmt"
	"postchi/internal/handlers"
	router "postchi/internal/routers"

	"postchi/internal/metrics"
	"postchi/pkg/db"
	"postchi/pkg/env"
	"postchi/pkg/kafka"
	"postchi/pkg/logger"

	"github.com/gofiber/fiber/v2"
	fiber_logger "github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {

	app := fiber.New()
	app.Use(fiber_logger.New())

	envs := env.ReadEnvs()
	logger, loggerErr := logger.Init(&envs)
	if loggerErr != nil {
		fmt.Println("logger error" + loggerErr.Error())
	}
	logger.StdLog("error", "[ar-0.0] postchi service started")
	metric := metrics.InitMetrics()

	kafkaClient, err := kafka.Init(envs.KAFKA_BROKERS, envs.KAFKA_TOPIC_SMS)
	if err != nil {
		logger.StdLog("error", fmt.Sprintf("[worker] kafka init failed: %v", err))
		panic("worker cannot run kafka not initialized with err " + err.Error())
	}

	DbClient, err := db.Init(envs.DB_DSN)
	if err != nil {
		logger.StdLog("error", fmt.Sprintf("[worker] kafka init failed: %v", err))
		panic("worker cannot run kafka not initialized with err " + err.Error())
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.StdLog("error", fmt.Sprintf("[ar-0.5] prometheus recovered from panic: %v", r))
			}
		}()
		metrics.StartMetricsServer(envs.PROMETHEUS_PORT)
	}()

	userHandler := handlers.UserHandlerInit(logger, &envs, metric, DbClient)
	smsHandler := handlers.SmsHandlerInit(logger, &envs, metric, kafkaClient,DbClient)

	router.SetupRoutes(app, userHandler, smsHandler)

	err = app.Listen(fmt.Sprintf(":%s", envs.APP_PORT))
	if err != nil {
		logger.StdLog("error", fmt.Sprintf("[ar-0.3] app listen error %s", err))
	}

}
