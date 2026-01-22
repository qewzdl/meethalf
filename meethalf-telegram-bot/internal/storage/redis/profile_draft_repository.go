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

type ProfileDraftRepository struct {
	client *redisgo.Client
	ttl    time.Duration
}

func NewProfileDraftRepository(client *redisgo.Client, ttl time.Duration) *ProfileDraftRepository {
	return &ProfileDraftRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *ProfileDraftRepository) Get(ctx context.Context, userID int64) (domain.ProfileDraft, bool, error) {
	if r == nil || r.client == nil {
		return domain.ProfileDraft{}, false, errors.New("redis profile draft repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return domain.ProfileDraft{}, false, err
	}

	value, err := r.client.Get(ctx, profileDraftKey(userID)).Result()
	if err == redisgo.Nil {
		return domain.ProfileDraft{}, false, nil
	}
	if err != nil {
		return domain.ProfileDraft{}, false, err
	}

	var draft domain.ProfileDraft
	if err := json.Unmarshal([]byte(value), &draft); err != nil {
		return domain.ProfileDraft{}, false, err
	}

	return draft, true, nil
}

func (r *ProfileDraftRepository) Save(ctx context.Context, draft domain.ProfileDraft) error {
	if r == nil || r.client == nil {
		return errors.New("redis profile draft repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	payload, err := json.Marshal(draft)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, profileDraftKey(draft.UserID), payload, r.ttl).Err()
}

func (r *ProfileDraftRepository) Delete(ctx context.Context, userID int64) error {
	if r == nil || r.client == nil {
		return errors.New("redis profile draft repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	return r.client.Del(ctx, profileDraftKey(userID)).Err()
}

func profileDraftKey(userID int64) string {
	return fmt.Sprintf("meethalf:profile_drafts:%d", userID)
}
