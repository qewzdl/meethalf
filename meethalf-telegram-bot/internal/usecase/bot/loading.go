package bot

import (
	"context"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

type LoadingMessageProvider interface {
	LoadingMessage(ctx context.Context, msg domain.IncomingMessage) (domain.OutgoingMessage, bool, error)
}

func (s *service) LoadingMessage(ctx context.Context, msg domain.IncomingMessage) (domain.OutgoingMessage, bool, error) {
	if s == nil || msg.ChatID == 0 {
		return domain.OutgoingMessage{}, false, nil
	}

	l := s.localizerForMessage(ctx, msg)
	if msg.Command != "" {
		loading, ok := s.loadingForCommand(ctx, l, msg)
		return loading, ok, nil
	}

	loading, ok, err := s.loadingForAdminAction(ctx, msg, l)
	if err != nil || ok {
		return loading, ok, err
	}

	return s.loadingForDraft(ctx, msg, l)
}

func (s *service) loadingForCommand(ctx context.Context, l localizer, msg domain.IncomingMessage) (domain.OutgoingMessage, bool) {
	if s == nil || s.profiles == nil || msg.User.ID == 0 {
		return domain.OutgoingMessage{}, false
	}

	switch msg.Command {
	case domain.CommandStart:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingStart)}, true
	case domain.CommandProfileView:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingProfileView)}, true
	case domain.CommandProfilePreview:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingProfilePreview)}, true
	case domain.CommandProfileEditName:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingEditName)}, true
	case domain.CommandProfileEditGender:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingEditGender)}, true
	case domain.CommandProfileEditBirthDate:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingEditBirthDate)}, true
	case domain.CommandProfileEditCountry:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingEditCountry)}, true
	case domain.CommandProfileEditCity:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingEditCity)}, true
	case domain.CommandProfileEditDesc:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingEditDesc)}, true
	case domain.CommandProfileEditEmoji:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingEditEmoji)}, true
	case domain.CommandProfileEditPhotos:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingEditPhotos)}, true
	case domain.CommandProfileVisibility:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingProfileVisibility)}, true
	case domain.CommandProfileLikeNotifications:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingProfileLikeNotifications)}, true
	case domain.CommandProfileDeleteConfirm:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgDeletingProfile)}, true
	case domain.CommandSearchAccuracy:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchStart)}, true
	case domain.CommandSearchGender:
		if s.sessionSearchAccuracyEnabled(ctx, msg.User.ID) {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchStart)}, true
	case domain.CommandSearchAI:
		if s.sessionAISearchPending(ctx, msg.User.ID) && strings.TrimSpace(msg.Text) != "" {
			return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchAI)}, true
		}
		return domain.OutgoingMessage{}, false
	case domain.CommandSearchRefresh:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchNext)}, true
	case domain.CommandMatchLike:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchNext)}, true
	case domain.CommandMatchDislike:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchNext)}, true
	case domain.CommandMatchReport:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchNext)}, true
	case domain.CommandMatchViewLike:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchHistoryAction)}, true
	case domain.CommandMatchViewDislike:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchHistoryAction)}, true
	case domain.CommandMatchViewReport:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchHistoryAction)}, true
	case domain.CommandMatchPrevious:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchPrev)}, true
	case domain.CommandMatchHistory:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchHistory)}, true
	case domain.CommandMatchHistoryView:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchHistoryProfile)}, true
	case domain.CommandMatchHistoryLike:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchHistoryAction)}, true
	case domain.CommandMatchHistoryDislike:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchHistoryAction)}, true
	case domain.CommandMatchHistoryReport:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchHistoryAction)}, true
	case domain.CommandMatchLikes:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingLikesList)}, true
	case domain.CommandMatchLikesView:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingLikesProfile)}, true
	case domain.CommandMatchLikesLike:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchHistoryAction)}, true
	case domain.CommandMatchLikesDislike:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchHistoryAction)}, true
	case domain.CommandMatchLikesReport:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingSearchHistoryAction)}, true
	case domain.CommandAdminUsers:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminUsers)}, true
	case domain.CommandAdminBannedUsers:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminBannedUsers)}, true
	case domain.CommandAdminShadowBannedUsers:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminShadowBannedUsers)}, true
	case domain.CommandAdminHiddenUsers:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminHiddenUsers)}, true
	case domain.CommandAdminModerators:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminModerators)}, true
	case domain.CommandAdminReports:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminReports)}, true
	case domain.CommandAdminBan:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminBan)}, true
	case domain.CommandAdminUnban:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminUnban)}, true
	case domain.CommandAdminShadowBan:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminShadowBan)}, true
	case domain.CommandAdminShadowUnban:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminShadowUnban)}, true
	case domain.CommandAdminHideProfile:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminHideProfile)}, true
	case domain.CommandAdminShowProfile:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminShowProfile)}, true
	case domain.CommandAdminDeleteProfile:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminDeleteProfile)}, true
	case domain.CommandAdminModerator:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminModerator)}, true
	case domain.CommandAdminUnmoderator:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminUnmoderator)}, true
	case domain.CommandAdminResetChoices:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminResetChoices)}, true
	case domain.CommandAdminResetStart:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminResetStart)}, true
	case domain.CommandAdminClearReports:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminClearReports)}, true
	case domain.CommandAdminPostAd:
		return domain.OutgoingMessage{}, false
	case domain.CommandAdminAdSend:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgLoadingAdminAd)}, true
	default:
		return domain.OutgoingMessage{}, false
	}
}

func (s *service) loadingForAdminAction(ctx context.Context, msg domain.IncomingMessage, l localizer) (domain.OutgoingMessage, bool, error) {
	if s == nil || s.adminActions == nil || msg.User.ID == 0 {
		return domain.OutgoingMessage{}, false, nil
	}
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil || !role.canAccessPanel() {
		return domain.OutgoingMessage{}, false, nil
	}

	action, found, err := s.adminActions.Get(ctx, msg.User.ID)
	if err != nil || !found {
		return domain.OutgoingMessage{}, false, err
	}
	switch action.Action {
	case domain.AdminActionBan:
	case domain.AdminActionUnban:
	case domain.AdminActionShadowBan:
	case domain.AdminActionUnshadowBan:
	case domain.AdminActionHideProfile:
	case domain.AdminActionShowProfile:
	case domain.AdminActionDeleteProfile:
	case domain.AdminActionModerator:
	case domain.AdminActionUnmoderator:
	case domain.AdminActionResetChoices:
	case domain.AdminActionResetStart:
	case domain.AdminActionClearReports:
	case domain.AdminActionPostAd:
	case domain.AdminActionPostAdButton:
	default:
		return domain.OutgoingMessage{}, false, nil
	}
	if !role.allowsAdminAction(action.Action) {
		return domain.OutgoingMessage{}, false, nil
	}
	if action.ChatID != 0 && msg.ChatID != 0 && action.ChatID != msg.ChatID {
		return domain.OutgoingMessage{}, false, nil
	}

	if action.Action == domain.AdminActionPostAd || action.Action == domain.AdminActionPostAdButton {
		return domain.OutgoingMessage{}, false, nil
	} else {
		if _, _, ok := s.parseAdminUserIdentifier(msg.Text); !ok {
			return domain.OutgoingMessage{}, false, nil
		}
	}

	loadingText := l.message(msgLoadingAdminBan)
	switch action.Action {
	case domain.AdminActionUnban:
		loadingText = l.message(msgLoadingAdminUnban)
	case domain.AdminActionShadowBan:
		loadingText = l.message(msgLoadingAdminShadowBan)
	case domain.AdminActionUnshadowBan:
		loadingText = l.message(msgLoadingAdminShadowUnban)
	case domain.AdminActionHideProfile:
		loadingText = l.message(msgLoadingAdminHideProfile)
	case domain.AdminActionShowProfile:
		loadingText = l.message(msgLoadingAdminShowProfile)
	case domain.AdminActionDeleteProfile:
		loadingText = l.message(msgLoadingAdminDeleteProfile)
	case domain.AdminActionModerator:
		loadingText = l.message(msgLoadingAdminModerator)
	case domain.AdminActionUnmoderator:
		loadingText = l.message(msgLoadingAdminUnmoderator)
	case domain.AdminActionResetChoices:
		loadingText = l.message(msgLoadingAdminResetChoices)
	case domain.AdminActionResetStart:
		loadingText = l.message(msgLoadingAdminResetStart)
	case domain.AdminActionClearReports:
		loadingText = l.message(msgLoadingAdminClearReports)
	case domain.AdminActionPostAd:
		loadingText = l.message(msgLoadingAdminAd)
	}
	return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingText}, true, nil
}

func (s *service) loadingForDraft(ctx context.Context, msg domain.IncomingMessage, l localizer) (domain.OutgoingMessage, bool, error) {
	if s == nil || s.drafts == nil || s.profiles == nil || msg.User.ID == 0 {
		return domain.OutgoingMessage{}, false, nil
	}

	draft, found, err := s.drafts.Get(ctx, msg.User.ID)
	if err != nil || !found {
		return domain.OutgoingMessage{}, false, err
	}

	if s.draftMode(draft) == domain.ProfileDraftModeEdit {
		if s.editDraftWillSave(msg, draft) {
			return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgUpdatingProfile)}, true, nil
		}
		return domain.OutgoingMessage{}, false, nil
	}

	if draft.Step == domain.ProfileDraftStepPhotos && s.albumDoneWillSave(msg, draft) {
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: l.message(msgCreatingProfile)}, true, nil
	}

	return domain.OutgoingMessage{}, false, nil
}

func (s *service) editDraftWillSave(msg domain.IncomingMessage, draft domain.ProfileDraft) bool {
	switch draft.Step {
	case domain.ProfileDraftStepName:
		value := strings.TrimSpace(msg.Text)
		if value == "" {
			return false
		}
		name := value
		if s.isAffirmative(value) {
			name = s.userFullName(msg.User)
			if name == "" {
				return false
			}
		}
		return len(name) <= maxNameLength
	case domain.ProfileDraftStepGender:
		_, ok := s.normalizeGender(msg.Text)
		return ok
	case domain.ProfileDraftStepBirthDate:
		birthDate, ok := s.parseBirthDate(msg.Text)
		if !ok {
			return false
		}
		age := s.ageFromBirthDate(birthDate, s.now(msg.ReceivedAt))
		if age < minAge || age > maxAge {
			return false
		}
		updated := draft
		updated.BirthDate = birthDate
		return s.nextRequiredStep(updated) == ""
	case domain.ProfileDraftStepCountry:
		country, ok := s.normalizeCountry(msg.Text)
		if !ok {
			return false
		}
		updated := draft
		previousCountry := updated.Country
		updated.Country = country
		if previousCountry != country {
			updated.City = ""
		} else if updated.City != "" {
			if normalizedCity, ok := s.normalizeCity(country, updated.City); ok {
				updated.City = normalizedCity
			} else {
				updated.City = ""
			}
		}
		return s.nextRequiredStep(updated) == ""
	case domain.ProfileDraftStepCity:
		city, ok := s.normalizeCity(draft.Country, msg.Text)
		if !ok {
			return false
		}
		updated := draft
		updated.City = city
		return s.nextRequiredStep(updated) == ""
	case domain.ProfileDraftStepDescription:
		description := strings.TrimSpace(msg.Text)
		if description == "" || len(description) > maxDescriptionLength {
			return false
		}
		return true
	case domain.ProfileDraftStepEmoji:
		code, ok := s.normalizeEmojiCode(msg.Text)
		if !ok {
			return false
		}
		updated := draft
		updated.EmojiCode = code
		return s.nextRequiredStep(updated) == ""
	case domain.ProfileDraftStepPhotos:
		return s.albumDoneWillSave(msg, draft)
	default:
		return false
	}
}

func (s *service) albumDoneWillSave(msg domain.IncomingMessage, draft domain.ProfileDraft) bool {
	if !s.isAlbumDone(msg.Text) {
		return false
	}

	photos, _ := s.mergePhotoIDs(draft.Photos, msg.PhotoIDs, maxPhotos)
	return len(photos) >= minPhotos
}
