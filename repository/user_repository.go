package repository

import (
	"bincang-visual/models"
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type UserRepository interface {
	RegisterUser(username string) (*models.User, error)
	GetUser(userId string) (*models.User, error)
	RemoveUser(userId string) error
}

type userRepositoryImpl struct {
	redisClient *redis.Client
}

func NewUserRepository(redisClient *redis.Client) UserRepository {
	return &userRepositoryImpl{
		redisClient: redisClient,
	}
}

func (r *userRepositoryImpl) RegisterUser(username string) (*models.User, error) {
	userId := uuid.New().String()
	ttl := 2 * 24 * time.Hour
	currentTime := time.Now()
	createdAt := currentTime.Format("15:04:05 01-02-2006")
	userModel := models.User{
		ID:        userId,
		CreatedAt: createdAt,
		Username:  username,
	}
	userJson, err := json.Marshal(userModel)
	if err != nil {
		return nil, err
	}
	err = r.redisClient.Set(context.Background(), userId, userJson, ttl).Err()
	if err != nil {
		return nil, err
	}
	return &userModel, nil
}

func (r *userRepositoryImpl) GetUser(userId string) (*models.User, error) {
	result, err := r.redisClient.Get(context.Background(), userId).Bytes()
	if err != nil {
		return nil, err
	}
	var user models.User
	err = json.Unmarshal(result, &user)
	return &user, nil
}

func (r *userRepositoryImpl) RemoveUser(userId string) error {
	err := r.redisClient.Del(context.Background(), userId).Err()
	if err != nil {
		return err
	}
	return nil
}
