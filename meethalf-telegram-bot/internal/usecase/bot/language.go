package bot

import (
	"context"
	"errors"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) resolveLanguage(ctx context.Context, msg domain.IncomingMessage) domain.Language {
	if lang := s.sessionLanguage(ctx, msg.User.ID); lang != "" {
		return lang
	}
	if lang := languageFromTelegramCode(msg.User.LanguageCode); lang != "" {
		return lang
	}
	return domain.DefaultLanguage
}

func (s *service) resolveLanguageForUser(ctx context.Context, userID int64, fallback domain.Language) domain.Language {
	if lang := s.sessionLanguage(ctx, userID); lang != "" {
		return lang
	}
	if fallback != "" {
		return normalizeLanguageValue(fallback)
	}
	return domain.DefaultLanguage
}

func (s *service) sessionLanguage(ctx context.Context, userID int64) domain.Language {
	if s == nil || s.sessions == nil || userID == 0 {
		return ""
	}
	session, found, err := s.sessions.Get(ctx, userID)
	if err != nil || !found {
		return ""
	}
	if session.Language == "" {
		return ""
	}
	return normalizeLanguageValue(session.Language)
}

func (s *service) setSessionLanguage(ctx context.Context, msg domain.IncomingMessage, lang domain.Language) error {
	if s == nil || s.sessions == nil || msg.User.ID == 0 {
		return nil
	}
	lang = normalizeLanguageValue(lang)
	return s.sessions.Touch(ctx, domain.Session{
		UserID:                msg.User.ID,
		ChatID:                msg.ChatID,
		Username:              msg.User.Username,
		Language:              lang,
		LastSeen:              s.now(msg.ReceivedAt),
		SearchAccuracyEnabled: s.sessionSearchAccuracyEnabled(ctx, msg.User.ID),
		PendingAISearch:       s.sessionAISearchPending(ctx, msg.User.ID),
	})
}

func parseLanguageInput(value string) (domain.Language, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "en", "eng", "english":
		return domain.LanguageEnglish, true
	case "ru", "rus", "russian":
		return domain.LanguageRussian, true
	default:
		return "", false
	}
}

func languageFromTelegramCode(code string) domain.Language {
	normalized := strings.ToLower(strings.TrimSpace(code))
	if normalized == "" {
		return ""
	}
	if strings.HasPrefix(normalized, "ru") {
		return domain.LanguageRussian
	}
	if strings.HasPrefix(normalized, "en") {
		return domain.LanguageEnglish
	}
	return ""
}

func (s *service) languageOnboardingMessage(l localizer) (string, *domain.InlineKeyboard) {
	if s == nil {
		return l.message(msgLanguagePrompt), nil
	}
	return l.message(msgLanguagePrompt), s.languageOnboardingInlineKeyboard(l)
}

func (s *service) languageOnboardingSelectionMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	value := strings.TrimSpace(msg.Arguments)
	if value == "" {
		return l.message(msgLanguagePrompt), s.languageOnboardingInlineKeyboard(l), nil
	}

	lang, ok := parseLanguageInput(value)
	if !ok {
		return l.message(msgLanguageUnsupported), s.languageOnboardingInlineKeyboard(l), nil
	}

	err := s.setSessionLanguage(ctx, msg, lang)
	updatedLocalizer := newLocalizer(lang)
	text, keyboard, startErr := s.startMessage(ctx, msg, updatedLocalizer)
	if text != "" {
		text = updatedLocalizer.message(msgLanguageUpdated) + "\n" + text
	} else {
		text = updatedLocalizer.message(msgLanguageUpdated)
	}
	return text, keyboard, errors.Join(err, startErr)
}

func (s *service) profileLanguageMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	value := strings.TrimSpace(msg.Arguments)
	if value == "" {
		if msg.Command == domain.CommandLanguage {
			value = ""
		} else {
			value = strings.TrimSpace(msg.Text)
		}
	}
	if value == "" || value == domain.CommandProfileLanguage || value == domain.CommandLanguage {
		return l.message(msgLanguagePrompt), s.languageInlineKeyboard(l), nil
	}

	lang, ok := parseLanguageInput(value)
	if !ok {
		return l.message(msgLanguageUnsupported), s.languageInlineKeyboard(l), nil
	}

	err := s.setSessionLanguage(ctx, msg, lang)
	updatedLocalizer := newLocalizer(lang)
	if s == nil || s.profiles == nil || msg.User.ID == 0 {
		return updatedLocalizer.message(msgLanguageUpdated), s.languageInlineKeyboard(updatedLocalizer), err
	}

	profile, found, profileErr := s.profiles.GetProfile(ctx, msg.User.ID)
	if profileErr != nil {
		if isBannedError(profileErr) {
			return s.userBannedText(updatedLocalizer), nil, errors.Join(err, profileErr)
		}
		return updatedLocalizer.message(msgProfileLoadFailed), s.languageInlineKeyboard(updatedLocalizer), errors.Join(err, profileErr)
	}

	if found {
		searchAccuracyEnabled := s.sessionSearchAccuracyEnabled(ctx, msg.User.ID)
		text := updatedLocalizer.message(msgLanguageUpdated) + "\n" + s.profileSettingsTextWithVisibility(updatedLocalizer, profile.IsHidden, searchAccuracyEnabled, profile.LikesNotificationsEnabled)
		return text, s.profileSettingsInlineKeyboard(updatedLocalizer, profile.IsHidden, searchAccuracyEnabled, profile.LikesNotificationsEnabled), err
	}

	text := updatedLocalizer.message(msgLanguageUpdated) + "\n" + s.profileSettingsText(updatedLocalizer)
	return text, s.profileSettingsGuestInlineKeyboard(updatedLocalizer), err
}
