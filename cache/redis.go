package cache

import (
	"encoding/json"
	"os"
	"time"

	log "github.com/acikkaynak/backend-api-go/pkg/logger"
	"github.com/go-redis/redis"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository() *RedisRepository {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("RedisAddr"),
		Password: os.Getenv("RedisPassword"),
		DB:       0,
	})

	return &RedisRepository{client: client}
}

func (repository *RedisRepository) SetKey(key string, value interface{}, ttl time.Duration) {
	status := repository.client.Set(key, value, ttl)
	_, err := status.Result()
	if err != nil {
		log.Logger().Info(err.Error())
	}
}

func (repository *RedisRepository) Get(key string) interface{} {
	status := repository.client.Get(key)
	if status.Err() != nil {
		return nil
	}

	stringResult, err := status.Result()

	var data interface{}
	if err = json.Unmarshal([]byte(stringResult), &data); err != nil {
		log.Logger().Info(err.Error())
		return nil
	}

	return data
}

func (repository *RedisRepository) Delete(key string) error {
	status := repository.client.Del(key)
	if status.Err() != nil {
		return status.Err()
	}

	return nil
}

func (repository *RedisRepository) Prune() error {
	resp := repository.client.FlushDB()
	return resp.Err()
}
