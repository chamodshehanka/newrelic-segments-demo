package routes

import (
	"nisansala/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/compute-untraced", handlers.ComputeUntraced)
	app.Get("/compute-traced", handlers.ComputeTraced)
}
