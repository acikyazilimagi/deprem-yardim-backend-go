package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

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

func (repository *RedisRepository) SetKey(key string, value interface{}, ttl int) {
	duration, _ := time.ParseDuration(strconv.FormatInt(int64(ttl), 10))
	status := repository.client.Set(key, value, duration)
	_, err := status.Result()
	if err != nil {
		fmt.Println(err)
	}
}

func (repository *RedisRepository) Get(key string) interface{} {
	status := repository.client.Get(key)
	if status.Err() != nil {
		fmt.Println(status.Err())
		return nil
	}

	stringResult, err := status.Result()

	var data interface{}
	if err = json.Unmarshal([]byte(stringResult), &data); err != nil {
		fmt.Println(err)
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
