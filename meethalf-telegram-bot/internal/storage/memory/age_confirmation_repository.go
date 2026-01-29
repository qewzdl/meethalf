package memory

import (
	"context"
	"errors"
	"strings"
	"sync"

	"meethalf-telegram-bot/internal/domain"
)

type AgeConfirmationRepository struct {
	mu            sync.RWMutex
	confirmations map[int64]domain.AgeConfirmation
	byUsername    map[string]int64
}

func NewAgeConfirmationRepository() *AgeConfirmationRepository {
	return &AgeConfirmationRepository{
		confirmations: make(map[int64]domain.AgeConfirmation),
		byUsername:    make(map[string]int64),
	}
}

func (r *AgeConfirmationRepository) Get(ctx context.Context, userID int64) (domain.AgeConfirmation, bool, error) {
	if r == nil {
		return domain.AgeConfirmation{}, false, errors.New("memory age confirmation repository is nil")
	}
	if err := ctx.Err(); err != nil {
		return domain.AgeConfirmation{}, false, err
	}

	r.mu.RLock()
	confirmation, ok := r.confirmations[userID]
	r.mu.RUnlock()

	return confirmation, ok, nil
}

func (r *AgeConfirmationRepository) Save(ctx context.Context, confirmation domain.AgeConfirmation) error {
	if r == nil {
		return errors.New("memory age confirmation repository is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	username := normalizeUsername(confirmation.Username)
	confirmation.Username = username

	r.mu.Lock()
	if existing, ok := r.confirmations[confirmation.UserID]; ok && existing.Username != "" {
		delete(r.byUsername, existing.Username)
	}
	r.confirmations[confirmation.UserID] = confirmation
	if username != "" {
		r.byUsername[username] = confirmation.UserID
	}
	r.mu.Unlock()

	return nil
}

func (r *AgeConfirmationRepository) Delete(ctx context.Context, userID int64) error {
	if r == nil {
		return errors.New("memory age confirmation repository is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	if existing, ok := r.confirmations[userID]; ok && existing.Username != "" {
		delete(r.byUsername, existing.Username)
	}
	delete(r.confirmations, userID)
	r.mu.Unlock()

	return nil
}

func (r *AgeConfirmationRepository) FindUserIDByUsername(ctx context.Context, username string) (int64, bool, error) {
	if r == nil {
		return 0, false, errors.New("memory age confirmation repository is nil")
	}
	if err := ctx.Err(); err != nil {
		return 0, false, err
	}

	normalized := normalizeUsername(username)
	if normalized == "" {
		return 0, false, nil
	}

	r.mu.RLock()
	userID, ok := r.byUsername[normalized]
	r.mu.RUnlock()
	if !ok {
		return 0, false, nil
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
