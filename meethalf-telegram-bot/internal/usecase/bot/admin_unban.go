package bot

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) adminUnbanMessage(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	if !s.isAdminUser(msg.User) {
		return s.adminAccessDeniedMessage(ctx, msg)
	}

	if s == nil || s.admin == nil {
		return s.adminUnbanFailedText(), s.adminMenuInlineKeyboard(), errors.New("admin service is not configured")
	}

	if strings.TrimSpace(msg.Arguments) == "" {
		return s.startAdminUnban(ctx, msg)
	}

	userID, username, ok := s.parseAdminUserIdentifier(msg.Arguments)
	if !ok {
		return s.adminUnbanUsageText(), s.adminUnbanInlineKeyboard(), nil
	}

	shouldClear := s.hasPendingAdminUnban(ctx, msg.User.ID)
	text, err := s.performAdminUnban(ctx, userID, username)
	if err == nil && shouldClear {
		_ = s.clearAdminAction(ctx, msg.User.ID)
	}

	return text, s.adminMenuInlineKeyboard(), err
}

func (s *service) startAdminUnban(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	if s == nil || s.adminActions == nil {
		return s.adminUnbanUsageText(), s.adminMenuInlineKeyboard(), errors.New("admin action repository is not configured")
	}

	action := domain.AdminActionState{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		Action:      domain.AdminActionUnban,
		RequestedAt: s.now(msg.ReceivedAt),
	}
	if err := s.adminActions.Save(ctx, action); err != nil {
		return s.adminUnbanFailedText(), s.adminMenuInlineKeyboard(), err
	}

	return s.adminUnbanUsageText(), s.adminUnbanInlineKeyboard(), nil
}

func (s *service) performAdminUnban(ctx context.Context, userID int64, username string) (string, error) {
	if s == nil || s.admin == nil {
		return s.adminUnbanFailedText(), errors.New("admin service is not configured")
	}

	var err error
	if username != "" {
		err = s.admin.UnbanUserByUsername(ctx, username)
	} else {
		err = s.admin.UnbanUser(ctx, userID)
	}
	if err != nil {
		text := s.adminUnbanFailedText()
		var status statusError
		if errors.As(err, &status) {
			switch status.StatusCode() {
			case http.StatusBadRequest:
				text = s.adminUnbanUsageText()
			case http.StatusNotFound:
				text = s.adminUserNotFoundText()
			}
		}
		return text, err
	}

	return s.adminUnbanSuccessText(s.adminUserIdentifierLabel(userID, username)), nil
}

func (s *service) hasPendingAdminUnban(ctx context.Context, userID int64) bool {
	if s == nil || s.adminActions == nil || userID == 0 {
		return false
	}

	action, found, err := s.adminActions.Get(ctx, userID)
	if err != nil || !found {
		return false
	}

	return action.Action == domain.AdminActionUnban
}
