package main

import (
	ctrl "bincang-visual/controllers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()

	// Middlewares
	app.Use(logger.New())

	ctrl.WebSocketSignalingController(app)

	app.Listen(":3000")
}
