package utils

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

func GenerateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func GenerateUniqueKey(redisClient *redis.Client, prefix string) (string, error) {
	for {
		randomStr := fmt.Sprintf("%s-%s-%s", GenerateRandomString(3), GenerateRandomString(4), GenerateRandomString(3))

		key := fmt.Sprintf("%s%s", prefix, randomStr)

		exists, err := redisClient.Exists(context.Background(), key).Result()
		if err != nil {
			return "", err
		}

		if exists == 0 {
			return key, nil
		}
	}
}

func RemovePrefix(input, prefix string) string {
	if len(input) > len(prefix) && input[:len(prefix)] == prefix {
		return input[len(prefix):]
	}
	return input
}
