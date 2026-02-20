package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"bincang-visual/internal/domain/entity"
	"bincang-visual/internal/domain/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockRoomRepository struct {
	mock.Mock
}

func (m *MockRoomRepository) Create(ctx context.Context, room *entity.Room, ttl time.Duration) error {
	args := m.Called(ctx, room, ttl)
	return args.Error(0)
}

func (m *MockRoomRepository) Get(ctx context.Context, roomID string) (*entity.Room, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Room), args.Error(1)
}

func (m *MockRoomRepository) Update(ctx context.Context, room *entity.Room) error {
	args := m.Called(ctx, room)
	return args.Error(0)
}

func (m *MockRoomRepository) Delete(ctx context.Context, roomID string) error {
	args := m.Called(ctx, roomID)
	return args.Error(0)
}

func (m *MockRoomRepository) Exists(ctx context.Context, roomID string) (bool, error) {
	args := m.Called(ctx, roomID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoomRepository) ExtendTTL(ctx context.Context, roomID string, duration time.Duration) error {
	args := m.Called(ctx, roomID, duration)
	return args.Error(0)
}

type MockParticipantRepository struct {
	mock.Mock
}

func (m *MockParticipantRepository) AddParticipant(ctx context.Context, participant *entity.Participant) error {
	args := m.Called(ctx, participant)
	return args.Error(0)
}

func (m *MockParticipantRepository) RemoveParticipant(ctx context.Context, roomID, userId string) error {
	args := m.Called(ctx, roomID, userId)
	return args.Error(0)
}

func (m *MockParticipantRepository) GetParticipants(ctx context.Context, roomID string) ([]*entity.Participant, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Participant), args.Error(1)
}

func (m *MockParticipantRepository) GetParticipant(ctx context.Context, roomID, userId string) (*entity.Participant, error) {
	args := m.Called(ctx, roomID, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Participant), args.Error(1)
}

func (m *MockParticipantRepository) UpdateParticipant(ctx context.Context, participant *entity.Participant) error {
	args := m.Called(ctx, participant)
	return args.Error(0)
}

func (m *MockParticipantRepository) GetParticipantCount(ctx context.Context, roomID string) (int, error) {
	args := m.Called(ctx, roomID)
	return args.Int(0), args.Error(1)
}

type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) SaveMessage(ctx context.Context, message *entity.ChatMessage) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockChatRepository) GetMessages(ctx context.Context, roomID string, limit int) ([]*entity.ChatMessage, error) {
	args := m.Called(ctx, roomID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ChatMessage), args.Error(1)
}

func (m *MockChatRepository) DeleteMessages(ctx context.Context, roomID string) error {
	args := m.Called(ctx, roomID)
	return args.Error(0)
}

type MockRecordingRepository struct {
	mock.Mock
}

func (m *MockRecordingRepository) Create(ctx context.Context, recording *entity.Recording) error {
	args := m.Called(ctx, recording)
	return args.Error(0)
}

func (m *MockRecordingRepository) Get(ctx context.Context, recordingID string) (*entity.Recording, error) {
	args := m.Called(ctx, recordingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Recording), args.Error(1)
}

func (m *MockRecordingRepository) Update(ctx context.Context, recording *entity.Recording) error {
	args := m.Called(ctx, recording)
	return args.Error(0)
}

func (m *MockRecordingRepository) GetByRoomID(ctx context.Context, roomID string) ([]*entity.Recording, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Recording), args.Error(1)
}

func (m *MockRecordingRepository) AddChunk(ctx context.Context, recordingID, chunkURL string) error {
	args := m.Called(ctx, recordingID, chunkURL)
	return args.Error(0)
}

// Tests
func TestCreateRoom(t *testing.T) {
	mockRoomRepo := new(MockRoomRepository)
	mockParticipantRepo := new(MockParticipantRepository)
	mockChatRepo := new(MockChatRepository)
	mockRecordingRepo := new(MockRecordingRepository)

	uc := usecase.NewRoomUseCase(mockRoomRepo, mockParticipantRepo, mockChatRepo, mockRecordingRepo)

	t.Run("successful room creation", func(t *testing.T) {
		mockRoomRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Room"), mock.AnythingOfType("time.Duration")).Return(nil).Once()

		input := usecase.CreateRoomInput{
			Name:            "Test Room",
			HostID:          "user123",
			MaxParticipants: 50,
			Settings: entity.RoomSettings{
				AllowScreenShare: true,
				AllowChat:        true,
			},
		}

		room, err := uc.CreateRoom(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, room)
		assert.Equal(t, "Test Room", room.Name)
		assert.Equal(t, "user123", room.HostID)
		assert.Equal(t, 50, room.MaxParticipants)
		mockRoomRepo.AssertExpectations(t)
	})

	t.Run("room creation with default max participants", func(t *testing.T) {
		mockRoomRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Room"), mock.AnythingOfType("time.Duration")).Return(nil).Once()

		input := usecase.CreateRoomInput{
			Name:   "Test Room",
			HostID: "user123",
		}

		room, err := uc.CreateRoom(context.Background(), input)

		assert.NoError(t, err)
		assert.Equal(t, 100, room.MaxParticipants)
		mockRoomRepo.AssertExpectations(t)
	})
}

func TestJoinRoom(t *testing.T) {
	mockRoomRepo := new(MockRoomRepository)
	mockParticipantRepo := new(MockParticipantRepository)
	mockChatRepo := new(MockChatRepository)
	mockRecordingRepo := new(MockRecordingRepository)

	uc := usecase.NewRoomUseCase(mockRoomRepo, mockParticipantRepo, mockChatRepo, mockRecordingRepo)

	t.Run("successful join", func(t *testing.T) {
		room := &entity.Room{
			ID:              "room123",
			HostID:          "host123",
			MaxParticipants: 100,
		}

		mockRoomRepo.On("Get", mock.Anything, "room123").Return(room, nil).Once()
		mockParticipantRepo.On("GetParticipantCount", mock.Anything, "room123").Return(5, nil).Once()
		mockParticipantRepo.On("AddParticipant", mock.Anything, mock.AnythingOfType("*entity.Participant")).Return(nil).Once()

		input := usecase.JoinRoomInput{
			RoomID:      "room123",
			UserID:      "user456",
			DisplayName: "Test User",
		}

		participant, err := uc.JoinRoom(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, participant)
		assert.Equal(t, "user456", participant.UserID)
		assert.Equal(t, "room123", participant.RoomID)
		assert.False(t, participant.IsHost)
		mockRoomRepo.AssertExpectations(t)
		mockParticipantRepo.AssertExpectations(t)
	})

	t.Run("join as host", func(t *testing.T) {
		room := &entity.Room{
			ID:              "room123",
			HostID:          "host123",
			MaxParticipants: 100,
		}

		mockRoomRepo.On("Get", mock.Anything, "room123").Return(room, nil).Once()
		mockParticipantRepo.On("GetParticipantCount", mock.Anything, "room123").Return(0, nil).Once()
		mockParticipantRepo.On("AddParticipant", mock.Anything, mock.AnythingOfType("*entity.Participant")).Return(nil).Once()

		input := usecase.JoinRoomInput{
			RoomID:      "room123",
			UserID:      "host123",
			DisplayName: "Host",
		}

		participant, err := uc.JoinRoom(context.Background(), input)

		assert.NoError(t, err)
		assert.True(t, participant.IsHost)
	})

	t.Run("room full", func(t *testing.T) {
		room := &entity.Room{
			ID:              "room123",
			MaxParticipants: 10,
		}

		mockRoomRepo.On("Get", mock.Anything, "room123").Return(room, nil).Once()
		mockParticipantRepo.On("GetParticipantCount", mock.Anything, "room123").Return(10, nil).Once()

		input := usecase.JoinRoomInput{
			RoomID:      "room123",
			UserID:      "user456",
			DisplayName: "Test User",
		}

		participant, err := uc.JoinRoom(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, participant)
		assert.Contains(t, err.Error(), "room is full")
	})

	t.Run("room not found", func(t *testing.T) {
		mockRoomRepo.On("Get", mock.Anything, "nonexistent").Return(nil, errors.New("room not found")).Once()

		input := usecase.JoinRoomInput{
			RoomID:      "nonexistent",
			UserID:      "user456",
			DisplayName: "Test User",
		}

		participant, err := uc.JoinRoom(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, participant)
	})
}

func TestStartRecording(t *testing.T) {
	mockRoomRepo := new(MockRoomRepository)
	mockParticipantRepo := new(MockParticipantRepository)
	mockChatRepo := new(MockChatRepository)
	mockRecordingRepo := new(MockRecordingRepository)

	uc := usecase.NewRoomUseCase(mockRoomRepo, mockParticipantRepo, mockChatRepo, mockRecordingRepo)

	t.Run("successful recording start", func(t *testing.T) {
		room := &entity.Room{
			ID:          "room123",
			HostID:      "host123",
			IsRecording: false,
		}

		mockRoomRepo.On("Get", mock.Anything, "room123").Return(room, nil).Once()
		mockRecordingRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Recording")).Return(nil).Once()
		mockRoomRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Room")).Return(nil).Once()

		recording, err := uc.StartRecording(context.Background(), "room123", "host123")

		assert.NoError(t, err)
		assert.NotNil(t, recording)
		assert.Equal(t, "room123", recording.RoomID)
		assert.Equal(t, "recording", recording.Status)
		mockRoomRepo.AssertExpectations(t)
		mockRecordingRepo.AssertExpectations(t)
	})

	t.Run("non-host cannot start recording", func(t *testing.T) {
		room := &entity.Room{
			ID:          "room123",
			HostID:      "host123",
			IsRecording: false,
		}

		mockRoomRepo.On("Get", mock.Anything, "room123").Return(room, nil).Once()

		recording, err := uc.StartRecording(context.Background(), "room123", "user456")

		assert.Error(t, err)
		assert.Nil(t, recording)
		assert.Contains(t, err.Error(), "only host can start recording")
	})

	t.Run("recording already in progress", func(t *testing.T) {
		room := &entity.Room{
			ID:          "room123",
			HostID:      "host123",
			IsRecording: true,
		}

		mockRoomRepo.On("Get", mock.Anything, "room123").Return(room, nil).Once()

		recording, err := uc.StartRecording(context.Background(), "room123", "host123")

		assert.Error(t, err)
		assert.Nil(t, recording)
		assert.Contains(t, err.Error(), "recording already in progress")
	})
}

func TestLeaveRoom(t *testing.T) {
	mockRoomRepo := new(MockRoomRepository)
	mockParticipantRepo := new(MockParticipantRepository)
	mockChatRepo := new(MockChatRepository)
	mockRecordingRepo := new(MockRecordingRepository)

	uc := usecase.NewRoomUseCase(mockRoomRepo, mockParticipantRepo, mockChatRepo, mockRecordingRepo)

	t.Run("leave room with remaining participants", func(t *testing.T) {
		mockParticipantRepo.On("RemoveParticipant", mock.Anything, "room123", "participant123").Return(nil).Once()
		mockParticipantRepo.On("GetParticipantCount", mock.Anything, "room123").Return(2, nil).Once()

		err := uc.LeaveRoom(context.Background(), "room123", "participant123")

		assert.NoError(t, err)
		mockParticipantRepo.AssertExpectations(t)
	})

	t.Run("last participant leaves - room deleted", func(t *testing.T) {
		mockParticipantRepo.On("RemoveParticipant", mock.Anything, "room123", "participant123").Return(nil).Once()
		mockParticipantRepo.On("GetParticipantCount", mock.Anything, "room123").Return(0, nil).Once()
		mockRoomRepo.On("Delete", mock.Anything, "room123").Return(nil).Once()
		mockChatRepo.On("DeleteMessages", mock.Anything, "room123").Return(nil).Maybe()

		err := uc.LeaveRoom(context.Background(), "room123", "participant123")

		assert.NoError(t, err)
		mockParticipantRepo.AssertExpectations(t)
		mockRoomRepo.AssertExpectations(t)
	})
}
