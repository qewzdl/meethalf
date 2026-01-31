package bot

import (
	"context"
	"errors"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

const (
	likeNotificationsEnableAction  = "enable"
	likeNotificationsDisableAction = "disable"
)

func (s *service) parseLikeNotificationsToggleAction(value string) (bool, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case likeNotificationsEnableAction:
		return true, true
	case likeNotificationsDisableAction:
		return false, true
	default:
		return false, false
	}
}

func (s *service) updateLikeNotificationsSetting(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, error) {
	if s == nil || s.profiles == nil {
		return s.likeNotificationsUpdateFailedText(l), errors.New("profile service is not configured")
	}

	if msg.User.ID == 0 {
		return s.likeNotificationsUpdateFailedText(l), errors.New("user id is missing")
	}

	enabled, ok := s.parseLikeNotificationsToggleAction(msg.Arguments)
	if !ok {
		return s.likeNotificationsUpdateFailedText(l), nil
	}

	updated, err := s.profiles.SetProfileLikeNotifications(ctx, msg.User.ID, enabled)
	if err != nil {
		if isBannedError(err) {
			return s.userBannedText(l), err
		}
		return s.likeNotificationsUpdateFailedText(l), err
	}
	if !updated {
		return l.message(msgProfileNotFoundCreateButton), nil
	}

	if enabled {
		return s.likeNotificationsEnabledText(l), nil
	}
	return s.likeNotificationsDisabledText(l), nil
}
