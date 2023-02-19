package cache

import (
	"context"
	"os"
	"time"

	log "github.com/acikkaynak/backend-api-go/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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

func (repository *RedisRepository) SetKey(key string, value []byte, ttl time.Duration) {
	status := repository.client.Set(context.Background(), key, value, ttl)
	_, err := status.Result()
	if err != nil {
		log.Logger().Info(err.Error())
	}
}

func (repository *RedisRepository) Get(key string) []byte {
	status := repository.client.Get(context.Background(), key)
	if status.Err() != nil {
		return nil
	}

	resp, err := status.Bytes()
	if err != nil {
		log.Logger().Error("redis cache get error", zap.Error(err))
	}

	return resp
}

func (repository *RedisRepository) Delete(key string) error {
	status := repository.client.Del(context.Background(), key)
	if status.Err() != nil {
		return status.Err()
	}

	return nil
}

func (repository *RedisRepository) Prune() error {
	resp := repository.client.FlushDB(context.Background())
	return resp.Err()
}
