package repository

import (
	"bincang-visual/models"
	"bincang-visual/utils"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RoomRepository interface {
	CreateRoom() (*models.Room, error)
	GetRoom(roomId string) (*models.Room, error)
	AddUser(roomId string, userModel models.User) error
	RemoveUser(roomId, userId string) error
}

type roomRepositoryImpl struct {
	redisClient *redis.Client
}

func NewRoomRepository(redisClient *redis.Client) RoomRepository {
	return &roomRepositoryImpl{redisClient: redisClient}
}

func (r roomRepositoryImpl) CreateRoom() (*models.Room, error) {
	// roomId := uuid.New().String()
	roomId := utils.GenerateRandomString()
	ttl := 2 * 24 * time.Hour

	// ttl := 20 * time.Second
	currentTime := time.Now()
	createdAt := currentTime.Format("15:04:05 01-02-2006")

	roomJson, err := json.Marshal(models.Room{
		RoomId:    roomId,
		CreatedAt: createdAt,
	})

	if err != nil {
		fmt.Println("Error marshalling", err)
		return nil, err
	}

	err = r.redisClient.Set(context.Background(), roomId, roomJson, ttl).Err()
	if err != nil {
		fmt.Println("Redis error: ", err)
		return nil, err
	}

	return &models.Room{
		RoomId:    roomId,
		CreatedAt: createdAt,
	}, nil
}

func (r roomRepositoryImpl) GetRoom(roomId string) (*models.Room, error) {
	val, err := r.redisClient.Get(context.Background(), roomId).Bytes()
	if err != nil {
		fmt.Println("Redis error: ", err)
		return nil, err
	}

	var roomObj models.Room
	err = json.Unmarshal(val, &roomObj)
	if err != nil {
		fmt.Println("Error unmarshalling Join: ", err)
	}

	return &roomObj, nil
}

func (r roomRepositoryImpl) AddUser(roomId string, userModel models.User) error {
	key := roomId

	// Get room JSON
	val, err := r.redisClient.Get(context.Background(), roomId).Bytes()
	if err != nil {
		return err
	}

	var room models.Room
	if err := json.Unmarshal(val, &room); err != nil {
		return err
	}

	// Add User
	if room.Users == nil {
		room.Users = make(map[string]models.User)
	}
	room.Users[userModel.ID] = userModel

	// Update Room
	updatedRoom, err := json.Marshal(room)
	if err != nil {
		return err
	}

	return r.redisClient.SetArgs(context.Background(), key, updatedRoom, redis.SetArgs{KeepTTL: true}).Err()
}

func (r roomRepositoryImpl) RemoveUser(roomId, userId string) error {
	key := roomId
	val, err := r.redisClient.Get(context.Background(), key).Bytes()
	if err != nil {
		return err
	}

	var room models.Room
	if err := json.Unmarshal(val, &room); err != nil {
		return err
	}

	// Delete user
	delete(room.Users, userId)

	updated, err := json.Marshal(room)

	if err != nil {
		return err
	}

	return r.redisClient.SetArgs(context.Background(), key, updated, redis.SetArgs{KeepTTL: true}).Err()
}
