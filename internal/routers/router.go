package router

import (
	"katana/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, h handlers.HandlerInterface) {

	app.Get("/health", func(c *fiber.Ctx) error {
		err := c.SendString("API is UP!")
		return err
	})

	app.Get("/:channel/live/playlist.m3u8", h.ServePlaylist)
	app.Get("/:channel/live/:stream/index.m3u8", h.ServeIndex)
}
