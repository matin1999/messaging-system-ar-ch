package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"postchi/internal/handlers/requests"
	"postchi/internal/helpers"
	"postchi/internal/metrics"
	"postchi/internal/sms"
	"postchi/pkg/db"
	"postchi/pkg/env"
	"postchi/pkg/kafka"
	"postchi/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

type SmsHandler struct {
	Envs        *env.Envs
	Metrics     *metrics.Metrics
	Logger      logger.LoggerInterface
	KafkaClient kafka.KafkaInterface
	Db          db.DataBaseInterface
}

type SmsHandlerInterface interface {
	SendExpressSms(c *fiber.Ctx) error
	SendAsyncSms(c *fiber.Ctx) error
}

func SmsHandlerInit(l logger.LoggerInterface, e *env.Envs, m *metrics.Metrics, k kafka.KafkaInterface, d db.DataBaseInterface) SmsHandlerInterface {
	return &SmsHandler{Envs: e, Logger: l, Metrics: m, KafkaClient: k, Db: d}
}

func (h *SmsHandler) SendExpressSms(c *fiber.Ctx) error {

	var req requests.SendSmsReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json body"})
	}

	userIDStr := c.Params("user_id")
	serviceIDStr := c.Params("service_id")
	var uid64, sid64 uint64
	if userIDStr != "" {
		var err error
		uid64, err = strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
		}
	}
	if serviceIDStr != "" {
		var err error
		sid64, err = strconv.ParseUint(serviceIDStr, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid service_id"})
		}
	}

	prov, err := sms.NewProvider(h.Envs, req.Provider)

	if err != nil {
		h.Logger.StdLog("error", fmt.Sprintf("[sms-express] provider init failed: %v", err))
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "provider unavailable"})
	}

	smsSerrvice := sms.NewService(prov)

	start := time.Now()
	status, msgID, sendErr := smsSerrvice.Send(req.To, req.Text)

	elapsed := time.Since(start).Seconds()
	if h.Metrics != nil {
		h.Metrics.SmsProviderResponseTimeHistogram.WithLabelValues(prov.GetName()).Observe(elapsed)
	}

	if sendErr != nil {
		// log and expose provider error metrics
		if h.Metrics != nil {
			h.Metrics.SmsProviderErrors.WithLabelValues(prov.GetName()).Inc()
		}
		h.Logger.StdLog("error", fmt.Sprintf("[sms-express] send failed: %v", sendErr))
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  status,
			"error":   sendErr.Error(),
			"message": "send failed",
		})
	}
	cost := helpers.CalculateCost(h.Envs, req.Text, "express")

	smsRecord := &db.Sms{
		Content:                  req.Text,
		Receptor:                 req.To,
		Status:                   "sent",
		SentTime:                 time.Now().Unix(),
		Cost:                     cost,
		ServiceProviderName:      prov.GetName(),
		ServiceProviderMessageId: msgID,
		ServiceId:                uint(sid64),
	}
	if err := h.Db.CreateSmsAndSpendCredit(uint(uid64), uint(sid64), smsRecord, 1); err != nil {
		h.Logger.StdLog("error", fmt.Sprintf("[sms-express] failed to persist SMS or deduct credit: %v", err))
		if err.Error() == "insufficient credits" {
			return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{"error": "insufficient credits"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db error"})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"Status": "ok",
	})
}

// regular async messages
func (h *SmsHandler) SendAsyncSms(c *fiber.Ctx) error {
	userIdParam := c.Params("user_id")
	serviceIdParam := c.Params("service_id")
	var req requests.SendSmsReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json body"})
	}
	if req.To == "" || req.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "'to' and 'text' are required"})
	}

	var smsRecordId uint
	var serviceId int
	var userId int

	var err error

	cost := uint(helpers.CalculateCost(h.Envs, req.Text, "async"))

	serviceId, err = strconv.Atoi(serviceIdParam)
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid service id param "})
	}
	userId, err = strconv.Atoi(userIdParam)
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id param "})
	}

	smsRecord := &db.Sms{
		Content:                  req.Text,
		Receptor:                 req.To,
		Status:                   "queued",
		SentTime:                 time.Now().Unix(),
		Cost:                     0,
		ServiceProviderName:      req.Provider,
		ServiceProviderMessageId: 0,
		ServiceId:                uint(serviceId),
	}
	if err := h.Db.CreateSmsAndSpendCredit(uint(userId), uint(serviceId), smsRecord, cost); err != nil {
		h.Logger.StdLog("error", fmt.Sprintf("[sms-async] failed to persist queued SMS record: %v", err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db error"})
	}
	smsRecordId = smsRecord.ID

	msgWithId := kafka.SmsKafkaMessage{
		To:        req.To,
		Content:   req.Text,
		Provider:  req.Provider,
		UserId:    uint(userId),
		ServiceId: uint(serviceId),
		SmsId:     smsRecordId,
	}
	kafkaValue, parseErr := json.Marshal(msgWithId)
	if parseErr != nil {
		h.Logger.StdLog("error", fmt.Sprintf("[sms-async] kafka message parser erro %s", parseErr))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json body"})
	}
	if err := h.KafkaClient.Publish(c.Context(), req.Provider, kafkaValue); err != nil {
		h.Logger.StdLog("error", "kafka publish failed: "+err.Error())
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to enqueue"})
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"status": "queued",
		"topic":  "sms_send",
		"to":     req.To,
	})
}
