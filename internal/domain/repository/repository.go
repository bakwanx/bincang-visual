package repository

import (
	"bincang-visual/internal/domain/entity"
	"context"
	"time"
)

type RoomRepository interface {
	Create(ctx context.Context, room *entity.Room, ttl time.Duration) error
	Get(ctx context.Context, roomID string) (*entity.Room, error)
	Update(ctx context.Context, room *entity.Room) error
	Delete(ctx context.Context, roomID string) error
	Exists(ctx context.Context, roomID string) (bool, error)
	ExtendTTL(ctx context.Context, roomID string, duration time.Duration) error
	SetScreenSharer(ctx context.Context, roomID, userID string) error
	GetScreenSharer(ctx context.Context, roomID string) (string, error)
	ClearScreenSharer(ctx context.Context, roomID string) error
}

type ParticipantRepository interface {
	AddParticipant(ctx context.Context, participant *entity.Participant) error
	RemoveParticipant(ctx context.Context, roomID, userId string) error
	GetParticipants(ctx context.Context, roomID string) ([]*entity.Participant, error)
	GetParticipant(ctx context.Context, roomID, userId string) (*entity.Participant, error)
	UpdateParticipant(ctx context.Context, participant *entity.Participant) error
	GetParticipantCount(ctx context.Context, roomID string) (int, error)
}

type ChatRepository interface {
	SaveMessage(ctx context.Context, message *entity.ChatMessage) error
	GetMessages(ctx context.Context, roomID string, limit int) ([]*entity.ChatMessage, error)
	DeleteMessages(ctx context.Context, roomID string) error
}

type RecordingRepository interface {
	Create(ctx context.Context, recording *entity.Recording) error
	Get(ctx context.Context, recordingID string) (*entity.Recording, error)
	Update(ctx context.Context, recording *entity.Recording) error
	GetByRoomID(ctx context.Context, roomID string) ([]*entity.Recording, error)
	AddChunk(ctx context.Context, recordingID, chunkURL string) error
}

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	Get(ctx context.Context, userID string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
}
type CalendarRepository interface {
	CreateEvent(ctx context.Context, event *entity.CalendarEvent) error
	GetEvent(ctx context.Context, eventID string) (*entity.CalendarEvent, error)
	UpdateEvent(ctx context.Context, event *entity.CalendarEvent) error
	DeleteEvent(ctx context.Context, eventID string) error
	GetUserEvents(ctx context.Context, userID string, from, to time.Time) ([]*entity.CalendarEvent, error)
}
