package handlers

import (
	"encoding/json"
	"fmt"
	"postchi/internal/handlers/requests"
	"postchi/internal/handlers/responses"
	"postchi/internal/metrics"
	"postchi/internal/sms"
	"postchi/pkg/env"
	"postchi/pkg/kafka"
	"postchi/pkg/logger"
	"time"

	"github.com/gofiber/fiber/v2"
)

type SmsHandler struct {
	Envs        *env.Envs
	Metrics     *metrics.Metrics
	Logger      logger.LoggerInterface
	KafkaClient kafka.KafkaInterface
}

type SmsHandlerInterface interface {
	SendExpressSms(c *fiber.Ctx) error
	SensAsyncSms(c *fiber.Ctx) error
}

func SmsHandlerInit(l logger.LoggerInterface, e *env.Envs, m *metrics.Metrics,k kafka.KafkaInterface) SmsHandlerInterface {
	return &SmsHandler{Envs: e, Logger: l, Metrics: m,KafkaClient: k}
}

func (h *SmsHandler) SendExpressSms(c *fiber.Ctx) error {

	var req requests.SendSmsReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json body"})
	}

	prov, err := sms.NewProvider(h.Envs, req.Provider)

	if err != nil {
		h.Logger.StdLog("error", fmt.Sprintf("[sms-express] provider init failed: %v", err))
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "provider unavailable"})
	}

	smsSerrvice := sms.NewService(prov)

	start := time.Now()
	status, _, sendErr := smsSerrvice.Send(req.To, req.Text)

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

	return c.Status(fiber.StatusAccepted).JSON(responses.SendSmsResp{
		Provider:  prov.GetName(),
		Status:    "ok",
		MessageID: "msgID",
	})
}

// regular async messages
func (h *SmsHandler) SensAsyncSms(c *fiber.Ctx) error {
	userId := c.Params("user_id")
	serviceId := c.Params("service_id")
	var req requests.SendSmsReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json body"})
	}
	if req.To == "" || req.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "'to' and 'text' are required"})
	}


	kafkaValue, parseErr := json.Marshal(kafka.SmsKafkaMessage{To: req.To, Content: req.Text, UserId: userId, ServiceId: serviceId})
	if parseErr != nil {
		h.Logger.StdLog("error", fmt.Sprintf("[ar-12] kafka message parser erro %s", parseErr))
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
