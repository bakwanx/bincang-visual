package server

import (
	httpdelivery "bincang-visual/delivery/http_delivery"
	websocketdelivery "bincang-visual/delivery/websocket_delivery"
	"bincang-visual/repository"

	"bincang-visual/usecase"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

type Server struct {
	redisClient *redis.Client
}

func Run() error {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found or failed to load")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: os.Getenv("REDIS_PASS"),
		DB:       0, // Use default DB
	})

	app := fiber.New()

	// Middlewares
	app.Use(logger.New())

	// Repository initialization
	roomRepo := repository.NewRoomRepository(rdb)
	userRepo := repository.NewUserRepository(rdb)

	// Usecase initialization
	roomUsecase := usecase.NewRoomUsecase(roomRepo)
	userUsecase := usecase.NewUserUsecase(userRepo)

	// Handle
	httpHandle := httpdelivery.NewHtppDataHandler(*roomUsecase, *userUsecase)
	websocketHandle := websocketdelivery.NewWebSocketHandler(*userUsecase, *roomUsecase)
	httpHandle.RegisterRoutes(app)
	websocketHandle.RegisterWebSocket(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app.Listen(":" + port)
	return nil
}
