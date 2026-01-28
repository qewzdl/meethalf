package bot

import (
	"context"
	"errors"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) showProfile(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, error) {
	profile, found, message, err := s.fetchProfile(ctx, msg, l)
	if err != nil || !found {
		return message, err
	}

	return s.profileDetails(l, profile), nil
}

func (s *service) showProfilePreview(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, error) {
	profile, found, message, err := s.fetchProfile(ctx, msg, l)
	if err != nil || !found {
		return message, err
	}

	return s.profilePreviewDetails(l, profile), nil
}

func (s *service) profileEditMenu(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, error) {
	if s != nil && s.drafts != nil && msg.User.ID != 0 {
		_ = s.drafts.Delete(ctx, msg.User.ID)
	}

	return s.profileEditMenuText(l), nil
}

func (s *service) fetchProfile(ctx context.Context, msg domain.IncomingMessage, l localizer) (domain.Profile, bool, string, error) {
	if s == nil || s.profiles == nil {
		return domain.Profile{}, false, l.message(msgProfileViewUnavailable), errors.New("profile service is not configured")
	}

	if msg.User.ID == 0 {
		return domain.Profile{}, false, l.message(msgProfileViewUnavailableChat), errors.New("user id is missing")
	}

	profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
	if err != nil {
		if isBannedError(err) {
			return domain.Profile{}, false, s.userBannedText(l), err
		}
		return domain.Profile{}, false, l.message(msgProfileLoadFailed), err
	}
	if !found {
		return domain.Profile{}, false, l.message(msgProfileNotFoundButtonBelow), nil
	}

	return profile, true, "", nil
}

func (s *service) profileAlbumMessages(chatID int64, profile domain.Profile, keyboard *domain.InlineKeyboard, l localizer) []domain.OutgoingMessage {
	return s.profileAlbumMessagesWithText(chatID, profile, s.profileDetails(l, profile), s.profileActionsText(l), keyboard)
}

func (s *service) profilePreviewAlbumMessages(chatID int64, profile domain.Profile, keyboard *domain.InlineKeyboard, l localizer) []domain.OutgoingMessage {
	return s.profileAlbumMessagesWithText(chatID, profile, s.profilePreviewDetails(l, profile), s.profilePreviewActionsText(l), keyboard)
}

func (s *service) profileAlbumMessagesWithText(chatID int64, profile domain.Profile, detailsText, actionsText string, keyboard *domain.InlineKeyboard) []domain.OutgoingMessage {
	text := detailsText
	if len(profile.Photos) == 0 {
		return []domain.OutgoingMessage{
			{
				ChatID:         chatID,
				Text:           text,
				InlineKeyboard: keyboard,
			},
		}
	}
	if len(profile.Photos) == 1 {
		return []domain.OutgoingMessage{
			{
				ChatID:         chatID,
				Text:           text,
				InlineKeyboard: keyboard,
				PhotoIDs:       profile.Photos,
			},
		}
	}

	messages := []domain.OutgoingMessage{
		{
			ChatID:   chatID,
			Text:     text,
			PhotoIDs: profile.Photos,
		},
	}
	if keyboard != nil {
		messages = append(messages, domain.OutgoingMessage{
			ChatID:         chatID,
			Text:           actionsText,
			InlineKeyboard: keyboard,
		})
	}

	return messages
}
