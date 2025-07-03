package repository

import (
	"bincang-visual/models"
	"bincang-visual/utils"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type CoturnConfigurationRepository interface {
	GetConfiguration() (*models.CoturnConfiguration, error)
}

type coturnConfigurationRepositoryImpl struct {
	redisClient *redis.Client
}

func NewCoturnConfigurationRepository(redisClient *redis.Client) CoturnConfigurationRepository {
	return &coturnConfigurationRepositoryImpl{redisClient: redisClient}
}

func (r coturnConfigurationRepositoryImpl) GetConfiguration() (*models.CoturnConfiguration, error) {
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

	return &coturn, nil
}

func (r coturnConfigurationRepositoryImpl) GetRoom(roomId string) (*models.Room, error) {
	key := roomPrefix + roomId
	val, err := r.redisClient.Get(context.Background(), key).Bytes()
	if err != nil {
		fmt.Println("Redis error: ", err)
		return nil, err
	}

	var roomObj models.Room
	err = json.Unmarshal(val, &roomObj)
	if err != nil {
		fmt.Println("Error unmarshalling Join: ", err)
	}
	roomObj.RoomId = utils.RemovePrefix(roomObj.RoomId, roomPrefix)
	return &roomObj, nil
}
