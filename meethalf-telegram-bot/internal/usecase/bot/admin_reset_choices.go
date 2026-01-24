package bot

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) adminResetChoicesMessage(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(), nil, roleErr
	}
	if !role.canModerateUsers() {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg)
		return text, keyboard, errors.Join(roleErr, err)
	}

	if s == nil || s.admin == nil {
		return s.adminResetChoicesFailedText(), s.adminMenuInlineKeyboard(role), errors.New("admin service is not configured")
	}

	if strings.TrimSpace(msg.Arguments) == "" {
		return s.startAdminResetChoices(ctx, msg, role)
	}

	userID, username, ok := s.parseAdminUserIdentifier(msg.Arguments)
	if !ok {
		return s.adminResetChoicesUsageText(), s.adminResetChoicesInlineKeyboard(), nil
	}

	restrictionText, restrictionErr := s.ensureModeratorCanModerateUser(ctx, role, userID, username, s.adminResetChoicesFailedText(), s.adminResetChoicesUsageText())
	if restrictionText != "" {
		return restrictionText, s.adminMenuInlineKeyboard(role), restrictionErr
	}

	shouldClear := s.hasPendingAdminResetChoices(ctx, msg.User.ID)
	text, err := s.performAdminResetChoices(ctx, userID, username)
	if err == nil && shouldClear {
		_ = s.clearAdminAction(ctx, msg.User.ID)
	}

	return text, s.adminMenuInlineKeyboard(role), err
}

func (s *service) startAdminResetChoices(ctx context.Context, msg domain.IncomingMessage, role adminRole) (string, *domain.InlineKeyboard, error) {
	if s == nil || s.adminActions == nil {
		return s.adminResetChoicesUsageText(), s.adminMenuInlineKeyboard(role), errors.New("admin action repository is not configured")
	}

	action := domain.AdminActionState{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		Action:      domain.AdminActionResetChoices,
		RequestedAt: s.now(msg.ReceivedAt),
	}
	if err := s.adminActions.Save(ctx, action); err != nil {
		return s.adminResetChoicesFailedText(), s.adminMenuInlineKeyboard(role), err
	}

	return s.adminResetChoicesUsageText(), s.adminResetChoicesInlineKeyboard(), nil
}

func (s *service) performAdminResetChoices(ctx context.Context, userID int64, username string) (string, error) {
	if s == nil || s.admin == nil {
		return s.adminResetChoicesFailedText(), errors.New("admin service is not configured")
	}

	var err error
	if username != "" {
		err = s.admin.ResetUserChoicesByUsername(ctx, username)
	} else {
		err = s.admin.ResetUserChoices(ctx, userID)
	}
	if err != nil {
		text := s.adminResetChoicesFailedText()
		var status statusError
		if errors.As(err, &status) {
			switch status.StatusCode() {
			case http.StatusBadRequest:
				text = s.adminResetChoicesUsageText()
			case http.StatusNotFound:
				text = s.adminUserNotFoundText()
			}
		}
		return text, err
	}

	return s.adminResetChoicesSuccessText(s.adminUserIdentifierLabel(userID, username)), nil
}

func (s *service) hasPendingAdminResetChoices(ctx context.Context, userID int64) bool {
	if s == nil || s.adminActions == nil || userID == 0 {
		return false
	}

	action, found, err := s.adminActions.Get(ctx, userID)
	if err != nil || !found {
		return false
	}

	return action.Action == domain.AdminActionResetChoices
}
