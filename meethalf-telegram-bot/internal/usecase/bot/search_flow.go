package bot

import (
	"context"
	"errors"
	"net/http"

	"meethalf-telegram-bot/internal/domain"
)

type statusError interface {
	error
	StatusCode() int
}

func (s *service) handleSearch(ctx context.Context, msg domain.IncomingMessage) ([]domain.OutgoingMessage, error) {
	if s == nil || s.search == nil {
		return []domain.OutgoingMessage{
			{ChatID: msg.ChatID, Text: s.searchUnavailableText()},
		}, errors.New("search service is not configured")
	}
	if msg.User.ID == 0 {
		return []domain.OutgoingMessage{
			{ChatID: msg.ChatID, Text: s.searchUnavailableText()},
		}, errors.New("user id is missing")
	}

	switch msg.Command {
	case domain.CommandSearchStart:
		hasProfile, err := s.viewerHasProfile(ctx, msg.User.ID)
		if err != nil {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText()}}, err
		}
		if !hasProfile {
			return []domain.OutgoingMessage{
				{
					ChatID:         msg.ChatID,
					Text:           s.searchProfileMissingText(),
					InlineKeyboard: s.profileCreateInlineKeyboard(),
				},
			}, nil
		}
		return []domain.OutgoingMessage{
			{
				ChatID:         msg.ChatID,
				Text:           s.searchGenderText(),
				InlineKeyboard: s.searchGenderInlineKeyboard(),
			},
		}, nil
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
					Text:           s.searchGenderText(),
					InlineKeyboard: s.searchGenderInlineKeyboard(),
				},
			}, nil
		}

		return []domain.OutgoingMessage{
			{
				ChatID:         msg.ChatID,
				Text:           s.searchAccuracyText(),
				InlineKeyboard: s.searchAccuracyInlineKeyboard(gender),
			},
		}, nil
	case domain.CommandSearchAccuracy:
		gender, accuracy, ok := s.parseSearchAccuracyArgs(msg.Arguments)
		if !ok {
			return []domain.OutgoingMessage{
				{
					ChatID:         msg.ChatID,
					Text:           s.searchGenderText(),
					InlineKeyboard: s.searchGenderInlineKeyboard(),
				},
			}, nil
		}

		candidate, found, err := s.search.StartSearch(ctx, msg.User.ID, gender, accuracy)
		if err != nil {
			return s.searchErrorResponse(msg.ChatID, err), err
		}
		if !found {
			return []domain.OutgoingMessage{{
				ChatID:         msg.ChatID,
				Text:           s.searchNoCandidatesText(),
				InlineKeyboard: s.searchNoCandidatesInlineKeyboard(),
			}}, nil
		}

		return s.matchProfileMessages(msg.ChatID, candidate), nil
	case domain.CommandSearchRefresh:
		return s.nextMatch(ctx, msg)
	case domain.CommandMatchLike:
		targetID, ok := s.parseTargetID(msg.Arguments)
		if !ok {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText()}}, nil
		}
		actionResult, err := s.search.RecordAction(ctx, msg.User.ID, targetID, domain.MatchActionLike)
		if err != nil {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText()}}, err
		}
		nextMessages, nextErr := s.nextMatch(ctx, msg)
		if actionResult.Matched {
			currentMatchMessages, targetMatchMessages, matchErr := s.mutualMatchMessages(ctx, msg, targetID)
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
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText()}}, nil
		}
		if _, err := s.search.RecordAction(ctx, msg.User.ID, targetID, domain.MatchActionDislike); err != nil {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText()}}, err
		}
		return s.nextMatch(ctx, msg)
	case domain.CommandMatchReport:
		targetID, ok := s.parseTargetID(msg.Arguments)
		if !ok {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText()}}, nil
		}
		if _, err := s.search.RecordAction(ctx, msg.User.ID, targetID, domain.MatchActionReport); err != nil {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText()}}, err
		}
		return s.nextMatch(ctx, msg)
	case domain.CommandMatchPrevious:
		candidate, found, err := s.search.PreviousCandidate(ctx, msg.User.ID)
		if err != nil {
			return s.searchErrorResponse(msg.ChatID, err), err
		}
		if !found {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchNoPreviousText()}}, nil
		}
		return s.matchProfileMessages(msg.ChatID, candidate), nil
	case domain.CommandMatchViewProfile:
		targetID, ok := s.parseTargetID(msg.Arguments)
		if !ok {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.matchProfileNotFoundText()}}, nil
		}
		hasProfile, err := s.viewerHasProfile(ctx, msg.User.ID)
		if err != nil {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchActionFailedText()}}, err
		}
		if !hasProfile {
			return []domain.OutgoingMessage{
				{
					ChatID:         msg.ChatID,
					Text:           s.profileViewRequiresProfileText(),
					InlineKeyboard: s.profileCreateInlineKeyboard(),
				},
			}, nil
		}
		profile, found, err := s.profiles.GetProfile(ctx, targetID)
		if err != nil {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.matchProfileNotFoundText()}}, err
		}
		if !found {
			return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.matchProfileNotFoundText()}}, nil
		}
		return s.matchProfileViewMessages(msg.ChatID, profile), nil
	default:
		return []domain.OutgoingMessage{{ChatID: msg.ChatID, Text: s.searchStartRequiredText()}}, nil
	}
}

func (s *service) nextMatch(ctx context.Context, msg domain.IncomingMessage) ([]domain.OutgoingMessage, error) {
	candidate, found, err := s.search.NextCandidate(ctx, msg.User.ID)
	if err != nil {
		return s.searchErrorResponse(msg.ChatID, err), err
	}
	if !found {
		return []domain.OutgoingMessage{{
			ChatID:         msg.ChatID,
			Text:           s.searchNoCandidatesText(),
			InlineKeyboard: s.searchNoCandidatesInlineKeyboard(),
		}}, nil
	}
	return s.matchProfileMessages(msg.ChatID, candidate), nil
}

func (s *service) searchErrorResponse(chatID int64, err error) []domain.OutgoingMessage {
	if err == nil {
		return []domain.OutgoingMessage{{ChatID: chatID, Text: s.searchActionFailedText()}}
	}

	var status statusError
	if errors.As(err, &status) {
		switch status.StatusCode() {
		case http.StatusNotFound:
			return []domain.OutgoingMessage{
				{
					ChatID:         chatID,
					Text:           s.searchProfileMissingText(),
					InlineKeyboard: s.profileCreateInlineKeyboard(),
				},
			}
		case http.StatusConflict:
			return []domain.OutgoingMessage{{ChatID: chatID, Text: s.searchStartRequiredText()}}
		}
	}

	return []domain.OutgoingMessage{{ChatID: chatID, Text: s.searchActionFailedText()}}
}

func (s *service) matchProfileMessages(chatID int64, candidate domain.MatchCandidate) []domain.OutgoingMessage {
	return s.profileAlbumMessagesWithText(
		chatID,
		candidate.Profile,
		s.profilePreviewCard(candidate.Profile),
		s.matchActionsText(),
		s.matchActionsInlineKeyboard(candidate.Profile.UserID, candidate.HasPrevious),
	)
}

func (s *service) matchProfileViewMessages(chatID int64, profile domain.Profile) []domain.OutgoingMessage {
	return s.profileAlbumMessagesWithText(
		chatID,
		profile,
		s.profilePreviewCard(profile),
		"",
		nil,
	)
}

func (s *service) pendingLikesFollowUp(ctx context.Context, msg domain.IncomingMessage) ([]domain.OutgoingMessage, error) {
	if msg.Command != domain.CommandStart && msg.Command != domain.CommandSearchStart {
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
			Text:           s.likeNotificationText(like),
			InlineKeyboard: s.matchViewInlineKeyboard(like.UserID),
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

	messages := make([]domain.OutgoingMessage, 0, len(likes))
	for _, like := range likes {
		messages = append(messages, domain.OutgoingMessage{
			ChatID:         session.ChatID,
			Kind:           domain.OutgoingMessageKindLikeNotification,
			Text:           s.likeNotificationText(like),
			InlineKeyboard: s.matchViewInlineKeyboard(like.UserID),
		})
	}

	return messages, nil
}

func (s *service) mutualMatchMessages(ctx context.Context, msg domain.IncomingMessage, targetID int64) ([]domain.OutgoingMessage, []domain.OutgoingMessage, error) {
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

	currentNickname := s.nicknameFromSession(targetSession, targetProfile)
	currentMessages := []domain.OutgoingMessage{
		{
			ChatID: msg.ChatID,
			Kind:   domain.OutgoingMessageKindMatchNotification,
			Text:   s.matchSuccessText(targetProfile, currentNickname),
		},
	}

	targetMessages := []domain.OutgoingMessage{}
	if targetSessionFound && targetSession.ChatID != 0 {
		viewerNickname := s.nicknameFromUser(msg.User, viewerProfile)
		targetMessages = append(targetMessages, domain.OutgoingMessage{
			ChatID: targetSession.ChatID,
			Kind:   domain.OutgoingMessageKindMatchNotification,
			Text:   s.matchSuccessText(viewerProfile, viewerNickname),
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
