package bot

import (
	"fmt"
	"strings"
	"time"

	"meethalf-telegram-bot/internal/domain"
)

const (
	defaultHelpText                   = "Use the buttons below to start searching, view or create your profile, or open settings."
	adminBadgeText                    = "Admin access enabled."
	moderatorBadgeText                = "Moderator access enabled."
	adminMenuText                     = "Admin panel. Choose an action."
	moderatorMenuText                 = "Moderator panel. Choose an action."
	adminAccessDeniedText             = "Admin access required."
	adminModerationRestrictedText     = "Moderators can only manage bans for regular users."
	adminUsersEmptyText               = "No users found."
	adminUsersLoadFailedText          = "Unable to load users. Please try again later."
	adminUsersPageTemplate            = "Users: %d total. Showing %d-%d."
	adminUsersEmptyPageTemplate       = "Users: %d total. No users found on this page."
	adminBannedUsersEmptyText         = "No banned users found."
	adminBannedUsersLoadFailedText    = "Unable to load banned users. Please try again later."
	adminBannedUsersPageTemplate      = "Banned users: %d total. Showing %d-%d."
	adminBannedUsersEmptyPageTemplate = "Banned users: %d total. No users found on this page."
	adminModeratorsEmptyText          = "No moderators found."
	adminModeratorsLoadFailedText     = "Unable to load moderators. Please try again later."
	adminModeratorsPageTemplate       = "Moderators: %d total. Showing %d-%d."
	adminModeratorsEmptyPageTemplate  = "Moderators: %d total. No users found on this page."
	adminReportsEmptyText             = "No reported users found."
	adminReportsLoadFailedText        = "Unable to load reported users. Please try again later."
	adminReportsPageTemplate          = "Reported users: %d total. Showing %d-%d."
	adminReportsEmptyPageTemplate     = "Reported users: %d total. No users found on this page."
	adminBanUsageText                 = "Send the user ID or @username to ban."
	adminBanFailedText                = "Unable to ban user. Please try again later."
	adminUserNotFoundText             = "User not found."
	adminBanSuccessTemplate           = "User %s has been banned."
	adminUnbanUsageText               = "Send the user ID or @username to unban."
	adminUnbanFailedText              = "Unable to unban user. Please try again later."
	adminUnbanSuccessTemplate         = "User %s has been unbanned."
	adminActionFailedText             = "Unable to complete the admin action. Please try again later."
	adminModeratorUsageText           = "Send the user ID or @username to assign the moderator role."
	adminModeratorFailedText          = "Unable to assign the moderator role. Please try again later."
	adminModeratorSuccessTemplate     = "User %s is now a moderator."
	adminUnmoderatorUsageText         = "Send the user ID or @username to remove the moderator role."
	adminUnmoderatorFailedText        = "Unable to remove the moderator role. Please try again later."
	adminUnmoderatorSuccessTemplate   = "User %s is no longer a moderator."
	adminResetChoicesUsageText        = "Send the user ID or @username to reset match choices."
	adminResetChoicesFailedText       = "Unable to reset match choices. Please try again later."
	adminResetChoicesSuccessTemplate  = "Match choices were reset for %s."
	adminClearReportsUsageText        = "Send the user ID or @username to clear reports."
	adminClearReportsFailedText       = "Unable to clear reports. Please try again later."
	adminClearReportsSuccessTemplate  = "Reports were cleared for %s."
	userBannedText                    = "Your account is banned. Contact support."
	profileCreatedText                = "Profile created."
	profileUpdatedText                = "Profile updated."
	loadingStartText                  = "Checking your profile..."
	loadingProfileViewText            = "Loading your profile..."
	loadingProfilePreviewText         = "Loading profile preview..."
	loadingEditNameText               = "Preparing name update..."
	loadingEditGenderText             = "Preparing gender update..."
	loadingEditBirthDateText          = "Preparing birth date update..."
	loadingEditCountryText            = "Preparing country update..."
	loadingEditCityText               = "Preparing city update..."
	loadingEditDescText               = "Preparing description update..."
	loadingEditEmojiText              = "Preparing emoji update..."
	loadingEditPhotosText             = "Preparing photo update..."
	loadingProfileVisibilityText      = "Updating search visibility..."
	loadingSearchStartText            = "Finding profiles..."
	loadingSearchNextText             = "Searching for the next profile..."
	loadingSearchPrevText             = "Opening the previous profile..."
	loadingSearchHistoryText          = "Loading search history..."
	loadingSearchHistoryProfileText   = "Opening history profile..."
	loadingSearchHistoryActionText    = "Updating decision..."
	loadingAdminUsersText             = "Loading users..."
	loadingAdminBanText               = "Banning user..."
	loadingAdminUnbanText             = "Unbanning user..."
	loadingAdminModeratorText         = "Assigning moderator role..."
	loadingAdminUnmoderatorText       = "Removing moderator role..."
	loadingAdminResetChoicesText      = "Resetting match choices..."
	loadingAdminClearReportsText      = "Clearing reports..."
	loadingAdminBannedUsersText       = "Loading banned users..."
	loadingAdminModeratorsText        = "Loading moderators..."
	loadingAdminReportsText           = "Loading reported users..."
	creatingProfileText               = "Creating your profile..."
	updatingProfileText               = "Updating your profile..."
	deletingProfileText               = "Deleting your profile..."
	profileDeleteConfirmText          = "Are you sure you want to delete your profile? This action cannot be undone."
	profileDeleteCanceledText         = "Profile deletion canceled."
	profileDeleteExpiredText          = "Profile deletion confirmation expired. Use Settings to start again."
	profileSetupCanceledText          = "Profile setup canceled."
	profileEditCanceledText           = "Profile edit canceled."
	actionCanceledText                = "Action canceled."
	profileHiddenText                 = "Profile is now hidden from search."
	profileVisibleText                = "Profile is now visible in search."
	profileVisibilityUpdateFailedText = "Unable to update search visibility. Please try again later."
	searchGenderPromptText            = "Select the gender to search for."
	searchAccuracyPromptText          = "Match accuracy (0–4): 0 wide/random → 4 strict/precise."
	searchNoCandidatesText            = "No matching profiles yet. Try again later."
	searchNoPreviousText              = "No previous profile."
	searchHistoryEmptyText            = "History is empty."
	searchStartRequiredText           = "Press \"Start search\" first."
	searchProfileMissingText          = "Create a profile first to start searching and view profiles."
	searchUnavailableText             = "Search is currently unavailable."
	searchActionFailedText            = "Unable to process the action. Try again later."
	searchHistoryPageTemplate         = "History: %d total. Showing %d-%d."
	searchHistoryEmptyPageTemplate    = "History: %d total. No profiles found on this page."
	searchHistoryActionTemplate       = "Current decision: %s.\nChoose an action."
	matchActionsText                  = "Choose an action."
	matchProfileNotFoundText          = "Profile not found."
	profileViewRequiresProfileText    = "Create a profile first to view other profiles."
	matchSuccessTemplate              = "It's a match! You and %s liked each other."
	matchNicknameTemplate             = "Nickname: %s"
)

func (s *service) namePrompt(user domain.User) string {
	header := s.stepHeader(domain.ProfileDraftStepName)
	telegramName := s.userFullName(user)
	if telegramName == "" {
		return header + "\nYour Telegram profile has no name set. Please type the name you want to use."
	}

	return fmt.Sprintf("%s\nCurrent Telegram name: %s\nUse the button below to use it, or send the name you prefer.", header, telegramName)
}

func (s *service) botCheckPrompt(question string) string {
	return s.botCheckRetryPrompt("", question)
}

func (s *service) botCheckRetryPrompt(reason, question string) string {
	text := fmt.Sprintf("To protect from bots, solve: %s\nChoose the correct answer below.", strings.TrimSpace(question))
	if strings.TrimSpace(reason) != "" {
		text = reason + "\n" + text
	}
	return s.stepText(domain.ProfileDraftStepBotCheck, text)
}

func (s *service) birthDatePrompt() string {
	return s.stepText(domain.ProfileDraftStepBirthDate, fmt.Sprintf("Enter your birth date in %s format (for example, 1990-04-23).", birthDateLayout))
}

func (s *service) genderPrompt() string {
	return s.stepText(domain.ProfileDraftStepGender, "Select your gender using the buttons below.")
}

func (s *service) countryPrompt() string {
	return s.stepText(domain.ProfileDraftStepCountry, "Select your country using the buttons below.")
}

func (s *service) cityPrompt() string {
	return s.stepText(domain.ProfileDraftStepCity, "Select your city using the buttons below.")
}

func (s *service) descriptionPrompt() string {
	return s.stepText(domain.ProfileDraftStepDescription, "Write a short description about yourself.")
}

func (s *service) emojiPrompt() string {
	return s.stepText(domain.ProfileDraftStepEmoji, "Select the emoji that describes you using the buttons below.")
}

func (s *service) photosPrompt() string {
	return s.stepText(domain.ProfileDraftStepPhotos, "Send 1-4 photos for your album. Use the Done button when finished.")
}

func (s *service) profileEditMenuText() string {
	return "Choose what you want to update in your profile."
}

func (s *service) adminMenuText() string {
	return adminMenuText
}

func (s *service) adminMenuTextForRole(role adminRole) string {
	if role == adminRoleModerator {
		return moderatorMenuText
	}
	return adminMenuText
}

func (s *service) adminAccessDeniedText() string {
	return adminAccessDeniedText
}

func (s *service) adminModerationRestrictedText() string {
	return adminModerationRestrictedText
}

func (s *service) adminUsersEmptyText() string {
	return adminUsersEmptyText
}

func (s *service) adminUsersLoadFailedText() string {
	return adminUsersLoadFailedText
}

func (s *service) adminBannedUsersEmptyText() string {
	return adminBannedUsersEmptyText
}

func (s *service) adminBannedUsersLoadFailedText() string {
	return adminBannedUsersLoadFailedText
}

func (s *service) adminModeratorsEmptyText() string {
	return adminModeratorsEmptyText
}

func (s *service) adminModeratorsLoadFailedText() string {
	return adminModeratorsLoadFailedText
}

func (s *service) adminReportsEmptyText() string {
	return adminReportsEmptyText
}

func (s *service) adminReportsLoadFailedText() string {
	return adminReportsLoadFailedText
}

func (s *service) adminBanUsageText() string {
	return adminBanUsageText
}

func (s *service) adminBanFailedText() string {
	return adminBanFailedText
}

func (s *service) adminUserNotFoundText() string {
	return adminUserNotFoundText
}

func (s *service) adminBanSuccessText(userRef string) string {
	return fmt.Sprintf(adminBanSuccessTemplate, userRef)
}

func (s *service) adminUnbanUsageText() string {
	return adminUnbanUsageText
}

func (s *service) adminUnbanFailedText() string {
	return adminUnbanFailedText
}

func (s *service) adminUnbanSuccessText(userRef string) string {
	return fmt.Sprintf(adminUnbanSuccessTemplate, userRef)
}

func (s *service) adminActionFailedText() string {
	return adminActionFailedText
}

func (s *service) adminModeratorUsageText() string {
	return adminModeratorUsageText
}

func (s *service) adminModeratorFailedText() string {
	return adminModeratorFailedText
}

func (s *service) adminModeratorSuccessText(userRef string) string {
	return fmt.Sprintf(adminModeratorSuccessTemplate, userRef)
}

func (s *service) adminUnmoderatorUsageText() string {
	return adminUnmoderatorUsageText
}

func (s *service) adminUnmoderatorFailedText() string {
	return adminUnmoderatorFailedText
}

func (s *service) adminUnmoderatorSuccessText(userRef string) string {
	return fmt.Sprintf(adminUnmoderatorSuccessTemplate, userRef)
}

func (s *service) adminResetChoicesUsageText() string {
	return adminResetChoicesUsageText
}

func (s *service) adminResetChoicesFailedText() string {
	return adminResetChoicesFailedText
}

func (s *service) adminResetChoicesSuccessText(userRef string) string {
	return fmt.Sprintf(adminResetChoicesSuccessTemplate, userRef)
}

func (s *service) adminClearReportsUsageText() string {
	return adminClearReportsUsageText
}

func (s *service) adminClearReportsFailedText() string {
	return adminClearReportsFailedText
}

func (s *service) adminClearReportsSuccessText(userRef string) string {
	return fmt.Sprintf(adminClearReportsSuccessTemplate, userRef)
}

func (s *service) userBannedText() string {
	return userBannedText
}

func (s *service) profileSettingsText() string {
	return "Profile settings. Use the buttons below to manage search visibility or delete your profile."
}

func (s *service) profileActionsText() string {
	return "Use the buttons below to preview or edit your profile."
}

func (s *service) profilePreviewActionsText() string {
	return "Use the buttons below to return to your profile or edit it."
}

func (s *service) editPrompt(step domain.ProfileDraftStep) string {
	switch step {
	case domain.ProfileDraftStepName:
		return s.editText(step, "Enter the new name.")
	case domain.ProfileDraftStepBirthDate:
		return s.editText(step, fmt.Sprintf("Enter the new birth date in %s format (for example, 1990-04-23).", birthDateLayout))
	case domain.ProfileDraftStepGender:
		return s.editText(step, "Select the new gender using the buttons below.")
	case domain.ProfileDraftStepCountry:
		return s.editText(step, "Select the new country using the buttons below.")
	case domain.ProfileDraftStepCity:
		return s.editText(step, "Select the new city using the buttons below.")
	case domain.ProfileDraftStepDescription:
		return s.editText(step, "Write the new description.")
	case domain.ProfileDraftStepEmoji:
		return s.editText(step, "Choose the new emoji using the buttons below.")
	case domain.ProfileDraftStepPhotos:
		return s.editText(step, "Send 1-4 photos to replace your album. Use the Done button when finished.")
	default:
		return s.editText(step, "Enter the updated value.")
	}
}

func (s *service) profileCreated() string {
	return profileCreatedText
}

func (s *service) profileUpdated() string {
	return profileUpdatedText
}

func (s *service) profileDeleteConfirmText() string {
	return profileDeleteConfirmText
}

func (s *service) profileDeleteCanceledText() string {
	return profileDeleteCanceledText
}

func (s *service) profileDeleteExpiredText() string {
	return profileDeleteExpiredText
}

func (s *service) profileSetupCanceledText() string {
	return profileSetupCanceledText
}

func (s *service) profileEditCanceledText() string {
	return profileEditCanceledText
}

func (s *service) actionCanceledText() string {
	return actionCanceledText
}

func (s *service) profileVisibilityUpdateFailedText() string {
	return profileVisibilityUpdateFailedText
}

func (s *service) profileSettingsTextWithVisibility(isHidden bool) string {
	status := s.profileVisibilityStatus(isHidden)
	return fmt.Sprintf("Profile settings.\nSearch visibility: %s.\nUse the buttons below to manage search visibility or delete your profile.", status)
}

func (s *service) profileVisibilityStatus(isHidden bool) string {
	if isHidden {
		return "Hidden from search"
	}
	return "Visible in search"
}

func (s *service) profileVisibilityUpdated(isHidden bool) string {
	if isHidden {
		return profileHiddenText
	}
	return profileVisibleText
}

func (s *service) searchGenderText() string {
	return searchGenderPromptText
}

func (s *service) searchAccuracyText() string {
	return searchAccuracyPromptText
}

func (s *service) searchNoCandidatesText() string {
	return searchNoCandidatesText
}

func (s *service) searchNoPreviousText() string {
	return searchNoPreviousText
}

func (s *service) searchHistoryEmptyText() string {
	return searchHistoryEmptyText
}

func (s *service) searchStartRequiredText() string {
	return searchStartRequiredText
}

func (s *service) searchProfileMissingText() string {
	return searchProfileMissingText
}

func (s *service) searchUnavailableText() string {
	return searchUnavailableText
}

func (s *service) searchActionFailedText() string {
	return searchActionFailedText
}

func (s *service) searchHistoryText(list domain.MatchHistoryList) string {
	if len(list.Items) == 0 {
		if list.Total == 0 {
			return s.searchHistoryEmptyText()
		}
		return fmt.Sprintf(searchHistoryEmptyPageTemplate, list.Total)
	}

	start := list.Offset + 1
	end := list.Offset + len(list.Items)
	if list.Total > 0 && end > list.Total {
		end = list.Total
	}

	lines := make([]string, 0, len(list.Items)+1)
	lines = append(lines, fmt.Sprintf(searchHistoryPageTemplate, list.Total, start, end))
	for i, item := range list.Items {
		lines = append(lines, s.historyItemLine(list.Offset+i+1, item))
	}

	return strings.Join(lines, "\n")
}

func (s *service) matchActionsText() string {
	return matchActionsText
}

func (s *service) historyActionsText(action domain.MatchAction) string {
	return fmt.Sprintf(searchHistoryActionTemplate, s.historyActionLabel(action))
}

func (s *service) matchProfileNotFoundText() string {
	return matchProfileNotFoundText
}

func (s *service) profileViewRequiresProfileText() string {
	return profileViewRequiresProfileText
}

func (s *service) matchSuccessText(profile domain.Profile, nickname string) string {
	name := strings.TrimSpace(profile.Name)
	if name == "" {
		name = "this user"
	}

	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return fmt.Sprintf(matchSuccessTemplate, name)
	}

	return fmt.Sprintf("%s\n%s", fmt.Sprintf(matchSuccessTemplate, name), fmt.Sprintf(matchNicknameTemplate, nickname))
}

func (s *service) historyItemLine(index int, item domain.MatchHistoryItem) string {
	name := strings.TrimSpace(item.Profile.Name)
	if name == "" {
		name = fmt.Sprintf("User %d", item.Profile.UserID)
	}

	age := s.ageFromBirthDate(item.Profile.BirthDate, time.Now().UTC())
	if age == 0 {
		age = item.Profile.Age
	}

	details := make([]string, 0, 2)
	if age > 0 {
		details = append(details, fmt.Sprintf("%d y.o.", age))
	}
	city := strings.TrimSpace(item.Profile.City)
	if city != "" {
		details = append(details, city)
	}

	label := name
	if len(details) > 0 {
		label = fmt.Sprintf("%s (%s)", name, strings.Join(details, ", "))
	}

	return fmt.Sprintf("%d. %s - %s", index, label, s.historyActionLabel(item.Action))
}

func (s *service) historyActionLabel(action domain.MatchAction) string {
	switch action {
	case domain.MatchActionLike:
		return "Like"
	case domain.MatchActionDislike:
		return "Dislike"
	case domain.MatchActionReport:
		return "Report"
	default:
		return "No decision"
	}
}

func (s *service) likeNotificationText(profile domain.Profile) string {
	name := strings.TrimSpace(profile.Name)
	if name == "" {
		return "You received a ❤️. View the profile?"
	}
	return fmt.Sprintf("You received a ❤️ from %s. View the profile?", name)
}

func (s *service) profileDetails(profile domain.Profile) string {
	return s.profileDetailsWithOptions(profile, profileDetailsOptions{
		header:            "Your profile:",
		includePhotoCount: true,
		includeTimestamps: true,
	})
}

func (s *service) profilePreviewDetails(profile domain.Profile) string {
	return s.profilePreviewCard(profile)
}

func (s *service) profilePreviewCard(profile domain.Profile) string {
	age := s.ageFromBirthDate(profile.BirthDate, time.Now().UTC())
	if age == 0 {
		age = profile.Age
	}

	name := strings.TrimSpace(profile.Name)
	emoji := s.emojiLabel(profile.EmojiCode)
	if emoji == "Not set" {
		emoji = ""
	}

	nameLine := strings.TrimSpace(strings.Join([]string{name, emoji}, " "))
	if nameLine == "" {
		nameLine = "Profile"
	}

	metaParts := make([]string, 0, 3)
	gender := s.genderLabel(profile.Gender)
	if gender != "" && gender != "Not set" {
		metaParts = append(metaParts, gender)
	}
	if age > 0 {
		metaParts = append(metaParts, fmt.Sprintf("%d y.o.", age))
	}
	metaLine := strings.Join(metaParts, " | ")

	locationParts := make([]string, 0, 2)
	city := strings.TrimSpace(profile.City)
	if city != "" {
		locationParts = append(locationParts, city)
	}
	country := s.countryLabel(profile.Country)
	if country != "" && country != "Not set" {
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

func (s *service) profileDetailsWithOptions(profile domain.Profile, options profileDetailsOptions) string {
	age := s.ageFromBirthDate(profile.BirthDate, time.Now().UTC())
	if age == 0 {
		age = profile.Age
	}
	city := strings.TrimSpace(profile.City)
	if city == "" {
		city = "Not set"
	}
	emoji := s.emojiLabel(profile.EmojiCode)
	header := strings.TrimSpace(options.header)
	lines := make([]string, 0, 10)
	if header != "" {
		lines = append(lines, header)
	}
	lines = append(lines,
		fmt.Sprintf("Name: %s", profile.Name),
		fmt.Sprintf("Emoji: %s", emoji),
		fmt.Sprintf("Gender: %s", s.genderLabel(profile.Gender)),
		fmt.Sprintf("Age: %d", age),
		fmt.Sprintf("Country: %s", s.countryLabel(profile.Country)),
		fmt.Sprintf("City: %s", city),
		fmt.Sprintf("Search visibility: %s", s.profileVisibilityStatus(profile.IsHidden)),
		fmt.Sprintf("Description: \n%s", profile.Description),
	)
	if options.includePhotoCount && len(profile.Photos) > 0 {
		lines = append(lines, fmt.Sprintf("Photos: %d", len(profile.Photos)))
	}

	if options.includeTimestamps {
		if !profile.CreatedAt.IsZero() {
			lines = append(lines, fmt.Sprintf("Created: %s", s.formatTime(profile.CreatedAt)))
		}
		if !profile.UpdatedAt.IsZero() {
			lines = append(lines, fmt.Sprintf("Updated: %s", s.formatTime(profile.UpdatedAt)))
		}
	}

	return strings.Join(lines, "\n")
}

func (s *service) editText(step domain.ProfileDraftStep, text string) string {
	header := s.editHeader(step)
	if text == "" {
		return header
	}

	return header + "\n" + text
}

func (s *service) photosPromptText(isEdit bool, text string) string {
	if isEdit {
		return s.editText(domain.ProfileDraftStepPhotos, text)
	}

	return s.stepText(domain.ProfileDraftStepPhotos, text)
}

func (s *service) editHeader(step domain.ProfileDraftStep) string {
	return fmt.Sprintf("Profile edit: %s", s.stepLabel(step))
}

func (s *service) stepText(step domain.ProfileDraftStep, text string) string {
	header := s.stepHeader(step)
	if text == "" {
		return header
	}

	return header + "\n" + text
}

func (s *service) stepHeader(step domain.ProfileDraftStep) string {
	return fmt.Sprintf("Profile setup (step %d/%d): %s\n%s", s.stepIndex(step), profileStepsTotal, s.stepLabel(step), s.profileSetupEstimateText())
}

func (s *service) profileSetupEstimateText() string {
	minutes := s.profileSetupTotalMinutes()
	return fmt.Sprintf("Estimated total time: ~%d min", minutes)
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

func (s *service) stepLabel(step domain.ProfileDraftStep) string {
	switch step {
	case domain.ProfileDraftStepBotCheck:
		return "Verification"
	case domain.ProfileDraftStepName:
		return "Name"
	case domain.ProfileDraftStepGender:
		return "Gender"
	case domain.ProfileDraftStepBirthDate:
		return "Birth date"
	case domain.ProfileDraftStepCountry:
		return "Country"
	case domain.ProfileDraftStepCity:
		return "City"
	case domain.ProfileDraftStepDescription:
		return "Description"
	case domain.ProfileDraftStepEmoji:
		return "Emoji"
	case domain.ProfileDraftStepPhotos:
		return "Photos"
	default:
		return "Profile"
	}
}

func (s *service) startGreeting(user domain.User, profile domain.Profile, status profileStatus) string {
	name := ""
	if status == profileStatusPresent {
		name = strings.TrimSpace(profile.Name)
	}
	if name == "" {
		name = s.userFullName(user)
	}
	if name == "" {
		return "Welcome to Meethalf bot."
	}

	return fmt.Sprintf("Welcome to Meethalf bot, %s.", name)
}
