package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"postchi/internal/handlers/requests"
	"postchi/internal/handlers/responses"
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
	SensAsyncSms(c *fiber.Ctx) error
}

func SmsHandlerInit(l logger.LoggerInterface, e *env.Envs, m *metrics.Metrics, k kafka.KafkaInterface, d db.DataBaseInterface) SmsHandlerInterface {
	return &SmsHandler{Envs: e, Logger: l, Metrics: m, KafkaClient: k, Db: d}
}

func (h *SmsHandler) SendExpressSms(c *fiber.Ctx) error {

	userIdParam := c.Params("user_id")
	serviceIdParam := c.Params("service_id")
	var req requests.SendSmsReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json body"})
	}

	serviceId, err := strconv.ParseUint(serviceIdParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
	}
	userId, err := strconv.ParseUint(userIdParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid service_id"})
	}
	if err := h.Db.SpendServiceCredit(uint(userId), uint(serviceId), 1); err != nil {
		h.Logger.StdLog("error", fmt.Sprintf("[sms-express] spend credit failed: %v", err))
		return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{"error": "insufficient credits"})
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

	serviceObj, dbErr := h.Db.GetService(uint(serviceId))
	if dbErr != nil {
		h.Logger.StdLog("error", "kafka publish failed: "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save message"})
	}

	if h.Db != nil {
		smsRecord := &db.Sms{
			Content:                  req.Text,
			SmsStatus:                "sent",
			Receptor:                 req.To,
			Status:                   "sent",
			SentTime:                 time.Now().Unix(),
			Cost:                     0,
			ServiceProviderName:      prov.GetName(),
			ServiceProviderMessageId: msgID,
			Service:                  *serviceObj,
		}
		if err := h.Db.CreateSmsRecord(smsRecord); err != nil {
			h.Logger.StdLog("error", fmt.Sprintf("[sms-express] failed to persist SMS record: %v", err))
		}
	}

	return c.Status(fiber.StatusAccepted).JSON(responses.SendSmsResp{
		Provider:  prov.GetName(),
		Status:    "ok",
		MessageID: fmt.Sprintf("%d", msgID),
	})
}

// regular async messages
func (h *SmsHandler) SensAsyncSms(c *fiber.Ctx) error {
	userIdParam := c.Params("user_id")
	serviceIdParam := c.Params("service_id")
	var req requests.SendSmsReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json body"})
	}
	if req.To == "" || req.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "'to' and 'text' are required"})
	}

	userId, err := strconv.ParseUint(userIdParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
	}
	serviceId, err := strconv.ParseUint(serviceIdParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid service_id"})
	}

	kafkaValue, parseErr := json.Marshal(kafka.SmsKafkaMessage{To: req.To, Content: req.Text, UserId: userIdParam, ServiceId: serviceIdParam})
	if parseErr != nil {
		h.Logger.StdLog("error", fmt.Sprintf("[ar-12] kafka message parser erro %s", parseErr))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json body"})
	}

	if err := h.KafkaClient.Publish(c.Context(), req.Provider, kafkaValue); err != nil {
		h.Logger.StdLog("error", "kafka publish failed: "+err.Error())
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to enqueue"})
	}

	serviceObj, dbErr := h.Db.GetService(uint(serviceId))
	if dbErr != nil {
		h.Logger.StdLog("error", "kafka publish failed: "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save message"})
	}

	smsRecord := &db.Sms{
		Content:                  req.Text,
		SmsStatus:                "queued",
		Receptor:                 req.To,
		Status:                   "queued",
		SentTime:                 time.Now().Unix(),
		Cost:                     0,
		ServiceProviderName:      req.Provider,
		ServiceProviderMessageId: 0,
		Service:                  *serviceObj,
	}
	if err := h.Db.CreateSmsRecord(smsRecord); err != nil {
		h.Logger.StdLog("error", fmt.Sprintf("[sms-async] failed to persist queued SMS record: %v", err))
	}

	if err := h.Db.SpendServiceCredit(uint(userId), uint(serviceId), 1); err != nil {
		h.Logger.StdLog("error", fmt.Sprintf("[sms-async] spend credit failed: %v", err))
		return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{"error": "insufficient credits"})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"status": "queued",
		"topic":  "sms_send",
		"to":     req.To,
	})
}
