package handlers

import (
	"bufio"
	"fmt"
	"io"
	"postchi/internal/metrics"
	"postchi/pkg/env"
	"postchi/pkg/logger"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Envs    *env.Envs
	Logger  logger.LoggerInterface
	Metrics *metrics.Metrics
}

type HandlerInterface interface {
	ServeIndex(c *fiber.Ctx) error
	ServePlaylist(c *fiber.Ctx) error
}

func HandlerInit(HandlerLogger logger.LoggerInterface, envs *env.Envs, metrices *metrics.Metrics) HandlerInterface {
	return &Handler{
		Envs:    envs,
		Logger:  HandlerLogger,
		Metrics: metrices,
	}
}

func (h *Handler) ServeIndex(c *fiber.Ctx) error {
	start := time.Now()
	channel := c.Params("channel")
	stream := c.Params("stream")

	indexDest := fmt.Sprintf("/live/%s/%s/index.m3u8", channel, stream)

	file, err := os.Open(indexDest)
	defer file.Close()
	if err != nil {
		h.Logger.StdLog("error", "[tw-postchi-4.0] index file not found err "+err.Error())
		h.Metrics.IndexStatusCodeCount.WithLabelValues("404", h.Envs.CHANNEL).Inc()
		duration := time.Since(start).Seconds()
		h.Metrics.IndexResponseTimeHistogram.WithLabelValues("404", h.Envs.CHANNEL).Observe(duration)
		return c.SendStatus(http.StatusNotFound)
	}

	reader := bufio.NewReader(file)
	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		h.Logger.StdLog("error", "[tw-postchi-4.1] index file bytes err "+err.Error())
		h.Metrics.IndexStatusCodeCount.WithLabelValues("500", h.Envs.CHANNEL).Inc()
		duration := time.Since(start).Seconds()
		h.Metrics.IndexResponseTimeHistogram.WithLabelValues("500", h.Envs.CHANNEL).Observe(duration)
		return c.SendStatus(http.StatusInternalServerError)
	}
	h.Metrics.IndexStatusCodeCount.WithLabelValues("200", h.Envs.CHANNEL).Inc()
	duration := time.Since(start).Seconds()
	h.Metrics.IndexResponseTimeHistogram.WithLabelValues("200", h.Envs.CHANNEL).Observe(duration)
	return c.SendString(string(bodyBytes))
}

func (h *Handler) ServePlaylist(c *fiber.Ctx) error {
	start := time.Now()
	channel := c.Params("channel")

	playlistDest := fmt.Sprintf("/live/%s/%s.m3u8", channel, channel)

	file, err := os.Open(playlistDest)
	defer file.Close()
	if err != nil {
		h.Logger.StdLog("error", "[tw-postchi-4.2] playlist file not found err "+err.Error())
		h.Metrics.PlaylistStatusCodeCount.WithLabelValues("404", h.Envs.CHANNEL).Inc()
		duration := time.Since(start).Seconds()
		h.Metrics.PlaylistResponseTimeHistogram.WithLabelValues("500", h.Envs.CHANNEL).Observe(duration)
		return c.SendStatus(http.StatusNotFound)
	}

	reader := bufio.NewReader(file)
	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		h.Logger.StdLog("error", "[tw-postchi-4.3] playlist file bytes err "+err.Error())
		h.Metrics.PlaylistStatusCodeCount.WithLabelValues("500", h.Envs.CHANNEL).Inc()
		duration := time.Since(start).Seconds()
		h.Metrics.PlaylistResponseTimeHistogram.WithLabelValues("500", h.Envs.CHANNEL).Observe(duration)
		return c.SendStatus(http.StatusInternalServerError)
	}
	h.Metrics.PlaylistStatusCodeCount.WithLabelValues("200", h.Envs.CHANNEL).Inc()
	duration := time.Since(start).Seconds()
	h.Metrics.PlaylistResponseTimeHistogram.WithLabelValues("200", h.Envs.CHANNEL).Observe(duration)
	return c.SendString(string(bodyBytes))
}
