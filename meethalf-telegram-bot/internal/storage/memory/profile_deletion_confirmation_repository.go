package memory

import (
	"context"
	"errors"
	"sync"

	"meethalf-telegram-bot/internal/domain"
)

type ProfileDeletionConfirmationRepository struct {
	mu            sync.RWMutex
	confirmations map[int64]domain.ProfileDeletionConfirmation
}

func NewProfileDeletionConfirmationRepository() *ProfileDeletionConfirmationRepository {
	return &ProfileDeletionConfirmationRepository{
		confirmations: make(map[int64]domain.ProfileDeletionConfirmation),
	}
}

func (r *ProfileDeletionConfirmationRepository) Get(ctx context.Context, userID int64) (domain.ProfileDeletionConfirmation, bool, error) {
	if r == nil {
		return domain.ProfileDeletionConfirmation{}, false, errors.New("memory profile delete confirmation repository is nil")
	}

	if err := ctx.Err(); err != nil {
		return domain.ProfileDeletionConfirmation{}, false, err
	}

	r.mu.RLock()
	confirmation, ok := r.confirmations[userID]
	r.mu.RUnlock()

	return confirmation, ok, nil
}

func (r *ProfileDeletionConfirmationRepository) Save(ctx context.Context, confirmation domain.ProfileDeletionConfirmation) error {
	if r == nil {
		return errors.New("memory profile delete confirmation repository is nil")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	r.confirmations[confirmation.UserID] = confirmation
	r.mu.Unlock()

	return nil
}

func (r *ProfileDeletionConfirmationRepository) Delete(ctx context.Context, userID int64) error {
	if r == nil {
		return errors.New("memory profile delete confirmation repository is nil")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	delete(r.confirmations, userID)
	r.mu.Unlock()

	return nil
}
