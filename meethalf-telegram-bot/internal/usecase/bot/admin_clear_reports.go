package bot

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) adminClearReportsMessage(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(), nil, roleErr
	}
	if !role.canModerateUsers() {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg)
		return text, keyboard, errors.Join(roleErr, err)
	}

	if s == nil || s.admin == nil {
		return s.adminClearReportsFailedText(), s.adminMenuInlineKeyboard(role), errors.New("admin service is not configured")
	}

	if strings.TrimSpace(msg.Arguments) == "" {
		return s.startAdminClearReports(ctx, msg, role)
	}

	userID, username, ok := s.parseAdminUserIdentifier(msg.Arguments)
	if !ok {
		return s.adminClearReportsUsageText(), s.adminClearReportsInlineKeyboard(), nil
	}

	restrictionText, restrictionErr := s.ensureModeratorCanModerateUser(ctx, role, userID, username, s.adminClearReportsFailedText(), s.adminClearReportsUsageText())
	if restrictionText != "" {
		return restrictionText, s.adminMenuInlineKeyboard(role), restrictionErr
	}

	shouldClear := s.hasPendingAdminClearReports(ctx, msg.User.ID)
	text, err := s.performAdminClearReports(ctx, userID, username)
	if err == nil && shouldClear {
		_ = s.clearAdminAction(ctx, msg.User.ID)
	}

	return text, s.adminMenuInlineKeyboard(role), err
}

func (s *service) startAdminClearReports(ctx context.Context, msg domain.IncomingMessage, role adminRole) (string, *domain.InlineKeyboard, error) {
	if s == nil || s.adminActions == nil {
		return s.adminClearReportsUsageText(), s.adminMenuInlineKeyboard(role), errors.New("admin action repository is not configured")
	}

	action := domain.AdminActionState{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		Action:      domain.AdminActionClearReports,
		RequestedAt: s.now(msg.ReceivedAt),
	}
	if err := s.adminActions.Save(ctx, action); err != nil {
		return s.adminClearReportsFailedText(), s.adminMenuInlineKeyboard(role), err
	}

	return s.adminClearReportsUsageText(), s.adminClearReportsInlineKeyboard(), nil
}

func (s *service) performAdminClearReports(ctx context.Context, userID int64, username string) (string, error) {
	if s == nil || s.admin == nil {
		return s.adminClearReportsFailedText(), errors.New("admin service is not configured")
	}

	var err error
	if username != "" {
		err = s.admin.ClearUserReportsByUsername(ctx, username)
	} else {
		err = s.admin.ClearUserReports(ctx, userID)
	}
	if err != nil {
		text := s.adminClearReportsFailedText()
		var status statusError
		if errors.As(err, &status) {
			switch status.StatusCode() {
			case http.StatusBadRequest:
				text = s.adminClearReportsUsageText()
			case http.StatusNotFound:
				text = s.adminUserNotFoundText()
			}
		}
		return text, err
	}

	return s.adminClearReportsSuccessText(s.adminUserIdentifierLabel(userID, username)), nil
}

func (s *service) hasPendingAdminClearReports(ctx context.Context, userID int64) bool {
	if s == nil || s.adminActions == nil || userID == 0 {
		return false
	}

	action, found, err := s.adminActions.Get(ctx, userID)
	if err != nil || !found {
		return false
	}

	return action.Action == domain.AdminActionClearReports
}
