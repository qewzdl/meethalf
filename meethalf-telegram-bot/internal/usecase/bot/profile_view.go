package bot

import (
	"context"
	"errors"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) showProfile(ctx context.Context, msg domain.IncomingMessage) (string, error) {
	profile, found, message, err := s.fetchProfile(ctx, msg)
	if err != nil || !found {
		return message, err
	}

	return s.profileDetails(profile), nil
}

func (s *service) showProfilePreview(ctx context.Context, msg domain.IncomingMessage) (string, error) {
	profile, found, message, err := s.fetchProfile(ctx, msg)
	if err != nil || !found {
		return message, err
	}

	return s.profilePreviewDetails(profile), nil
}

func (s *service) profileEditMenu(ctx context.Context, msg domain.IncomingMessage) (string, error) {
	if s != nil && s.drafts != nil && msg.User.ID != 0 {
		_ = s.drafts.Delete(ctx, msg.User.ID)
	}

	return s.profileEditMenuText(), nil
}

func (s *service) fetchProfile(ctx context.Context, msg domain.IncomingMessage) (domain.Profile, bool, string, error) {
	if s == nil || s.profiles == nil {
		return domain.Profile{}, false, "Profile service is not available right now.", errors.New("profile service is not configured")
	}

	if msg.User.ID == 0 {
		return domain.Profile{}, false, "Profile is not available for this chat.", errors.New("user id is missing")
	}

	profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
	if err != nil {
		return domain.Profile{}, false, "Unable to load profile. Please try again later.", err
	}
	if !found {
		return domain.Profile{}, false, "Profile not found. Use the button below to create it.", nil
	}

	return profile, true, "", nil
}

func (s *service) profileAlbumMessages(chatID int64, profile domain.Profile, keyboard *domain.InlineKeyboard) []domain.OutgoingMessage {
	return s.profileAlbumMessagesWithText(chatID, profile, s.profileDetails(profile), s.profileActionsText(), keyboard)
}

func (s *service) profilePreviewAlbumMessages(chatID int64, profile domain.Profile, keyboard *domain.InlineKeyboard) []domain.OutgoingMessage {
	return s.profileAlbumMessagesWithText(chatID, profile, s.profilePreviewDetails(profile), s.profilePreviewActionsText(), keyboard)
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
