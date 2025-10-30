package handlers

import (
	"postchi/internal/metrics"
	"postchi/pkg/env"
	"postchi/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

type UserManagementHandler struct {
	Envs    *env.Envs
	Logger  logger.LoggerInterface
	Metrics *metrics.Metrics
}

type UserHandlerInterface interface {
	GetUserServiceStatus(c *fiber.Ctx) error
}

func UserHandlerInit(HandlerLogger logger.LoggerInterface, envs *env.Envs, metrices *metrics.Metrics) UserHandlerInterface {
	return &UserManagementHandler{
		Envs:    envs,
		Logger:  HandlerLogger,
		Metrics: metrices,
	}
}

func (h *UserManagementHandler) GetUserServiceStatus(c *fiber.Ctx) error {
	
}
