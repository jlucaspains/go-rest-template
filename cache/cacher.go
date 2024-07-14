package cache

import (
	"context"
	"encoding/json"
	"goapi-template/config"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cacher interface {
	Get(ctx context.Context, key string) ([]byte, error)
	GetString(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value []byte) error
	SetString(ctx context.Context, key string, value string) error
	DeleteKey(ctx context.Context, key string) error
}

type RedisCacher struct {
	client     *redis.Client
	expiration time.Duration
}

// func (t *RedisCacher) Get(ctx context.Context, key string) ([]byte, error) {
// 	return t.client.Get(ctx, key).Bytes()
// }

func (t *RedisCacher) GetString(ctx context.Context, key string) (string, error) {
	return t.client.Get(ctx, key).Result()
}

// func (t *RedisCacher) Set(ctx context.Context, key string, value []byte) error {
// 	return t.client.Set(ctx, key, value, t.expiration).Err()
// }

func (t *RedisCacher) SetString(ctx context.Context, key string, value string) error {
	return t.client.Set(ctx, key, value, t.expiration).Err()
}

func (t *RedisCacher) DeleteKey(ctx context.Context, key string) error {
	return t.client.Del(ctx, key).Err()
}

func (t *RedisCacher) Close() {
	err := t.client.Close()

	slog.Error("Error closing redis connection", "error", err)
}

func NewRawCacher(config *config.CacheConfiguration) *RedisCacher {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddress,
		Password: config.RedisPassword,
		DB:       config.RedisDb,
	})

	return &RedisCacher{client: rdb, expiration: config.Expiration}
}

func GetObject[T any](cacher Cacher, ctx context.Context, key string) (*T, error) {
	p, err := cacher.GetString(ctx, key)
	if err != nil {
		return nil, err
	}

	result := new(T)
	err = json.Unmarshal([]byte(p), result)

	return result, err
}

func SetObject[T any](cacher Cacher, ctx context.Context, key string, value *T) error {
	p, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return cacher.SetString(ctx, key, string(p))
}
