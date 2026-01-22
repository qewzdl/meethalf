package memory

import (
	"context"
	"errors"
	"sync"

	"meethalf-telegram-bot/internal/domain"
)

type ProfileDraftRepository struct {
	mu     sync.RWMutex
	drafts map[int64]domain.ProfileDraft
}

func NewProfileDraftRepository() *ProfileDraftRepository {
	return &ProfileDraftRepository{
		drafts: make(map[int64]domain.ProfileDraft),
	}
}

func (r *ProfileDraftRepository) Get(ctx context.Context, userID int64) (domain.ProfileDraft, bool, error) {
	if r == nil {
		return domain.ProfileDraft{}, false, errors.New("memory profile draft repository is nil")
	}

	if err := ctx.Err(); err != nil {
		return domain.ProfileDraft{}, false, err
	}

	r.mu.RLock()
	draft, ok := r.drafts[userID]
	r.mu.RUnlock()

	return draft, ok, nil
}

func (r *ProfileDraftRepository) Save(ctx context.Context, draft domain.ProfileDraft) error {
	if r == nil {
		return errors.New("memory profile draft repository is nil")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	r.drafts[draft.UserID] = draft
	r.mu.Unlock()

	return nil
}

func (r *ProfileDraftRepository) Delete(ctx context.Context, userID int64) error {
	if r == nil {
		return errors.New("memory profile draft repository is nil")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	delete(r.drafts, userID)
	r.mu.Unlock()

	return nil
}
