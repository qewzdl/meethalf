package bot

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) adminShowProfileMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if !role.canModerateUsers() {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}

	if s == nil || s.admin == nil {
		return s.adminShowProfileFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin service is not configured")
	}

	if strings.TrimSpace(msg.Arguments) == "" {
		return s.startAdminShowProfile(ctx, msg, role, l)
	}

	userID, username, ok := s.parseAdminUserIdentifier(msg.Arguments)
	if !ok {
		return s.adminShowProfileUsageText(l), s.adminShowProfileInlineKeyboard(l), nil
	}

	restrictionText, restrictionErr := s.ensureModeratorCanModerateUser(ctx, role, userID, username, s.adminShowProfileFailedText(l), s.adminShowProfileUsageText(l), l)
	if restrictionText != "" {
		return restrictionText, s.adminMenuInlineKeyboard(l, role), restrictionErr
	}

	shouldClear := s.hasPendingAdminShowProfile(ctx, msg.User.ID)
	text, err := s.performAdminShowProfile(ctx, userID, username, l)
	if err == nil && shouldClear {
		_ = s.clearAdminAction(ctx, msg.User.ID)
	}

	return text, s.adminMenuInlineKeyboard(l, role), err
}

func (s *service) startAdminShowProfile(ctx context.Context, msg domain.IncomingMessage, role adminRole, l localizer) (string, *domain.InlineKeyboard, error) {
	if s == nil || s.adminActions == nil {
		return s.adminShowProfileUsageText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin action repository is not configured")
	}

	action := domain.AdminActionState{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		Action:      domain.AdminActionShowProfile,
		RequestedAt: s.now(msg.ReceivedAt),
	}
	if err := s.adminActions.Save(ctx, action); err != nil {
		return s.adminShowProfileFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}

	return s.adminShowProfileUsageText(l), s.adminShowProfileInlineKeyboard(l), nil
}

func (s *service) performAdminShowProfile(ctx context.Context, userID int64, username string, l localizer) (string, error) {
	if s == nil || s.admin == nil {
		return s.adminShowProfileFailedText(l), errors.New("admin service is not configured")
	}

	var err error
	if username != "" {
		err = s.admin.ShowUserByUsername(ctx, username)
	} else {
		err = s.admin.ShowUser(ctx, userID)
	}
	if err != nil {
		text := s.adminShowProfileFailedText(l)
		var status statusError
		if errors.As(err, &status) {
			switch status.StatusCode() {
			case http.StatusBadRequest:
				text = s.adminShowProfileUsageText(l)
			case http.StatusNotFound:
				text = s.adminUserNotFoundText(l)
			}
		}
		return text, err
	}

	return s.adminShowProfileSuccessText(l, s.adminUserIdentifierLabel(userID, username)), nil
}

func (s *service) hasPendingAdminShowProfile(ctx context.Context, userID int64) bool {
	if s == nil || s.adminActions == nil || userID == 0 {
		return false
	}

	action, found, err := s.adminActions.Get(ctx, userID)
	if err != nil || !found {
		return false
	}

	return action.Action == domain.AdminActionShowProfile
}
