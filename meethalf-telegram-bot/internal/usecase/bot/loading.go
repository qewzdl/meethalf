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

	if msg.Command != "" {
		loading, ok := s.loadingForCommand(msg)
		return loading, ok, nil
	}

	loading, ok, err := s.loadingForAdminAction(ctx, msg)
	if err != nil || ok {
		return loading, ok, err
	}

	return s.loadingForDraft(ctx, msg)
}

func (s *service) loadingForCommand(msg domain.IncomingMessage) (domain.OutgoingMessage, bool) {
	if s == nil || s.profiles == nil || msg.User.ID == 0 {
		return domain.OutgoingMessage{}, false
	}

	switch msg.Command {
	case domain.CommandStart:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingStartText}, true
	case domain.CommandProfileView:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingProfileViewText}, true
	case domain.CommandProfilePreview:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingProfilePreviewText}, true
	case domain.CommandProfileEditName:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingEditNameText}, true
	case domain.CommandProfileEditGender:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingEditGenderText}, true
	case domain.CommandProfileEditBirthDate:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingEditBirthDateText}, true
	case domain.CommandProfileEditCountry:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingEditCountryText}, true
	case domain.CommandProfileEditCity:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingEditCityText}, true
	case domain.CommandProfileEditDesc:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingEditDescText}, true
	case domain.CommandProfileEditEmoji:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingEditEmojiText}, true
	case domain.CommandProfileEditPhotos:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingEditPhotosText}, true
	case domain.CommandProfileVisibility:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingProfileVisibilityText}, true
	case domain.CommandProfileDeleteConfirm:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: deletingProfileText}, true
	case domain.CommandSearchAccuracy:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingSearchStartText}, true
	case domain.CommandSearchRefresh:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingSearchNextText}, true
	case domain.CommandMatchLike:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingSearchNextText}, true
	case domain.CommandMatchDislike:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingSearchNextText}, true
	case domain.CommandMatchReport:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingSearchNextText}, true
	case domain.CommandMatchPrevious:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingSearchPrevText}, true
	case domain.CommandMatchHistory:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingSearchHistoryText}, true
	case domain.CommandMatchHistoryView:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingSearchHistoryProfileText}, true
	case domain.CommandMatchHistoryLike:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingSearchHistoryActionText}, true
	case domain.CommandMatchHistoryDislike:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingSearchHistoryActionText}, true
	case domain.CommandMatchHistoryReport:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingSearchHistoryActionText}, true
	case domain.CommandAdminUsers:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingAdminUsersText}, true
	case domain.CommandAdminBannedUsers:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingAdminBannedUsersText}, true
	case domain.CommandAdminModerators:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingAdminModeratorsText}, true
	case domain.CommandAdminReports:
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingAdminReportsText}, true
	case domain.CommandAdminBan:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingAdminBanText}, true
	case domain.CommandAdminUnban:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingAdminUnbanText}, true
	case domain.CommandAdminModerator:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingAdminModeratorText}, true
	case domain.CommandAdminUnmoderator:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingAdminUnmoderatorText}, true
	case domain.CommandAdminResetChoices:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingAdminResetChoicesText}, true
	case domain.CommandAdminClearReports:
		if strings.TrimSpace(msg.Arguments) == "" {
			return domain.OutgoingMessage{}, false
		}
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingAdminClearReportsText}, true
	default:
		return domain.OutgoingMessage{}, false
	}
}

func (s *service) loadingForAdminAction(ctx context.Context, msg domain.IncomingMessage) (domain.OutgoingMessage, bool, error) {
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
	case domain.AdminActionModerator:
	case domain.AdminActionUnmoderator:
	case domain.AdminActionResetChoices:
	case domain.AdminActionClearReports:
	default:
		return domain.OutgoingMessage{}, false, nil
	}
	if !role.allowsAdminAction(action.Action) {
		return domain.OutgoingMessage{}, false, nil
	}
	if action.ChatID != 0 && msg.ChatID != 0 && action.ChatID != msg.ChatID {
		return domain.OutgoingMessage{}, false, nil
	}

	if _, _, ok := s.parseAdminUserIdentifier(msg.Text); !ok {
		return domain.OutgoingMessage{}, false, nil
	}

	loadingText := loadingAdminBanText
	switch action.Action {
	case domain.AdminActionUnban:
		loadingText = loadingAdminUnbanText
	case domain.AdminActionModerator:
		loadingText = loadingAdminModeratorText
	case domain.AdminActionUnmoderator:
		loadingText = loadingAdminUnmoderatorText
	case domain.AdminActionResetChoices:
		loadingText = loadingAdminResetChoicesText
	case domain.AdminActionClearReports:
		loadingText = loadingAdminClearReportsText
	}
	return domain.OutgoingMessage{ChatID: msg.ChatID, Text: loadingText}, true, nil
}

func (s *service) loadingForDraft(ctx context.Context, msg domain.IncomingMessage) (domain.OutgoingMessage, bool, error) {
	if s == nil || s.drafts == nil || s.profiles == nil || msg.User.ID == 0 {
		return domain.OutgoingMessage{}, false, nil
	}

	draft, found, err := s.drafts.Get(ctx, msg.User.ID)
	if err != nil || !found {
		return domain.OutgoingMessage{}, false, err
	}

	if s.draftMode(draft) == domain.ProfileDraftModeEdit {
		if s.editDraftWillSave(msg, draft) {
			return domain.OutgoingMessage{ChatID: msg.ChatID, Text: updatingProfileText}, true, nil
		}
		return domain.OutgoingMessage{}, false, nil
	}

	if draft.Step == domain.ProfileDraftStepPhotos && s.albumDoneWillSave(msg, draft) {
		return domain.OutgoingMessage{ChatID: msg.ChatID, Text: creatingProfileText}, true, nil
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
