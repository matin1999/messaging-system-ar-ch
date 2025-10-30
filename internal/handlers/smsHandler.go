package handlers

import (
	"errors"
	"postchi/internal/metrics"
	"postchi/pkg/env"
	"postchi/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

type SmsHandler struct {
	Envs    *env.Envs
	Logger  logger.LoggerInterface
	Metrics *metrics.Metrics
}

type SmsHandlerInterface interface {
	SendExpressSms(c *fiber.Ctx) error
}

func SmsHandlerInit(HandlerLogger logger.LoggerInterface, envs *env.Envs, metrices *metrics.Metrics) SmsHandlerInterface {
	return &SmsHandler{
		Envs:    envs,
		Logger:  HandlerLogger,
		Metrics: metrices,
	}
}

func (h *SmsHandler) SendExpressSms(c *fiber.Ctx) error {
	return errors.New("test")
}
