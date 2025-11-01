package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"postchi/internal/handlers/requests"
	"postchi/internal/helpers"
	"postchi/internal/metrics"
	"postchi/pkg/db"
	"postchi/pkg/env"
	"postchi/pkg/logger"
)

type UserManagementHandler struct {
	Envs    *env.Envs
	Logger  logger.LoggerInterface
	Metrics *metrics.Metrics
	Db      db.DataBaseInterface
}

type UserHandlerInterface interface {
	CreateUser(c *fiber.Ctx) error
	CreateServiceForUser(c *fiber.Ctx) error
	ChargeService(c *fiber.Ctx) error
	GetUserServiceStatus(c *fiber.Ctx) error
}

func UserHandlerInit(l logger.LoggerInterface, envs *env.Envs, m *metrics.Metrics, db db.DataBaseInterface) UserHandlerInterface {
	return &UserManagementHandler{
		Envs:    envs,
		Logger:  l,
		Metrics: m,
		Db:      db,
	}
}

// POST /account
// body: { "name": "admin", "password": "admin" }
func (h *UserManagementHandler) CreateUser(c *fiber.Ctx) error {
	var req requests.CreateUserReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json"})
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Password = strings.TrimSpace(req.Password)
	if req.Name == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name and api_key are required"})
	}

	var u db.User
	if err := h.Db.DB().Where("name = ?", req.Name).First(&u).Error; err == nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"result": "user exist",
		})
	}
	if err := h.Db.CreateUser(req.Name, req.Password); err != nil {
		h.Logger.StdLog("error", "CreateUser: "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user_id": u.ID, "name": u.Name, "created": true,
	})
}

// POST /account/:user_id/services/create
// body: { "type": "express" | "async", "initial_credits": 1000 }
func (h *UserManagementHandler) CreateServiceForUser(c *fiber.Ctx) error {

	userID, err := helpers.ParseUintParam(c, "user_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	var req requests.CreateServiceReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json"})
	}
	serviceType, err := helpers.ToServiceType(req.Type)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	err = h.Db.CreateUserService(userID, serviceType, int(req.InitialCredit))
	if err != nil {
		h.Logger.StdLog("error", "CreateServiceForUser: "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create service"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "service created"})
}

// POST /account/:user_id/services/charge
// body: { "credits_amount": 5000 }
func (h *UserManagementHandler) ChargeService(c *fiber.Ctx) error {
	userID, err := helpers.ParseUintParam(c, "user_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	serviceId := c.Query("service_id")
	if serviceId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing service_id"})
	}
	sid64, err := strconv.ParseUint(serviceId, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid service_id"})
	}
	serviceID := uint(sid64)

	var req requests.ChargeReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json"})
	}
	if req.CreditAmount == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "credits_delta must be non-zero"})
	}

	if err := h.Db.UpdateServiceCredit(userID, serviceID, int(req.CreditAmount)); err != nil {
		h.Logger.StdLog("error", "ChargeService: "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update credits"})
	}

	
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":    "charged",
	})
}

// GET /account/:user_id/services/status
func (h *UserManagementHandler) GetUserServiceStatus(c *fiber.Ctx) error {
	userID, err := helpers.ParseUintParam(c, "user_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	svcs, err := h.Db.GetUserServices(userID)
	if err != nil {
		h.Logger.StdLog("error", "GetUserServiceStatus: "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db error"})
	}

	resp := make([]fiber.Map, 0, len(svcs))
	for _, s := range svcs {
		resp = append(resp, fiber.Map{
			"id":      s.ID,
			"type":    s.Type,
			"status":  s.Status,
			"credits": s.Credits,
		})
	}
	return c.JSON(fiber.Map{
		"user_id":  userID,
		"services": resp,
	})
}
