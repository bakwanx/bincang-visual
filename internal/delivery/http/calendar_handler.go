package http

import (
	"bincang-visual/internal/domain/entity"
	"bincang-visual/internal/domain/repository"
	"bincang-visual/internal/domain/usecase"

	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CalendarHandler struct {
	calendarRepo repository.CalendarRepository
	roomUseCase  *usecase.RoomUseCase
}

func NewCalendarHandler(
	calendarRepo repository.CalendarRepository,
	roomUseCase *usecase.RoomUseCase,
) *CalendarHandler {
	return &CalendarHandler{
		calendarRepo: calendarRepo,
		roomUseCase:  roomUseCase,
	}
}

type CreateScheduledMeetingRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	Attendees   []string  `json:"attendees"`
}

func (h *CalendarHandler) CreateScheduledMeeting(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req CreateScheduledMeetingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate times
	if req.StartTime.Before(time.Now()) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Start time must be in the future",
		})
	}

	if req.EndTime.Before(req.StartTime) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "End time must be after start time",
		})
	}

	room, err := h.roomUseCase.CreateRoom(c.Context(), usecase.CreateRoomInput{
		Name:            req.Title,
		HostID:          userID,
		MaxParticipants: 100,
		Settings: entity.RoomSettings{
			AllowScreenShare: true,
			AllowChat:        true,
			RecordingEnabled: true,
		},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create room",
		})
	}

	calendarEvent := &entity.CalendarEvent{
		ID:          uuid.New().String(),
		RoomID:      room.ID,
		Title:       req.Title,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Attendees:   req.Attendees,
		CreatorID:   userID,
	}

	err = h.calendarRepo.CreateEvent(c.Context(), calendarEvent)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create calendar event",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"event":   calendarEvent,
		"room":    room,
		"joinUrl": fmt.Sprintf("https://my-domain.com/join/%s", room.ID),
	})
}

func (h *CalendarHandler) GetUpcomingMeetings(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	now := time.Now()
	oneMonthLater := now.AddDate(0, 1, 0)

	events, err := h.calendarRepo.GetUserEvents(
		c.Context(),
		userID,
		now,
		oneMonthLater,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get events " + err.Error(),
		})
	}

	return c.JSON(events)
}

func (h *CalendarHandler) CancelMeeting(c *fiber.Ctx) error {
	eventID := c.Params("eventId")

	err := h.calendarRepo.DeleteEvent(c.Context(), eventID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete event",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Meeting cancelled successfully",
	})
}
