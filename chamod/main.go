package main

import (
	"chamod/configs"
	"chamod/routes"
	"chamod/utils"
	"fmt"
	"log"

	"github.com/gofiber/contrib/fibernewrelic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	config := configs.GetConfig()
	port := config.Port
	if port == 0 {
		panic("Port is not set in the configuration")
	}
	utils.SetLogLevel(0)

	app := fiber.New()
	newrelicApp := utils.SetupNewRelic(config)
	if newrelicApp != nil {
		app.Use(fibernewrelic.New(fibernewrelic.Config{
			Application: newrelicApp,
			Enabled:     true,
		}))
	}
	app.Use(requestid.New())

	routes.SetupRoutes(app)
	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
