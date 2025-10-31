package handlers

import (
	"log"
	"postchi/internal/metrics"
	"postchi/internal/sms"
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
	provider, err := sms.NewProvider(h.Envs,"kavenegar")
	if err != nil {
		log.Fatal(err)
		h.Logger.StdLog("error",err.Error())
	}

	smsService := sms.NewService(provider)

	messageId,messageStatus,messageErr := smsService.Send("+989123456789", "Hello from Go SMS Gateway!")
	if messageErr != nil {
		log.Println("Failed to send:", err)
	} else {
		log.Println("Message sent successfully!")
	}
}
