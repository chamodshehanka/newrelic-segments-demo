package main

import (
	"fmt"
	"log"
	"nisansala/configs"
	"nisansala/utils"

	"nisansala/routes"

	"github.com/gofiber/contrib/fibernewrelic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	// create fiber app
	config := configs.GetConfig()
	port := config.Port
	if port == 0 {
		panic("Port is not set in the configuration")
	}
	utils.SetLogLevel(0)
	app := fiber.New()

	// Setup New Relic using environment variables
	newRelicApp := utils.SetupNewRelic(config)
	if newRelicApp != nil {
		app.Use(fibernewrelic.New(fibernewrelic.Config{
			Application: newRelicApp,
			Enabled:     true,
		}))
	}
	app.Use(func(c *fiber.Ctx) error {
		log.Printf("Incoming headers at entry: newrelic=%s, traceparent=%s, tracestate=%s",
			c.Get("newrelic"), c.Get("traceparent"), c.Get("tracestate"))
		return c.Next()
	})
	app.Use(requestid.New())

	routes.SetupRoutes(app)

	log.Printf("nisansala-service listening on %d", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
