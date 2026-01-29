package bot

import (
	"context"
	"errors"
	"net/http"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) ensureModeratorCanModerateUser(ctx context.Context, role adminRole, userID int64, username, failText, usageText string, l localizer) (string, error) {
	if role != adminRoleModerator {
		return "", nil
	}

	if s == nil || s.admin == nil {
		return failText, errors.New("admin service is not configured")
	}

	var (
		user domain.UserSummary
		err  error
	)
	if username != "" {
		user, err = s.admin.GetUserByUsername(ctx, username)
	} else {
		user, err = s.admin.GetUser(ctx, userID)
	}
	if err != nil {
		text := failText
		var status statusError
		if errors.As(err, &status) {
			switch status.StatusCode() {
			case http.StatusBadRequest:
				text = usageText
			case http.StatusNotFound:
				text = s.adminUserNotFoundText(l)
			}
		}
		return text, err
	}

	if user.IsModerator || s.isAdminUsername(user.Username) {
		return s.adminModerationRestrictedText(l), nil
	}

	return "", nil
}

func (s *service) ensureModeratorCanResetStart(ctx context.Context, role adminRole, userID int64, username string, l localizer) (string, error) {
	if role != adminRoleModerator {
		return "", nil
	}

	if username != "" && s.isAdminUsername(username) {
		return s.adminModerationRestrictedText(l), nil
	}

	if s == nil || s.admin == nil {
		return s.adminResetStartFailedText(l), errors.New("admin service is not configured")
	}

	var (
		user domain.UserSummary
		err  error
	)
	if username != "" {
		user, err = s.admin.GetUserByUsername(ctx, username)
	} else {
		user, err = s.admin.GetUser(ctx, userID)
	}
	if err != nil {
		var status statusError
		if errors.As(err, &status) {
			switch status.StatusCode() {
			case http.StatusBadRequest:
				return s.adminResetStartUsageText(l), err
			case http.StatusNotFound:
				return "", nil
			}
		}
		return s.adminResetStartFailedText(l), err
	}

	if user.IsModerator || s.isAdminUsername(user.Username) {
		return s.adminModerationRestrictedText(l), nil
	}

	return "", nil
}
