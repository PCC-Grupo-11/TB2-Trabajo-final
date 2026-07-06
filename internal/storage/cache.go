package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strconv"
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
	h := fnv.New64a()
	h.Write([]byte(s))
	return strconv.FormatUint(h.Sum64(), 16)
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

func (c *Cache) GetPredictionsPipeline(ctx context.Context, hashes []string) (map[string][]byte, error) {
	if len(hashes) == 0 {
		return nil, nil
	}
	cmds, err := c.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, h := range hashes {
			pipe.Get(ctx, "PRED:"+h)
		}
		return nil
	})
	if err != nil && err != redis.Nil {
		return nil, err
	}
	out := make(map[string][]byte, len(hashes))
	for i, cmd := range cmds {
		if getCmd, ok := cmd.(*redis.StringCmd); ok {
			data, err := getCmd.Bytes()
			if err == nil {
				out[hashes[i]] = data
			}
		}
	}
	return out, nil
}

func (c *Cache) SetPredictionsPipeline(ctx context.Context, entries map[string][]byte) error {
	if len(entries) == 0 {
		return nil
	}
	_, err := c.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for hash, data := range entries {
			pipe.Set(ctx, "PRED:"+hash, data, c.ttl)
		}
		return nil
	})
	return err
}

func (c *Cache) IncrMetricsPipeline(ctx context.Context, counters map[string]int64, floats map[string]float64) error {
	if len(counters) == 0 && len(floats) == 0 {
		return nil
	}
	_, err := c.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for k, v := range counters {
			pipe.IncrBy(ctx, k, v)
		}
		for k, v := range floats {
			pipe.IncrByFloat(ctx, k, v)
		}
		return nil
	})
	return err
}

func (c *Cache) ResetMetrics(ctx context.Context) error {
	return c.client.Del(ctx, "predictions_count", "latency_sum", "cache_hits", "cache_misses").Err()
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Cache) Close() error {
	return c.client.Close()
}
