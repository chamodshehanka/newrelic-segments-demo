package main

import (
	"fmt"
	"log"
	"nisansala/configs"
	"nisansala/utils"

	"github.com/gofiber/contrib/fibernewrelic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"nisansala/routes"
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
	app.Use(requestid.New())

	// Setup New Relic using environment variables
	newRelicApp := utils.SetupNewRelic(config)
	if newRelicApp != nil {
		app.Use(fibernewrelic.New(fibernewrelic.Config{Application: newRelicApp}))
	}

	routes.SetupRoutes(app)

	log.Printf("nisansala-service listening on %d", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
