package bot

import (
	"fmt"
	"strings"
	"time"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) namePrompt(l localizer, user domain.User) string {
	header := s.stepHeader(l, domain.ProfileDraftStepName)
	telegramName := s.userFullName(user)
	if telegramName == "" {
		return header + "\n" + l.message(msgNamePromptNoTelegram)
	}

	return fmt.Sprintf("%s\n%s", header, l.message(msgNamePromptWithTelegram, telegramName))
}

func (s *service) botCheckPrompt(l localizer, question string) string {
	return s.botCheckRetryPrompt(l, "", question)
}

func (s *service) botCheckRetryPrompt(l localizer, reason, question string) string {
	text := l.message(msgBotCheckPrompt, strings.TrimSpace(question))
	if strings.TrimSpace(reason) != "" {
		text = reason + "\n" + text
	}
	return s.stepText(l, domain.ProfileDraftStepBotCheck, text)
}

func (s *service) birthDatePrompt(l localizer) string {
	return s.stepText(l, domain.ProfileDraftStepBirthDate, l.message(msgBirthDatePrompt, birthDateLayout, birthDateExample))
}

func (s *service) genderPrompt(l localizer) string {
	return s.stepText(l, domain.ProfileDraftStepGender, l.message(msgGenderPrompt))
}

func (s *service) countryPrompt(l localizer) string {
	return s.stepText(l, domain.ProfileDraftStepCountry, l.message(msgCountryPrompt))
}

func (s *service) cityPrompt(l localizer) string {
	return s.stepText(l, domain.ProfileDraftStepCity, l.message(msgCityPrompt))
}

func (s *service) descriptionPrompt(l localizer) string {
	return s.stepText(l, domain.ProfileDraftStepDescription, l.message(msgDescriptionPrompt))
}

func (s *service) emojiPrompt(l localizer) string {
	return s.stepText(l, domain.ProfileDraftStepEmoji, l.message(msgEmojiPrompt))
}

func (s *service) photosPrompt(l localizer) string {
	return s.stepText(l, domain.ProfileDraftStepPhotos, l.message(msgPhotosPrompt))
}

func (s *service) profileEditMenuText(l localizer) string {
	return l.message(msgProfileEditMenu)
}

func (s *service) adminMenuText(l localizer) string {
	return l.message(msgAdminMenu)
}

func (s *service) adminMenuTextForRole(l localizer, role adminRole) string {
	if role == adminRoleModerator {
		return l.message(msgModeratorMenu)
	}
	return l.message(msgAdminMenu)
}

func (s *service) adminAccessDeniedText(l localizer) string {
	return l.message(msgAdminAccessDenied)
}

func (s *service) adminModerationRestrictedText(l localizer) string {
	return l.message(msgAdminModerationRestricted)
}

func (s *service) adminUsersEmptyText(l localizer) string {
	return l.message(msgAdminUsersEmpty)
}

func (s *service) adminUsersLoadFailedText(l localizer) string {
	return l.message(msgAdminUsersLoadFailed)
}

func (s *service) adminBannedUsersEmptyText(l localizer) string {
	return l.message(msgAdminBannedUsersEmpty)
}

func (s *service) adminBannedUsersLoadFailedText(l localizer) string {
	return l.message(msgAdminBannedUsersLoadFailed)
}

func (s *service) adminShadowBannedUsersEmptyText(l localizer) string {
	return l.message(msgAdminShadowBannedUsersEmpty)
}

func (s *service) adminShadowBannedUsersLoadFailedText(l localizer) string {
	return l.message(msgAdminShadowBannedUsersLoadFailed)
}

func (s *service) adminModeratorsEmptyText(l localizer) string {
	return l.message(msgAdminModeratorsEmpty)
}

func (s *service) adminModeratorsLoadFailedText(l localizer) string {
	return l.message(msgAdminModeratorsLoadFailed)
}

func (s *service) adminReportsEmptyText(l localizer) string {
	return l.message(msgAdminReportsEmpty)
}

func (s *service) adminReportsLoadFailedText(l localizer) string {
	return l.message(msgAdminReportsLoadFailed)
}

func (s *service) adminHiddenUsersLoadFailedText(l localizer) string {
	return l.message(msgAdminHiddenUsersLoadFailed)
}

func (s *service) adminBanUsageText(l localizer) string {
	return l.message(msgAdminBanUsage)
}

func (s *service) adminBanFailedText(l localizer) string {
	return l.message(msgAdminBanFailed)
}

func (s *service) adminUserNotFoundText(l localizer) string {
	return l.message(msgAdminUserNotFound)
}

func (s *service) adminBanSuccessText(l localizer, userRef string) string {
	return l.message(msgAdminBanSuccess, userRef)
}

func (s *service) adminUnbanUsageText(l localizer) string {
	return l.message(msgAdminUnbanUsage)
}

func (s *service) adminUnbanFailedText(l localizer) string {
	return l.message(msgAdminUnbanFailed)
}

func (s *service) adminUnbanSuccessText(l localizer, userRef string) string {
	return l.message(msgAdminUnbanSuccess, userRef)
}

func (s *service) adminShadowBanUsageText(l localizer) string {
	return l.message(msgAdminShadowBanUsage)
}

func (s *service) adminShadowBanFailedText(l localizer) string {
	return l.message(msgAdminShadowBanFailed)
}

func (s *service) adminShadowBanSuccessText(l localizer, userRef string) string {
	return l.message(msgAdminShadowBanSuccess, userRef)
}

func (s *service) adminShadowUnbanUsageText(l localizer) string {
	return l.message(msgAdminShadowUnbanUsage)
}

func (s *service) adminShadowUnbanFailedText(l localizer) string {
	return l.message(msgAdminShadowUnbanFailed)
}

func (s *service) adminShadowUnbanSuccessText(l localizer, userRef string) string {
	return l.message(msgAdminShadowUnbanSuccess, userRef)
}

func (s *service) adminHideProfileUsageText(l localizer) string {
	return l.message(msgAdminHideProfileUsage)
}

func (s *service) adminHideProfileFailedText(l localizer) string {
	return l.message(msgAdminHideProfileFailed)
}

func (s *service) adminHideProfileSuccessText(l localizer, userRef string) string {
	return l.message(msgAdminHideProfileSuccess, userRef)
}

func (s *service) adminShowProfileUsageText(l localizer) string {
	return l.message(msgAdminShowProfileUsage)
}

func (s *service) adminShowProfileFailedText(l localizer) string {
	return l.message(msgAdminShowProfileFailed)
}

func (s *service) adminShowProfileSuccessText(l localizer, userRef string) string {
	return l.message(msgAdminShowProfileSuccess, userRef)
}

func (s *service) adminActionFailedText(l localizer) string {
	return l.message(msgAdminActionFailed)
}

func (s *service) adminModeratorUsageText(l localizer) string {
	return l.message(msgAdminModeratorUsage)
}

func (s *service) adminModeratorFailedText(l localizer) string {
	return l.message(msgAdminModeratorFailed)
}

func (s *service) adminModeratorSuccessText(l localizer, userRef string) string {
	return l.message(msgAdminModeratorSuccess, userRef)
}

func (s *service) adminUnmoderatorUsageText(l localizer) string {
	return l.message(msgAdminUnmoderatorUsage)
}

func (s *service) adminUnmoderatorFailedText(l localizer) string {
	return l.message(msgAdminUnmoderatorFailed)
}

func (s *service) adminUnmoderatorSuccessText(l localizer, userRef string) string {
	return l.message(msgAdminUnmoderatorSuccess, userRef)
}

func (s *service) adminResetChoicesUsageText(l localizer) string {
	return l.message(msgAdminResetChoicesUsage)
}

func (s *service) adminResetChoicesFailedText(l localizer) string {
	return l.message(msgAdminResetChoicesFailed)
}

func (s *service) adminResetChoicesSuccessText(l localizer, userRef string) string {
	return l.message(msgAdminResetChoicesSuccess, userRef)
}

func (s *service) adminResetStartUsageText(l localizer) string {
	return l.message(msgAdminResetStartUsage)
}

func (s *service) adminResetStartFailedText(l localizer) string {
	return l.message(msgAdminResetStartFailed)
}

func (s *service) adminResetStartSuccessText(l localizer, userRef string) string {
	return l.message(msgAdminResetStartSuccess, userRef)
}

func (s *service) adminClearReportsUsageText(l localizer) string {
	return l.message(msgAdminClearReportsUsage)
}

func (s *service) adminClearReportsFailedText(l localizer) string {
	return l.message(msgAdminClearReportsFailed)
}

func (s *service) adminClearReportsSuccessText(l localizer, userRef string) string {
	return l.message(msgAdminClearReportsSuccess, userRef)
}

func (s *service) userBannedText(l localizer) string {
	return l.message(msgUserBanned)
}

func (s *service) ageConfirmationPromptText(l localizer) string {
	return l.message(msgAgeConfirmPrompt, minAge)
}

func (s *service) ageConfirmationDeclinedText(l localizer) string {
	return l.message(msgAgeConfirmDeclined, minAge)
}

func (s *service) ageConfirmationFailedText(l localizer) string {
	return l.message(msgAgeConfirmSaveFailed)
}

func (s *service) profileSettingsText(l localizer) string {
	return l.message(msgProfileSettingsLanguageHint)
}

func (s *service) profileActionsText(l localizer) string {
	return l.message(msgProfileActions)
}

func (s *service) profilePreviewActionsText(l localizer) string {
	return l.message(msgProfilePreviewActions)
}

func (s *service) editPrompt(l localizer, step domain.ProfileDraftStep) string {
	switch step {
	case domain.ProfileDraftStepName:
		return s.editText(l, step, l.message(msgEditNamePrompt))
	case domain.ProfileDraftStepBirthDate:
		return s.editText(l, step, l.message(msgEditBirthDatePrompt, birthDateLayout, birthDateExample))
	case domain.ProfileDraftStepGender:
		return s.editText(l, step, l.message(msgEditGenderPrompt))
	case domain.ProfileDraftStepCountry:
		return s.editText(l, step, l.message(msgEditCountryPrompt))
	case domain.ProfileDraftStepCity:
		return s.editText(l, step, l.message(msgEditCityPrompt))
	case domain.ProfileDraftStepDescription:
		return s.editText(l, step, l.message(msgEditDescriptionPrompt))
	case domain.ProfileDraftStepEmoji:
		return s.editText(l, step, l.message(msgEditEmojiPrompt))
	case domain.ProfileDraftStepPhotos:
		return s.editText(l, step, l.message(msgEditPhotosPrompt))
	default:
		return s.editText(l, step, l.message(msgEditDefaultPrompt))
	}
}

func (s *service) profileCreated(l localizer) string {
	return l.message(msgProfileCreated)
}

func (s *service) profileUpdated(l localizer) string {
	return l.message(msgProfileUpdated)
}

func (s *service) profileDeleteConfirmText(l localizer) string {
	return l.message(msgProfileDeleteConfirm)
}

func (s *service) profileDeleteCanceledText(l localizer) string {
	return l.message(msgProfileDeleteCanceled)
}

func (s *service) profileDeleteExpiredText(l localizer) string {
	return l.message(msgProfileDeleteExpired)
}

func (s *service) profileSetupCanceledText(l localizer) string {
	return l.message(msgProfileSetupCanceled)
}

func (s *service) profileEditCanceledText(l localizer) string {
	return l.message(msgProfileEditCanceled)
}

func (s *service) actionCanceledText(l localizer) string {
	return l.message(msgActionCanceled)
}

func (s *service) profileVisibilityUpdateFailedText(l localizer) string {
	return l.message(msgProfileVisibilityUpdateFailed)
}

func (s *service) profileSettingsTextWithVisibility(l localizer, isHidden bool, searchAccuracyEnabled bool) string {
	status := s.profileVisibilityStatus(l, isHidden)
	searchStatus := s.searchAccuracyStatus(l, searchAccuracyEnabled)
	return l.message(msgProfileSettingsWithVisibility, status, searchStatus)
}

func (s *service) profileVisibilityStatus(l localizer, isHidden bool) string {
	if isHidden {
		return l.label(labelStatusHidden)
	}
	return l.label(labelStatusVisible)
}

func (s *service) searchAccuracyStatus(l localizer, enabled bool) string {
	if enabled {
		return l.label(labelStatusEnabled)
	}
	return l.label(labelStatusDisabled)
}

func (s *service) profileVisibilityUpdated(l localizer, isHidden bool) string {
	if isHidden {
		return l.message(msgProfileHidden)
	}
	return l.message(msgProfileVisible)
}

func (s *service) searchGenderText(l localizer) string {
	return l.message(msgSearchGenderPrompt)
}

func (s *service) searchAccuracyText(l localizer) string {
	return l.message(msgSearchAccuracyPrompt)
}

func (s *service) searchAccuracyEnabledText(l localizer) string {
	return l.message(msgSearchAccuracyEnabled)
}

func (s *service) searchAccuracyDisabledText(l localizer) string {
	return l.message(msgSearchAccuracyDisabled)
}

func (s *service) searchAccuracyUpdateFailedText(l localizer) string {
	return l.message(msgSearchAccuracyUpdateFailed)
}

func (s *service) searchAIPromptText(l localizer) string {
	return l.message(msgSearchAIPrompt)
}

func (s *service) searchAITooShortText(l localizer) string {
	return l.message(msgSearchAITooShort)
}

func (s *service) searchAIUnavailableText(l localizer) string {
	return l.message(msgSearchAIUnavailable)
}

func (s *service) searchNoCandidatesText(l localizer) string {
	return l.message(msgSearchNoCandidates)
}

func (s *service) searchNoPreviousText(l localizer) string {
	return l.message(msgSearchNoPrevious)
}

func (s *service) searchHistoryEmptyText(l localizer) string {
	return l.message(msgSearchHistoryEmpty)
}

func (s *service) searchStartRequiredText(l localizer) string {
	return l.message(msgSearchStartRequired)
}

func (s *service) searchProfileMissingText(l localizer) string {
	return l.message(msgSearchProfileMissing)
}

func (s *service) searchUnavailableText(l localizer) string {
	return l.message(msgSearchUnavailable)
}

func (s *service) searchActionFailedText(l localizer) string {
	return l.message(msgSearchActionFailed)
}

func (s *service) searchHistoryText(l localizer, list domain.MatchHistoryList) string {
	if len(list.Items) == 0 {
		if list.Total == 0 {
			return s.searchHistoryEmptyText(l)
		}
		return l.message(msgSearchHistoryEmptyPage, list.Total)
	}

	start := list.Offset + 1
	end := list.Offset + len(list.Items)
	if list.Total > 0 && end > list.Total {
		end = list.Total
	}

	lines := make([]string, 0, len(list.Items)+1)
	lines = append(lines, l.message(msgSearchHistoryPage, list.Total, start, end))
	for i, item := range list.Items {
		lines = append(lines, s.historyItemLine(l, list.Offset+i+1, item))
	}

	return strings.Join(lines, "\n")
}

func (s *service) likesListText(l localizer, list domain.MatchLikesList) string {
	if len(list.Items) == 0 {
		if list.Total == 0 {
			return l.message(msgLikesListEmpty)
		}
		return l.message(msgLikesListEmptyPage, list.Total)
	}

	start := list.Offset + 1
	end := list.Offset + len(list.Items)
	if list.Total > 0 && end > list.Total {
		end = list.Total
	}

	lines := make([]string, 0, len(list.Items)+1)
	lines = append(lines, l.message(msgLikesListPage, list.Total, start, end))
	for i, profile := range list.Items {
		lines = append(lines, s.likesItemLine(l, list.Offset+i+1, profile))
	}

	return strings.Join(lines, "\n")
}

func (s *service) matchActionsText(l localizer) string {
	return l.message(msgMatchActions)
}

func (s *service) matchActionSavedText(l localizer) string {
	return l.message(msgMatchActionSaved)
}

func (s *service) historyActionsText(l localizer, action domain.MatchAction) string {
	return l.message(msgSearchHistoryAction, l.historyActionLabel(action))
}

func (s *service) matchProfileNotFoundText(l localizer) string {
	return l.message(msgMatchProfileNotFound)
}

func (s *service) profileViewRequiresProfileText(l localizer) string {
	return l.message(msgProfileViewRequiresProfile)
}

func (s *service) matchSuccessText(l localizer, profile domain.Profile, nickname string) string {
	name := strings.TrimSpace(profile.Name)
	if name == "" {
		name = l.label(labelThisUser)
	}

	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return l.message(msgMatchSuccess, name)
	}

	return fmt.Sprintf("%s\n%s", l.message(msgMatchSuccess, name), l.message(msgMatchNickname, nickname))
}

func (s *service) historyItemLine(l localizer, index int, item domain.MatchHistoryItem) string {
	name := strings.TrimSpace(item.Profile.Name)
	if name == "" {
		name = l.message(msgSearchHistoryUserFallback, item.Profile.UserID)
	}

	age := s.ageFromBirthDate(item.Profile.BirthDate, time.Now().UTC())
	if age == 0 {
		age = item.Profile.Age
	}

	details := make([]string, 0, 2)
	if age > 0 {
		details = append(details, l.message(msgAgeShort, age))
	}
	city := strings.TrimSpace(item.Profile.City)
	if city != "" {
		details = append(details, l.cityLabel(city))
	}

	label := name
	if len(details) > 0 {
		label = l.message(msgSearchHistoryLabel, name, strings.Join(details, ", "))
	}

	return l.message(msgSearchHistoryLine, index, label, l.historyActionLabel(item.Action))
}

func (s *service) likesItemLine(l localizer, index int, profile domain.Profile) string {
	name := strings.TrimSpace(profile.Name)
	if name == "" {
		name = l.message(msgSearchHistoryUserFallback, profile.UserID)
	}

	age := s.ageFromBirthDate(profile.BirthDate, time.Now().UTC())
	if age == 0 {
		age = profile.Age
	}

	details := make([]string, 0, 2)
	if age > 0 {
		details = append(details, l.message(msgAgeShort, age))
	}
	city := strings.TrimSpace(profile.City)
	if city != "" {
		details = append(details, l.cityLabel(city))
	}

	label := name
	if len(details) > 0 {
		label = l.message(msgSearchHistoryLabel, name, strings.Join(details, ", "))
	}

	return l.message(msgLikesListLine, index, label)
}

func (s *service) historyActionLabel(action domain.MatchAction) string {
	return newLocalizer(domain.DefaultLanguage).historyActionLabel(action)
}

func (s *service) likeNotificationText(l localizer, profile domain.Profile) string {
	name := strings.TrimSpace(profile.Name)
	if name == "" {
		return l.message(msgLikeNotification)
	}
	return l.message(msgLikeNotificationFrom, name)
}

func (s *service) profileDetails(l localizer, profile domain.Profile) string {
	return s.profileDetailsWithOptions(l, profile, profileDetailsOptions{
		header:            l.message(msgProfileDetailsHeader),
		includePhotoCount: true,
		includeTimestamps: true,
	})
}

func (s *service) profilePreviewDetails(l localizer, profile domain.Profile) string {
	return s.profilePreviewCard(l, profile)
}

func (s *service) profilePreviewCard(l localizer, profile domain.Profile) string {
	age := s.ageFromBirthDate(profile.BirthDate, time.Now().UTC())
	if age == 0 {
		age = profile.Age
	}

	name := strings.TrimSpace(profile.Name)
	emoji := s.emojiLabel(l, profile.EmojiCode)
	if emoji == l.label(labelNotSet) {
		emoji = ""
	}

	nameLine := strings.TrimSpace(strings.Join([]string{name, emoji}, " "))
	if nameLine == "" {
		nameLine = l.message(msgProfilePreviewFallbackName)
	}

	metaParts := make([]string, 0, 3)
	gender := l.genderLabel(profile.Gender)
	if gender != "" && gender != l.label(labelNotSet) {
		metaParts = append(metaParts, gender)
	}
	if age > 0 {
		metaParts = append(metaParts, l.message(msgAgeShort, age))
	}
	metaLine := strings.Join(metaParts, " | ")

	locationParts := make([]string, 0, 2)
	city := strings.TrimSpace(profile.City)
	if city != "" {
		locationParts = append(locationParts, l.cityLabel(city))
	}
	country := l.countryLabel(profile.Country)
	if country != "" && country != l.label(labelNotSet) {
		locationParts = append(locationParts, country)
	}
	locationLine := strings.Join(locationParts, ", ")

	description := strings.TrimSpace(profile.Description)

	lines := make([]string, 0, 4)
	if nameLine != "" {
		lines = append(lines, nameLine)
	}
	if metaLine != "" {
		lines = append(lines, metaLine)
	}
	if locationLine != "" {
		lines = append(lines, locationLine)
	}
	if description != "" {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, description)
	}

	return strings.Join(lines, "\n")
}

type profileDetailsOptions struct {
	header            string
	includePhotoCount bool
	includeTimestamps bool
}

func (s *service) profileDetailsWithOptions(l localizer, profile domain.Profile, options profileDetailsOptions) string {
	age := s.ageFromBirthDate(profile.BirthDate, time.Now().UTC())
	if age == 0 {
		age = profile.Age
	}
	city := strings.TrimSpace(profile.City)
	if city == "" {
		city = l.label(labelNotSet)
	} else {
		city = l.cityLabel(city)
	}
	emoji := s.emojiLabel(l, profile.EmojiCode)
	header := strings.TrimSpace(options.header)
	lines := make([]string, 0, 10)
	if header != "" {
		lines = append(lines, header)
	}
	lines = append(lines,
		l.message(msgProfileDetailsLine, l.label(labelName), profile.Name),
		l.message(msgProfileDetailsLine, l.label(labelEmoji), emoji),
		l.message(msgProfileDetailsLine, l.label(labelGender), l.genderLabel(profile.Gender)),
		l.message(msgProfileDetailsLine, l.label(labelAge), fmt.Sprintf("%d", age)),
		l.message(msgProfileDetailsLine, l.label(labelCountry), l.countryLabel(profile.Country)),
		l.message(msgProfileDetailsLine, l.label(labelCity), city),
		l.message(msgProfileDetailsLine, l.label(labelSearchVisibility), s.profileVisibilityStatus(l, profile.IsHidden)),
		l.message(msgProfileDetailsDescriptionLabel, profile.Description),
	)
	if options.includePhotoCount && len(profile.Photos) > 0 {
		lines = append(lines, l.message(msgProfileDetailsPhotos, len(profile.Photos)))
	}

	if options.includeTimestamps {
		if !profile.CreatedAt.IsZero() {
			lines = append(lines, l.message(msgProfileDetailsCreated, s.formatTime(profile.CreatedAt)))
		}
		if !profile.UpdatedAt.IsZero() {
			lines = append(lines, l.message(msgProfileDetailsUpdated, s.formatTime(profile.UpdatedAt)))
		}
	}

	return strings.Join(lines, "\n")
}

func (s *service) editText(l localizer, step domain.ProfileDraftStep, text string) string {
	header := s.editHeader(l, step)
	if text == "" {
		return header
	}

	return header + "\n" + text
}

func (s *service) photosPromptText(l localizer, isEdit bool, text string) string {
	if isEdit {
		return s.editText(l, domain.ProfileDraftStepPhotos, text)
	}

	return s.stepText(l, domain.ProfileDraftStepPhotos, text)
}

func (s *service) editHeader(l localizer, step domain.ProfileDraftStep) string {
	return l.message(msgProfileEditHeader, l.stepLabel(step))
}

func (s *service) stepText(l localizer, step domain.ProfileDraftStep, text string) string {
	header := s.stepHeader(l, step)
	if text == "" {
		return header
	}

	return header + "\n" + text
}

func (s *service) stepHeader(l localizer, step domain.ProfileDraftStep) string {
	return l.message(msgProfileSetupHeader, s.stepIndex(step), profileStepsTotal, l.stepLabel(step), s.profileSetupEstimateText(l))
}

func (s *service) profileSetupEstimateText(l localizer) string {
	minutes := s.profileSetupTotalMinutes()
	return l.message(msgProfileSetupEstimate, minutes)
}

func (s *service) stepIndex(step domain.ProfileDraftStep) int {
	switch step {
	case domain.ProfileDraftStepBotCheck:
		return 1
	case domain.ProfileDraftStepName:
		return 2
	case domain.ProfileDraftStepGender:
		return 3
	case domain.ProfileDraftStepBirthDate:
		return 4
	case domain.ProfileDraftStepCountry:
		return 5
	case domain.ProfileDraftStepCity:
		return 6
	case domain.ProfileDraftStepDescription:
		return 7
	case domain.ProfileDraftStepEmoji:
		return 8
	case domain.ProfileDraftStepPhotos:
		return 9
	default:
		return 1
	}
}

func (s *service) stepLabel(l localizer, step domain.ProfileDraftStep) string {
	return l.stepLabel(step)
}

func (s *service) startGreeting(l localizer, user domain.User, profile domain.Profile, status profileStatus) string {
	name := ""
	if status == profileStatusPresent {
		name = strings.TrimSpace(profile.Name)
	}
	if name == "" {
		name = s.userFullName(user)
	}
	if name == "" {
		return l.message(msgStartGreeting)
	}

	return l.message(msgStartGreetingNamed, name)
}
