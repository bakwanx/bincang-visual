package http

import (
	"bincang-visual/internal/domain/usecase"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AnalyticsHandler struct {
	roomUseCase *usecase.RoomUseCase
}

func NewAnalyticsHandler(roomUseCase *usecase.RoomUseCase) *AnalyticsHandler {
	return &AnalyticsHandler{
		roomUseCase: roomUseCase,
	}
}

func (h *AnalyticsHandler) GetRoomStatistics(c *fiber.Ctx) error {
	roomID := c.Params("roomId")

	room, err := h.roomUseCase.GetRoom(c.Context(), roomID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Room not found",
		})
	}

	participants, err := h.roomUseCase.GetParticipants(c.Context(), roomID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get participants",
		})
	}

	duration := time.Since(room.CreatedAt)

	stats := fiber.Map{
		"roomId":           roomID,
		"roomName":         room.Name,
		"participantCount": len(participants),
		"maxParticipants":  room.MaxParticipants,
		"duration":         duration.Minutes(),
		"isRecording":      room.IsRecording,
		"createdAt":        room.CreatedAt,
	}

	return c.JSON(stats)
}

func (h *AnalyticsHandler) GetUserStatistics(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	// Its used for user statistics in production. For now return dummy data
	stats := fiber.Map{
		"userId":          userID,
		"totalMeetings":   0,
		"totalDuration":   0,
		"averageDuration": 0,
	}

	return c.JSON(stats)
}
