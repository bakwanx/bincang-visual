package routes

import (
	ds "bincang-visual/datasource"
	"bincang-visual/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type WebSocketDataHandler interface {
	RegisterRoutes(app *fiber.App)
	GetRooms(c *fiber.Ctx) error
	GetConnections(c *fiber.Ctx) error
	RegisterUser(c *fiber.Ctx) error
	GetUsers(c *fiber.Ctx) error
	CheckAndRemoveData(c *fiber.Ctx) error
}

type WebSocketDataHandlerImpl struct{}

func NewWebSocketDataHandler() WebSocketDataHandler {
	return &WebSocketDataHandlerImpl{}
}

func (i *WebSocketDataHandlerImpl) RegisterRoutes(app *fiber.App) {
	app.Get("/rooms", i.GetRooms)
	app.Get("/connections", i.GetConnections)
	app.Post("/register-user", i.RegisterUser)
	app.Get("/user", i.GetUsers)
	app.Get("/remove", i.CheckAndRemoveData)
}

func (i WebSocketDataHandlerImpl) GetRooms(c *fiber.Ctx) error {
	rooms := ds.Rooms
	if len(rooms) != 0 {
		return c.Status(fiber.StatusOK).JSON(rooms)
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rooms empty"})
}

func (i WebSocketDataHandlerImpl) GetConnections(c *fiber.Ctx) error {
	clients := ds.Clients
	if len(clients) != 0 {
		return c.Status(fiber.StatusOK).JSON(clients)
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "clients empty"})
}

func (i WebSocketDataHandlerImpl) RegisterUser(c *fiber.Ctx) error {
	userModel := new(models.User)
	if err := c.BodyParser(userModel); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if userModel.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Username empty"})
	}
	userId := uuid.New().String()
	currentTime := time.Now()
	userModel.ID = userId
	userModel.CreatedAt = currentTime.Format("01-02-2006")
	ds.Users[userId] = *userModel
	return c.Status(fiber.StatusOK).JSON(userModel)
}

func (i WebSocketDataHandlerImpl) GetUsers(c *fiber.Ctx) error {
	users := ds.Users
	if len(users) != 0 {
		return c.Status(fiber.StatusOK).JSON(users)
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "users empty"})
}

func (i WebSocketDataHandlerImpl) CheckAndRemoveData(c *fiber.Ctx) error {
	currentTime := time.Now()
	strCurrentDate := currentTime.Format("01-02-2006")
	for _, room := range ds.Rooms {
		for userId := range room {
			if ds.Clients[userId] == nil && room[userId].CreatedAt != strCurrentDate {
				delete(room, userId)
				delete(ds.Users, userId)
			}
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"clients": ds.Clients, "rooms": ds.Rooms, "users": ds.Users})
}
