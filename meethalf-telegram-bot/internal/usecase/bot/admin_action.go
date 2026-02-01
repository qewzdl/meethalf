package bot

import (
	"context"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) applyAdminAction(ctx context.Context, msg domain.IncomingMessage, l localizer) (domain.IncomingMessage, *domain.OutgoingMessage, bool, error) {
	if s == nil || s.adminActions == nil || msg.User.ID == 0 {
		return msg, nil, false, nil
	}
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil || !role.canAccessPanel() {
		_ = s.adminActions.Delete(ctx, msg.User.ID)
		return msg, nil, false, nil
	}

	action, found, err := s.adminActions.Get(ctx, msg.User.ID)
	if err != nil {
		response := domain.OutgoingMessage{
			ChatID:         msg.ChatID,
			Text:           s.adminActionFailedText(l),
			InlineKeyboard: s.adminMenuInlineKeyboard(l, role),
		}
		return msg, &response, true, err
	}
	if !found {
		return msg, nil, false, nil
	}
	if !role.allowsAdminAction(action.Action) {
		_ = s.adminActions.Delete(ctx, msg.User.ID)
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
				Text:           s.adminBanUsageText(l),
				InlineKeyboard: s.adminBanInlineKeyboard(l),
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
				Text:           s.adminUnbanUsageText(l),
				InlineKeyboard: s.adminUnbanInlineKeyboard(l),
			}
			return msg, &response, true, nil
		}
		msg.Command = domain.CommandAdminUnban
		msg.Arguments = s.adminUserIdentifierLabel(userID, username)
		return msg, nil, false, nil
	case domain.AdminActionShadowBan:
		userID, username, ok := s.parseAdminUserIdentifier(msg.Text)
		if !ok {
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminShadowBanUsageText(l),
				InlineKeyboard: s.adminShadowBanInlineKeyboard(l),
			}
			return msg, &response, true, nil
		}
		msg.Command = domain.CommandAdminShadowBan
		msg.Arguments = s.adminUserIdentifierLabel(userID, username)
		return msg, nil, false, nil
	case domain.AdminActionUnshadowBan:
		userID, username, ok := s.parseAdminUserIdentifier(msg.Text)
		if !ok {
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminShadowUnbanUsageText(l),
				InlineKeyboard: s.adminShadowUnbanInlineKeyboard(l),
			}
			return msg, &response, true, nil
		}
		msg.Command = domain.CommandAdminShadowUnban
		msg.Arguments = s.adminUserIdentifierLabel(userID, username)
		return msg, nil, false, nil
	case domain.AdminActionHideProfile:
		userID, username, ok := s.parseAdminUserIdentifier(msg.Text)
		if !ok {
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminHideProfileUsageText(l),
				InlineKeyboard: s.adminHideProfileInlineKeyboard(l),
			}
			return msg, &response, true, nil
		}
		msg.Command = domain.CommandAdminHideProfile
		msg.Arguments = s.adminUserIdentifierLabel(userID, username)
		return msg, nil, false, nil
	case domain.AdminActionShowProfile:
		userID, username, ok := s.parseAdminUserIdentifier(msg.Text)
		if !ok {
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminShowProfileUsageText(l),
				InlineKeyboard: s.adminShowProfileInlineKeyboard(l),
			}
			return msg, &response, true, nil
		}
		msg.Command = domain.CommandAdminShowProfile
		msg.Arguments = s.adminUserIdentifierLabel(userID, username)
		return msg, nil, false, nil
	case domain.AdminActionModerator:
		userID, username, ok := s.parseAdminUserIdentifier(msg.Text)
		if !ok {
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminModeratorUsageText(l),
				InlineKeyboard: s.adminModeratorInlineKeyboard(l),
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
				Text:           s.adminUnmoderatorUsageText(l),
				InlineKeyboard: s.adminUnmoderatorInlineKeyboard(l),
			}
			return msg, &response, true, nil
		}
		msg.Command = domain.CommandAdminUnmoderator
		msg.Arguments = s.adminUserIdentifierLabel(userID, username)
		return msg, nil, false, nil
	case domain.AdminActionResetChoices:
		userID, username, ok := s.parseAdminUserIdentifier(msg.Text)
		if !ok {
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminResetChoicesUsageText(l),
				InlineKeyboard: s.adminResetChoicesInlineKeyboard(l),
			}
			return msg, &response, true, nil
		}
		msg.Command = domain.CommandAdminResetChoices
		msg.Arguments = s.adminUserIdentifierLabel(userID, username)
		return msg, nil, false, nil
	case domain.AdminActionResetStart:
		userID, username, ok := s.parseAdminUserIdentifier(msg.Text)
		if !ok {
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminResetStartUsageText(l),
				InlineKeyboard: s.adminResetStartInlineKeyboard(l),
			}
			return msg, &response, true, nil
		}
		msg.Command = domain.CommandAdminResetStart
		msg.Arguments = s.adminUserIdentifierLabel(userID, username)
		return msg, nil, false, nil
	case domain.AdminActionClearReports:
		userID, username, ok := s.parseAdminUserIdentifier(msg.Text)
		if !ok {
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminClearReportsUsageText(l),
				InlineKeyboard: s.adminClearReportsInlineKeyboard(l),
			}
			return msg, &response, true, nil
		}
		msg.Command = domain.CommandAdminClearReports
		msg.Arguments = s.adminUserIdentifierLabel(userID, username)
		return msg, nil, false, nil
	case domain.AdminActionPostAd:
		text, photoIDs, buttons, payloadErr := s.adminAdPayload(msg)
		if payloadErr != nil {
			_, hasPhoto, buttonCount := adminAdDraftMeta(action)
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminAdUsageText(l),
				InlineKeyboard: s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0),
			}
			return msg, &response, true, nil
		}
		if text != "" && len([]rune(text)) > maxAdTextLength {
			_, hasPhoto, buttonCount := adminAdDraftMeta(action)
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminAdTooLongText(l, maxAdTextLength),
				InlineKeyboard: s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0),
			}
			return msg, &response, true, nil
		}

		updated := s.applyAdminAdDraftUpdate(&action, text, photoIDs, buttons)
		if action.ChatID == 0 {
			action.ChatID = msg.ChatID
		}
		action.Action = domain.AdminActionPostAd
		if updated {
			action.RequestedAt = s.now(msg.ReceivedAt)
			if err := s.adminActions.Save(ctx, action); err != nil {
				response := domain.OutgoingMessage{
					ChatID:         msg.ChatID,
					Text:           s.adminAdFailedText(l),
					InlineKeyboard: s.adminMenuInlineKeyboard(l, role),
				}
				return msg, &response, true, err
			}
		}

		hasText, hasPhoto, buttonCount := adminAdDraftMeta(action)
		if !hasText && !hasPhoto {
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminAdUsageText(l),
				InlineKeyboard: s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0),
			}
			return msg, &response, true, nil
		}
		response := domain.OutgoingMessage{
			ChatID:         msg.ChatID,
			Text:           s.adminAdDraftStatusText(l, hasText, hasPhoto, buttonCount),
			InlineKeyboard: s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0),
		}
		return msg, &response, true, nil
	case domain.AdminActionPostAdButton:
		button, err := parseAdButtonInput(msg.Text)
		if err != nil {
			_, hasPhoto, buttonCount := adminAdDraftMeta(action)
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminAdButtonUsageText(l),
				InlineKeyboard: s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0),
			}
			return msg, &response, true, nil
		}
		action.AdButtons = append(action.AdButtons, button)
		action.Action = domain.AdminActionPostAd
		if action.ChatID == 0 {
			action.ChatID = msg.ChatID
		}
		action.RequestedAt = s.now(msg.ReceivedAt)
		if err := s.adminActions.Save(ctx, action); err != nil {
			response := domain.OutgoingMessage{
				ChatID:         msg.ChatID,
				Text:           s.adminAdFailedText(l),
				InlineKeyboard: s.adminMenuInlineKeyboard(l, role),
			}
			return msg, &response, true, err
		}
		hasText, hasPhoto, buttonCount := adminAdDraftMeta(action)
		response := domain.OutgoingMessage{
			ChatID:         msg.ChatID,
			Text:           s.adminAdButtonAddedText(l, buttonCount) + "\n" + s.adminAdDraftStatusText(l, hasText, hasPhoto, buttonCount),
			InlineKeyboard: s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0),
		}
		return msg, &response, true, nil
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
