package bot

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) adminUnmoderatorMessage(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(), nil, roleErr
	}
	if !role.canManageModerators() {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg)
		return text, keyboard, errors.Join(roleErr, err)
	}

	if s == nil || s.admin == nil {
		return s.adminUnmoderatorFailedText(), s.adminMenuInlineKeyboard(role), errors.New("admin service is not configured")
	}

	if strings.TrimSpace(msg.Arguments) == "" {
		return s.startAdminUnmoderator(ctx, msg, role)
	}

	userID, username, ok := s.parseAdminUserIdentifier(msg.Arguments)
	if !ok {
		return s.adminUnmoderatorUsageText(), s.adminUnmoderatorInlineKeyboard(), nil
	}

	shouldClear := s.hasPendingAdminUnmoderator(ctx, msg.User.ID)
	text, err := s.performAdminUnmoderator(ctx, userID, username)
	if err == nil && shouldClear {
		_ = s.clearAdminAction(ctx, msg.User.ID)
	}

	return text, s.adminMenuInlineKeyboard(role), err
}

func (s *service) startAdminUnmoderator(ctx context.Context, msg domain.IncomingMessage, role adminRole) (string, *domain.InlineKeyboard, error) {
	if s == nil || s.adminActions == nil {
		return s.adminUnmoderatorUsageText(), s.adminMenuInlineKeyboard(role), errors.New("admin action repository is not configured")
	}

	action := domain.AdminActionState{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		Action:      domain.AdminActionUnmoderator,
		RequestedAt: s.now(msg.ReceivedAt),
	}
	if err := s.adminActions.Save(ctx, action); err != nil {
		return s.adminUnmoderatorFailedText(), s.adminMenuInlineKeyboard(role), err
	}

	return s.adminUnmoderatorUsageText(), s.adminUnmoderatorInlineKeyboard(), nil
}

func (s *service) performAdminUnmoderator(ctx context.Context, userID int64, username string) (string, error) {
	if s == nil || s.admin == nil {
		return s.adminUnmoderatorFailedText(), errors.New("admin service is not configured")
	}

	var err error
	if username != "" {
		err = s.admin.RemoveModeratorByUsername(ctx, username)
	} else {
		err = s.admin.RemoveModerator(ctx, userID)
	}
	if err != nil {
		text := s.adminUnmoderatorFailedText()
		var status statusError
		if errors.As(err, &status) {
			switch status.StatusCode() {
			case http.StatusBadRequest:
				text = s.adminUnmoderatorUsageText()
			case http.StatusNotFound:
				text = s.adminUserNotFoundText()
			}
		}
		return text, err
	}

	return s.adminUnmoderatorSuccessText(s.adminUserIdentifierLabel(userID, username)), nil
}

func (s *service) hasPendingAdminUnmoderator(ctx context.Context, userID int64) bool {
	if s == nil || s.adminActions == nil || userID == 0 {
		return false
	}

	action, found, err := s.adminActions.Get(ctx, userID)
	if err != nil || !found {
		return false
	}

	return action.Action == domain.AdminActionUnmoderator
}
