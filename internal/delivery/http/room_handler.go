package http

import (
	"bincang-visual/internal/config"
	"bincang-visual/internal/domain/entity"
	"bincang-visual/internal/domain/usecase"
	"fmt"
	"log"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

type RoomHandler struct {
	roomUseCase *usecase.RoomUseCase
	baseURL     string
}

func NewRoomHandler(roomUseCase *usecase.RoomUseCase, baseURL string) *RoomHandler {
	return &RoomHandler{
		roomUseCase: roomUseCase,
		baseURL:     baseURL,
	}
}

type CreateRoomRequest struct {
	Name            string              `json:"name"`
	MaxParticipants int                 `json:"maxParticipants"`
	Settings        entity.RoomSettings `json:"settings"`
}

type CreateRoomResponse struct {
	RoomID    string `json:"roomId"`
	RoomURL   string `json:"roomUrl"`
	JoinURL   string `json:"joinUrl"`
	HostID    string `json:"hostId"`
	CreatedAt string `json:"createdAt"`
}

func (h *RoomHandler) CreateRoom(c *fiber.Ctx) error {
	// Get user from context (set by auth middleware)
	userID := "anonymous"
	if uid := c.Locals("userID"); uid != nil {
		userID = uid.(string)
	}

	var req CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	room, err := h.roomUseCase.CreateRoom(c.Context(), usecase.CreateRoomInput{
		Name:            req.Name,
		HostID:          userID,
		MaxParticipants: req.MaxParticipants,
		Settings:        req.Settings,
	})
	if err != nil {
		log.Printf("[Handler] Failed to create room: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create room",
		})
	}

	roomURL := fmt.Sprintf("%s/room/%s", h.baseURL, room.ID)
	joinURL := fmt.Sprintf("%s/join/%s", h.baseURL, room.ID)

	return c.Status(fiber.StatusCreated).JSON(CreateRoomResponse{
		RoomID:    room.ID,
		RoomURL:   roomURL,
		JoinURL:   joinURL,
		HostID:    room.HostID,
		CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GET /api/rooms/:roomId
func (h *RoomHandler) GetRoom(c *fiber.Ctx) error {
	roomID := c.Params("roomId")

	room, err := h.roomUseCase.GetRoom(c.Context(), roomID)
	if err != nil {
		log.Printf("[Handler] Room not found: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Room not found",
		})
	}

	return c.JSON(room)
}

// GET /api/rooms/:roomId/participants
func (h *RoomHandler) GetParticipants(c *fiber.Ctx) error {
	roomID := c.Params("roomId")

	participants, err := h.roomUseCase.GetParticipants(c.Context(), roomID)
	if err != nil {
		log.Printf("[Handler] Failed to get participants: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get participants",
		})
	}

	return c.JSON(participants)
}

// GET /api/rooms/:roomId/chat
func (h *RoomHandler) GetChatHistory(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	limit := c.QueryInt("limit", 100)

	messages, err := h.roomUseCase.GetChatHistory(c.Context(), roomID, limit)
	if err != nil {
		log.Printf("[Handler] Failed to get chat history: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get chat history",
		})
	}

	return c.JSON(messages)
}

type StartRecordingRequest struct {
	RoomID string `json:"roomId"`
}

// POST /api/recordings/start
func (h *RoomHandler) StartRecording(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req StartRecordingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	recording, err := h.roomUseCase.StartRecording(c.Context(), req.RoomID, userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(recording)
}

type StopRecordingRequest struct {
	RoomID      string `json:"roomId"`
	RecordingID string `json:"recordingId"`
}

// POST /api/recordings/stop
func (h *RoomHandler) StopRecording(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req StopRecordingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.roomUseCase.StopRecording(c.Context(), req.RoomID, req.RecordingID, userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Recording stopped successfully",
	})
}

type UploadChunkRequest struct {
	RecordingID string `json:"recordingId"`
	ChunkURL    string `json:"chunkUrl"`
}

// POST /api/recordings/upload-chunk
func (h *RoomHandler) UploadChunk(c *fiber.Ctx) error {
	var req UploadChunkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.roomUseCase.AddRecordingChunk(c.Context(), req.RecordingID, req.ChunkURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add chunk",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Chunk uploaded successfully",
	})
}

// GET /api/recordings/:recordingId
func (h *RoomHandler) GetRecording(c *fiber.Ctx) error {
	recordingID := c.Params("recordingId")

	recording, err := h.roomUseCase.GetRecording(c.Context(), recordingID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Recording not found",
		})
	}

	return c.JSON(recording)
}

// GET /api/ice-servers
func (h *RoomHandler) GetICEServers(c *fiber.Ctx) error {
	// TODO: generate short-lived TURN credentials
	configuration := h.roomUseCase.GetConfiguration()
	iceServers := []config.TurnStunConfig{
		{
			URLs: []string{
				"stun:stun.l.google.com:19302",
				"stun:stun1.l.google.com:19302",
				"stun:202.10.42.100:3478",
				"stun:stun.flashdance.cx:3478",
			},
		},
		configuration,
	}

	config := entity.RoomConfig{
		ICEServers:       iceServers,
		MaxBitrate:       2500000, // 2.5 Mbps
		CodecPreferences: []string{"VP9", "VP8", "H264"},
	}

	return c.JSON(config)
}

func (h *RoomHandler) DeleteRoom(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	roomID := c.Params("roomId")

	room, err := h.roomUseCase.GetRoom(c.Context(), roomID)
	if err != nil {
		log.Printf("[Handler] Room not found: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Room not found",
		})
	}

	if room.HostID != userID {
		log.Printf("[Handler] Only host can delete the room: %v", err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only host can delete the room",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Room will be deleted when all participants leave",
	})
}

func (h *RoomHandler) GenerateRoomLink(c *fiber.Ctx) error {
	roomID := c.Params("roomId")

	room, err := h.roomUseCase.GetRoom(c.Context(), roomID)
	if err != nil {
		log.Printf("[Handler] Room not found: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Room not found",
		})
	}

	link := fmt.Sprintf("%s/join/%s", h.baseURL, roomID)

	return c.JSON(fiber.Map{
		"roomId":   roomID,
		"roomName": room.Name,
		"link":     link,
		"qrCode":   generateQRCodeURL(link),
	})
}

func (h *RoomHandler) ValidateRoom(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	if roomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID is required",
		})
	}

	room, err := h.roomUseCase.GetRoom(c.Context(), roomID)
	if err != nil {
		log.Printf("[Handler] Room not found: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Room not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": room,
	})
}

func generateQRCodeURL(link string) string {
	return fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=%s",
		url.QueryEscape(link))
}
