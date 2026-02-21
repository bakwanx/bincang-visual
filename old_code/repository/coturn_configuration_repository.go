package repository

import (
	"bincang-visual/old_code/models"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type CoturnConfigurationRepository interface {
	GetConfiguration() (*string, error)
}

type coturnConfigurationRepositoryImpl struct {
	redisClient *redis.Client
}

func NewCoturnConfigurationRepository(redisClient *redis.Client) CoturnConfigurationRepository {
	return &coturnConfigurationRepositoryImpl{redisClient: redisClient}
}

func (r coturnConfigurationRepositoryImpl) GetConfiguration() (*string, error) {
	key := "config:coturn"
	val, err := r.redisClient.Get(context.Background(), key).Bytes()
	if err != nil {
		fmt.Println("Redis error: ", err)
		return nil, err
	}

	var coturn models.CoturnConfiguration
	err = json.Unmarshal(val, &coturn)
	if err != nil {
		fmt.Println("Error unmarshalling Join: ", err)
	}
	result := string(val)

	return &result, nil
}
