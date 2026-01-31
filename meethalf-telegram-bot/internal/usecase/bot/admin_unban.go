package bot

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) adminUnbanMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if !role.canModerateUsers() {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}

	if s == nil || s.admin == nil {
		return s.adminUnbanFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin service is not configured")
	}

	if strings.TrimSpace(msg.Arguments) == "" {
		return s.startAdminUnban(ctx, msg, role, l)
	}

	userID, username, ok := s.parseAdminUserIdentifier(msg.Arguments)
	if !ok {
		return s.adminUnbanUsageText(l), s.adminUnbanInlineKeyboard(l), nil
	}

	restrictionText, restrictionErr := s.ensureModeratorCanModerateUser(ctx, role, userID, username, s.adminUnbanFailedText(l), s.adminUnbanUsageText(l), l)
	if restrictionText != "" {
		return restrictionText, s.adminMenuInlineKeyboard(l, role), restrictionErr
	}

	shouldClear := s.hasPendingAdminUnban(ctx, msg.User.ID)
	text, err := s.performAdminUnban(ctx, userID, username, l)
	if err == nil && shouldClear {
		_ = s.clearAdminAction(ctx, msg.User.ID)
		s.registerAdminActionCleanup(msg)
	}

	return text, s.adminMenuInlineKeyboard(l, role), err
}

func (s *service) startAdminUnban(ctx context.Context, msg domain.IncomingMessage, role adminRole, l localizer) (string, *domain.InlineKeyboard, error) {
	if s == nil || s.adminActions == nil {
		return s.adminUnbanUsageText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin action repository is not configured")
	}

	action := domain.AdminActionState{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		Action:      domain.AdminActionUnban,
		RequestedAt: s.now(msg.ReceivedAt),
	}
	if err := s.adminActions.Save(ctx, action); err != nil {
		return s.adminUnbanFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}

	return s.adminUnbanUsageText(l), s.adminUnbanInlineKeyboard(l), nil
}

func (s *service) performAdminUnban(ctx context.Context, userID int64, username string, l localizer) (string, error) {
	if s == nil || s.admin == nil {
		return s.adminUnbanFailedText(l), errors.New("admin service is not configured")
	}

	var err error
	if username != "" {
		err = s.admin.UnbanUserByUsername(ctx, username)
	} else {
		err = s.admin.UnbanUser(ctx, userID)
	}
	if err != nil {
		text := s.adminUnbanFailedText(l)
		var status statusError
		if errors.As(err, &status) {
			switch status.StatusCode() {
			case http.StatusBadRequest:
				text = s.adminUnbanUsageText(l)
			case http.StatusNotFound:
				text = s.adminUserNotFoundText(l)
			}
		}
		return text, err
	}

	return s.adminUnbanSuccessText(l, s.adminUserIdentifierLabel(userID, username)), nil
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
