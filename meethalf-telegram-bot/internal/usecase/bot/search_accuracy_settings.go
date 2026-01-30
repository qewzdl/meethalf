package bot

import (
	"context"
	"errors"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

const (
	searchAccuracyEnableAction  = "enable"
	searchAccuracyDisableAction = "disable"
)

func (s *service) sessionSearchAccuracyEnabled(ctx context.Context, userID int64) bool {
	if s == nil || s.sessions == nil || userID == 0 {
		return false
	}

	session, found, err := s.sessions.Get(ctx, userID)
	if err != nil || !found {
		return false
	}

	return session.SearchAccuracyEnabled
}

func (s *service) setSessionSearchAccuracyEnabled(ctx context.Context, msg domain.IncomingMessage, enabled bool) error {
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
		SearchAccuracyEnabled: enabled,
		PendingAISearch:       s.sessionAISearchPending(ctx, msg.User.ID),
	})
}

func (s *service) parseSearchAccuracyToggleAction(value string) (bool, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case searchAccuracyEnableAction:
		return true, true
	case searchAccuracyDisableAction:
		return false, true
	default:
		return false, false
	}
}

func (s *service) updateSearchAccuracySetting(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, error) {
	if s == nil || s.sessions == nil {
		return s.searchAccuracyUpdateFailedText(l), errors.New("session repository is not configured")
	}

	if msg.User.ID == 0 {
		return s.searchAccuracyUpdateFailedText(l), errors.New("user id is missing")
	}

	enabled, ok := s.parseSearchAccuracyToggleAction(msg.Arguments)
	if !ok {
		return s.searchAccuracyUpdateFailedText(l), nil
	}

	if err := s.setSessionSearchAccuracyEnabled(ctx, msg, enabled); err != nil {
		return s.searchAccuracyUpdateFailedText(l), err
	}

	if enabled {
		return s.searchAccuracyEnabledText(l), nil
	}
	return s.searchAccuracyDisabledText(l), nil
}
