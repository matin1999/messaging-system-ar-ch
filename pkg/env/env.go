package env

import (
	"log"
	"os"
	"strconv"
)

type Envs struct {
	PROMETHEUS_PORT       string
	APP_PORT              string
	LOG_LEVEL             string
	KAVENEGAR_SMS_API_KEY string
	KAVENEGAR_SMS_NUMBER  string
	KAFKA_BROKERS         string
	KAFKA_TOPIC_SMS       string
	KAFKA_CONSUMER_GROUP  string
	SMS_WORKER_COUNT      int
	DB_DSN                string
	COST_PER_CHAR_EXPRESS int
	COST_PER_CHAR_ASYNC   int
}

func ReadEnvs() Envs {

	envs := Envs{}
	envs.APP_PORT = os.Getenv("APP_PORT")
	envs.PROMETHEUS_PORT = os.Getenv("PROMETHEUS_PORT")
	envs.LOG_LEVEL = os.Getenv("LOG_LEVEL")
	envs.KAVENEGAR_SMS_API_KEY = os.Getenv("KAVENEGAR_SMS_API_KEY")
	envs.KAVENEGAR_SMS_NUMBER = os.Getenv("KAVENEGAR_SMS_NUMBER")
	envs.KAFKA_BROKERS = os.Getenv("KAFKA_BROKERS")
	envs.KAFKA_TOPIC_SMS = os.Getenv("KAFKA_TOPIC_SMS")
	envs.KAFKA_CONSUMER_GROUP = os.Getenv("KAFKA_CONSUMER_GROUP")
	envs.DB_DSN = os.Getenv("DB_DSN")

	workerCount, convErr := strconv.Atoi(os.Getenv("SMS_WORKER_COUNT"))
	if convErr != nil {
		log.Fatalf("Failed to parse worker count")
		envs.SMS_WORKER_COUNT = 10

	} else {
		envs.SMS_WORKER_COUNT = workerCount
	}

	asyncCostperChar, convErr := strconv.Atoi(os.Getenv("COST_PER_CHAR_ASYNC"))
	if convErr != nil {
		log.Fatalf("Failed to parse COST_PER_CHAR_ASYNC")
		envs.COST_PER_CHAR_ASYNC = 1

	} else {
		envs.COST_PER_CHAR_ASYNC = asyncCostperChar
	}

	expressCostPerChar, convErr := strconv.Atoi(os.Getenv("COST_PER_CHAR_EXPRESS"))
	if convErr != nil {
		log.Fatalf("Failed to parse COST_PER_CHAR_EXPRESS")
		envs.SMS_WORKER_COUNT = 3
	} else {
		envs.COST_PER_CHAR_EXPRESS = expressCostPerChar
	}

	return envs
}
