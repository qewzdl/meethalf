package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"meethalf-telegram-bot/internal/domain"
)

type AdminActionRepository struct {
	mu      sync.RWMutex
	actions map[int64]domain.AdminActionState
	ttl     time.Duration
}

func NewAdminActionRepository(ttl time.Duration) *AdminActionRepository {
	return &AdminActionRepository{
		actions: make(map[int64]domain.AdminActionState),
		ttl:     ttl,
	}
}

func (r *AdminActionRepository) Get(ctx context.Context, userID int64) (domain.AdminActionState, bool, error) {
	if r == nil {
		return domain.AdminActionState{}, false, errors.New("memory admin action repository is nil")
	}

	if err := ctx.Err(); err != nil {
		return domain.AdminActionState{}, false, err
	}

	r.mu.RLock()
	action, ok := r.actions[userID]
	r.mu.RUnlock()
	if !ok {
		return domain.AdminActionState{}, false, nil
	}

	if r.ttl > 0 {
		expired := action.RequestedAt.IsZero() || time.Since(action.RequestedAt) > r.ttl
		if expired {
			_ = r.Delete(ctx, userID)
			return domain.AdminActionState{}, false, nil
		}
	}

	return action, true, nil
}

func (r *AdminActionRepository) Save(ctx context.Context, action domain.AdminActionState) error {
	if r == nil {
		return errors.New("memory admin action repository is nil")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	r.actions[action.UserID] = action
	r.mu.Unlock()

	return nil
}

func (r *AdminActionRepository) Delete(ctx context.Context, userID int64) error {
	if r == nil {
		return errors.New("memory admin action repository is nil")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	delete(r.actions, userID)
	r.mu.Unlock()

	return nil
}
