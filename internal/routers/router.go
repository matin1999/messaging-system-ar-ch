package router

import (
	"postchi/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, userH handlers.UserHandlerInterface, smsH handlers.SmsHandlerInterface) {

	app.Get("/health", func(c *fiber.Ctx) error {
		err := c.SendString("API is UP!")
		return err
	})

	// user account
	// 1 user status account charge + count of sms sens + accont type + message rate
	// 2 sms that acount send by pagination
	// 3 charging account
	app.Get("account/:user_id/services/status", userH.GetUserServiceStatus)
	app.Get("account/:user_id/services/create", userH.GetUserServiceStatus)
	app.Get("account/:user_id/services/charge", userH.GetUserServiceStatus)

	// user send message
	// 1 send regulur message
	// 2 express message
	// 3 chack status of specific message
	app.Post("/sms/:user_id/:service_id/express/send", smsH.SendExpressSms)
	app.Post("/sms/:user_id/:service_id/indirect/send", smsH.SendExpressSms)

}
