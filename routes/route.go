package routes

import (
	"bincang-visual/controllers"

	"github.com/gofiber/fiber/v2"
)

type WebSocketDataHandler interface {
	RegisterRoutes(app *fiber.App)
	GetRooms(c *fiber.Ctx) error
	GetConnections(c *fiber.Ctx) error
}

type WebSocketDataHandlerImpl struct{}

func NewWebSocketDataHandler() WebSocketDataHandler {
	return &WebSocketDataHandlerImpl{}
}

func (i *WebSocketDataHandlerImpl) RegisterRoutes(app *fiber.App) {
	app.Get("/rooms", i.GetRooms)
	app.Get("/connections", i.GetConnections)
}

func (i WebSocketDataHandlerImpl) GetRooms(c *fiber.Ctx) error {
	rooms := controllers.Rooms
	if len(rooms) != 0 {
		return c.Status(fiber.StatusOK).JSON(rooms)
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rooms empty"})
}

func (i WebSocketDataHandlerImpl) GetConnections(c *fiber.Ctx) error {
	clients := controllers.Clients
	if len(clients) != 0 {
		return c.Status(fiber.StatusOK).JSON(clients)
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "clients empty"})
}
