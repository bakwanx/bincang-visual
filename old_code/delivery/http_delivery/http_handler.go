package httpdelivery

import (
	ds "bincang-visual/old_code/datasource"
	"bincang-visual/old_code/models"
	"bincang-visual/old_code/usecase"
	"bincang-visual/utils"

	"github.com/gofiber/fiber/v2"
)

type HttpDataHandler interface {
	RegisterRoutes(app *fiber.App)
	CreateRoom(c *fiber.Ctx) error
	GetRoom(c *fiber.Ctx) error
	GetConnections(c *fiber.Ctx) error
	RegisterUser(c *fiber.Ctx) error
	GetUser(c *fiber.Ctx) error
}

type httpDataHandlerImpl struct {
	roomUsecase                usecase.RoomUsecase
	userUsecase                usecase.UserUsecase
	coturnConfigurationUsecase usecase.CoturnConfigurationUsecase
}

func NewHtppDataHandler(roomUsecase usecase.RoomUsecase, userUsecase usecase.UserUsecase, coturnConfigurationUsecase usecase.CoturnConfigurationUsecase) HttpDataHandler {
	return &httpDataHandlerImpl{
		roomUsecase:                roomUsecase,
		userUsecase:                userUsecase,
		coturnConfigurationUsecase: coturnConfigurationUsecase,
	}
}

func (i httpDataHandlerImpl) RegisterRoutes(app *fiber.App) {

	app.Get("/connections", i.GetConnections)
	app.Post("/register-user", i.RegisterUser)
	// app.Get("/users", i.GetUsers)
	app.Get("/user", i.GetUser)

	app.Post("/create-room", i.CreateRoom)
	app.Get("/room", i.GetRoom)
	app.Get("/coturn", i.GetCoturnConfiguration)
}

func (i httpDataHandlerImpl) CreateRoom(c *fiber.Ctx) error {
	result, err := i.roomUsecase.CreateRoom()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "room created",
		"data":    result,
	})
}

func (i httpDataHandlerImpl) GetRoom(c *fiber.Ctx) error {
	roomId := c.Query("roomId")
	result, err := i.roomUsecase.GetRoom(roomId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Room not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "success",
		"data":    result,
	})
}

func (i httpDataHandlerImpl) GetCoturnConfiguration(c *fiber.Ctx) error {
	result, err := i.coturnConfigurationUsecase.GetConfiguration()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Failed to retrieve configurations"})
	}
	encryptedConfiguration, err := utils.EncryptText(*result)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "success",
		"data":    encryptedConfiguration,
	})
}

func (i httpDataHandlerImpl) GetConnections(c *fiber.Ctx) error {
	clients := ds.Clients
	if len(clients) != 0 {
		return c.Status(fiber.StatusOK).JSON(clients)
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "clients empty"})
}

func (i httpDataHandlerImpl) RegisterUser(c *fiber.Ctx) error {
	var userModel = models.User{}
	if err := c.BodyParser(&userModel); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	if userModel.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Username empty"})
	}
	result, err := i.userUsecase.RegisterUser(userModel.Username)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(result)
}

func (i httpDataHandlerImpl) GetUser(c *fiber.Ctx) error {
	userId := c.Query("userId")
	result, err := i.userUsecase.GetUser(userId)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "success",
		"data":    result,
	})
}

// func (i httpDataHandlerImpl) GetRooms(c *fiber.Ctx) error {
// 	rooms := ds.Rooms
// 	if len(rooms) != 0 {
// 		return c.Status(fiber.StatusOK).JSON(rooms)
// 	}

// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "rooms empty"})
// }

// func (i httpDataHandlerImpl) GetUsers(c *fiber.Ctx) error {

// 	users := ds.Users
// 	if len(users) != 0 {
// 		return c.Status(fiber.StatusOK).JSON(users)
// 	}

// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "users empty"})
// }

// func (i httpDataHandlerImpl) CheckAndRemoveData(c *fiber.Ctx) error {
// 	currentTime := time.Now()
// 	strCurrentDate := currentTime.Format("01-02-2006")
// 	for _, room := range ds.Rooms {
// 		for userId := range room {
// 			if ds.Clients[userId] == nil && room[userId].CreatedAt != strCurrentDate {
// 				delete(room, userId)
// 				delete(ds.Users, userId)
// 			}
// 		}
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{"clients": ds.Clients, "rooms": ds.Rooms, "users": ds.Users})
// }
