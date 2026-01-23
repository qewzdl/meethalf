package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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
	if err := r.client.HSet(ctx, key, "chat_id", session.ChatID, "last_seen", lastSeen, "username", session.Username).Err(); err != nil {
		return err
	}

	if r.ttl > 0 {
		if err := r.client.Expire(ctx, key, r.ttl).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (r *SessionRepository) Get(ctx context.Context, userID int64) (domain.Session, bool, error) {
	if r == nil || r.client == nil {
		return domain.Session{}, false, errors.New("redis session repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return domain.Session{}, false, err
	}
	if userID <= 0 {
		return domain.Session{}, false, errors.New("user id is required")
	}

	key := fmt.Sprintf("meethalf:sessions:%d", userID)
	values, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return domain.Session{}, false, err
	}
	if len(values) == 0 {
		return domain.Session{}, false, nil
	}

	chatID := int64(0)
	if raw, ok := values["chat_id"]; ok && raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return domain.Session{}, false, err
		}
		chatID = parsed
	}

	lastSeen := time.Time{}
	if raw, ok := values["last_seen"]; ok && raw != "" {
		parsed, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return domain.Session{}, false, err
		}
		lastSeen = parsed
	}

	return domain.Session{
		UserID:   userID,
		ChatID:   chatID,
		Username: values["username"],
		LastSeen: lastSeen,
	}, true, nil
}
