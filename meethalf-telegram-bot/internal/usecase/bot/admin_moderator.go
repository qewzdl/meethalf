package bot

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) adminModeratorMessage(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	if !s.isAdminUser(msg.User) {
		return s.adminAccessDeniedMessage(ctx, msg)
	}

	if s == nil || s.admin == nil {
		return s.adminModeratorFailedText(), s.adminMenuInlineKeyboard(), errors.New("admin service is not configured")
	}

	if strings.TrimSpace(msg.Arguments) == "" {
		return s.startAdminModerator(ctx, msg)
	}

	userID, username, ok := s.parseAdminUserIdentifier(msg.Arguments)
	if !ok {
		return s.adminModeratorUsageText(), s.adminModeratorInlineKeyboard(), nil
	}

	shouldClear := s.hasPendingAdminModerator(ctx, msg.User.ID)
	text, err := s.performAdminModerator(ctx, userID, username)
	if err == nil && shouldClear {
		_ = s.clearAdminAction(ctx, msg.User.ID)
	}

	return text, s.adminMenuInlineKeyboard(), err
}

func (s *service) startAdminModerator(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	if s == nil || s.adminActions == nil {
		return s.adminModeratorUsageText(), s.adminMenuInlineKeyboard(), errors.New("admin action repository is not configured")
	}

	action := domain.AdminActionState{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		Action:      domain.AdminActionModerator,
		RequestedAt: s.now(msg.ReceivedAt),
	}
	if err := s.adminActions.Save(ctx, action); err != nil {
		return s.adminModeratorFailedText(), s.adminMenuInlineKeyboard(), err
	}

	return s.adminModeratorUsageText(), s.adminModeratorInlineKeyboard(), nil
}

func (s *service) performAdminModerator(ctx context.Context, userID int64, username string) (string, error) {
	if s == nil || s.admin == nil {
		return s.adminModeratorFailedText(), errors.New("admin service is not configured")
	}

	var err error
	if username != "" {
		err = s.admin.MakeModeratorByUsername(ctx, username)
	} else {
		err = s.admin.MakeModerator(ctx, userID)
	}
	if err != nil {
		text := s.adminModeratorFailedText()
		var status statusError
		if errors.As(err, &status) {
			switch status.StatusCode() {
			case http.StatusBadRequest:
				text = s.adminModeratorUsageText()
			case http.StatusNotFound:
				text = s.adminUserNotFoundText()
			}
		}
		return text, err
	}

	return s.adminModeratorSuccessText(s.adminUserIdentifierLabel(userID, username)), nil
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
