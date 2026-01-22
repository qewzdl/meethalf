package bot

import (
	"fmt"
	"strings"
	"time"

	"meethalf-telegram-bot/internal/domain"
)

const (
	defaultHelpText           = "Use the button below to view or create your profile. Open Settings to delete it."
	profileCreatedText        = "Profile created."
	profileUpdatedText        = "Profile updated."
	loadingStartText          = "Checking your profile..."
	loadingProfileViewText    = "Loading your profile..."
	loadingProfilePreviewText = "Loading profile preview..."
	loadingEditNameText       = "Preparing name update..."
	loadingEditGenderText     = "Preparing gender update..."
	loadingEditBirthDateText  = "Preparing birth date update..."
	loadingEditCountryText    = "Preparing country update..."
	loadingEditCityText       = "Preparing city update..."
	loadingEditDescText       = "Preparing description update..."
	loadingEditEmojiText      = "Preparing emoji update..."
	loadingEditPhotosText     = "Preparing photo update..."
	creatingProfileText       = "Creating your profile..."
	updatingProfileText       = "Updating your profile..."
	deletingProfileText       = "Deleting your profile..."
	profileDeleteConfirmText  = "Are you sure you want to delete your profile? This action cannot be undone."
	profileDeleteCanceledText = "Profile deletion canceled."
	profileDeleteExpiredText  = "Profile deletion confirmation expired. Use Settings to start again."
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
	text := fmt.Sprintf("To protect from bots, solve: %s\nReply with the result.", strings.TrimSpace(question))
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

func (s *service) profileSettingsText() string {
	return "Profile settings. Use the button below to delete your profile. You will be asked to confirm."
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
