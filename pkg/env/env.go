package env

import (
	"os"
)

type Envs struct {
	PROMETHEUS_PORT string
	APP_PORT        string
	LOG_LEVEL       string
}

func ReadEnvs() Envs {
	envs := Envs{}
	envs.APP_PORT = os.Getenv("APP_PORT")
	envs.PROMETHEUS_PORT = os.Getenv("PROMETHEUS_PORT")
	envs.LOG_LEVEL = os.Getenv("LOG_LEVEL")

	return envs
}
