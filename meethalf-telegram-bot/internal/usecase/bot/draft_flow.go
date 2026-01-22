package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) handleDraft(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft) (string, error) {
	if s.draftMode(draft) == domain.ProfileDraftModeEdit {
		switch draft.Step {
		case domain.ProfileDraftStepName:
			return s.applyName(ctx, msg, draft)
		case domain.ProfileDraftStepGender:
			return s.applyGender(ctx, msg, draft)
		case domain.ProfileDraftStepBirthDate:
			return s.applyBirthDate(ctx, msg, draft)
		case domain.ProfileDraftStepCountry:
			return s.applyCountry(ctx, msg, draft)
		case domain.ProfileDraftStepCity:
			return s.applyCity(ctx, msg, draft)
		case domain.ProfileDraftStepDescription:
			return s.applyDescription(ctx, msg, draft)
		case domain.ProfileDraftStepEmoji:
			return s.applyEmoji(ctx, msg, draft)
		case domain.ProfileDraftStepPhotos:
			return s.applyPhotos(ctx, msg, draft)
		default:
			return s.profileEditMenuText(), nil
		}
	}

	switch draft.Step {
	case domain.ProfileDraftStepBotCheck:
		return s.applyBotCheck(ctx, msg, draft)
	case domain.ProfileDraftStepName:
		return s.applyName(ctx, msg, draft)
	case domain.ProfileDraftStepGender:
		return s.applyGender(ctx, msg, draft)
	case domain.ProfileDraftStepBirthDate:
		return s.applyBirthDate(ctx, msg, draft)
	case domain.ProfileDraftStepCountry:
		return s.applyCountry(ctx, msg, draft)
	case domain.ProfileDraftStepCity:
		return s.applyCity(ctx, msg, draft)
	case domain.ProfileDraftStepDescription:
		return s.applyDescription(ctx, msg, draft)
	case domain.ProfileDraftStepEmoji:
		return s.applyEmoji(ctx, msg, draft)
	case domain.ProfileDraftStepPhotos:
		return s.applyPhotos(ctx, msg, draft)
	default:
		return s.startProfileSetup(ctx, msg)
	}
}

func (s *service) startProfileSetup(ctx context.Context, msg domain.IncomingMessage) (string, error) {
	if s == nil || s.drafts == nil {
		return "Profile setup is not available right now.", errors.New("profile draft repository is not configured")
	}

	if msg.User.ID == 0 {
		return "Profile setup is not available for this chat.", errors.New("user id is missing")
	}

	draft := domain.ProfileDraft{
		UserID:    msg.User.ID,
		ChatID:    msg.ChatID,
		Step:      domain.ProfileDraftStepBotCheck,
		Mode:      domain.ProfileDraftModeCreate,
		UpdatedAt: s.now(msg.ReceivedAt),
	}
	s.resetBotCheck(&draft, msg.ReceivedAt)

	if err := s.drafts.Save(ctx, draft); err != nil {
		return "Failed to start profile setup. Please try again later.", err
	}

	return s.botCheckPrompt(draft.BotCheckQuestion), nil
}

func (s *service) startProfileEdit(ctx context.Context, msg domain.IncomingMessage, step domain.ProfileDraftStep) (string, error) {
	if s == nil || s.drafts == nil {
		return "Profile edit is not available right now.", errors.New("profile draft repository is not configured")
	}

	if s.profiles == nil {
		return "Profile service is not available right now.", errors.New("profile service is not configured")
	}

	if msg.User.ID == 0 {
		return "Profile edit is not available for this chat.", errors.New("user id is missing")
	}

	profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
	if err != nil {
		return "Unable to load profile. Please try again later.", err
	}
	if !found {
		return "Profile not found. Use /profile to create it.", nil
	}

	stepToEdit := step
	customPrompt := ""
	if profile.BirthDate.IsZero() && step != domain.ProfileDraftStepBirthDate {
		stepToEdit = domain.ProfileDraftStepBirthDate
		customPrompt = s.editText(stepToEdit, fmt.Sprintf("Birth date is required before editing other fields. Enter it in %s format (for example, 1990-04-23).", birthDateLayout))
	}
	if customPrompt == "" && !profile.BirthDate.IsZero() && profile.Country == "" && step != domain.ProfileDraftStepCountry {
		stepToEdit = domain.ProfileDraftStepCountry
		customPrompt = s.editText(stepToEdit, "Country is required before editing other fields. Select it using the buttons below.")
	}
	if customPrompt == "" && !profile.BirthDate.IsZero() && profile.Country != "" && strings.TrimSpace(profile.City) == "" && step != domain.ProfileDraftStepCity {
		stepToEdit = domain.ProfileDraftStepCity
		customPrompt = s.editText(stepToEdit, "City is required before editing other fields. Enter it.")
	}
	if customPrompt == "" && profile.EmojiCode == "" && step != domain.ProfileDraftStepEmoji {
		stepToEdit = domain.ProfileDraftStepEmoji
		customPrompt = s.editText(stepToEdit, "Emoji is required before editing other fields. Select it using the buttons below.")
	}

	draft := domain.ProfileDraft{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		Step:        stepToEdit,
		Name:        profile.Name,
		Gender:      profile.Gender,
		BirthDate:   profile.BirthDate,
		Country:     profile.Country,
		City:        profile.City,
		Description: profile.Description,
		EmojiCode:   profile.EmojiCode,
		Photos:      profile.Photos,
		Mode:        domain.ProfileDraftModeEdit,
		UpdatedAt:   s.now(msg.ReceivedAt),
	}
	if stepToEdit == domain.ProfileDraftStepPhotos {
		draft.Photos = nil
	}

	if err := s.drafts.Save(ctx, draft); err != nil {
		return "Failed to start profile edit. Please try again later.", err
	}

	if customPrompt != "" {
		return customPrompt, nil
	}

	return s.editPrompt(stepToEdit), nil
}

func (s *service) deleteProfile(ctx context.Context, msg domain.IncomingMessage) (string, error) {
	if s == nil || s.profiles == nil {
		return "Profile service is not available right now.", errors.New("profile service is not configured")
	}

	if msg.User.ID == 0 {
		return "Profile delete is not available for this chat.", errors.New("user id is missing")
	}

	deleted, err := s.profiles.DeleteProfile(ctx, msg.User.ID)
	if err != nil {
		return "Unable to delete profile. Please try again later.", err
	}
	if !deleted {
		return "Profile not found. Use the Create Profile button to create it.", nil
	}

	if s.drafts != nil {
		if err := s.drafts.Delete(ctx, msg.User.ID); err != nil {
			return "Profile deleted. Note: could not clear the profile draft.", err
		}
	}

	return "Profile deleted. Use the Create Profile button to create a new one.", nil
}

func (s *service) applyBotCheck(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft) (string, error) {
	if s.drafts == nil {
		return "Profile setup is not available right now.", errors.New("profile draft repository is not configured")
	}

	s.ensureBotCheck(&draft, msg.ReceivedAt)
	answer := strings.TrimSpace(msg.Text)
	if answer == "" {
		return s.botCheckPrompt(draft.BotCheckQuestion), nil
	}

	if s.botCheckMatches(draft, answer) {
		draft.Step = domain.ProfileDraftStepName
		draft.BotCheckQuestion = ""
		draft.BotCheckAnswer = 0
		draft.BotCheckAttempts = 0
		draft.UpdatedAt = s.now(msg.ReceivedAt)
		if err := s.drafts.Save(ctx, draft); err != nil {
			return "Failed to start profile setup. Please try again later.", err
		}

		return s.namePrompt(msg.User), nil
	}

	draft.BotCheckAttempts++
	if draft.BotCheckAttempts >= botCheckMaxAttempts {
		s.resetBotCheck(&draft, msg.ReceivedAt)
		if err := s.drafts.Save(ctx, draft); err != nil {
			return "Failed to save profile setup. Please try again later.", err
		}
		return s.botCheckRetryPrompt("Too many attempts. Let's try a new check.", draft.BotCheckQuestion), nil
	}

	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if err := s.drafts.Save(ctx, draft); err != nil {
		return "Failed to save profile setup. Please try again later.", err
	}

	return s.botCheckRetryPrompt("Incorrect answer. Try again.", draft.BotCheckQuestion), nil
}

func (s *service) applyName(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft) (string, error) {
	value := strings.TrimSpace(msg.Text)
	draft.Mode = s.draftMode(draft)
	isEdit := draft.Mode == domain.ProfileDraftModeEdit
	if value == "" {
		if isEdit {
			return s.editText(domain.ProfileDraftStepName, "Please enter a name."), nil
		}
		return s.stepText(domain.ProfileDraftStepName, "Please enter a name or use the button below to use your Telegram name."), nil
	}

	name := value
	if s.isAffirmative(value) {
		telegramName := s.userFullName(msg.User)
		if telegramName == "" {
			if isEdit {
				return s.editText(domain.ProfileDraftStepName, "Your Telegram profile has no name set. Please type the name you want to use."), nil
			}
			return s.stepText(domain.ProfileDraftStepName, "Your Telegram profile has no name set. Please type the name you want to use."), nil
		}
		name = telegramName
	}

	if len(name) > maxNameLength {
		if isEdit {
			return s.editText(domain.ProfileDraftStepName, fmt.Sprintf("Name is too long (max %d characters).", maxNameLength)), nil
		}
		return s.stepText(domain.ProfileDraftStepName, fmt.Sprintf("Name is too long (max %d characters).", maxNameLength)), nil
	}

	draft.Name = name
	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if isEdit {
		if err := s.drafts.Save(ctx, draft); err != nil {
			return "Failed to save profile edit. Please try again later.", err
		}

		return s.saveProfile(ctx, draft)
	}

	draft.Step = domain.ProfileDraftStepGender
	if err := s.drafts.Save(ctx, draft); err != nil {
		return "Failed to save profile setup. Please try again later.", err
	}

	return s.genderPrompt(), nil
}

func (s *service) applyGender(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft) (string, error) {
	value := strings.TrimSpace(msg.Text)
	draft.Mode = s.draftMode(draft)
	isEdit := draft.Mode == domain.ProfileDraftModeEdit
	gender, ok := s.normalizeGender(value)
	if !ok {
		text := "Gender must be one of: male, female, or other."
		if isEdit {
			return s.editText(domain.ProfileDraftStepGender, text), nil
		}
		return s.stepText(domain.ProfileDraftStepGender, text+" Try again."), nil
	}

	draft.Gender = gender
	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if isEdit {
		if err := s.drafts.Save(ctx, draft); err != nil {
			return "Failed to save profile edit. Please try again later.", err
		}

		return s.saveProfile(ctx, draft)
	}

	draft.Step = domain.ProfileDraftStepBirthDate
	if err := s.drafts.Save(ctx, draft); err != nil {
		return "Failed to save profile setup. Please try again later.", err
	}

	return s.birthDatePrompt(), nil
}

func (s *service) applyBirthDate(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft) (string, error) {
	value := strings.TrimSpace(msg.Text)
	draft.Mode = s.draftMode(draft)
	isEdit := draft.Mode == domain.ProfileDraftModeEdit
	birthDate, ok := s.parseBirthDate(value)
	if !ok {
		if isEdit {
			return s.editText(domain.ProfileDraftStepBirthDate, fmt.Sprintf("Birth date must be in %s format (for example, 1990-04-23). Try again.", birthDateLayout)), nil
		}
		return s.stepText(domain.ProfileDraftStepBirthDate, fmt.Sprintf("Birth date must be in %s format (for example, 1990-04-23). Try again.", birthDateLayout)), nil
	}

	age := s.ageFromBirthDate(birthDate, s.now(msg.ReceivedAt))
	if age < minAge || age > maxAge {
		if isEdit {
			return s.editText(domain.ProfileDraftStepBirthDate, fmt.Sprintf("Age must be between %d and %d years. Please check the birth date.", minAge, maxAge)), nil
		}
		return s.stepText(domain.ProfileDraftStepBirthDate, fmt.Sprintf("Age must be between %d and %d years. Please check the birth date.", minAge, maxAge)), nil
	}

	draft.BirthDate = birthDate
	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if isEdit {
		if missingStep := s.nextRequiredStep(draft); missingStep != "" {
			draft.Step = missingStep
			if err := s.drafts.Save(ctx, draft); err != nil {
				return "Failed to save profile edit. Please try again later.", err
			}
			return s.editPrompt(missingStep), nil
		}

		if err := s.drafts.Save(ctx, draft); err != nil {
			return "Failed to save profile edit. Please try again later.", err
		}

		return s.saveProfile(ctx, draft)
	}

	draft.Step = domain.ProfileDraftStepCountry
	if err := s.drafts.Save(ctx, draft); err != nil {
		return "Failed to save profile setup. Please try again later.", err
	}

	return s.countryPrompt(), nil
}

func (s *service) applyCountry(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft) (string, error) {
	value := strings.TrimSpace(msg.Text)
	draft.Mode = s.draftMode(draft)
	isEdit := draft.Mode == domain.ProfileDraftModeEdit
	country, ok := s.normalizeCountry(value)
	if !ok {
		text := "Country must be one of: Russia, Kazakhstan, or Belarus."
		if isEdit {
			return s.editText(domain.ProfileDraftStepCountry, text), nil
		}
		return s.stepText(domain.ProfileDraftStepCountry, text+" Try again."), nil
	}

	previousCountry := draft.Country
	draft.Country = country
	if previousCountry != country {
		draft.City = ""
	} else if draft.City != "" {
		if normalizedCity, ok := s.normalizeCity(country, draft.City); ok {
			draft.City = normalizedCity
		} else {
			draft.City = ""
		}
	}
	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if isEdit {
		if missingStep := s.nextRequiredStep(draft); missingStep != "" {
			draft.Step = missingStep
			if err := s.drafts.Save(ctx, draft); err != nil {
				return "Failed to save profile edit. Please try again later.", err
			}
			return s.editPrompt(missingStep), nil
		}

		if err := s.drafts.Save(ctx, draft); err != nil {
			return "Failed to save profile edit. Please try again later.", err
		}

		return s.saveProfile(ctx, draft)
	}

	draft.Step = domain.ProfileDraftStepCity
	if err := s.drafts.Save(ctx, draft); err != nil {
		return "Failed to save profile setup. Please try again later.", err
	}

	return s.cityPrompt(), nil
}

func (s *service) applyCity(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft) (string, error) {
	draft.Mode = s.draftMode(draft)
	isEdit := draft.Mode == domain.ProfileDraftModeEdit
	city, ok := s.normalizeCity(draft.Country, msg.Text)
	if !ok {
		text := "City must be selected from the list for your country."
		if isEdit {
			return s.editText(domain.ProfileDraftStepCity, text), nil
		}
		return s.stepText(domain.ProfileDraftStepCity, text+" Try again."), nil
	}

	draft.City = city
	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if isEdit {
		if missingStep := s.nextRequiredStep(draft); missingStep != "" {
			draft.Step = missingStep
			if err := s.drafts.Save(ctx, draft); err != nil {
				return "Failed to save profile edit. Please try again later.", err
			}
			return s.editPrompt(missingStep), nil
		}

		if err := s.drafts.Save(ctx, draft); err != nil {
			return "Failed to save profile edit. Please try again later.", err
		}

		return s.saveProfile(ctx, draft)
	}

	draft.Step = domain.ProfileDraftStepDescription
	if err := s.drafts.Save(ctx, draft); err != nil {
		return "Failed to save profile setup. Please try again later.", err
	}

	return s.descriptionPrompt(), nil
}

func (s *service) applyDescription(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft) (string, error) {
	description := strings.TrimSpace(msg.Text)
	draft.Mode = s.draftMode(draft)
	isEdit := draft.Mode == domain.ProfileDraftModeEdit
	if description == "" {
		if isEdit {
			return s.editText(domain.ProfileDraftStepDescription, "Description cannot be empty. Please try again."), nil
		}
		return s.stepText(domain.ProfileDraftStepDescription, "Description cannot be empty. Please try again."), nil
	}
	if len(description) > maxDescriptionLength {
		if isEdit {
			return s.editText(domain.ProfileDraftStepDescription, fmt.Sprintf("Description is too long (max %d characters).", maxDescriptionLength)), nil
		}
		return s.stepText(domain.ProfileDraftStepDescription, fmt.Sprintf("Description is too long (max %d characters).", maxDescriptionLength)), nil
	}

	draft.Description = description
	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if err := s.drafts.Save(ctx, draft); err != nil {
		if isEdit {
			return "Failed to save profile edit. Please try again later.", err
		}
		return "Failed to save profile setup. Please try again later.", err
	}

	if isEdit {
		return s.saveProfile(ctx, draft)
	}

	draft.Step = domain.ProfileDraftStepEmoji
	if err := s.drafts.Save(ctx, draft); err != nil {
		return "Failed to save profile setup. Please try again later.", err
	}

	return s.emojiPrompt(), nil
}

func (s *service) applyEmoji(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft) (string, error) {
	value := strings.TrimSpace(msg.Text)
	draft.Mode = s.draftMode(draft)
	isEdit := draft.Mode == domain.ProfileDraftModeEdit
	code, ok := s.normalizeEmojiCode(value)
	if !ok {
		text := "Emoji must be selected from the list."
		if isEdit {
			return s.editText(domain.ProfileDraftStepEmoji, text), nil
		}
		return s.stepText(domain.ProfileDraftStepEmoji, text+" Try again."), nil
	}

	draft.EmojiCode = code
	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if isEdit {
		if missingStep := s.nextRequiredStep(draft); missingStep != "" {
			draft.Step = missingStep
			if err := s.drafts.Save(ctx, draft); err != nil {
				return "Failed to save profile edit. Please try again later.", err
			}
			return s.editPrompt(missingStep), nil
		}

		if err := s.drafts.Save(ctx, draft); err != nil {
			return "Failed to save profile edit. Please try again later.", err
		}

		return s.saveProfile(ctx, draft)
	}

	draft.Step = domain.ProfileDraftStepPhotos
	if err := s.drafts.Save(ctx, draft); err != nil {
		return "Failed to save profile setup. Please try again later.", err
	}

	return s.photosPrompt(), nil
}

func (s *service) saveProfile(ctx context.Context, draft domain.ProfileDraft) (string, error) {
	if s.profiles == nil {
		return "Profile service is not available right now.", errors.New("profile service is not configured")
	}

	if err := s.profiles.CreateProfile(ctx, domain.Profile{
		UserID:      draft.UserID,
		Name:        draft.Name,
		Gender:      draft.Gender,
		BirthDate:   draft.BirthDate,
		Country:     draft.Country,
		City:        draft.City,
		Description: draft.Description,
		EmojiCode:   draft.EmojiCode,
		Photos:      draft.Photos,
	}); err != nil {
		return "Failed to save profile. Please try again later.", err
	}

	success := s.profileCreated()
	if s.draftMode(draft) == domain.ProfileDraftModeEdit {
		success = s.profileUpdated()
	}

	if err := s.drafts.Delete(ctx, draft.UserID); err != nil {
		return success + "\nNote: could not clear the profile draft.", err
	}

	return success, nil
}

func (s *service) applyPhotos(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft) (string, error) {
	draft.Mode = s.draftMode(draft)
	isEdit := draft.Mode == domain.ProfileDraftModeEdit

	addedPhotos := false
	if len(msg.PhotoIDs) > 0 {
		draft.Photos, _ = s.mergePhotoIDs(draft.Photos, msg.PhotoIDs, maxPhotos)
		draft.UpdatedAt = s.now(msg.ReceivedAt)
		if err := s.drafts.Save(ctx, draft); err != nil {
			return "Failed to save profile setup. Please try again later.", err
		}
		addedPhotos = true
	}

	if s.isAlbumDone(msg.Text) {
		if len(draft.Photos) < minPhotos {
			return s.photosPromptText(isEdit, fmt.Sprintf("Please send at least %d photo before finishing.", minPhotos)), nil
		}

		if !addedPhotos {
			draft.UpdatedAt = s.now(msg.ReceivedAt)
			if err := s.drafts.Save(ctx, draft); err != nil {
				return "Failed to save profile setup. Please try again later.", err
			}
		}

		return s.saveProfile(ctx, draft)
	}

	if !addedPhotos {
		return s.photosPromptText(isEdit, "Send 1-4 photos for your album. Use the Done button when finished."), nil
	}

	if len(draft.Photos) >= maxPhotos {
		return s.photosPromptText(isEdit, fmt.Sprintf("You reached the limit of %d photos. Use the Done button to finish.", maxPhotos)), nil
	}

	return s.photosPromptText(isEdit, fmt.Sprintf("Photos in album: %d/%d. Send more or use the Done button.", len(draft.Photos), maxPhotos)), nil
}

func (s *service) nextRequiredStep(draft domain.ProfileDraft) domain.ProfileDraftStep {
	if draft.BirthDate.IsZero() {
		return domain.ProfileDraftStepBirthDate
	}
	if draft.Country == "" {
		return domain.ProfileDraftStepCountry
	}
	if strings.TrimSpace(draft.City) == "" {
		return domain.ProfileDraftStepCity
	}
	if draft.EmojiCode == "" {
		return domain.ProfileDraftStepEmoji
	}

	return ""
}
