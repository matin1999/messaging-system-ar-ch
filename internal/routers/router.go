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

	app.Get("account/createuser", userH.CreateUser)

	app.Get("account/:user_id/services/status", userH.GetUserServiceStatus)
	app.Get("account/:user_id/services/create", userH.CreateServiceForUser)
	app.Post("account/:user_id/services/charge", userH.ChargeService)

	app.Post("/sms/:user_id/:service_id/express/send", smsH.SendExpressSms)
	app.Post("/sms/:user_id/:service_id/async/send", smsH.SensAsyncSms)

}
