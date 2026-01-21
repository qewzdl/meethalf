package memory

import (
	"context"
	"errors"
	"sync"

	"meethalf-telegram-bot/internal/domain"
)

type SessionRepository struct {
	mu       sync.RWMutex
	sessions map[int64]domain.Session
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		sessions: make(map[int64]domain.Session),
	}
}

func (r *SessionRepository) Touch(ctx context.Context, session domain.Session) error {
	if r == nil {
		return errors.New("memory session repository is nil")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	r.sessions[session.UserID] = session
	r.mu.Unlock()

	return nil
}
