package bot

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

type statusError interface {
	error
	StatusCode() int
}

const (
	historyPageSize = 5
	likesPageSize   = 5
)

func (s *service) handleSearch(ctx context.Context, msg domain.IncomingMessage, l localizer) ([]domain.OutgoingMessage, error) {
	if s == nil || s.search == nil {
		return []domain.OutgoingMessage{
			{ChatID: msg.ChatID, Text: s.searchUnavailableText(l)},
		}, errors.New("search service is not configured")
	}
	if msg.User.ID == 0 {
		return []domain.OutgoingMessage{
			{ChatID: msg.ChatID, Text: s.searchUnavailableText(l)},
		}, errors.New("user id is missing")
	}

	switch msg.Command {
	case domain.CommandSearchStart:
		hasProfile, err := s.viewerHasProfile(ctx, msg.User.ID)
		if err != nil {
			if isBannedError(err) {
				return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.userBannedText(l)}}, err
			}
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, err
		}
		if !hasProfile {
			return []domain.OutgoingMessage{
				{
					ChatID:         msg.ChatID,
					Text:           s.searchProfileMissingText(l),
					InlineKeyboard: s.profileCreateInlineKeyboardWithAdmin(ctx, msg, l),
				},
			}, nil
		}
		return []domain.OutgoingMessage{
			{
				ChatID:         msg.ChatID,
				Text:           s.searchGenderText(l),
				InlineKeyboard: s.searchGenderInlineKeyboard(l),
			},
		}, nil
	case domain.CommandSearchAI:
		hasProfile, err := s.viewerHasProfile(ctx, msg.User.ID)
		if err != nil {
			_ = s.setSessionAISearchPending(ctx, msg, false)
			if isBannedError(err) {
				return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.userBannedText(l)}}, err
			}
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, err
		}
		if !hasProfile {
			_ = s.setSessionAISearchPending(ctx, msg, false)
			return []domain.OutgoingMessage{
				{
					ChatID:         msg.ChatID,
					Text:           s.searchProfileMissingText(l),
					InlineKeyboard: s.profileCreateInlineKeyboardWithAdmin(ctx, msg, l),
				},
			}, nil
		}

		pending := s.sessionAISearchPending(ctx, msg.User.ID)
		if !pending {
			available, err := s.search.AIAvailable(ctx, msg.User.ID)
			if err != nil {
				_ = s.setSessionAISearchPending(ctx, msg, false)
				return s.searchErrorResponse(ctx, msg, l, err), err
			}
			if !available {
				_ = s.setSessionAISearchPending(ctx, msg, false)
				return []domain.OutgoingMessage{{
					ChatID:         msg.ChatID,
					Text:           s.searchAIUnavailableText(l),
					InlineKeyboard: s.searchAIUnavailableInlineKeyboard(l),
				}}, nil
			}
		}

		if pending {
			prompt := strings.TrimSpace(msg.Text)
			if len([]rune(prompt)) < aiSearchMinPromptLength {
				return []domain.OutgoingMessage{
					{
						ChatID:         msg.ChatID,
						Text:           s.searchAITooShortText(l) + "\n" + s.searchAIPromptText(l),
						InlineKeyboard: withCancelInlineKeyboard(l, nil),
					},
				}, nil
			}

			_ = s.setSessionAISearchPending(ctx, msg, false)
			candidate, found, err := s.search.SearchWithAI(ctx, msg.User.ID, prompt)
			if err != nil {
				return s.searchErrorResponse(ctx, msg, l, err), err
			}
			if !found {
				return []domain.OutgoingMessage{{
					ChatID:         msg.ChatID,
					Text:           s.searchNoCandidatesText(l),
					InlineKeyboard: s.searchNoCandidatesInlineKeyboard(l),
				}}, nil
			}

			return s.matchProfileMessages(msg.ChatID, candidate, l), nil
		}

		pendingErr := s.setSessionAISearchPending(ctx, msg, true)
		return []domain.OutgoingMessage{
			{
				ChatID:         msg.ChatID,
				Text:           s.searchAIPromptText(l),
				InlineKeyboard: withCancelInlineKeyboard(l, nil),
			},
		}, pendingErr
	case domain.CommandSearchGender:
		value := msg.Arguments
		if value == "" {
			value = msg.Text
		}
		gender, ok := s.normalizeSearchGender(value)
		if !ok {
			return []domain.OutgoingMessage{
				{
					ChatID:         msg.ChatID,
					Text:           s.searchGenderText(l),
					InlineKeyboard: s.searchGenderInlineKeyboard(l),
				},
			}, nil
		}

		if s.sessionSearchAccuracyEnabled(ctx, msg.User.ID) {
			return []domain.OutgoingMessage{
				{
					ChatID:         msg.ChatID,
					Text:           s.searchAccuracyText(l),
					InlineKeyboard: s.searchAccuracyInlineKeyboard(l, gender),
				},
			}, nil
		}

		return s.startSearch(ctx, msg, l, gender, searchAccuracyDefault)
	case domain.CommandSearchAccuracy:
		gender, accuracy, ok := s.parseSearchAccuracyArgs(msg.Arguments)
		if !ok {
			return []domain.OutgoingMessage{
				{
					ChatID:         msg.ChatID,
					Text:           s.searchGenderText(l),
					InlineKeyboard: s.searchGenderInlineKeyboard(l),
				},
			}, nil
		}

		if !s.sessionSearchAccuracyEnabled(ctx, msg.User.ID) {
			accuracy = searchAccuracyDefault
		}

		return s.startSearch(ctx, msg, l, gender, accuracy)
	case domain.CommandSearchRefresh:
		return s.nextMatch(ctx, msg, l)
	case domain.CommandMatchLike:
		targetID, ok := s.parseTargetID(msg.Arguments)
		if !ok {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, nil
		}
		actionResult, err := s.search.RecordAction(ctx, msg.User.ID, targetID, domain.MatchActionLike)
		if err != nil {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, err
		}
		nextMessages, nextErr := s.nextMatch(ctx, msg, l)
		if actionResult.Matched {
			currentMatchMessages, targetMatchMessages, matchErr := s.mutualMatchMessages(ctx, msg, targetID, l)
			messages := append(currentMatchMessages, nextMessages...)
			if len(targetMatchMessages) > 0 {
				messages = append(messages, targetMatchMessages...)
			}
			return messages, errors.Join(nextErr, matchErr)
		}

		likeNotifications, likeErr := s.notifyPendingLikesForUser(ctx, targetID)
		if len(likeNotifications) > 0 {
			nextMessages = append(nextMessages, likeNotifications...)
		}
		return nextMessages, errors.Join(nextErr, likeErr)
	case domain.CommandMatchDislike:
		targetID, ok := s.parseTargetID(msg.Arguments)
		if !ok {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, nil
		}
		if _, err := s.search.RecordAction(ctx, msg.User.ID, targetID, domain.MatchActionDislike); err != nil {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, err
		}
		return s.nextMatch(ctx, msg, l)
	case domain.CommandMatchReport:
		targetID, ok := s.parseTargetID(msg.Arguments)
		if !ok {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, nil
		}
		if _, err := s.search.RecordAction(ctx, msg.User.ID, targetID, domain.MatchActionReport); err != nil {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, err
		}
		return s.nextMatch(ctx, msg, l)
	case domain.CommandMatchPrevious:
		candidate, found, err := s.search.PreviousCandidate(ctx, msg.User.ID)
		if err != nil {
			return s.searchErrorResponse(ctx, msg, l, err), err
		}
		if !found {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchNoPreviousText(l)}}, nil
		}
		return s.matchProfileMessages(msg.ChatID, candidate, l), nil
	case domain.CommandMatchHistory:
		return s.historyListMessages(ctx, msg, l)
	case domain.CommandMatchHistoryView:
		return s.historyViewMessages(ctx, msg, l)
	case domain.CommandMatchHistoryLike:
		return s.historyActionMessages(ctx, msg, domain.MatchActionLike, l)
	case domain.CommandMatchHistoryDislike:
		return s.historyActionMessages(ctx, msg, domain.MatchActionDislike, l)
	case domain.CommandMatchHistoryReport:
		return s.historyActionMessages(ctx, msg, domain.MatchActionReport, l)
	case domain.CommandMatchLikes:
		return s.likesListMessages(ctx, msg, l)
	case domain.CommandMatchLikesView:
		return s.likesViewMessages(ctx, msg, l)
	case domain.CommandMatchLikesLike:
		return s.likesActionMessages(ctx, msg, domain.MatchActionLike, l)
	case domain.CommandMatchLikesDislike:
		return s.likesActionMessages(ctx, msg, domain.MatchActionDislike, l)
	case domain.CommandMatchLikesReport:
		return s.likesActionMessages(ctx, msg, domain.MatchActionReport, l)
	case domain.CommandMatchViewLike:
		return s.matchViewActionMessages(ctx, msg, domain.MatchActionLike, l)
	case domain.CommandMatchViewDislike:
		return s.matchViewActionMessages(ctx, msg, domain.MatchActionDislike, l)
	case domain.CommandMatchViewReport:
		return s.matchViewActionMessages(ctx, msg, domain.MatchActionReport, l)
	case domain.CommandMatchViewProfile:
		targetID, ok := s.parseTargetID(msg.Arguments)
		if !ok {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.matchProfileNotFoundText(l)}}, nil
		}
		hasProfile, err := s.viewerHasProfile(ctx, msg.User.ID)
		if err != nil {
			if isBannedError(err) {
				return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.userBannedText(l)}}, err
			}
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, err
		}
		if !hasProfile {
			return []domain.OutgoingMessage{
				{
					ChatID:         msg.ChatID,
					Text:           s.profileViewRequiresProfileText(l),
					InlineKeyboard: s.profileCreateInlineKeyboardWithAdmin(ctx, msg, l),
				},
			}, nil
		}
		profile, found, err := s.profiles.GetProfile(ctx, targetID)
		if err != nil {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.matchProfileNotFoundText(l)}}, err
		}
		if !found {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.matchProfileNotFoundText(l)}}, nil
		}
		return s.matchProfileViewMessages(msg.ChatID, profile, l), nil
	default:
		return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchStartRequiredText(l)}}, nil
	}
}

func (s *service) startSearch(ctx context.Context, msg domain.IncomingMessage, l localizer, gender domain.Gender, accuracy int) ([]domain.OutgoingMessage, error) {
	candidate, found, err := s.search.StartSearch(ctx, msg.User.ID, gender, accuracy)
	if err != nil {
		return s.searchErrorResponse(ctx, msg, l, err), err
	}
	if !found {
		return []domain.OutgoingMessage{{
			ChatID:         msg.ChatID,
			Text:           s.searchNoCandidatesText(l),
			InlineKeyboard: s.searchNoCandidatesInlineKeyboard(l),
		}}, nil
	}

	return s.matchProfileMessages(msg.ChatID, candidate, l), nil
}

func (s *service) nextMatch(ctx context.Context, msg domain.IncomingMessage, l localizer) ([]domain.OutgoingMessage, error) {
	candidate, found, err := s.search.NextCandidate(ctx, msg.User.ID)
	if err != nil {
		return s.searchErrorResponse(ctx, msg, l, err), err
	}
	if !found {
		return []domain.OutgoingMessage{{
			ChatID:         msg.ChatID,
			Text:           s.searchNoCandidatesText(l),
			InlineKeyboard: s.searchNoCandidatesInlineKeyboard(l),
		}}, nil
	}
	return s.matchProfileMessages(msg.ChatID, candidate, l), nil
}

func (s *service) searchErrorResponse(ctx context.Context, msg domain.IncomingMessage, l localizer, err error) []domain.OutgoingMessage {
	chatID := msg.ChatID
	if err == nil {
		return []domain.OutgoingMessage{{ChatID: chatID, Text: s.searchActionFailedText(l)}}
	}

	var status statusError
	if errors.As(err, &status) {
		switch status.StatusCode() {
		case http.StatusForbidden:
			return []domain.OutgoingMessage{{ChatID: chatID, Text: s.userBannedText(l)}}
		case http.StatusNotFound:
			return []domain.OutgoingMessage{
				{
					ChatID:         chatID,
					Text:           s.searchProfileMissingText(l),
					InlineKeyboard: s.profileCreateInlineKeyboardWithAdmin(ctx, msg, l),
				},
			}
		case http.StatusConflict:
			return []domain.OutgoingMessage{{ChatID: chatID, Text: s.searchStartRequiredText(l)}}
		case http.StatusServiceUnavailable:
			return []domain.OutgoingMessage{{
				ChatID:         chatID,
				Text:           s.searchAIUnavailableText(l),
				InlineKeyboard: s.searchAIUnavailableInlineKeyboard(l),
			}}
		}
	}

	return []domain.OutgoingMessage{{ChatID: chatID, Text: s.searchActionFailedText(l)}}
}

func (s *service) profileCreateInlineKeyboardWithAdmin(ctx context.Context, msg domain.IncomingMessage, l localizer) *domain.InlineKeyboard {
	role, _ := s.resolveAdminRole(ctx, msg.User)
	return s.withAdminMenuInlineKeyboard(l, s.profileCreateInlineKeyboard(l), role)
}

func (s *service) matchProfileMessages(chatID int64, candidate domain.MatchCandidate, l localizer) []domain.OutgoingMessage {
	return s.profileAlbumMessagesWithText(
		chatID,
		candidate.Profile,
		s.profilePreviewCard(l, candidate.Profile),
		s.matchActionsText(l),
		s.matchActionsInlineKeyboard(l, candidate.Profile.UserID, candidate.HasPrevious),
	)
}

func (s *service) matchProfileViewMessages(chatID int64, profile domain.Profile, l localizer) []domain.OutgoingMessage {
	return s.profileAlbumMessagesWithText(
		chatID,
		profile,
		s.profilePreviewCard(l, profile),
		s.matchActionsText(l),
		s.matchViewActionsInlineKeyboard(l, profile.UserID),
	)
}

func (s *service) historyListMessages(ctx context.Context, msg domain.IncomingMessage, l localizer) ([]domain.OutgoingMessage, error) {
	offset := s.parseHistoryOffset(msg.Arguments)
	list, err := s.search.History(ctx, msg.User.ID, historyPageSize, offset)
	if err != nil {
		return s.searchErrorResponse(ctx, msg, l, err), err
	}

	text := s.searchHistoryText(l, list)
	keyboard := s.historyInlineKeyboard(l, list)
	return []domain.OutgoingMessage{
		{
			ChatID:         msg.ChatID,
			Text:           text,
			InlineKeyboard: keyboard,
		},
	}, nil
}

func (s *service) historyViewMessages(ctx context.Context, msg domain.IncomingMessage, l localizer) ([]domain.OutgoingMessage, error) {
	targetID, offset, ok := s.parseHistoryTarget(msg.Arguments)
	if !ok {
		return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.matchProfileNotFoundText(l)}}, nil
	}

	item, list, found, err := s.historyItem(ctx, msg.User.ID, targetID, offset)
	if err != nil {
		return s.searchErrorResponse(ctx, msg, l, err), err
	}
	if !found {
		text := s.searchHistoryText(l, list)
		keyboard := s.historyInlineKeyboard(l, list)
		return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: text, InlineKeyboard: keyboard}}, nil
	}

	return s.historyProfileMessages(msg.ChatID, item, list.Offset, l), nil
}

func (s *service) historyActionMessages(ctx context.Context, msg domain.IncomingMessage, action domain.MatchAction, l localizer) ([]domain.OutgoingMessage, error) {
	targetID, offset, ok := s.parseHistoryTarget(msg.Arguments)
	if !ok {
		return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, nil
	}

	actionResult, err := s.search.RecordAction(ctx, msg.User.ID, targetID, action)
	if err != nil {
		return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, err
	}

	item, list, found, historyErr := s.historyItem(ctx, msg.User.ID, targetID, offset)
	if !found || historyErr != nil {
		if historyErr != nil {
			return s.searchErrorResponse(ctx, msg, l, historyErr), historyErr
		}
		text := s.searchHistoryText(l, list)
		keyboard := s.historyInlineKeyboard(l, list)
		return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: text, InlineKeyboard: keyboard}}, nil
	}

	if action != "" {
		item.Action = action
	}

	historyMessages := s.historyProfileMessages(msg.ChatID, item, list.Offset, l)
	if action != domain.MatchActionLike {
		return historyMessages, historyErr
	}
	if !actionResult.Matched {
		likeNotifications, likeErr := s.notifyPendingLikesForUser(ctx, targetID)
		if len(likeNotifications) > 0 {
			historyMessages = append(historyMessages, likeNotifications...)
		}
		return historyMessages, errors.Join(historyErr, likeErr)
	}

	currentMatchMessages, targetMatchMessages, matchErr := s.mutualMatchMessages(ctx, msg, targetID, l)
	messages := append(currentMatchMessages, historyMessages...)
	if len(targetMatchMessages) > 0 {
		messages = append(messages, targetMatchMessages...)
	}

	return messages, errors.Join(historyErr, matchErr)
}

func (s *service) likesListMessages(ctx context.Context, msg domain.IncomingMessage, l localizer) ([]domain.OutgoingMessage, error) {
	offset := s.parseHistoryOffset(msg.Arguments)
	return s.likesListMessagesWithOffset(ctx, msg, l, offset)
}

func (s *service) likesListMessagesWithOffset(ctx context.Context, msg domain.IncomingMessage, l localizer, offset int) ([]domain.OutgoingMessage, error) {
	hasProfile, err := s.viewerHasProfile(ctx, msg.User.ID)
	if err != nil {
		if isBannedError(err) {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.userBannedText(l)}}, err
		}
		return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, err
	}
	if !hasProfile {
		return []domain.OutgoingMessage{
			{
				ChatID:         msg.ChatID,
				Text:           s.searchProfileMissingText(l),
				InlineKeyboard: s.profileCreateInlineKeyboardWithAdmin(ctx, msg, l),
			},
		}, nil
	}

	list, err := s.search.ReceivedLikes(ctx, msg.User.ID, likesPageSize, offset)
	if err != nil {
		return s.searchErrorResponse(ctx, msg, l, err), err
	}

	if list.Total > 0 && list.Limit > 0 && len(list.Items) == 0 && list.Offset >= list.Total {
		lastOffset := (list.Total - 1) / list.Limit * list.Limit
		if lastOffset < 0 {
			lastOffset = 0
		}
		if lastOffset != list.Offset {
			list, err = s.search.ReceivedLikes(ctx, msg.User.ID, list.Limit, lastOffset)
			if err != nil {
				return s.searchErrorResponse(ctx, msg, l, err), err
			}
		}
	}

	text := s.likesListText(l, list)
	keyboard := s.likesInlineKeyboard(l, list)
	return []domain.OutgoingMessage{
		{
			ChatID:         msg.ChatID,
			Text:           text,
			InlineKeyboard: keyboard,
		},
	}, nil
}

func (s *service) likesViewMessages(ctx context.Context, msg domain.IncomingMessage, l localizer) ([]domain.OutgoingMessage, error) {
	targetID, offset, ok := s.parseHistoryTarget(msg.Arguments)
	if !ok {
		return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.matchProfileNotFoundText(l)}}, nil
	}

	hasProfile, err := s.viewerHasProfile(ctx, msg.User.ID)
	if err != nil {
		if isBannedError(err) {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.userBannedText(l)}}, err
		}
		return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, err
	}
	if !hasProfile {
		return []domain.OutgoingMessage{
			{
				ChatID:         msg.ChatID,
				Text:           s.profileViewRequiresProfileText(l),
				InlineKeyboard: s.profileCreateInlineKeyboardWithAdmin(ctx, msg, l),
			},
		}, nil
	}

	profile, found, err := s.profiles.GetProfile(ctx, targetID)
	if err != nil {
		return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.matchProfileNotFoundText(l)}}, err
	}
	if !found {
		return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.matchProfileNotFoundText(l)}}, nil
	}

	return s.likesProfileMessages(msg.ChatID, profile, offset, l), nil
}

func (s *service) likesActionMessages(ctx context.Context, msg domain.IncomingMessage, action domain.MatchAction, l localizer) ([]domain.OutgoingMessage, error) {
	targetID, offset, ok := s.parseHistoryTarget(msg.Arguments)
	if !ok {
		return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, nil
	}

	actionResult, err := s.search.RecordAction(ctx, msg.User.ID, targetID, action)
	if err != nil {
		return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, err
	}

	listMessages, listErr := s.likesListMessagesWithOffset(ctx, msg, l, offset)
	if action != domain.MatchActionLike {
		return listMessages, listErr
	}

	if actionResult.Matched {
		currentMatchMessages, targetMatchMessages, matchErr := s.mutualMatchMessages(ctx, msg, targetID, l)
		messages := append(currentMatchMessages, listMessages...)
		if len(targetMatchMessages) > 0 {
			messages = append(messages, targetMatchMessages...)
		}
		return messages, errors.Join(listErr, matchErr)
	}

	likeNotifications, likeErr := s.notifyPendingLikesForUser(ctx, targetID)
	if len(likeNotifications) > 0 {
		listMessages = append(listMessages, likeNotifications...)
	}
	return listMessages, errors.Join(listErr, likeErr)
}

func (s *service) matchViewActionMessages(ctx context.Context, msg domain.IncomingMessage, action domain.MatchAction, l localizer) ([]domain.OutgoingMessage, error) {
	targetID, ok := s.parseTargetID(msg.Arguments)
	if !ok {
		return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, nil
	}

	actionResult, err := s.search.RecordAction(ctx, msg.User.ID, targetID, action)
	if err != nil {
		return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText(l)}}, err
	}

	if action == domain.MatchActionLike && actionResult.Matched {
		currentMatchMessages, targetMatchMessages, matchErr := s.mutualMatchMessages(ctx, msg, targetID, l)
		messages := append([]domain.OutgoingMessage{}, currentMatchMessages...)
		if len(targetMatchMessages) > 0 {
			messages = append(messages, targetMatchMessages...)
		}
		return messages, matchErr
	}

	messages := []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.matchActionSavedText(l)}}
	if action != domain.MatchActionLike {
		return messages, nil
	}

	likeNotifications, likeErr := s.notifyPendingLikesForUser(ctx, targetID)
	if len(likeNotifications) > 0 {
		messages = append(messages, likeNotifications...)
	}
	return messages, likeErr
}

func (s *service) historyItem(ctx context.Context, userID, targetID int64, offset int) (domain.MatchHistoryItem, domain.MatchHistoryList, bool, error) {
	list, err := s.search.History(ctx, userID, historyPageSize, offset)
	if err != nil {
		return domain.MatchHistoryItem{}, domain.MatchHistoryList{}, false, err
	}

	for _, item := range list.Items {
		if item.Profile.UserID == targetID {
			return item, list, true, nil
		}
	}

	return domain.MatchHistoryItem{}, list, false, nil
}

func (s *service) historyProfileMessages(chatID int64, item domain.MatchHistoryItem, offset int, l localizer) []domain.OutgoingMessage {
	return s.profileAlbumMessagesWithText(
		chatID,
		item.Profile,
		s.profilePreviewCard(l, item.Profile),
		s.historyActionsText(l, item.Action),
		s.historyActionsInlineKeyboard(l, item.Profile.UserID, offset, item.Action),
	)
}

func (s *service) likesProfileMessages(chatID int64, profile domain.Profile, offset int, l localizer) []domain.OutgoingMessage {
	return s.profileAlbumMessagesWithText(
		chatID,
		profile,
		s.profilePreviewCard(l, profile),
		s.matchActionsText(l),
		s.likesActionsInlineKeyboard(l, profile.UserID, offset),
	)
}

func (s *service) pendingLikesFollowUp(ctx context.Context, msg domain.IncomingMessage, l localizer) ([]domain.OutgoingMessage, error) {
	if msg.Command != domain.CommandStart && msg.Command != domain.CommandSearchStart && msg.Command != domain.CommandSearchAI {
		return nil, nil
	}
	if s == nil || s.search == nil {
		return nil, nil
	}
	if msg.User.ID == 0 {
		return nil, nil
	}

	hasProfile, err := s.viewerHasProfile(ctx, msg.User.ID)
	if err != nil {
		return nil, err
	}
	if !hasProfile {
		return nil, nil
	}

	likes, err := s.search.PendingLikes(ctx, msg.User.ID)
	if err != nil {
		return nil, err
	}
	if len(likes) == 0 {
		return nil, nil
	}

	messages := make([]domain.OutgoingMessage, 0, len(likes))
	for _, like := range likes {
		messages = append(messages, domain.OutgoingMessage{
			ChatID:         msg.ChatID,
			Kind:           domain.OutgoingMessageKindLikeNotification,
			Text:           s.likeNotificationText(l, like),
			InlineKeyboard: s.matchViewInlineKeyboard(l, like.UserID),
		})
	}

	return messages, nil
}

func (s *service) notifyPendingLikesForUser(ctx context.Context, userID int64) ([]domain.OutgoingMessage, error) {
	if s == nil || s.search == nil || s.sessions == nil {
		return nil, nil
	}
	if userID <= 0 {
		return nil, nil
	}

	session, found, err := s.sessions.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !found || session.ChatID == 0 {
		return nil, nil
	}

	likes, err := s.search.PendingLikes(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(likes) == 0 {
		return nil, nil
	}

	l := s.localizerForUser(ctx, userID, domain.DefaultLanguage)
	messages := make([]domain.OutgoingMessage, 0, len(likes))
	for _, like := range likes {
		messages = append(messages, domain.OutgoingMessage{
			ChatID:         session.ChatID,
			Kind:           domain.OutgoingMessageKindLikeNotification,
			Text:           s.likeNotificationText(l, like),
			InlineKeyboard: s.matchViewInlineKeyboard(l, like.UserID),
		})
	}

	return messages, nil
}

func (s *service) mutualMatchMessages(ctx context.Context, msg domain.IncomingMessage, targetID int64, l localizer) ([]domain.OutgoingMessage, []domain.OutgoingMessage, error) {
	var errs []error

	viewerProfile := domain.Profile{UserID: msg.User.ID}
	targetProfile := domain.Profile{UserID: targetID}
	if s == nil || s.profiles == nil {
		errs = append(errs, errors.New("profile service is not configured"))
	} else {
		if profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID); err != nil {
			errs = append(errs, err)
		} else if found {
			viewerProfile = profile
		}
		if profile, found, err := s.profiles.GetProfile(ctx, targetID); err != nil {
			errs = append(errs, err)
		} else if found {
			targetProfile = profile
		}
	}

	targetSession := domain.Session{}
	targetSessionFound := false
	if s == nil || s.sessions == nil {
		errs = append(errs, errors.New("session repository is not configured"))
	} else {
		session, found, err := s.sessions.Get(ctx, targetID)
		if err != nil {
			errs = append(errs, err)
		} else {
			targetSession = session
			targetSessionFound = found
		}
	}

	currentNickname := s.nicknameFromSession(l, targetSession, targetProfile)
	currentMessages := []domain.OutgoingMessage{
		{
			ChatID: msg.ChatID,
			Kind:   domain.OutgoingMessageKindMatchNotification,
			Text:   s.matchSuccessText(l, targetProfile, currentNickname),
		},
	}

	targetMessages := []domain.OutgoingMessage{}
	if targetSessionFound && targetSession.ChatID != 0 {
		targetLocalizer := s.localizerForUser(ctx, targetID, domain.DefaultLanguage)
		viewerNickname := s.nicknameFromUser(targetLocalizer, msg.User, viewerProfile)
		targetMessages = append(targetMessages, domain.OutgoingMessage{
			ChatID: targetSession.ChatID,
			Kind:   domain.OutgoingMessageKindMatchNotification,
			Text:   s.matchSuccessText(targetLocalizer, viewerProfile, viewerNickname),
		})
	}

	return currentMessages, targetMessages, errors.Join(errs...)
}

func (s *service) viewerHasProfile(ctx context.Context, userID int64) (bool, error) {
	if s == nil || s.profiles == nil {
		return false, errors.New("profile service is not configured")
	}
	if userID <= 0 {
		return false, errors.New("user id is missing")
	}

	_, found, err := s.profiles.GetProfile(ctx, userID)
	if err != nil {
		return false, err
	}

	return found, nil
}
