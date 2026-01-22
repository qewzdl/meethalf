package bot

import (
	"context"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) profileDetailsFollowUp(ctx context.Context, msg domain.IncomingMessage, text string, replyErr error) ([]domain.OutgoingMessage, error) {
	if replyErr != nil || !s.needsProfileDetailsFollowUp(text) {
		return nil, nil
	}

	if s == nil || s.profiles == nil {
		return nil, nil
	}

	if msg.User.ID == 0 {
		return nil, nil
	}

	profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	return s.profileAlbumMessages(msg.ChatID, profile, s.profileViewInlineKeyboard()), nil
}

func (s *service) needsProfileDetailsFollowUp(text string) bool {
	switch text {
	case profileCreatedText, profileUpdatedText:
		return true
	default:
		return false
	}
}
