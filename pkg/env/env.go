package env

import (
	"os"
)

type Envs struct {
	PROMETHEUS_PORT       string
	APP_PORT              string
	LOG_LEVEL             string
	KAVENEGAR_SMS_API_KEY string
	KAVENEGAR_SMS_NUMBER  string
	KAFKA_BROKERS         string
	KAFKA_TOPIC_SMS       string
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

	return envs
}
