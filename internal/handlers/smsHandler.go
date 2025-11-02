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
	return &SmsHandler{Envs: e, Logger: l, Metrics: m, KafkaClient: k,Db: d}
}

func (h *SmsHandler) SendExpressSms(c *fiber.Ctx) error {

	var req requests.SendSmsReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json body"})
	}

    // Parse user_id and service_id from the URL and attempt to spend one credit
    // before sending the message.  If the service does not have enough
    // credits, return an error.
    userIDStr := c.Params("user_id")
    serviceIDStr := c.Params("service_id")
    var uid64, sid64 uint64
    // If a user and service are specified, attempt to spend one credit
    if userIDStr != "" && serviceIDStr != "" && h.Db != nil {
        var err error
        uid64, err = strconv.ParseUint(userIDStr, 10, 64)
        if err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
        }
        sid64, err = strconv.ParseUint(serviceIDStr, 10, 64)
        if err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid service_id"})
        }
        // spend one credit (cost per SMS)
        if err := h.Db.SpendServiceCredit(uint(uid64), uint(sid64), 1); err != nil {
            h.Logger.StdLog("error", fmt.Sprintf("[sms-express] spend credit failed: %v", err))
            return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{"error": "insufficient credits"})
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

    // persist the message into the database
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
            ServiceID:                uint(sid64),
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

    // Attempt to spend one credit before enqueuing the message.  Reject if no
    // credits remain.
    if userId != "" && serviceId != "" && h.Db != nil {
        uid64, err := strconv.ParseUint(userId, 10, 64)
        if err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
        }
        sid64, err := strconv.ParseUint(serviceId, 10, 64)
        if err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid service_id"})
        }
        if err := h.Db.SpendServiceCredit(uint(uid64), uint(sid64), 1); err != nil {
            h.Logger.StdLog("error", fmt.Sprintf("[sms-async] spend credit failed: %v", err))
            return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{"error": "insufficient credits"})
        }
    }

    if err := h.KafkaClient.Publish(c.Context(), req.Provider, kafkaValue); err != nil {
        h.Logger.StdLog("error", "kafka publish failed: "+err.Error())
        return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to enqueue"})
    }

    // persist the queued message into the database
    if h.Db != nil {
        // parse service id to set on the record (already parsed above for credit)
        var sidUint uint
        if serviceId != "" {
            if sid64, err := strconv.ParseUint(serviceId, 10, 64); err == nil {
                sidUint = uint(sid64)
            }
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
            ServiceID:                sidUint,
        }
        if err := h.Db.CreateSmsRecord(smsRecord); err != nil {
            h.Logger.StdLog("error", fmt.Sprintf("[sms-async] failed to persist queued SMS record: %v", err))
        }
    }

    return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
        "status": "queued",
        "topic":  "sms_send",
        "to":     req.To,
    })
}
