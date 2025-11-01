package handlers

import (
	"strconv"
	"github.com/gofiber/fiber/v2"
	"postchi/internal/metrics"
	"postchi/pkg/env"
	"postchi/pkg/logger"
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
	uid := c.Params("user_id")
	if uid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing user_id"})
	}
	_, _ = strconv.Atoi(uid) 
	return c.JSON(fiber.Map{
		"user_id": uid,
		"services": []fiber.Map{
			{"type": "express", "status": "active"},
			{"type": "indirect", "status": "active"},
		},
	})
}
