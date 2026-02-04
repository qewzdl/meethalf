package bot

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) adminViewProfileMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if !role.canModerateUsers() {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}

	if s == nil || s.admin == nil {
		return s.adminViewProfileFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin service is not configured")
	}

	if strings.TrimSpace(msg.Arguments) == "" {
		return s.startAdminViewProfile(ctx, msg, role, l)
	}

	userID, username, ok := s.parseAdminUserIdentifier(msg.Arguments)
	if !ok {
		return s.adminViewProfileUsageText(l), s.adminViewProfileInlineKeyboard(l), nil
	}

	restrictionText, restrictionErr := s.ensureModeratorCanModerateUser(ctx, role, userID, username, s.adminViewProfileFailedText(l), s.adminViewProfileUsageText(l), l)
	if restrictionText != "" {
		return restrictionText, s.adminMenuInlineKeyboard(l, role), restrictionErr
	}

	shouldClear := s.hasPendingAdminViewProfile(ctx, msg.User.ID)
	profile, text, err := s.performAdminViewProfile(ctx, userID, username, l)
	if err != nil {
		return text, s.adminMenuInlineKeyboard(l, role), err
	}
	if shouldClear {
		_ = s.clearAdminAction(ctx, msg.User.ID)
		s.registerAdminActionCleanup(msg)
	}

	return s.adminProfileDetails(l, profile), s.adminMenuInlineKeyboard(l, role), nil
}

func (s *service) startAdminViewProfile(ctx context.Context, msg domain.IncomingMessage, role adminRole, l localizer) (string, *domain.InlineKeyboard, error) {
	if s == nil || s.adminActions == nil {
		return s.adminViewProfileUsageText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin action repository is not configured")
	}

	action := domain.AdminActionState{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		Action:      domain.AdminActionViewProfile,
		RequestedAt: s.now(msg.ReceivedAt),
	}
	if err := s.adminActions.Save(ctx, action); err != nil {
		return s.adminViewProfileFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}

	return s.adminViewProfileUsageText(l), s.adminViewProfileInlineKeyboard(l), nil
}

func (s *service) performAdminViewProfile(ctx context.Context, userID int64, username string, l localizer) (domain.Profile, string, error) {
	if s == nil || s.admin == nil {
		return domain.Profile{}, s.adminViewProfileFailedText(l), errors.New("admin service is not configured")
	}

	var (
		profile domain.Profile
		err     error
	)
	if username != "" {
		profile, err = s.admin.GetProfileByUsername(ctx, username)
	} else {
		profile, err = s.admin.GetProfile(ctx, userID)
	}
	if err != nil {
		text := s.adminViewProfileFailedText(l)
		var status statusError
		if errors.As(err, &status) {
			switch status.StatusCode() {
			case http.StatusBadRequest:
				text = s.adminViewProfileUsageText(l)
			case http.StatusNotFound:
				text = s.adminUserNotFoundText(l)
			}
		}
		return domain.Profile{}, text, err
	}

	return profile, "", nil
}

func (s *service) hasPendingAdminViewProfile(ctx context.Context, userID int64) bool {
	if s == nil || s.adminActions == nil || userID == 0 {
		return false
	}

	action, found, err := s.adminActions.Get(ctx, userID)
	if err != nil || !found {
		return false
	}

	return action.Action == domain.AdminActionViewProfile
}
