package main

import (
	"bincang-visual/internal/config"
	"bincang-visual/internal/delivery/http"
	"bincang-visual/internal/domain/usecase"
	"bincang-visual/internal/middleware"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	wsHandler "bincang-visual/internal/delivery/websocket"
	calendarRepo "bincang-visual/internal/repository/calendar"
	redisRepo "bincang-visual/internal/repository/redis"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	googleOAuthConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/calendar",
		},
		Endpoint: google.Endpoint,
	}

	roomRepo := redisRepo.NewRoomRepository(redisClient)
	participantRepo := redisRepo.NewParticipantRepository(redisClient)
	chatRepo := redisRepo.NewChatRepository(redisClient)
	recordingRepo := redisRepo.NewRecordingRepository(redisClient)
	userRepo := redisRepo.NewUserRepository(redisClient)
	calendarRepository := calendarRepo.NewGoogleCalendarRepository(googleOAuthConfig)

	roomUseCase := usecase.NewRoomUseCase(
		roomRepo,
		participantRepo,
		chatRepo,
		recordingRepo,
		*cfg,
	)

	signalingHub := wsHandler.NewSignalingHub(roomUseCase)
	go signalingHub.Run()
	log.Println("Signaling hub started")

	app := fiber.New(fiber.Config{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		BodyLimit:    50 * 1024 * 1024, // 50MB for recordings
		ServerHeader: "Bincang-Visual",
		AppName:      "Bincang Visual v1.0.0",
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} (${latency})\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "https://bakwanx.github.io",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	roomHandler := http.NewRoomHandler(roomUseCase, cfg.Server.BaseURL)
	authHandler := http.NewAuthHandler(
		userRepo,
		googleOAuthConfig,
		cfg.JWT.Secret,
		cfg.JWT.Expiration,
		redisClient,
	)
	calendarHandler := http.NewCalendarHandler(calendarRepository, roomUseCase)
	analyticsHandler := http.NewAnalyticsHandler(roomUseCase)
	uploadHandler := http.NewUploadHandler(cfg.Storage.LocalPath)

	api := app.Group("/api")

	// auth routes (public)
	auth := api.Group("/auth")
	auth.Get("/google", authHandler.GoogleLogin)
	auth.Get("/google/callback", authHandler.GoogleCallback)

	// direct token sign-in (for mobile)
	auth.Post("/google/signin", authHandler.GoogleTokenSignIn)

	// ice servers (public or with optional auth)
	api.Post("/rooms", roomHandler.CreateRoom)
	api.Get("/rooms/:roomId", roomHandler.GetRoom)
	api.Get("/rooms/:roomId/validate", roomHandler.ValidateRoom)
	api.Get("/ice-servers", roomHandler.GetICEServers)

	// protected routes (require JWT)
	protected := api.Group("", middleware.JWTMiddleware(cfg.JWT.Secret))

	// auth - authenticated
	protected.Post("/auth/refresh", authHandler.RefreshToken)
	protected.Get("/auth/me", authHandler.GetCurrentUser)

	// room management
	protected.Get("/rooms/:roomId/participants", roomHandler.GetParticipants)
	protected.Get("/rooms/:roomId/chat", roomHandler.GetChatHistory)
	protected.Delete("/rooms/:roomId", roomHandler.DeleteRoom)

	// recording
	protected.Post("/recordings/start", roomHandler.StartRecording)
	protected.Post("/recordings/stop", roomHandler.StopRecording)
	protected.Post("/recordings/upload-chunk", uploadHandler.UploadRecordingChunk)
	protected.Get("/recordings/:recordingId", roomHandler.GetRecording)

	// calendar integration
	calendarRoutes := protected.Group("/calendar", middleware.OAuthMiddleware(redisClient))
	calendarRoutes.Post("/schedule", calendarHandler.CreateScheduledMeeting)
	calendarRoutes.Get("/upcoming", calendarHandler.GetUpcomingMeetings)
	calendarRoutes.Delete("/:eventId", calendarHandler.CancelMeeting)

	// analytics
	protected.Get("/analytics/room/:roomId", analyticsHandler.GetRoomStatistics)
	protected.Get("/analytics/user", analyticsHandler.GetUserStatistics)

	// storage (serve recording files)
	app.Static("/storage", cfg.Storage.LocalPath)

	// webSocket routes
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/room/:roomId", websocket.New(func(c *websocket.Conn) {
		// get user info from query params or headers
		// use validate JWT token when auth has been implemented
		roomID := c.Params("roomId")
		userID := c.Query("userId")
		displayName := c.Query("displayName")

		if userID == "" {
			userID = "anon-" + uuid.New().String()
		}
		if displayName == "" {
			displayName = "Guest"
		}
		clientID := uuid.New().String()

		log.Printf("WebSocket connection: roomID=%s, userID=%s, clientID=%s", roomID, userID, clientID)

		signalingHub.HandleWebSocket(c, roomID, userID, clientID, displayName)
	}))

	// setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")

		signalingHub.Shutdown()

		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		}

		if err := app.Shutdown(); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}

		log.Println("Server stopped")
		os.Exit(0)
	}()

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Environment: %s", cfg.Server.Environment)

	if err := app.Listen(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
