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
		UserID:   msg.User.ID,
		ChatID:   msg.ChatID,
		Username: msg.User.Username,
		Language: lang,
		LastSeen: s.now(msg.ReceivedAt),
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

func (s *service) profileLanguageMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	value := strings.TrimSpace(msg.Arguments)
	if value == "" {
		value = strings.TrimSpace(msg.Text)
	}
	if value == "" || value == domain.CommandProfileLanguage {
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
		text := updatedLocalizer.message(msgLanguageUpdated) + "\n" + s.profileSettingsTextWithVisibility(updatedLocalizer, profile.IsHidden)
		return text, s.profileSettingsInlineKeyboard(updatedLocalizer, profile.IsHidden), err
	}

	text := updatedLocalizer.message(msgLanguageUpdated) + "\n" + updatedLocalizer.message(msgProfileNotFoundCreateButton)
	return text, s.profileCreateInlineKeyboard(updatedLocalizer), err
}
