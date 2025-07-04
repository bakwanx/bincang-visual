package server

import (
	httpdelivery "bincang-visual/delivery/http_delivery"
	websocketdelivery "bincang-visual/delivery/websocket_delivery"
	"bincang-visual/repository"
	"context"

	"bincang-visual/usecase"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       0,
	})

	ctx := context.Background()
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Redis connection failed: %v", err))
	}

	app := fiber.New()

	// Middlewares
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowMethods:     "GET, POST, PUT, DELETE",
		AllowCredentials: true,
	}))

	// Repository initialization
	roomRepo := repository.NewRoomRepository(rdb)
	userRepo := repository.NewUserRepository(rdb)
	coturnRepo := repository.NewCoturnConfigurationRepository(rdb)

	// Usecase initialization
	roomUsecase := usecase.NewRoomUsecase(roomRepo)
	userUsecase := usecase.NewUserUsecase(userRepo)
	coturnUsecase := usecase.NewCoturnConfigurationUsecase(coturnRepo)
	websocketUsecase := usecase.NewWebsocketUsecase(userRepo, roomRepo)

	// Handle
	httpHandle := httpdelivery.NewHtppDataHandler(*roomUsecase, *userUsecase, *coturnUsecase)
	websocketHandle := websocketdelivery.NewWebSocketHandler(*userUsecase, *roomUsecase, *websocketUsecase)
	httpHandle.RegisterRoutes(app)
	websocketHandle.RegisterWebSocket(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app.Listen(":" + port)
	return nil
}
