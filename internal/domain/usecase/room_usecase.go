package usecase

import (
	"bincang-visual/internal/config"
	"bincang-visual/internal/domain/entity"
	"bincang-visual/internal/domain/repository"
	"bincang-visual/utils"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type RoomUseCase struct {
	roomRepo        repository.RoomRepository
	participantRepo repository.ParticipantRepository
	chatRepo        repository.ChatRepository
	recordingRepo   repository.RecordingRepository
	defaultTTL      time.Duration
	config          config.Config
}

func NewRoomUseCase(
	roomRepo repository.RoomRepository,
	participantRepo repository.ParticipantRepository,
	chatRepo repository.ChatRepository,
	recordingRepo repository.RecordingRepository,
	config config.Config,
) *RoomUseCase {
	return &RoomUseCase{
		roomRepo:        roomRepo,
		participantRepo: participantRepo,
		chatRepo:        chatRepo,
		recordingRepo:   recordingRepo,
		defaultTTL:      24 * time.Hour, // 24 hours default
		config:          config,
	}
}

type CreateRoomInput struct {
	Name            string
	HostID          string
	MaxParticipants int
	Settings        entity.RoomSettings
}

func (uc *RoomUseCase) CreateRoom(ctx context.Context, input CreateRoomInput) (*entity.Room, error) {
	room := &entity.Room{
		ID:              uuid.New().String(),
		Name:            input.Name,
		HostID:          input.HostID,
		CreatedAt:       time.Now(),
		MaxParticipants: input.MaxParticipants,
		IsRecording:     false,
		Settings:        input.Settings,
	}

	if room.MaxParticipants == 0 {
		room.MaxParticipants = 100
	}

	ttl := 24 * time.Hour
	if err := uc.roomRepo.Create(ctx, room, ttl); err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	return room, nil
}

func (uc *RoomUseCase) GetRoom(ctx context.Context, roomID string) (*entity.Room, error) {
	room, err := uc.roomRepo.Get(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	return room, nil
}

type JoinRoomInput struct {
	RoomID      string
	UserID      string
	DisplayName string
}

func (uc *RoomUseCase) JoinRoom(ctx context.Context, input JoinRoomInput) (*entity.Participant, error) {

	room, err := uc.roomRepo.Get(ctx, input.RoomID)
	if err != nil {
		return nil, fmt.Errorf("room not found: %w", err)
	}

	count, err := uc.participantRepo.GetParticipantCount(ctx, input.RoomID)
	if err != nil {
		return nil, err
	}

	if count >= room.MaxParticipants {
		return nil, fmt.Errorf("room is full")
	}

	participant := &entity.Participant{
		UserID:      input.UserID,
		RoomID:      input.RoomID,
		DisplayName: input.DisplayName,
		JoinedAt:    time.Now(),
		IsHost:      input.UserID == room.HostID,
		IsMuted:     false,
		IsVideoOff:  false,
	}

	if err := uc.participantRepo.AddParticipant(ctx, participant); err != nil {
		return nil, fmt.Errorf("failed to add participant: %w", err)
	}

	return participant, nil
}

func (uc *RoomUseCase) LeaveRoom(ctx context.Context, roomID, userId string) error {
	if err := uc.participantRepo.RemoveParticipant(ctx, roomID, userId); err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	// count, err := uc.participantRepo.GetParticipantCount(ctx, roomID)
	// if err != nil {
	// 	return err
	// }

	// if count == 0 {
	// 	if err := uc.roomRepo.Delete(ctx, roomID); err != nil {
	// 		return err
	// 	}

	// 	_ = uc.chatRepo.DeleteMessages(ctx, roomID)
	// }

	return nil
}

func (uc *RoomUseCase) GetParticipants(ctx context.Context, roomID string) ([]*entity.Participant, error) {
	participants, err := uc.participantRepo.GetParticipants(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}

	return participants, nil
}

func (uc *RoomUseCase) UpdateParticipantState(ctx context.Context, participant *entity.Participant) error {
	if err := uc.participantRepo.UpdateParticipant(ctx, participant); err != nil {
		return fmt.Errorf("failed to update participant: %w", err)
	}

	return nil
}

func (uc *RoomUseCase) SendChatMessage(ctx context.Context, message *entity.ChatMessage) error {
	message.ID = uuid.New().String()
	message.Timestamp = time.Now()

	if err := uc.chatRepo.SaveMessage(ctx, message); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}

func (uc *RoomUseCase) GetChatHistory(ctx context.Context, roomID string, limit int) ([]*entity.ChatMessage, error) {
	messages, err := uc.chatRepo.GetMessages(ctx, roomID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

func (uc *RoomUseCase) StartRecording(ctx context.Context, roomID, hostID string) (*entity.Recording, error) {

	room, err := uc.roomRepo.Get(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("room not found: %w", err)
	}

	if room.HostID != hostID {
		return nil, fmt.Errorf("only host can start recording")
	}

	if room.IsRecording {
		return nil, fmt.Errorf("recording already in progress")
	}

	recording := &entity.Recording{
		ID:        uuid.New().String(),
		RoomID:    roomID,
		StartTime: time.Now(),
		Status:    "recording",
		Chunks:    []string{},
	}

	if err := uc.recordingRepo.Create(ctx, recording); err != nil {
		return nil, fmt.Errorf("failed to create recording: %w", err)
	}

	room.IsRecording = true
	if err := uc.roomRepo.Update(ctx, room); err != nil {
		return nil, err
	}

	return recording, nil
}

func (uc *RoomUseCase) StopRecording(ctx context.Context, roomID, recordingID, hostID string) error {

	room, err := uc.roomRepo.Get(ctx, roomID)
	if err != nil {
		return fmt.Errorf("room not found: %w", err)
	}

	if room.HostID != hostID {
		return fmt.Errorf("only host can stop recording")
	}

	recording, err := uc.recordingRepo.Get(ctx, recordingID)
	if err != nil {
		return fmt.Errorf("recording not found: %w", err)
	}

	recording.EndTime = time.Now()
	recording.Duration = int(recording.EndTime.Sub(recording.StartTime).Seconds())
	recording.Status = "processing"

	if err := uc.recordingRepo.Update(ctx, recording); err != nil {
		return err
	}

	room.IsRecording = false
	if err := uc.roomRepo.Update(ctx, room); err != nil {
		return err
	}

	return nil
}

func (uc *RoomUseCase) AddRecordingChunk(ctx context.Context, recordingID, chunkURL string) error {
	if err := uc.recordingRepo.AddChunk(ctx, recordingID, chunkURL); err != nil {
		return fmt.Errorf("failed to add chunk: %w", err)
	}

	return nil
}

func (uc *RoomUseCase) GetRecording(ctx context.Context, recordingID string) (*entity.Recording, error) {
	recording, err := uc.recordingRepo.Get(ctx, recordingID)
	if err != nil {
		return nil, fmt.Errorf("recording not found: %w", err)
	}

	return recording, nil
}

func (uc *RoomUseCase) ExtendRoomDuration(ctx context.Context, roomID string, duration time.Duration) error {
	if err := uc.roomRepo.ExtendTTL(ctx, roomID, duration); err != nil {
		return fmt.Errorf("failed to extend room duration: %w", err)
	}

	return nil
}

func (uc *RoomUseCase) GetConfiguration() config.TurnStunConfig {
	config := uc.config.TurnStun
	if uc.config.IsProduction() {
		config.Username, _ = utils.EncryptText(config.Username)
		config.Credential, _ = utils.EncryptText(config.Credential)
	}

	return config
}

func (uc *RoomUseCase) StartScreenShare(ctx context.Context, roomID, userID string) error {

	currentSharer, err := uc.roomRepo.GetScreenSharer(ctx, roomID)
	if err != nil {
		return fmt.Errorf("failed to check screen sharer: %w", err)
	}

	if currentSharer != "" && currentSharer != userID {
		return fmt.Errorf("screen share already in progress by another user")
	}

	if err := uc.roomRepo.SetScreenSharer(ctx, roomID, userID); err != nil {
		return fmt.Errorf("failed to set screen sharer: %w", err)
	}

	log.Printf("[UseCase] User %s started screen share in room %s", userID, roomID)
	return nil
}

func (uc *RoomUseCase) StopScreenShare(ctx context.Context, roomID, userID string) error {

	currentSharer, err := uc.roomRepo.GetScreenSharer(ctx, roomID)
	if err != nil {
		return fmt.Errorf("failed to check screen sharer: %w", err)
	}

	if currentSharer != userID {
		return fmt.Errorf("you are not the current screen sharer")
	}

	if err := uc.roomRepo.ClearScreenSharer(ctx, roomID); err != nil {
		return fmt.Errorf("failed to clear screen sharer: %w", err)
	}

	log.Printf("[UseCase] User %s stopped screen share in room %s", userID, roomID)
	return nil
}

func (uc *RoomUseCase) GetScreenSharer(ctx context.Context, roomID string) (string, error) {
	return uc.roomRepo.GetScreenSharer(ctx, roomID)
}
