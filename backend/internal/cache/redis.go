package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
)

func InitRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func Set(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return redisClient.Set(ctx, key, jsonValue, expiration).Err()
}

func Get(key string, dest interface{}) error {
	ctx := context.Background()
	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}
