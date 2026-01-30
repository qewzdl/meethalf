package bot

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

const adminUsersPageSize = 20

func (s *service) adminMenuMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if !role.canAccessPanel() {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}

	if s != nil && s.adminActions != nil && msg.User.ID != 0 {
		_ = s.adminActions.Delete(ctx, msg.User.ID)
	}

	return s.adminMenuTextForRole(l, role), s.adminMenuInlineKeyboard(l, role), roleErr
}

func (s *service) adminUsersMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if !role.canModerateUsers() {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}

	if s == nil || s.admin == nil {
		return s.adminUsersLoadFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin service is not configured")
	}

	offset := s.parseAdminUsersOffset(msg.Arguments)
	list, err := s.admin.ListUsers(ctx, adminUsersPageSize, offset, false, false, false)
	if err != nil {
		return s.adminUsersLoadFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}

	text := s.adminUsersText(l, list)
	keyboard := s.adminUsersInlineKeyboard(l, list.Offset, list.Limit, list.Total)
	return text, keyboard, nil
}

func (s *service) adminBannedUsersMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if !role.canModerateUsers() {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}

	if s == nil || s.admin == nil {
		return s.adminBannedUsersLoadFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin service is not configured")
	}

	offset := s.parseAdminUsersOffset(msg.Arguments)
	list, err := s.admin.ListUsers(ctx, adminUsersPageSize, offset, true, false, false)
	if err != nil {
		return s.adminBannedUsersLoadFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}

	text := s.adminBannedUsersText(l, list)
	keyboard := s.adminBannedUsersInlineKeyboard(l, list.Offset, list.Limit, list.Total)
	return text, keyboard, nil
}

func (s *service) adminModeratorsMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if !role.canManageModerators() {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}

	if s == nil || s.admin == nil {
		return s.adminModeratorsLoadFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin service is not configured")
	}

	offset := s.parseAdminUsersOffset(msg.Arguments)
	list, err := s.admin.ListUsers(ctx, adminUsersPageSize, offset, false, true, false)
	if err != nil {
		return s.adminModeratorsLoadFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}

	text := s.adminModeratorsText(l, list)
	keyboard := s.adminModeratorsInlineKeyboard(l, list.Offset, list.Limit, list.Total)
	return text, keyboard, nil
}

func (s *service) adminReportsMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if !role.canModerateUsers() {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}

	if s == nil || s.admin == nil {
		return s.adminReportsLoadFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin service is not configured")
	}

	offset := s.parseAdminUsersOffset(msg.Arguments)
	list, err := s.admin.ListReportedUsers(ctx, adminUsersPageSize, offset)
	if err != nil {
		return s.adminReportsLoadFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}

	text := s.adminReportsText(l, list)
	keyboard := s.adminReportsInlineKeyboard(l, list.Offset, list.Limit, list.Total)
	return text, keyboard, nil
}

func (s *service) adminHiddenUsersMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if !role.canModerateUsers() {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}

	if s == nil || s.admin == nil {
		return s.adminHiddenUsersLoadFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin service is not configured")
	}

	offset := s.parseAdminUsersOffset(msg.Arguments)
	list, err := s.admin.ListUsers(ctx, adminUsersPageSize, offset, false, false, true)
	if err != nil {
		return s.adminHiddenUsersLoadFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}

	text := s.adminHiddenUsersText(l, list)
	keyboard := s.adminHiddenUsersInlineKeyboard(l, list.Offset, list.Limit, list.Total)
	return text, keyboard, nil
}

func (s *service) adminAccessDeniedMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	text := s.adminAccessDeniedText(l)
	if s == nil {
		return text, nil, errors.New("service is not configured")
	}

	profile, status, statusErr := s.resolveProfileStatus(ctx, msg.User.ID)
	if isBannedError(statusErr) {
		return s.userBannedText(l), nil, statusErr
	}

	role := s.adminRoleForProfile(msg.User, profile, status)
	text = text + "\n" + s.helpTextFor(l, role)
	return text, s.startInlineKeyboardByStatus(l, status, role), statusErr
}

func (s *service) parseAdminUsersOffset(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}

	parts := strings.Split(value, ":")
	if len(parts) > 0 {
		value = strings.TrimSpace(parts[0])
	}

	offset, err := strconv.Atoi(value)
	if err != nil || offset < 0 {
		return 0
	}

	return offset
}

func (s *service) adminUsersText(l localizer, list domain.UserList) string {
	return s.adminUsersTextWithTemplates(l, list, msgAdminUsersPage, msgAdminUsersEmpty, msgAdminUsersEmptyPage)
}

func (s *service) adminBannedUsersText(l localizer, list domain.UserList) string {
	return s.adminUsersTextWithTemplates(l, list, msgAdminBannedUsersPage, msgAdminBannedUsersEmpty, msgAdminBannedUsersEmptyPage)
}

func (s *service) adminModeratorsText(l localizer, list domain.UserList) string {
	return s.adminUsersTextWithTemplates(l, list, msgAdminModeratorsPage, msgAdminModeratorsEmpty, msgAdminModeratorsEmptyPage)
}

func (s *service) adminReportsText(l localizer, list domain.ReportedUserList) string {
	return s.adminReportedUsersTextWithTemplates(l, list, msgAdminReportsPage, msgAdminReportsEmpty, msgAdminReportsEmptyPage)
}

func (s *service) adminHiddenUsersText(l localizer, list domain.UserList) string {
	return s.adminUsersTextWithTemplates(l, list, msgAdminHiddenUsersPage, msgAdminHiddenUsersEmpty, msgAdminHiddenUsersEmptyPage)
}

func (s *service) adminUsersTextWithTemplates(l localizer, list domain.UserList, pageKey, emptyKey, emptyPageKey messageKey) string {
	if len(list.Users) == 0 {
		if list.Total == 0 {
			return l.message(emptyKey)
		}
		return l.message(emptyPageKey, list.Total)
	}

	start := list.Offset + 1
	end := list.Offset + len(list.Users)
	if list.Total > 0 && end > list.Total {
		end = list.Total
	}

	lines := make([]string, 0, len(list.Users)+1)
	lines = append(lines, l.message(pageKey, list.Total, start, end))

	for i, user := range list.Users {
		name := strings.TrimSpace(user.Name)
		if name == "" {
			name = l.message(msgAdminUserListNoName)
		}
		usernameLabel := l.message(msgAdminUserListNA)
		if username := strings.TrimSpace(user.Username); username != "" {
			usernameLabel = s.formatUsername(username)
		}
		status := l.adminStatusLabel(user.IsBanned, user.IsModerator, user.IsHidden)
		lines = append(lines, l.message(msgAdminUserListLine, list.Offset+i+1, name, user.UserID, usernameLabel, status))
	}

	return strings.Join(lines, "\n")
}

func (s *service) adminReportedUsersTextWithTemplates(l localizer, list domain.ReportedUserList, pageKey, emptyKey, emptyPageKey messageKey) string {
	if len(list.Users) == 0 {
		if list.Total == 0 {
			return l.message(emptyKey)
		}
		return l.message(emptyPageKey, list.Total)
	}

	start := list.Offset + 1
	end := list.Offset + len(list.Users)
	if list.Total > 0 && end > list.Total {
		end = list.Total
	}

	lines := make([]string, 0, len(list.Users)+1)
	lines = append(lines, l.message(pageKey, list.Total, start, end))

	for i, user := range list.Users {
		name := strings.TrimSpace(user.Name)
		if name == "" {
			name = l.message(msgAdminUserListNoName)
		}
		usernameLabel := l.message(msgAdminUserListNA)
		if username := strings.TrimSpace(user.Username); username != "" {
			usernameLabel = s.formatUsername(username)
		}
		status := l.adminStatusLabel(user.IsBanned, user.IsModerator, user.IsHidden)
		lines = append(lines, l.message(msgAdminReportedUserListLine, list.Offset+i+1, name, user.UserID, usernameLabel, user.ReportCount, status))
	}

	return strings.Join(lines, "\n")
}
