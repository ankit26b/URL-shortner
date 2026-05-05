package cache

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

type URLCache interface {
	Get(ctx context.Context, shortCode string) (string, error)
	Set(ctx context.Context, shortCode, longURL string) error
}

type RedisURLCache struct {
	client *goredis.Client
	ttl    time.Duration
}

func NewRedisURLCache(client *goredis.Client, ttl time.Duration) *RedisURLCache {
	return &RedisURLCache{client: client, ttl: ttl}
}

func (c *RedisURLCache) Get(ctx context.Context, shortCode string) (string, error) {
	return c.client.Get(ctx, c.key(shortCode)).Result()
}

func (c *RedisURLCache) Set(ctx context.Context, shortCode, longURL string) error {
	return c.client.Set(ctx, c.key(shortCode), longURL, c.ttl).Err()
}

func (c *RedisURLCache) key(shortCode string) string {
	return fmt.Sprintf("url:%s", shortCode)
}
