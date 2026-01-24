package bot

import (
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

func (s *service) isAdminUser(user domain.User) bool {
	if s == nil || len(s.adminUsernames) == 0 {
		return false
	}
	normalized := normalizeAdminUsername(user.Username)
	if normalized == "" {
		return false
	}
	_, ok := s.adminUsernames[normalized]
	return ok
}

func (s *service) helpTextFor(user domain.User) string {
	base := strings.TrimSpace(s.helpText)
	if !s.isAdminUser(user) {
		return base
	}
	if base == "" {
		return adminBadgeText
	}
	return adminBadgeText + "\n" + base
}
