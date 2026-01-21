package redis

import (
	"context"
	"fmt"
	"net"

	"meethalf-telegram-bot/internal/config"

	redisgo "github.com/redis/go-redis/v9"
)

func New(cfg config.RedisConfig) (*redisgo.Client, error) {
	host := cfg.Host
	if host == "" {
		host = "localhost"
	}

	port := cfg.Port
	if port == "" {
		port = "6379"
	}

	options := &redisgo.Options{
		Addr:     net.JoinHostPort(host, port),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	if cfg.ConnectTimeout > 0 {
		options.DialTimeout = cfg.ConnectTimeout
	}
	if cfg.ReadTimeout > 0 {
		options.ReadTimeout = cfg.ReadTimeout
	}
	if cfg.WriteTimeout > 0 {
		options.WriteTimeout = cfg.WriteTimeout
	}
	if cfg.PoolSize > 0 {
		options.PoolSize = cfg.PoolSize
	}
	if cfg.MinIdleConns > 0 {
		options.MinIdleConns = cfg.MinIdleConns
	}

	client := redisgo.NewClient(options)

	ctx := context.Background()
	if cfg.ConnectTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.ConnectTimeout)
		defer cancel()
	}

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}
