package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"meethalf-telegram-bot/internal/domain"

	redisgo "github.com/redis/go-redis/v9"
)

type AdminActionRepository struct {
	client *redisgo.Client
	ttl    time.Duration
}

func NewAdminActionRepository(client *redisgo.Client, ttl time.Duration) *AdminActionRepository {
	return &AdminActionRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *AdminActionRepository) Get(ctx context.Context, userID int64) (domain.AdminActionState, bool, error) {
	if r == nil || r.client == nil {
		return domain.AdminActionState{}, false, errors.New("redis admin action repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return domain.AdminActionState{}, false, err
	}

	value, err := r.client.Get(ctx, adminActionKey(userID)).Result()
	if err == redisgo.Nil {
		return domain.AdminActionState{}, false, nil
	}
	if err != nil {
		return domain.AdminActionState{}, false, err
	}

	var action domain.AdminActionState
	if err := json.Unmarshal([]byte(value), &action); err != nil {
		return domain.AdminActionState{}, false, err
	}

	return action, true, nil
}

func (r *AdminActionRepository) Save(ctx context.Context, action domain.AdminActionState) error {
	if r == nil || r.client == nil {
		return errors.New("redis admin action repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	payload, err := json.Marshal(action)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, adminActionKey(action.UserID), payload, r.ttl).Err()
}

func (r *AdminActionRepository) Delete(ctx context.Context, userID int64) error {
	if r == nil || r.client == nil {
		return errors.New("redis admin action repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	return r.client.Del(ctx, adminActionKey(userID)).Err()
}

func adminActionKey(userID int64) string {
	return fmt.Sprintf("meethalf:admin_actions:%d", userID)
}
