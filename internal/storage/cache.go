package storage

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCache(addr string, ttl time.Duration) (*Cache, error) {
	client := redis.NewClient(&redis.Options{Addr: addr})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return &Cache{client: client, ttl: ttl}, nil
}

func HashString(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

func CacheKey(input any) string {
	data, _ := json.Marshal(input)
	return HashString(string(data))
}

func (c *Cache) GetPrediction(ctx context.Context, hash string) ([]byte, error) {
	return c.client.Get(ctx, "PRED:"+hash).Bytes()
}

func (c *Cache) SetPrediction(ctx context.Context, hash string, data []byte) error {
	return c.client.Set(ctx, "PRED:"+hash, data, c.ttl).Err()
}

func (c *Cache) IncrCounter(ctx context.Context, key string) error {
	return c.client.Incr(ctx, key).Err()
}

func (c *Cache) IncrBy(ctx context.Context, key string, count int64) error {
	return c.client.IncrBy(ctx, key, count).Err()
}

func (c *Cache) IncrByFloat(ctx context.Context, key string, value float64) error {
	return c.client.IncrByFloat(ctx, key, value).Err()
}

func (c *Cache) GetCounter(ctx context.Context, key string) (int64, error) {
	return c.client.Get(ctx, key).Int64()
}

func (c *Cache) GetFloat64(ctx context.Context, key string) (float64, error) {
	return c.client.Get(ctx, key).Float64()
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Cache) Close() error {
	return c.client.Close()
}
