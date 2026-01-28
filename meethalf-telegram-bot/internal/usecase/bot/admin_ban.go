package bot

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) adminBanMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if !role.canModerateUsers() {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}

	if s == nil || s.admin == nil {
		return s.adminBanFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin service is not configured")
	}

	if strings.TrimSpace(msg.Arguments) == "" {
		return s.startAdminBan(ctx, msg, role, l)
	}

	userID, username, ok := s.parseAdminUserIdentifier(msg.Arguments)
	if !ok {
		return s.adminBanUsageText(l), s.adminBanInlineKeyboard(l), nil
	}

	restrictionText, restrictionErr := s.ensureModeratorCanModerateUser(ctx, role, userID, username, s.adminBanFailedText(l), s.adminBanUsageText(l), l)
	if restrictionText != "" {
		return restrictionText, s.adminMenuInlineKeyboard(l, role), restrictionErr
	}

	shouldClear := s.hasPendingAdminBan(ctx, msg.User.ID)
	text, err := s.performAdminBan(ctx, userID, username, l)
	if err == nil && shouldClear {
		_ = s.clearAdminAction(ctx, msg.User.ID)
	}

	return text, s.adminMenuInlineKeyboard(l, role), err
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

func (s *service) startAdminBan(ctx context.Context, msg domain.IncomingMessage, role adminRole, l localizer) (string, *domain.InlineKeyboard, error) {
	if s == nil || s.adminActions == nil {
		return s.adminBanUsageText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin action repository is not configured")
	}

	action := domain.AdminActionState{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		Action:      domain.AdminActionBan,
		RequestedAt: s.now(msg.ReceivedAt),
	}
	if err := s.adminActions.Save(ctx, action); err != nil {
		return s.adminBanFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}

	return s.adminBanUsageText(l), s.adminBanInlineKeyboard(l), nil
}

func (s *service) performAdminBan(ctx context.Context, userID int64, username string, l localizer) (string, error) {
	if s == nil || s.admin == nil {
		return s.adminBanFailedText(l), errors.New("admin service is not configured")
	}

	var err error
	if username != "" {
		err = s.admin.BanUserByUsername(ctx, username)
	} else {
		err = s.admin.BanUser(ctx, userID)
	}
	if err != nil {
		text := s.adminBanFailedText(l)
		var status statusError
		if errors.As(err, &status) {
			switch status.StatusCode() {
			case http.StatusBadRequest:
				text = s.adminBanUsageText(l)
			case http.StatusNotFound:
				text = s.adminUserNotFoundText(l)
			}
		}
		return text, err
	}

	return s.adminBanSuccessText(l, s.adminUserIdentifierLabel(userID, username)), nil
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
