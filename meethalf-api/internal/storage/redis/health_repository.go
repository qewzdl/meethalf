package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type HealthRepository struct {
	client  *redis.Client
	timeout time.Duration
}

func NewHealthRepository(client *redis.Client, timeout time.Duration) *HealthRepository {
	return &HealthRepository{
		client:  client,
		timeout: timeout,
	}
}

func (r *HealthRepository) Name() string {
	return "redis"
}

func (r *HealthRepository) Timeout() time.Duration {
	if r == nil {
		return 0
	}

	return r.timeout
}

func (r *HealthRepository) Ping(ctx context.Context) error {
	if r == nil || r.client == nil {
		return errors.New("redis client is not initialized")
	}

	return r.client.Ping(ctx).Err()
}
