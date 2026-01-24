package bot

import (
	"context"
	"errors"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) updateProfileVisibility(ctx context.Context, msg domain.IncomingMessage) (string, error) {
	if s == nil || s.profiles == nil {
		return s.profileVisibilityUpdateFailedText(), errors.New("profile service is not configured")
	}

	if msg.User.ID == 0 {
		return s.profileVisibilityUpdateFailedText(), errors.New("user id is missing")
	}

	isHidden, ok := s.parseProfileVisibilityAction(msg.Arguments)
	if !ok {
		return s.profileVisibilityUpdateFailedText(), nil
	}

	updated, err := s.profiles.SetProfileVisibility(ctx, msg.User.ID, isHidden)
	if err != nil {
		if isBannedError(err) {
			return s.userBannedText(), err
		}
		return s.profileVisibilityUpdateFailedText(), err
	}
	if !updated {
		return "Profile not found. Use the Create Profile button to create it.", nil
	}

	return s.profileVisibilityUpdated(isHidden), nil
}
