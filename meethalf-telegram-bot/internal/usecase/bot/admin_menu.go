package bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

const adminUsersPageSize = 20

func (s *service) adminMenuMessage(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	if !s.isAdminUser(msg.User) {
		return s.adminAccessDeniedMessage(ctx, msg)
	}

	if s != nil && s.adminActions != nil && msg.User.ID != 0 {
		_ = s.adminActions.Delete(ctx, msg.User.ID)
	}

	return s.adminMenuText(), s.adminMenuInlineKeyboard(), nil
}

func (s *service) adminUsersMessage(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	if !s.isAdminUser(msg.User) {
		return s.adminAccessDeniedMessage(ctx, msg)
	}

	if s == nil || s.admin == nil {
		return s.adminUsersLoadFailedText(), s.adminMenuInlineKeyboard(), errors.New("admin service is not configured")
	}

	offset := s.parseAdminUsersOffset(msg.Arguments)
	list, err := s.admin.ListUsers(ctx, adminUsersPageSize, offset, false, false)
	if err != nil {
		return s.adminUsersLoadFailedText(), s.adminMenuInlineKeyboard(), err
	}

	text := s.adminUsersText(list)
	keyboard := s.adminUsersInlineKeyboard(list.Offset, list.Limit, list.Total)
	return text, keyboard, nil
}

func (s *service) adminBannedUsersMessage(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	if !s.isAdminUser(msg.User) {
		return s.adminAccessDeniedMessage(ctx, msg)
	}

	if s == nil || s.admin == nil {
		return s.adminBannedUsersLoadFailedText(), s.adminMenuInlineKeyboard(), errors.New("admin service is not configured")
	}

	offset := s.parseAdminUsersOffset(msg.Arguments)
	list, err := s.admin.ListUsers(ctx, adminUsersPageSize, offset, true, false)
	if err != nil {
		return s.adminBannedUsersLoadFailedText(), s.adminMenuInlineKeyboard(), err
	}

	text := s.adminBannedUsersText(list)
	keyboard := s.adminBannedUsersInlineKeyboard(list.Offset, list.Limit, list.Total)
	return text, keyboard, nil
}

func (s *service) adminModeratorsMessage(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	if !s.isAdminUser(msg.User) {
		return s.adminAccessDeniedMessage(ctx, msg)
	}

	if s == nil || s.admin == nil {
		return s.adminModeratorsLoadFailedText(), s.adminMenuInlineKeyboard(), errors.New("admin service is not configured")
	}

	offset := s.parseAdminUsersOffset(msg.Arguments)
	list, err := s.admin.ListUsers(ctx, adminUsersPageSize, offset, false, true)
	if err != nil {
		return s.adminModeratorsLoadFailedText(), s.adminMenuInlineKeyboard(), err
	}

	text := s.adminModeratorsText(list)
	keyboard := s.adminModeratorsInlineKeyboard(list.Offset, list.Limit, list.Total)
	return text, keyboard, nil
}

func (s *service) adminReportsMessage(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	if !s.isAdminUser(msg.User) {
		return s.adminAccessDeniedMessage(ctx, msg)
	}

	if s == nil || s.admin == nil {
		return s.adminReportsLoadFailedText(), s.adminMenuInlineKeyboard(), errors.New("admin service is not configured")
	}

	offset := s.parseAdminUsersOffset(msg.Arguments)
	list, err := s.admin.ListReportedUsers(ctx, adminUsersPageSize, offset)
	if err != nil {
		return s.adminReportsLoadFailedText(), s.adminMenuInlineKeyboard(), err
	}

	text := s.adminReportsText(list)
	keyboard := s.adminReportsInlineKeyboard(list.Offset, list.Limit, list.Total)
	return text, keyboard, nil
}

func (s *service) adminAccessDeniedMessage(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	text := s.adminAccessDeniedText()
	if s == nil {
		return text, nil, errors.New("service is not configured")
	}

	_, status, statusErr := s.resolveProfileStatus(ctx, msg.User.ID)
	if isBannedError(statusErr) {
		return s.userBannedText(), nil, statusErr
	}
	if s.helpText != "" {
		text = text + "\n" + s.helpText
	}

	return text, s.startInlineKeyboardByStatus(status, msg.User), statusErr
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

func (s *service) adminUsersText(list domain.UserList) string {
	return s.adminUsersTextWithTemplates(list, adminUsersPageTemplate, s.adminUsersEmptyText(), adminUsersEmptyPageTemplate)
}

func (s *service) adminBannedUsersText(list domain.UserList) string {
	return s.adminUsersTextWithTemplates(list, adminBannedUsersPageTemplate, s.adminBannedUsersEmptyText(), adminBannedUsersEmptyPageTemplate)
}

func (s *service) adminModeratorsText(list domain.UserList) string {
	return s.adminUsersTextWithTemplates(list, adminModeratorsPageTemplate, s.adminModeratorsEmptyText(), adminModeratorsEmptyPageTemplate)
}

func (s *service) adminReportsText(list domain.ReportedUserList) string {
	return s.adminReportedUsersTextWithTemplates(list, adminReportsPageTemplate, s.adminReportsEmptyText(), adminReportsEmptyPageTemplate)
}

func (s *service) adminUsersTextWithTemplates(list domain.UserList, pageTemplate, emptyText, emptyPageTemplate string) string {
	if len(list.Users) == 0 {
		if list.Total == 0 {
			return emptyText
		}
		return fmt.Sprintf(emptyPageTemplate, list.Total)
	}

	start := list.Offset + 1
	end := list.Offset + len(list.Users)
	if list.Total > 0 && end > list.Total {
		end = list.Total
	}

	lines := make([]string, 0, len(list.Users)+1)
	lines = append(lines, fmt.Sprintf(pageTemplate, list.Total, start, end))

	for i, user := range list.Users {
		name := strings.TrimSpace(user.Name)
		if name == "" {
			name = "No name"
		}
		usernameLabel := "n/a"
		if username := strings.TrimSpace(user.Username); username != "" {
			usernameLabel = s.formatUsername(username)
		}
		statusParts := make([]string, 0, 3)
		if user.IsBanned {
			statusParts = append(statusParts, "banned")
		}
		if user.IsModerator {
			statusParts = append(statusParts, "moderator")
		}
		if user.IsHidden {
			statusParts = append(statusParts, "hidden")
		}
		status := "visible"
		if len(statusParts) > 0 {
			status = strings.Join(statusParts, ", ")
		}
		lines = append(lines, fmt.Sprintf("%d. %s (ID: %d, username: %s, %s)", list.Offset+i+1, name, user.UserID, usernameLabel, status))
	}

	return strings.Join(lines, "\n")
}

func (s *service) adminReportedUsersTextWithTemplates(list domain.ReportedUserList, pageTemplate, emptyText, emptyPageTemplate string) string {
	if len(list.Users) == 0 {
		if list.Total == 0 {
			return emptyText
		}
		return fmt.Sprintf(emptyPageTemplate, list.Total)
	}

	start := list.Offset + 1
	end := list.Offset + len(list.Users)
	if list.Total > 0 && end > list.Total {
		end = list.Total
	}

	lines := make([]string, 0, len(list.Users)+1)
	lines = append(lines, fmt.Sprintf(pageTemplate, list.Total, start, end))

	for i, user := range list.Users {
		name := strings.TrimSpace(user.Name)
		if name == "" {
			name = "No name"
		}
		usernameLabel := "n/a"
		if username := strings.TrimSpace(user.Username); username != "" {
			usernameLabel = s.formatUsername(username)
		}
		statusParts := make([]string, 0, 3)
		if user.IsBanned {
			statusParts = append(statusParts, "banned")
		}
		if user.IsModerator {
			statusParts = append(statusParts, "moderator")
		}
		if user.IsHidden {
			statusParts = append(statusParts, "hidden")
		}
		status := "visible"
		if len(statusParts) > 0 {
			status = strings.Join(statusParts, ", ")
		}
		lines = append(lines, fmt.Sprintf("%d. %s (ID: %d, username: %s, reports: %d, %s)", list.Offset+i+1, name, user.UserID, usernameLabel, user.ReportCount, status))
	}

	return strings.Join(lines, "\n")
}
