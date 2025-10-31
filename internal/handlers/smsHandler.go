package handlers

import (
	"context"
	"fmt"
	"postchi/internal/handlers/requests"
	"postchi/internal/handlers/responses"
	"postchi/internal/metrics"
	"postchi/internal/sms"
	"time"
	"github.com/gofiber/fiber/v2"
	"postchi/pkg/env"
	"postchi/pkg/logger"
)

type SmsHandler struct {
	Envs    *env.Envs
	Logger  logger.LoggerInterface
	Metrics *metrics.Metrics
}

type SmsHandlerInterface interface {
	SendExpressSms(c *fiber.Ctx) error
	SensAsyncSms(c *fiber.Ctx) error
}

func SmsHandlerInit(l logger.LoggerInterface, e *env.Envs, m *metrics.Metrics) SmsHandlerInterface {
	return &SmsHandler{Envs: e, Logger: l, Metrics: m}
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

	return c.Status(fiber.StatusAccepted).JSON(responses.SendSmsResp{
		Provider:  prov.GetName(),
		Status:    status,
		MessageID: msgID,
	})
}

// regular async messages
func (h *SmsHandler) SensAsyncSms(c *fiber.Ctx) error {
	var req requests.SendSmsReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json body"})
	}
	if req.To == "" || req.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "'to' and 'text' are required"})
	}



	prov, err := sms.NewProvider(h.Envs, req.Provider)
	if err != nil {
		h.Logger.StdLog("error", fmt.Sprintf("[sms-indirect] provider init failed: %v", err))
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "provider unavailable"})
	}
	svc := sms.NewService(prov)

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	select {
	case <-time.After(250 * time.Millisecond):
	case <-ctx.Done():
		return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{"error": "timed out before dispatch"})
	}

	start := time.Now()
	status, msgID, sendErr := svc.Send(req.To, req.Text)
	elapsed := time.Since(start).Seconds()

	if h.Metrics != nil {
		h.Metrics.SmsProviderResponseTimeHistogram.WithLabelValues(prov.GetName()).Observe(elapsed)
	}

	if sendErr != nil {
		if h.Metrics != nil {
			h.Metrics.SmsProviderErrors.WithLabelValues(prov.GetName()).Inc()
		}
		h.Logger.StdLog("error", fmt.Sprintf("[sms-indirect] send failed: %v", sendErr))
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  status,
			"error":   sendErr.Error(),
			"message": "send failed",
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(responses.SendSmsResp{
		Provider:  prov.GetName(),
		Status:    status,
		MessageID: msgID,
	})
}
