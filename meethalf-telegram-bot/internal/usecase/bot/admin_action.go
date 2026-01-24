package bot

import (
	"context"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) applyAdminAction(ctx context.Context, msg domain.IncomingMessage) (domain.IncomingMessage, *domain.OutgoingMessage, bool, error) {
	if s == nil || s.adminActions == nil || msg.User.ID == 0 {
		return msg, nil, false, nil
	}
	if !s.isAdminUser(msg.User) {
		_ = s.adminActions.Delete(ctx, msg.User.ID)
		return msg, nil, false, nil
	}

	action, found, err := s.adminActions.Get(ctx, msg.User.ID)
	if err != nil {
		response := domain.OutgoingMessage{
			ChatID:         msg.ChatID,
			Text:           s.adminActionFailedText(),
			InlineKeyboard: s.adminMenuInlineKeyboard(),
		}
		return msg, &response, true, err
	}
	if !found {
		return msg, nil, false, nil
	}
	if action.ChatID != 0 && msg.ChatID != 0 && action.ChatID != msg.ChatID {
		return msg, nil, false, nil
	}

	switch action.Action {
	case domain.AdminActionBan:
		userID, username, ok := s.parseAdminUserIdentifier(msg.Text)
		if !ok {
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminBanUsageText(),
				InlineKeyboard: s.adminBanInlineKeyboard(),
			}
			return msg, &response, true, nil
		}
		msg.Command = domain.CommandAdminBan
		msg.Arguments = s.adminUserIdentifierLabel(userID, username)
		return msg, nil, false, nil
	case domain.AdminActionUnban:
		userID, username, ok := s.parseAdminUserIdentifier(msg.Text)
		if !ok {
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminUnbanUsageText(),
				InlineKeyboard: s.adminUnbanInlineKeyboard(),
			}
			return msg, &response, true, nil
		}
		msg.Command = domain.CommandAdminUnban
		msg.Arguments = s.adminUserIdentifierLabel(userID, username)
		return msg, nil, false, nil
	case domain.AdminActionModerator:
		userID, username, ok := s.parseAdminUserIdentifier(msg.Text)
		if !ok {
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminModeratorUsageText(),
				InlineKeyboard: s.adminModeratorInlineKeyboard(),
			}
			return msg, &response, true, nil
		}
		msg.Command = domain.CommandAdminModerator
		msg.Arguments = s.adminUserIdentifierLabel(userID, username)
		return msg, nil, false, nil
	case domain.AdminActionUnmoderator:
		userID, username, ok := s.parseAdminUserIdentifier(msg.Text)
		if !ok {
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminUnmoderatorUsageText(),
				InlineKeyboard: s.adminUnmoderatorInlineKeyboard(),
			}
			return msg, &response, true, nil
		}
		msg.Command = domain.CommandAdminUnmoderator
		msg.Arguments = s.adminUserIdentifierLabel(userID, username)
		return msg, nil, false, nil
	default:
		_ = s.adminActions.Delete(ctx, msg.User.ID)
		return msg, nil, false, nil
	}
}

func (s *service) clearAdminAction(ctx context.Context, userID int64) error {
	if s == nil || s.adminActions == nil || userID == 0 {
		return nil
	}

	return s.adminActions.Delete(ctx, userID)
}
