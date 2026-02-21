package redis

import (
	"bincang-visual/internal/domain/entity"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	roomPrefix        = "room:"
	participantPrefix = "room:%s:participants"
	chatPrefix        = "room:%s:chat"
	recordingPrefix   = "recording:"
	userPrefix        = "user:"
)

// ============= ROOM REPOSITORY =============

type RoomRepositoryImpl struct {
	client *redis.Client
}

func NewRoomRepository(client *redis.Client) *RoomRepositoryImpl {
	return &RoomRepositoryImpl{client: client}
}

func (r *RoomRepositoryImpl) Create(ctx context.Context, room *entity.Room, ttl time.Duration) error {
	key := roomPrefix + room.ID
	data, err := json.Marshal(room)
	if err != nil {
		return fmt.Errorf("failed to marshal room: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RoomRepositoryImpl) Get(ctx context.Context, roomID string) (*entity.Room, error) {
	key := roomPrefix + roomID
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("room not found")
		}
		return nil, err
	}

	var room entity.Room
	if err := json.Unmarshal(data, &room); err != nil {
		return nil, fmt.Errorf("failed to unmarshal room: %w", err)
	}
	return &room, nil
}

func (r *RoomRepositoryImpl) Update(ctx context.Context, room *entity.Room) error {
	key := roomPrefix + room.ID
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return err
	}
	data, err := json.Marshal(room)
	if err != nil {
		return fmt.Errorf("failed to marshal room: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RoomRepositoryImpl) Delete(ctx context.Context, roomID string) error {
	key := roomPrefix + roomID
	return r.client.Del(ctx, key).Err()
}

func (r *RoomRepositoryImpl) Exists(ctx context.Context, roomID string) (bool, error) {
	key := roomPrefix + roomID
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

func (r *RoomRepositoryImpl) ExtendTTL(ctx context.Context, roomID string, duration time.Duration) error {
	key := roomPrefix + roomID
	return r.client.Expire(ctx, key, duration).Err()
}

func (r *RoomRepositoryImpl) SetScreenSharer(ctx context.Context, roomID, userID string) error {
	key := fmt.Sprintf("room:%s:screen_sharer", roomID)

	return r.client.Set(ctx, key, userID, 24*time.Hour).Err()
}

func (r *RoomRepositoryImpl) GetScreenSharer(ctx context.Context, roomID string) (string, error) {
	key := fmt.Sprintf("room:%s:screen_sharer", roomID)
	userID, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // No one sharing
	}
	return userID, err
}

func (r *RoomRepositoryImpl) ClearScreenSharer(ctx context.Context, roomID string) error {
	key := fmt.Sprintf("room:%s:screen_sharer", roomID)
	return r.client.Del(ctx, key).Err()
}

// ============= PARTICIPANT REPOSITORY =============

type ParticipantRepositoryImpl struct {
	client *redis.Client
}

func NewParticipantRepository(client *redis.Client) *ParticipantRepositoryImpl {
	return &ParticipantRepositoryImpl{client: client}
}

func (r *ParticipantRepositoryImpl) AddParticipant(ctx context.Context, participant *entity.Participant) error {
	key := fmt.Sprintf(participantPrefix, participant.RoomID)
	data, err := json.Marshal(participant)
	if err != nil {
		return err
	}
	return r.client.HSet(ctx, key, participant.UserID, data).Err()
}

func (r *ParticipantRepositoryImpl) RemoveParticipant(ctx context.Context, roomID, userId string) error {
	key := fmt.Sprintf(participantPrefix, roomID)
	return r.client.HDel(ctx, key, userId).Err()
}

func (r *ParticipantRepositoryImpl) GetParticipants(ctx context.Context, roomID string) ([]*entity.Participant, error) {
	key := fmt.Sprintf(participantPrefix, roomID)
	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	participants := make([]*entity.Participant, 0, len(data))
	for _, v := range data {
		var p entity.Participant
		if err := json.Unmarshal([]byte(v), &p); err != nil {
			continue
		}
		participants = append(participants, &p)
	}
	return participants, nil
}

func (r *ParticipantRepositoryImpl) GetParticipant(ctx context.Context, roomID, userId string) (*entity.Participant, error) {
	key := fmt.Sprintf(participantPrefix, roomID)
	data, err := r.client.HGet(ctx, key, userId).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("participant not found")
		}
		return nil, err
	}

	var participant entity.Participant
	if err := json.Unmarshal(data, &participant); err != nil {
		return nil, err
	}
	return &participant, nil
}

func (r *ParticipantRepositoryImpl) UpdateParticipant(ctx context.Context, participant *entity.Participant) error {
	return r.AddParticipant(ctx, participant)
}

func (r *ParticipantRepositoryImpl) GetParticipantCount(ctx context.Context, roomID string) (int, error) {
	key := fmt.Sprintf(participantPrefix, roomID)
	count, err := r.client.HLen(ctx, key).Result()
	return int(count), err
}

// ============= CHAT REPOSITORY =============

type ChatRepositoryImpl struct {
	client *redis.Client
}

func NewChatRepository(client *redis.Client) *ChatRepositoryImpl {
	return &ChatRepositoryImpl{client: client}
}

func (r *ChatRepositoryImpl) SaveMessage(ctx context.Context, message *entity.ChatMessage) error {
	key := fmt.Sprintf(chatPrefix, message.RoomID)
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return r.client.RPush(ctx, key, data).Err()
}

func (r *ChatRepositoryImpl) GetMessages(ctx context.Context, roomID string, limit int) ([]*entity.ChatMessage, error) {
	key := fmt.Sprintf(chatPrefix, roomID)
	data, err := r.client.LRange(ctx, key, -int64(limit), -1).Result()
	if err != nil {
		return nil, err
	}

	messages := make([]*entity.ChatMessage, 0, len(data))
	for _, v := range data {
		var msg entity.ChatMessage
		if err := json.Unmarshal([]byte(v), &msg); err != nil {
			continue
		}
		messages = append(messages, &msg)
	}
	return messages, nil
}

func (r *ChatRepositoryImpl) DeleteMessages(ctx context.Context, roomID string) error {
	key := fmt.Sprintf(chatPrefix, roomID)
	return r.client.Del(ctx, key).Err()
}

// ============= RECORDING REPOSITORY =============

type RecordingRepositoryImpl struct {
	client *redis.Client
}

func NewRecordingRepository(client *redis.Client) *RecordingRepositoryImpl {
	return &RecordingRepositoryImpl{client: client}
}

func (r *RecordingRepositoryImpl) Create(ctx context.Context, recording *entity.Recording) error {
	key := recordingPrefix + recording.ID
	data, err := json.Marshal(recording)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, 7*24*time.Hour).Err()
}

func (r *RecordingRepositoryImpl) Get(ctx context.Context, recordingID string) (*entity.Recording, error) {
	key := recordingPrefix + recordingID
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("recording not found")
		}
		return nil, err
	}

	var recording entity.Recording
	if err := json.Unmarshal(data, &recording); err != nil {
		return nil, err
	}
	return &recording, nil
}

func (r *RecordingRepositoryImpl) Update(ctx context.Context, recording *entity.Recording) error {
	key := recordingPrefix + recording.ID
	data, err := json.Marshal(recording)
	if err != nil {
		return err
	}
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RecordingRepositoryImpl) GetByRoomID(ctx context.Context, roomID string) ([]*entity.Recording, error) {
	pattern := recordingPrefix + "*"
	var recordings []*entity.Recording

	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		data, err := r.client.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue
		}

		var rec entity.Recording
		if err := json.Unmarshal(data, &rec); err != nil {
			continue
		}

		if rec.RoomID == roomID {
			recordings = append(recordings, &rec)
		}
	}
	return recordings, iter.Err()
}

func (r *RecordingRepositoryImpl) AddChunk(ctx context.Context, recordingID, chunkURL string) error {
	recording, err := r.Get(ctx, recordingID)
	if err != nil {
		return err
	}
	recording.Chunks = append(recording.Chunks, chunkURL)
	return r.Update(ctx, recording)
}

// ============= USER REPOSITORY =============

type UserRepositoryImpl struct {
	client *redis.Client
}

func NewUserRepository(client *redis.Client) *UserRepositoryImpl {
	return &UserRepositoryImpl{client: client}
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	key := userPrefix + user.ID
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, 0).Err()
}

func (r *UserRepositoryImpl) Get(ctx context.Context, userID string) (*entity.User, error) {
	key := userPrefix + userID
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	var user entity.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	pattern := userPrefix + "*"

	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		data, err := r.client.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue
		}

		var user entity.User
		if err := json.Unmarshal(data, &user); err != nil {
			continue
		}

		if user.Email == email {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (r *UserRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	return r.Create(ctx, user)
}
