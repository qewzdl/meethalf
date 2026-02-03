package bot

import (
	"context"
	"time"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) enforceAgeAccess(ctx context.Context, msg domain.IncomingMessage, l localizer) (*domain.OutgoingMessage, bool, error) {
	if s == nil || msg.User.ID == 0 {
		return nil, false, nil
	}

	if response, handled, err := s.handleAgeConfirmationAction(ctx, msg, l); handled || err != nil {
		return response, handled, err
	}

	confirmed, err := s.isAgeConfirmed(ctx, msg.User.ID)
	if err != nil {
		return nil, false, err
	}

	profile, found, err := s.profileForAgeGate(ctx, msg.User.ID)
	if err != nil {
		if isBannedError(err) {
			return &domain.OutgoingMessage{ChatID: msg.ChatID, Text: s.userBannedText(l)}, true, err
		}
		return nil, false, err
	}

	if found {
		age := s.profileAge(profile, s.now(msg.ReceivedAt))
		if age < minAge {
			return &domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgAgeAccessDenied, minAge)}, true, nil
		}
		if !confirmed && s.ageConfirmations != nil {
			if saveErr := s.saveAgeConfirmation(ctx, msg); saveErr != nil {
				return &domain.OutgoingMessage{ChatID: msg.ChatID, Text: s.ageConfirmationFailedText(l)}, true, saveErr
			}
			confirmed = true
		}
	}

	if !confirmed && s.ageConfirmations != nil {
		return &domain.OutgoingMessage{
			ChatID:         msg.ChatID,
			Text:           s.ageConfirmationPromptText(l),
			InlineKeyboard: s.ageConfirmationInlineKeyboard(l),
		}, true, nil
	}

	return nil, false, nil
}

func (s *service) handleAgeConfirmationAction(ctx context.Context, msg domain.IncomingMessage, l localizer) (*domain.OutgoingMessage, bool, error) {
	switch msg.Command {
	case domain.CommandAgeConfirmYes:
		if saveErr := s.saveAgeConfirmation(ctx, msg); saveErr != nil {
			return &domain.OutgoingMessage{ChatID: msg.ChatID, Text: s.ageConfirmationFailedText(l)}, true, saveErr
		}
		text, keyboard := s.languageOnboardingMessage(l)
		return &domain.OutgoingMessage{ChatID: msg.ChatID, Text: text, InlineKeyboard: keyboard}, true, nil
	case domain.CommandAgeConfirmNo:
		return &domain.OutgoingMessage{
			ChatID:         msg.ChatID,
			Text:           s.ageConfirmationDeclinedText(l),
			InlineKeyboard: s.ageConfirmationInlineKeyboard(l),
		}, true, nil
	default:
		return nil, false, nil
	}
}

func (s *service) isAgeConfirmed(ctx context.Context, userID int64) (bool, error) {
	if s == nil || s.ageConfirmations == nil || userID == 0 {
		return false, nil
	}

	_, found, err := s.ageConfirmations.Get(ctx, userID)
	if err != nil {
		return false, err
	}

	return found, nil
}

func (s *service) saveAgeConfirmation(ctx context.Context, msg domain.IncomingMessage) error {
	if s == nil || s.ageConfirmations == nil || msg.User.ID == 0 {
		return nil
	}

	confirmation := domain.AgeConfirmation{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		Username:    normalizeUsername(msg.User.Username),
		ConfirmedAt: s.now(msg.ReceivedAt),
	}

	return s.ageConfirmations.Save(ctx, confirmation)
}

func (s *service) profileForAgeGate(ctx context.Context, userID int64) (domain.Profile, bool, error) {
	if s == nil || s.profiles == nil || userID == 0 {
		return domain.Profile{}, false, nil
	}

	return s.profiles.GetProfile(ctx, userID)
}

func (s *service) profileAge(profile domain.Profile, now time.Time) int {
	age := s.ageFromBirthDate(profile.BirthDate, now)
	if age == 0 {
		age = profile.Age
	}
	return age
}
