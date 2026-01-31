package bot

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) adminResetStartMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if !role.canModerateUsers() {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}

	if strings.TrimSpace(msg.Arguments) == "" {
		return s.startAdminResetStart(ctx, msg, role, l)
	}

	userID, username, ok := s.parseAdminUserIdentifier(msg.Arguments)
	if !ok {
		return s.adminResetStartUsageText(l), s.adminResetStartInlineKeyboard(l), nil
	}

	restrictionText, restrictionErr := s.ensureModeratorCanResetStart(ctx, role, userID, username, l)
	if restrictionText != "" {
		return restrictionText, s.adminMenuInlineKeyboard(l, role), restrictionErr
	}

	shouldClear := s.hasPendingAdminResetStart(ctx, msg.User.ID)
	text, err := s.performAdminResetStart(ctx, userID, username, l)
	if err == nil && shouldClear {
		_ = s.clearAdminAction(ctx, msg.User.ID)
		s.registerAdminActionCleanup(msg)
	}

	return text, s.adminMenuInlineKeyboard(l, role), err
}

func (s *service) startAdminResetStart(ctx context.Context, msg domain.IncomingMessage, role adminRole, l localizer) (string, *domain.InlineKeyboard, error) {
	if s == nil || s.adminActions == nil {
		return s.adminResetStartUsageText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin action repository is not configured")
	}

	action := domain.AdminActionState{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		Action:      domain.AdminActionResetStart,
		RequestedAt: s.now(msg.ReceivedAt),
	}
	if err := s.adminActions.Save(ctx, action); err != nil {
		return s.adminResetStartFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}

	return s.adminResetStartUsageText(l), s.adminResetStartInlineKeyboard(l), nil
}

func (s *service) performAdminResetStart(ctx context.Context, userID int64, username string, l localizer) (string, error) {
	if s == nil || s.ageConfirmations == nil {
		return s.adminResetStartFailedText(l), errors.New("age confirmation repository is not configured")
	}

	targetID := userID
	normalizedUsername := normalizeUsername(username)
	if normalizedUsername != "" {
		resolvedID, found, err := s.ageConfirmations.FindUserIDByUsername(ctx, normalizedUsername)
		if err != nil {
			return s.adminResetStartFailedText(l), err
		}
		if found {
			targetID = resolvedID
		} else {
			if s.admin == nil {
				return s.adminResetStartFailedText(l), errors.New("admin service is not configured")
			}

			user, err := s.admin.GetUserByUsername(ctx, normalizedUsername)
			if err != nil {
				text := s.adminResetStartFailedText(l)
				var status statusError
				if errors.As(err, &status) {
					switch status.StatusCode() {
					case http.StatusBadRequest:
						text = s.adminResetStartUsageText(l)
					case http.StatusNotFound:
						text = s.adminUserNotFoundText(l)
					}
				}
				return text, err
			}
			targetID = user.UserID
		}
	}

	if targetID <= 0 {
		return s.adminResetStartUsageText(l), nil
	}

	if err := s.ageConfirmations.Delete(ctx, targetID); err != nil {
		return s.adminResetStartFailedText(l), err
	}

	userRef := username
	if normalizedUsername != "" {
		userRef = normalizedUsername
	}

	return s.adminResetStartSuccessText(l, s.adminUserIdentifierLabel(targetID, userRef)), nil
}

func (s *service) hasPendingAdminResetStart(ctx context.Context, userID int64) bool {
	if s == nil || s.adminActions == nil || userID == 0 {
		return false
	}

	action, found, err := s.adminActions.Get(ctx, userID)
	if err != nil || !found {
		return false
	}

	return action.Action == domain.AdminActionResetStart
}
