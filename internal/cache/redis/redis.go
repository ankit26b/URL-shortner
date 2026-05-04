package redis

import (
	"context"
	"fmt"

	goredis "github.com/go-redis/redis/v8"

	"url-shortener/internal/config"
)

func New(cfg *config.Config) (*goredis.Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  cfg.RedisDialTimeout,
		ReadTimeout:  cfg.RedisReadTimeout,
		WriteTimeout: cfg.RedisWriteTimeout,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
