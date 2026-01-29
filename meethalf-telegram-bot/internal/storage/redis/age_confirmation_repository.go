package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"meethalf-telegram-bot/internal/domain"

	redisgo "github.com/redis/go-redis/v9"
)

type AgeConfirmationRepository struct {
	client *redisgo.Client
}

func NewAgeConfirmationRepository(client *redisgo.Client) *AgeConfirmationRepository {
	return &AgeConfirmationRepository{
		client: client,
	}
}

func (r *AgeConfirmationRepository) Get(ctx context.Context, userID int64) (domain.AgeConfirmation, bool, error) {
	if r == nil || r.client == nil {
		return domain.AgeConfirmation{}, false, errors.New("redis age confirmation repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return domain.AgeConfirmation{}, false, err
	}
	if userID <= 0 {
		return domain.AgeConfirmation{}, false, errors.New("user id is required")
	}

	value, err := r.client.Get(ctx, ageConfirmationKey(userID)).Result()
	if err == redisgo.Nil {
		return domain.AgeConfirmation{}, false, nil
	}
	if err != nil {
		return domain.AgeConfirmation{}, false, err
	}

	var confirmation domain.AgeConfirmation
	if err := json.Unmarshal([]byte(value), &confirmation); err != nil {
		return domain.AgeConfirmation{}, false, err
	}

	return confirmation, true, nil
}

func (r *AgeConfirmationRepository) Save(ctx context.Context, confirmation domain.AgeConfirmation) error {
	if r == nil || r.client == nil {
		return errors.New("redis age confirmation repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if confirmation.UserID <= 0 {
		return errors.New("user id is required")
	}

	username := normalizeUsername(confirmation.Username)
	confirmation.Username = username

	existing, found, err := r.Get(ctx, confirmation.UserID)
	if err != nil {
		return err
	}
	if found {
		oldUsername := normalizeUsername(existing.Username)
		if oldUsername != "" && oldUsername != username {
			if err := r.client.Del(ctx, ageConfirmationUsernameKey(oldUsername)).Err(); err != nil {
				return err
			}
		}
		if username == "" && oldUsername != "" {
			if err := r.client.Del(ctx, ageConfirmationUsernameKey(oldUsername)).Err(); err != nil {
				return err
			}
		}
	}

	payload, err := json.Marshal(confirmation)
	if err != nil {
		return err
	}

	if err := r.client.Set(ctx, ageConfirmationKey(confirmation.UserID), payload, 0).Err(); err != nil {
		return err
	}
	if username == "" {
		return nil
	}

	return r.client.Set(ctx, ageConfirmationUsernameKey(username), confirmation.UserID, 0).Err()
}

func (r *AgeConfirmationRepository) Delete(ctx context.Context, userID int64) error {
	if r == nil || r.client == nil {
		return errors.New("redis age confirmation repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if userID <= 0 {
		return errors.New("user id is required")
	}

	confirmation, found, err := r.Get(ctx, userID)
	if err != nil {
		return err
	}
	if found {
		username := normalizeUsername(confirmation.Username)
		if username != "" {
			if err := r.client.Del(ctx, ageConfirmationUsernameKey(username)).Err(); err != nil {
				return err
			}
		}
	}

	return r.client.Del(ctx, ageConfirmationKey(userID)).Err()
}

func ageConfirmationKey(userID int64) string {
	return fmt.Sprintf("meethalf:age_confirmations:%d", userID)
}

func ageConfirmationUsernameKey(username string) string {
	return fmt.Sprintf("meethalf:age_confirmations_by_username:%s", username)
}

func (r *AgeConfirmationRepository) FindUserIDByUsername(ctx context.Context, username string) (int64, bool, error) {
	if r == nil || r.client == nil {
		return 0, false, errors.New("redis age confirmation repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return 0, false, err
	}

	normalized := normalizeUsername(username)
	if normalized == "" {
		return 0, false, nil
	}

	value, err := r.client.Get(ctx, ageConfirmationUsernameKey(normalized)).Result()
	if err == redisgo.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	userID, err := strconv.ParseInt(value, 10, 64)
	if err != nil || userID <= 0 {
		if err == nil {
			err = errors.New("invalid user id value")
		}
		return 0, false, err
	}

	return userID, true, nil
}

func normalizeUsername(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return ""
	}
	normalized = strings.TrimPrefix(normalized, "@")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return ""
	}
	return strings.ToLower(normalized)
}
