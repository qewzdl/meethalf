package bot

import (
	"context"

	"meethalf-telegram-bot/internal/domain"
)

const aiSearchMinPromptLength = 3

func (s *service) sessionAISearchPending(ctx context.Context, userID int64) bool {
	if s == nil || s.sessions == nil || userID == 0 {
		return false
	}

	session, found, err := s.sessions.Get(ctx, userID)
	if err != nil || !found {
		return false
	}

	return session.PendingAISearch
}

func (s *service) setSessionAISearchPending(ctx context.Context, msg domain.IncomingMessage, pending bool) error {
	if s == nil || s.sessions == nil || msg.User.ID == 0 {
		return nil
	}

	lang := msg.Language
	if lang == "" {
		lang = s.resolveLanguage(ctx, msg)
	}

	return s.sessions.Touch(ctx, domain.Session{
		UserID:                msg.User.ID,
		ChatID:                msg.ChatID,
		Username:              msg.User.Username,
		Language:              normalizeLanguageValue(lang),
		LastSeen:              s.now(msg.ReceivedAt),
		SearchAccuracyEnabled: s.sessionSearchAccuracyEnabled(ctx, msg.User.ID),
		PendingAISearch:       pending,
	})
}

func (s *service) hasActiveDraft(ctx context.Context, userID int64) bool {
	if s == nil || s.drafts == nil || userID == 0 {
		return false
	}

	_, found, err := s.drafts.Get(ctx, userID)
	if err != nil || !found {
		return false
	}
	return true
}
