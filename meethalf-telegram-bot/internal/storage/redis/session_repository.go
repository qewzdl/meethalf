package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"meethalf-telegram-bot/internal/domain"

	redisgo "github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	client *redisgo.Client
	ttl    time.Duration
}

func NewSessionRepository(client *redisgo.Client, ttl time.Duration) *SessionRepository {
	return &SessionRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *SessionRepository) Touch(ctx context.Context, session domain.Session) error {
	if r == nil || r.client == nil {
		return errors.New("redis session repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	key := fmt.Sprintf("meethalf:sessions:%d", session.UserID)
	lastSeen := session.LastSeen.UTC().Format(time.RFC3339Nano)
	if err := r.client.HSet(ctx, key, "chat_id", session.ChatID, "last_seen", lastSeen).Err(); err != nil {
		return err
	}

	if r.ttl > 0 {
		if err := r.client.Expire(ctx, key, r.ttl).Err(); err != nil {
			return err
		}
	}

	return nil
}
