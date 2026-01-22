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

type ProfileDeletionConfirmationRepository struct {
	client *redisgo.Client
	ttl    time.Duration
}

func NewProfileDeletionConfirmationRepository(client *redisgo.Client, ttl time.Duration) *ProfileDeletionConfirmationRepository {
	return &ProfileDeletionConfirmationRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *ProfileDeletionConfirmationRepository) Get(ctx context.Context, userID int64) (domain.ProfileDeletionConfirmation, bool, error) {
	if r == nil || r.client == nil {
		return domain.ProfileDeletionConfirmation{}, false, errors.New("redis profile delete confirmation repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return domain.ProfileDeletionConfirmation{}, false, err
	}

	value, err := r.client.Get(ctx, profileDeleteConfirmationKey(userID)).Result()
	if err == redisgo.Nil {
		return domain.ProfileDeletionConfirmation{}, false, nil
	}
	if err != nil {
		return domain.ProfileDeletionConfirmation{}, false, err
	}

	var confirmation domain.ProfileDeletionConfirmation
	if err := json.Unmarshal([]byte(value), &confirmation); err != nil {
		return domain.ProfileDeletionConfirmation{}, false, err
	}

	return confirmation, true, nil
}

func (r *ProfileDeletionConfirmationRepository) Save(ctx context.Context, confirmation domain.ProfileDeletionConfirmation) error {
	if r == nil || r.client == nil {
		return errors.New("redis profile delete confirmation repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	payload, err := json.Marshal(confirmation)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, profileDeleteConfirmationKey(confirmation.UserID), payload, r.ttl).Err()
}

func (r *ProfileDeletionConfirmationRepository) Delete(ctx context.Context, userID int64) error {
	if r == nil || r.client == nil {
		return errors.New("redis profile delete confirmation repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	return r.client.Del(ctx, profileDeleteConfirmationKey(userID)).Err()
}

func profileDeleteConfirmationKey(userID int64) string {
	return fmt.Sprintf("meethalf:profile_delete_confirmations:%d", userID)
}
