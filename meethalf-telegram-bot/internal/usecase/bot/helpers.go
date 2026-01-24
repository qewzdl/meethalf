package bot

import (
	"context"
	"strconv"
	"strings"
	"time"

	"meethalf-telegram-bot/internal/domain"
)

const (
	albumDoneButtonText      = "Done"
	albumDoneCallbackData    = "done"
	telegramNameButtonText   = "Use Telegram name"
	telegramNameCallbackData = "yes"
)

func (s *service) userFullName(user domain.User) string {
	first := strings.TrimSpace(user.FirstName)
	last := strings.TrimSpace(user.LastName)

	if first == "" && last == "" {
		return ""
	}
	if last == "" {
		return first
	}
	if first == "" {
		return last
	}

	return first + " " + last
}

func (s *service) isAffirmative(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case telegramNameCallbackData, "y", "ok":
		return true
	default:
		return false
	}
}

func (s *service) normalizeGender(value string) (domain.Gender, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "male", "m":
		return domain.GenderMale, true
	case "female", "f":
		return domain.GenderFemale, true
	case "other", "o":
		return domain.GenderOther, true
	case "unspecified", "unknown", "not set":
		return domain.GenderUnspecified, true
	default:
		return "", false
	}
}

func (s *service) genderLabel(gender domain.Gender) string {
	switch gender {
	case domain.GenderMale:
		return "Male"
	case domain.GenderFemale:
		return "Female"
	case domain.GenderOther:
		return "Other"
	default:
		return "Not set"
	}
}

func (s *service) normalizeCountry(value string) (domain.Country, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "russia":
		return domain.CountryRussia, true
	case "kazakhstan":
		return domain.CountryKazakhstan, true
	case "belarus":
		return domain.CountryBelarus, true
	default:
		return "", false
	}
}

func (s *service) countryLabel(country domain.Country) string {
	switch country {
	case domain.CountryRussia:
		return "Russia"
	case domain.CountryKazakhstan:
		return "Kazakhstan"
	case domain.CountryBelarus:
		return "Belarus"
	default:
		return "Not set"
	}
}

func (s *service) isProfileShortcut(text string) bool {
	normalized := strings.ToLower(strings.TrimSpace(text))
	return normalized == "profile" || normalized == "/profile"
}

func (s *service) isAlbumDone(text string) bool {
	switch strings.ToLower(strings.TrimSpace(text)) {
	case albumDoneCallbackData, "finish", "save":
		return true
	default:
		return false
	}
}

func (s *service) formatTime(value time.Time) string {
	return value.UTC().Format("2006-01-02 15:04 UTC")
}

func (s *service) draftMode(draft domain.ProfileDraft) domain.ProfileDraftMode {
	if draft.Mode == "" {
		return domain.ProfileDraftModeCreate
	}

	return draft.Mode
}

func (s *service) now(fallback time.Time) time.Time {
	if fallback.IsZero() {
		return time.Now().UTC()
	}

	return fallback.UTC()
}

func (s *service) mergePhotoIDs(existing, incoming []string, limit int) ([]string, int) {
	if limit <= 0 {
		return existing, 0
	}

	out := make([]string, 0, len(existing)+len(incoming))
	seen := make(map[string]struct{}, len(existing))
	for _, id := range existing {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
		if len(out) >= limit {
			return out[:limit], 0
		}
	}

	added := 0
	for _, id := range incoming {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		if len(out) >= limit {
			break
		}
		seen[id] = struct{}{}
		out = append(out, id)
		added++
	}

	return out, added
}

func (s *service) parseBirthDate(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}

	layouts := []string{
		birthDateLayout,
		"2006-1-2",
		"02.01.2006",
		"2.1.2006",
		"02/01/2006",
		"2/1/2006",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err != nil {
			continue
		}
		normalized := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
		return normalized, true
	}

	return time.Time{}, false
}

func (s *service) ageFromBirthDate(birthDate time.Time, now time.Time) int {
	if birthDate.IsZero() {
		return 0
	}

	birthDate = birthDate.UTC()
	now = now.UTC()

	age := now.Year() - birthDate.Year()
	if now.Month() < birthDate.Month() || (now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		age--
	}

	return age
}

func (s *service) isProfileEditAction(command string) bool {
	switch command {
	case domain.CommandProfileEditName,
		domain.CommandProfileEditGender,
		domain.CommandProfileEditBirthDate,
		domain.CommandProfileEditCountry,
		domain.CommandProfileEditCity,
		domain.CommandProfileEditDesc,
		domain.CommandProfileEditEmoji,
		domain.CommandProfileEditPhotos:
		return true
	default:
		return false
	}
}

func (s *service) isDraftCommand(command string) bool {
	if command == "" {
		return true
	}

	switch command {
	case domain.CommandProfile,
		domain.CommandProfileSetupBack:
		return true
	default:
		return s.isProfileEditAction(command)
	}
}

func (s *service) isSearchCommand(command string) bool {
	switch command {
	case domain.CommandSearchStart,
		domain.CommandSearchRefresh,
		domain.CommandSearchGender,
		domain.CommandSearchAccuracy,
		domain.CommandMatchLike,
		domain.CommandMatchDislike,
		domain.CommandMatchReport,
		domain.CommandMatchPrevious,
		domain.CommandMatchViewProfile,
		domain.CommandMatchHistory,
		domain.CommandMatchHistoryView,
		domain.CommandMatchHistoryLike,
		domain.CommandMatchHistoryDislike,
		domain.CommandMatchHistoryReport:
		return true
	default:
		return false
	}
}

func (s *service) normalizeSearchGender(value string) (domain.Gender, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "any", "all":
		return domain.GenderUnspecified, true
	default:
		return s.normalizeGender(value)
	}
}

func (s *service) parseSearchAccuracy(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	accuracy, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	if accuracy < searchAccuracyMin || accuracy > searchAccuracyMax {
		return 0, false
	}
	return accuracy, true
}

func (s *service) parseSearchAccuracyArgs(value string) (domain.Gender, int, bool) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) != 2 {
		return "", 0, false
	}
	gender, ok := s.normalizeSearchGender(parts[0])
	if !ok {
		return "", 0, false
	}
	accuracy, ok := s.parseSearchAccuracy(parts[1])
	if !ok {
		return "", 0, false
	}
	return gender, accuracy, true
}

func (s *service) parseProfileVisibilityAction(value string) (bool, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case profileVisibilityHideAction:
		return true, true
	case profileVisibilityShowAction:
		return false, true
	default:
		return false, false
	}
}

func (s *service) parseTargetID(value string) (int64, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func (s *service) parseHistoryOffset(value string) int {
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

func (s *service) parseHistoryTarget(value string) (int64, int, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, 0, false
	}

	parts := strings.Split(value, ":")
	if len(parts) == 0 {
		return 0, 0, false
	}

	targetID, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil || targetID <= 0 {
		return 0, 0, false
	}

	offset := 0
	if len(parts) > 1 {
		offset, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
		if offset < 0 {
			offset = 0
		}
	}

	return targetID, offset, true
}

func (s *service) formatUsername(username string) string {
	normalized := strings.TrimSpace(username)
	if normalized == "" {
		return ""
	}
	if strings.HasPrefix(normalized, "@") {
		return normalized
	}
	return "@" + normalized
}

func (s *service) profileUsername(ctx context.Context, userID int64) string {
	if s == nil || s.sessions == nil || userID == 0 {
		return ""
	}

	session, found, err := s.sessions.Get(ctx, userID)
	if err != nil || !found {
		return ""
	}

	return strings.TrimSpace(session.Username)
}

func (s *service) adminUserIdentifierLabel(userID int64, username string) string {
	if username != "" {
		return s.formatUsername(username)
	}
	if userID <= 0 {
		return ""
	}
	return strconv.FormatInt(userID, 10)
}

func (s *service) nicknameFromUser(user domain.User, profile domain.Profile) string {
	if nickname := s.formatUsername(user.Username); nickname != "" {
		return nickname
	}
	if name := strings.TrimSpace(profile.Name); name != "" {
		return name
	}
	if fullName := s.userFullName(user); fullName != "" {
		return fullName
	}
	return "Unknown"
}

func (s *service) nicknameFromSession(session domain.Session, profile domain.Profile) string {
	if nickname := s.formatUsername(session.Username); nickname != "" {
		return nickname
	}
	if name := strings.TrimSpace(profile.Name); name != "" {
		return name
	}
	return "Unknown"
}
