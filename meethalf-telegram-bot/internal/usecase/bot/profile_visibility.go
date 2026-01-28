package bot

import (
	"context"
	"errors"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) updateProfileVisibility(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, error) {
	if s == nil || s.profiles == nil {
		return s.profileVisibilityUpdateFailedText(l), errors.New("profile service is not configured")
	}

	if msg.User.ID == 0 {
		return s.profileVisibilityUpdateFailedText(l), errors.New("user id is missing")
	}

	isHidden, ok := s.parseProfileVisibilityAction(msg.Arguments)
	if !ok {
		return s.profileVisibilityUpdateFailedText(l), nil
	}

	updated, err := s.profiles.SetProfileVisibility(ctx, msg.User.ID, isHidden)
	if err != nil {
		if isBannedError(err) {
			return s.userBannedText(l), err
		}
		return s.profileVisibilityUpdateFailedText(l), err
	}
	if !updated {
		return l.message(msgProfileNotFoundCreateButton), nil
	}

	return s.profileVisibilityUpdated(l, isHidden), nil
}
