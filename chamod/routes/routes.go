package routes

import (
	"chamod/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/process-untraced", handlers.UntracedHandler)
	app.Get("/process-traced", handlers.TracedHandler)
}
