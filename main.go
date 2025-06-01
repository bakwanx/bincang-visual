package main

import (
	ctrl "bincang-visual/controllers"
	"bincang-visual/routes"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found or failed to load")
	}
	app := fiber.New()

	// Middlewares
	app.Use(logger.New())

	// Routes
	routes := routes.NewWebSocketDataHandler()
	routes.RegisterRoutes(app)

	ctrl.WebSocketSignalingController(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app.Listen(":" + port)
}
