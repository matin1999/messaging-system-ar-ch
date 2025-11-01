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
// body: { "name": "Alice", "api_key": "api_alice_123" }
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
// body: { "type": "express" | "indirect", "initial_credits": 1000 }
func (h *UserManagementHandler) CreateServiceForUser(c *fiber.Ctx) error {

	userID, err := helpers.ParseUintParam(c, "user_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	var req requests.CreateServiceReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json"})
	}
	typ, err := helpers.ToServiceType(req.Type)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	svc, err := h.Db.CreateUserService(userID, typ)
	if err != nil {
		h.Logger.StdLog("error", "CreateServiceForUser: "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create service"})
	}

	if req.InitialCreds > 0 && svc.Credits < req.InitialCreds {
		if e := h.Store.DB().Model(&db.Service{}).
			Where("id = ?", svc.ID).
			Update("credits", req.InitialCreds).Error; e != nil {
			h.Logger.StdLog("warn", "initial credit bump failed: "+e.Error())
		}
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"service_id": svc.ID,
		"user_id":    svc.UserID,
		"type":       svc.Type,
		"status":     svc.Status,
		"credits":    svc.Credits,
	})
}

// POST /account/:user_id/services/charge
// body: { "credits_delta": 500 }  (negative to deduct)
func (h *UserManagementHandler) ChargeService(c *fiber.Ctx) error {
	if h.Store == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "db unavailable"})
	}
	userID, err := parseUintParam(c, "user_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	serviceIDRaw := c.Query("service_id") // allow as query ?service_id=...
	if serviceIDRaw == "" {
		serviceIDRaw = c.Params("service_id") // or path if you prefer /:service_id
	}
	if serviceIDRaw == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing service_id"})
	}
	sid64, err := strconv.ParseUint(serviceIDRaw, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid service_id"})
	}
	serviceID := uint(sid64)

	var req requests.ChargeReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json"})
	}
	if req.CreditsDelta == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "credits_delta must be non-zero"})
	}

	if err := h.Store.ChargeUserService(userID, serviceID, req.CreditsDelta); err != nil {
		h.Logger.StdLog("error", "ChargeService: "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update credits"})
	}

	// return the new balance
	var svc db.Service
	if e := h.Store.DB().Where("id = ? AND user_id = ?", serviceID, userID).First(&svc).Error; e != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "charged", "credits": "unknown"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":    "charged",
		"user_id":    userID,
		"service_id": serviceID,
		"credits":    svc.Credits,
	})
}

// GET /account/:user_id/services/status
func (h *UserManagementHandler) GetUserServiceStatus(c *fiber.Ctx) error {
	if h.Store == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "db unavailable"})
	}
	userID, err := parseUintParam(c, "user_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	svcs, err := h.Store.GetUserServices(userID)
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
