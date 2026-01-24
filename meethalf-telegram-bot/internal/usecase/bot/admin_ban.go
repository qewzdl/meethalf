package bot

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) adminBanMessage(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	if !s.isAdminUser(msg.User) {
		return s.adminAccessDeniedMessage(ctx, msg)
	}

	if s == nil || s.admin == nil {
		return s.adminBanFailedText(), s.adminMenuInlineKeyboard(), errors.New("admin service is not configured")
	}

	if strings.TrimSpace(msg.Arguments) == "" {
		return s.startAdminBan(ctx, msg)
	}

	userID, username, ok := s.parseAdminUserIdentifier(msg.Arguments)
	if !ok {
		return s.adminBanUsageText(), s.adminBanInlineKeyboard(), nil
	}

	shouldClear := s.hasPendingAdminBan(ctx, msg.User.ID)
	text, err := s.performAdminBan(ctx, userID, username)
	if err == nil && shouldClear {
		_ = s.clearAdminAction(ctx, msg.User.ID)
	}

	return text, s.adminMenuInlineKeyboard(), err
}

func (s *service) parseAdminUserIdentifier(value string) (int64, string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, "", false
	}

	parts := strings.Fields(value)
	if len(parts) == 0 {
		return 0, "", false
	}

	token := strings.TrimSpace(parts[0])
	if strings.HasPrefix(token, "@") {
		username := strings.TrimSpace(strings.TrimPrefix(token, "@"))
		if username == "" {
			return 0, "", false
		}
		return 0, username, true
	}

	userID, err := strconv.ParseInt(token, 10, 64)
	if err == nil && userID > 0 {
		return userID, "", true
	}

	if token == "" {
		return 0, "", false
	}

	return 0, token, true
}

func (s *service) startAdminBan(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	if s == nil || s.adminActions == nil {
		return s.adminBanUsageText(), s.adminMenuInlineKeyboard(), errors.New("admin action repository is not configured")
	}

	action := domain.AdminActionState{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		Action:      domain.AdminActionBan,
		RequestedAt: s.now(msg.ReceivedAt),
	}
	if err := s.adminActions.Save(ctx, action); err != nil {
		return s.adminBanFailedText(), s.adminMenuInlineKeyboard(), err
	}

	return s.adminBanUsageText(), s.adminBanInlineKeyboard(), nil
}

func (s *service) performAdminBan(ctx context.Context, userID int64, username string) (string, error) {
	if s == nil || s.admin == nil {
		return s.adminBanFailedText(), errors.New("admin service is not configured")
	}

	var err error
	if username != "" {
		err = s.admin.BanUserByUsername(ctx, username)
	} else {
		err = s.admin.BanUser(ctx, userID)
	}
	if err != nil {
		text := s.adminBanFailedText()
		var status statusError
		if errors.As(err, &status) {
			switch status.StatusCode() {
			case http.StatusBadRequest:
				text = s.adminBanUsageText()
			case http.StatusNotFound:
				text = s.adminUserNotFoundText()
			}
		}
		return text, err
	}

	return s.adminBanSuccessText(s.adminUserIdentifierLabel(userID, username)), nil
}

func (s *service) hasPendingAdminBan(ctx context.Context, userID int64) bool {
	if s == nil || s.adminActions == nil || userID == 0 {
		return false
	}

	action, found, err := s.adminActions.Get(ctx, userID)
	if err != nil || !found {
		return false
	}

	return action.Action == domain.AdminActionBan
}
