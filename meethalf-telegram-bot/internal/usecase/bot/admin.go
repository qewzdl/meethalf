package bot

import (
	"context"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

func normalizeAdminUsername(username string) string {
	normalized := strings.TrimSpace(username)
	if normalized == "" {
		return ""
	}
	normalized = strings.TrimPrefix(normalized, "@")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return ""
	}
	return strings.ToLower(normalized)
}

func normalizeAdminUsernames(usernames []string) map[string]struct{} {
	if len(usernames) == 0 {
		return nil
	}

	unique := make(map[string]struct{}, len(usernames))
	for _, username := range usernames {
		normalized := normalizeAdminUsername(username)
		if normalized == "" {
			continue
		}
		unique[normalized] = struct{}{}
	}

	if len(unique) == 0 {
		return nil
	}

	return unique
}

type adminRole int

const (
	adminRoleNone adminRole = iota
	adminRoleModerator
	adminRoleAdmin
)

func (role adminRole) canAccessPanel() bool {
	return role >= adminRoleModerator
}

func (role adminRole) canModerateUsers() bool {
	return role >= adminRoleModerator
}

func (role adminRole) canManageModerators() bool {
	return role == adminRoleAdmin
}

func (role adminRole) allowsAdminAction(action domain.AdminActionType) bool {
	switch action {
	case domain.AdminActionBan, domain.AdminActionUnban:
		return role.canModerateUsers()
	case domain.AdminActionModerator, domain.AdminActionUnmoderator:
		return role.canManageModerators()
	default:
		return false
	}
}

func (s *service) isAdminUser(user domain.User) bool {
	return s.isAdminUsername(user.Username)
}

func (s *service) isAdminUsername(username string) bool {
	if s == nil || len(s.adminUsernames) == 0 {
		return false
	}
	normalized := normalizeAdminUsername(username)
	if normalized == "" {
		return false
	}
	_, ok := s.adminUsernames[normalized]
	return ok
}

func (s *service) adminRoleForProfile(user domain.User, profile domain.Profile, status profileStatus) adminRole {
	if s.isAdminUser(user) {
		return adminRoleAdmin
	}
	if status == profileStatusPresent && profile.IsModerator {
		return adminRoleModerator
	}
	return adminRoleNone
}

func (s *service) resolveAdminRole(ctx context.Context, user domain.User) (adminRole, error) {
	if s.isAdminUser(user) {
		return adminRoleAdmin, nil
	}
	if s == nil || s.profiles == nil || user.ID == 0 {
		return adminRoleNone, nil
	}

	profile, found, err := s.profiles.GetProfile(ctx, user.ID)
	if err != nil {
		return adminRoleNone, err
	}
	if found && profile.IsModerator {
		return adminRoleModerator, nil
	}
	return adminRoleNone, nil
}

func (s *service) helpTextFor(user domain.User, role adminRole) string {
	base := strings.TrimSpace(s.helpText)
	badge := ""
	switch role {
	case adminRoleAdmin:
		badge = adminBadgeText
	case adminRoleModerator:
		badge = moderatorBadgeText
	}

	if badge == "" {
		return base
	}
	if base == "" {
		return badge
	}
	return badge + "\n" + base
}
