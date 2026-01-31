package bot

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) adminModeratorMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if !role.canManageModerators() {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}

	if s == nil || s.admin == nil {
		return s.adminModeratorFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin service is not configured")
	}

	if strings.TrimSpace(msg.Arguments) == "" {
		return s.startAdminModerator(ctx, msg, role, l)
	}

	userID, username, ok := s.parseAdminUserIdentifier(msg.Arguments)
	if !ok {
		return s.adminModeratorUsageText(l), s.adminModeratorInlineKeyboard(l), nil
	}

	shouldClear := s.hasPendingAdminModerator(ctx, msg.User.ID)
	text, err := s.performAdminModerator(ctx, userID, username, l)
	if err == nil && shouldClear {
		_ = s.clearAdminAction(ctx, msg.User.ID)
		s.registerAdminActionCleanup(msg)
	}

	return text, s.adminMenuInlineKeyboard(l, role), err
}

func (s *service) startAdminModerator(ctx context.Context, msg domain.IncomingMessage, role adminRole, l localizer) (string, *domain.InlineKeyboard, error) {
	if s == nil || s.adminActions == nil {
		return s.adminModeratorUsageText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin action repository is not configured")
	}

	action := domain.AdminActionState{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		Action:      domain.AdminActionModerator,
		RequestedAt: s.now(msg.ReceivedAt),
	}
	if err := s.adminActions.Save(ctx, action); err != nil {
		return s.adminModeratorFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}

	return s.adminModeratorUsageText(l), s.adminModeratorInlineKeyboard(l), nil
}

func (s *service) performAdminModerator(ctx context.Context, userID int64, username string, l localizer) (string, error) {
	if s == nil || s.admin == nil {
		return s.adminModeratorFailedText(l), errors.New("admin service is not configured")
	}

	var err error
	if username != "" {
		err = s.admin.MakeModeratorByUsername(ctx, username)
	} else {
		err = s.admin.MakeModerator(ctx, userID)
	}
	if err != nil {
		text := s.adminModeratorFailedText(l)
		var status statusError
		if errors.As(err, &status) {
			switch status.StatusCode() {
			case http.StatusBadRequest:
				text = s.adminModeratorUsageText(l)
			case http.StatusNotFound:
				text = s.adminUserNotFoundText(l)
			}
		}
		return text, err
	}

	return s.adminModeratorSuccessText(l, s.adminUserIdentifierLabel(userID, username)), nil
}

func (s *service) hasPendingAdminModerator(ctx context.Context, userID int64) bool {
	if s == nil || s.adminActions == nil || userID == 0 {
		return false
	}

	action, found, err := s.adminActions.Get(ctx, userID)
	if err != nil || !found {
		return false
	}

	return action.Action == domain.AdminActionModerator
}
