package bot

import (
	"context"
	"errors"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) handleDraft(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft, l localizer) (string, error) {
	draft = s.ensureProfileSetupStart(ctx, msg, draft)
	if s.draftMode(draft) == domain.ProfileDraftModeEdit {
		switch draft.Step {
		case domain.ProfileDraftStepName:
			return s.applyName(ctx, msg, draft, l)
		case domain.ProfileDraftStepGender:
			return s.applyGender(ctx, msg, draft, l)
		case domain.ProfileDraftStepBirthDate:
			return s.applyBirthDate(ctx, msg, draft, l)
		case domain.ProfileDraftStepCountry:
			return s.applyCountry(ctx, msg, draft, l)
		case domain.ProfileDraftStepCity:
			return s.applyCity(ctx, msg, draft, l)
		case domain.ProfileDraftStepDescription:
			return s.applyDescription(ctx, msg, draft, l)
		case domain.ProfileDraftStepEmoji:
			return s.applyEmoji(ctx, msg, draft, l)
		case domain.ProfileDraftStepPhotos:
			return s.applyPhotos(ctx, msg, draft, l)
		default:
			return s.profileEditMenuText(l), nil
		}
	}

	switch draft.Step {
	case domain.ProfileDraftStepBotCheck:
		return s.applyBotCheck(ctx, msg, draft, l)
	case domain.ProfileDraftStepName:
		return s.applyName(ctx, msg, draft, l)
	case domain.ProfileDraftStepGender:
		return s.applyGender(ctx, msg, draft, l)
	case domain.ProfileDraftStepBirthDate:
		return s.applyBirthDate(ctx, msg, draft, l)
	case domain.ProfileDraftStepCountry:
		return s.applyCountry(ctx, msg, draft, l)
	case domain.ProfileDraftStepCity:
		return s.applyCity(ctx, msg, draft, l)
	case domain.ProfileDraftStepDescription:
		return s.applyDescription(ctx, msg, draft, l)
	case domain.ProfileDraftStepEmoji:
		return s.applyEmoji(ctx, msg, draft, l)
	case domain.ProfileDraftStepPhotos:
		return s.applyPhotos(ctx, msg, draft, l)
	default:
		return s.startProfileSetup(ctx, msg, l)
	}
}

func (s *service) startProfileSetup(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, error) {
	if s == nil || s.drafts == nil {
		return l.message(msgProfileSetupUnavailable), errors.New("profile draft repository is not configured")
	}

	if msg.User.ID == 0 {
		return l.message(msgProfileSetupUnavailableChat), errors.New("user id is missing")
	}

	draft := domain.ProfileDraft{
		UserID:              msg.User.ID,
		ChatID:              msg.ChatID,
		SetupStartMessageID: msg.MessageID,
		Step:                domain.ProfileDraftStepBotCheck,
		Mode:                domain.ProfileDraftModeCreate,
		UpdatedAt:           s.now(msg.ReceivedAt),
	}
	s.resetBotCheck(&draft, msg.ReceivedAt)

	if err := s.drafts.Save(ctx, draft); err != nil {
		return l.message(msgProfileSetupStartFailed), err
	}

	return s.botCheckPrompt(l, draft.BotCheckQuestion), nil
}

func (s *service) profileSetupBack(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, error) {
	if s == nil || s.drafts == nil {
		return l.message(msgProfileSetupUnavailable), errors.New("profile draft repository is not configured")
	}

	if msg.User.ID == 0 {
		return l.message(msgProfileSetupUnavailableChat), errors.New("user id is missing")
	}

	draft, found, err := s.drafts.Get(ctx, msg.User.ID)
	if err != nil {
		return l.message(msgProfileSetupLoadFailed), err
	}
	if !found {
		return l.message(msgProfileSetupNotActive), nil
	}

	if s.draftMode(draft) != domain.ProfileDraftModeCreate {
		return s.editPrompt(l, draft.Step), nil
	}

	previousStep := s.previousProfileSetupStep(draft.Step)
	if previousStep == "" {
		return s.profileSetupPrompt(l, draft.Step, draft, msg.User), nil
	}

	draft.Step = previousStep
	if previousStep == domain.ProfileDraftStepBotCheck {
		s.ensureBotCheck(&draft, msg.ReceivedAt)
	}
	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if err := s.drafts.Save(ctx, draft); err != nil {
		return l.message(msgProfileSetupSaveFailed), err
	}

	return s.profileSetupPrompt(l, previousStep, draft, msg.User), nil
}

func (s *service) startProfileEdit(ctx context.Context, msg domain.IncomingMessage, step domain.ProfileDraftStep, l localizer) (string, error) {
	if s == nil || s.drafts == nil {
		return l.message(msgProfileEditUnavailable), errors.New("profile draft repository is not configured")
	}

	if s.profiles == nil {
		return l.message(msgProfileServiceUnavailable), errors.New("profile service is not configured")
	}

	if msg.User.ID == 0 {
		return l.message(msgProfileEditUnavailableChat), errors.New("user id is missing")
	}

	profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
	if err != nil {
		if isBannedError(err) {
			return s.userBannedText(l), err
		}
		return l.message(msgProfileLoadFailed), err
	}
	if !found {
		return l.message(msgProfileNotFoundUseProfile), nil
	}

	stepToEdit := step
	if profile.BirthDate.IsZero() && step != domain.ProfileDraftStepBirthDate {
		stepToEdit = domain.ProfileDraftStepBirthDate
	} else if !profile.BirthDate.IsZero() && profile.Country == "" && step != domain.ProfileDraftStepCountry {
		stepToEdit = domain.ProfileDraftStepCountry
	} else if profile.Country != "" && strings.TrimSpace(profile.City) == "" && step != domain.ProfileDraftStepCity {
		stepToEdit = domain.ProfileDraftStepCity
	} else if profile.EmojiCode == "" && step != domain.ProfileDraftStepEmoji {
		stepToEdit = domain.ProfileDraftStepEmoji
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
		IsHidden:    profile.IsHidden,
		Mode:        domain.ProfileDraftModeEdit,
		UpdatedAt:   s.now(msg.ReceivedAt),
	}
	if stepToEdit == domain.ProfileDraftStepPhotos {
		draft.Photos = nil
	}

	if err := s.drafts.Save(ctx, draft); err != nil {
		return l.message(msgProfileEditStartFailed), err
	}

	return s.editPrompt(l, stepToEdit), nil
}

func (s *service) deleteProfile(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, error) {
	if s == nil || s.profiles == nil {
		return l.message(msgProfileServiceUnavailable), errors.New("profile service is not configured")
	}

	if msg.User.ID == 0 {
		return l.message(msgProfileDeleteUnavailableChat), errors.New("user id is missing")
	}

	deleted, err := s.profiles.DeleteProfile(ctx, msg.User.ID)
	if err != nil {
		if isBannedError(err) {
			return s.userBannedText(l), err
		}
		return l.message(msgProfileDeleteFailed), err
	}
	if !deleted {
		return l.message(msgProfileNotFoundCreateButton), nil
	}

	if s.drafts != nil {
		if err := s.drafts.Delete(ctx, msg.User.ID); err != nil {
			return l.message(msgProfileDeletedDraftWarning), err
		}
	}

	return l.message(msgProfileDeleted), nil
}

func (s *service) applyBotCheck(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft, l localizer) (string, error) {
	if s.drafts == nil {
		return l.message(msgProfileSetupUnavailable), errors.New("profile draft repository is not configured")
	}

	s.ensureBotCheck(&draft, msg.ReceivedAt)
	answer := strings.TrimSpace(msg.Text)
	if answer == "" {
		return s.botCheckPrompt(l, draft.BotCheckQuestion), nil
	}

	if s.botCheckMatches(draft, answer) {
		draft.Step = domain.ProfileDraftStepName
		draft.BotCheckQuestion = ""
		draft.BotCheckAnswer = 0
		draft.BotCheckAttempts = 0
		draft.UpdatedAt = s.now(msg.ReceivedAt)
		if err := s.drafts.Save(ctx, draft); err != nil {
			return l.message(msgProfileSetupSaveFailed), err
		}

		return s.namePrompt(l, msg.User), nil
	}

	draft.BotCheckAttempts++
	if draft.BotCheckAttempts >= botCheckMaxAttempts {
		s.resetBotCheck(&draft, msg.ReceivedAt)
		if err := s.drafts.Save(ctx, draft); err != nil {
			return l.message(msgProfileSetupSaveFailed), err
		}
		return s.botCheckRetryPrompt(l, l.message(msgBotCheckTooManyAttempts), draft.BotCheckQuestion), nil
	}

	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if err := s.drafts.Save(ctx, draft); err != nil {
		return l.message(msgProfileSetupSaveFailed), err
	}

	return s.botCheckRetryPrompt(l, l.message(msgBotCheckIncorrect), draft.BotCheckQuestion), nil
}

func (s *service) applyName(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft, l localizer) (string, error) {
	value := strings.TrimSpace(msg.Text)
	draft.Mode = s.draftMode(draft)
	isEdit := draft.Mode == domain.ProfileDraftModeEdit
	if value == "" {
		if isEdit {
			return s.editText(l, domain.ProfileDraftStepName, l.message(msgNamePromptEmpty)), nil
		}
		return s.stepText(l, domain.ProfileDraftStepName, l.message(msgNamePromptEmptyCreate)), nil
	}

	name := value
	if s.isAffirmative(value) {
		telegramName := s.userFullName(msg.User)
		if telegramName == "" {
			if isEdit {
				return s.editText(l, domain.ProfileDraftStepName, l.message(msgNamePromptTelegramMissing)), nil
			}
			return s.stepText(l, domain.ProfileDraftStepName, l.message(msgNamePromptTelegramMissing)), nil
		}
		name = telegramName
	}

	if len(name) > maxNameLength {
		if isEdit {
			return s.editText(l, domain.ProfileDraftStepName, l.message(msgNameTooLong, maxNameLength)), nil
		}
		return s.stepText(l, domain.ProfileDraftStepName, l.message(msgNameTooLong, maxNameLength)), nil
	}

	draft.Name = name
	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if isEdit {
		if err := s.drafts.Save(ctx, draft); err != nil {
			return l.message(msgProfileEditSaveFailed), err
		}

		return s.saveProfile(ctx, draft, l)
	}

	draft.Step = domain.ProfileDraftStepGender
	if err := s.drafts.Save(ctx, draft); err != nil {
		return l.message(msgProfileSetupSaveFailed), err
	}

	return s.genderPrompt(l), nil
}

func (s *service) applyGender(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft, l localizer) (string, error) {
	value := strings.TrimSpace(msg.Text)
	draft.Mode = s.draftMode(draft)
	isEdit := draft.Mode == domain.ProfileDraftModeEdit
	gender, ok := s.normalizeGender(value)
	if !ok {
		if isEdit {
			return s.editText(l, domain.ProfileDraftStepGender, l.message(msgGenderInvalid)), nil
		}
		return s.stepText(l, domain.ProfileDraftStepGender, l.message(msgGenderInvalid)), nil
	}

	draft.Gender = gender
	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if isEdit {
		if err := s.drafts.Save(ctx, draft); err != nil {
			return l.message(msgProfileEditSaveFailed), err
		}

		return s.saveProfile(ctx, draft, l)
	}

	draft.Step = domain.ProfileDraftStepBirthDate
	if err := s.drafts.Save(ctx, draft); err != nil {
		return l.message(msgProfileSetupSaveFailed), err
	}

	return s.birthDatePrompt(l), nil
}

func (s *service) applyBirthDate(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft, l localizer) (string, error) {
	value := strings.TrimSpace(msg.Text)
	draft.Mode = s.draftMode(draft)
	isEdit := draft.Mode == domain.ProfileDraftModeEdit
	birthDate, ok := s.parseBirthDate(value)
	if !ok {
		if isEdit {
			return s.editText(l, domain.ProfileDraftStepBirthDate, l.message(msgBirthDateInvalid, birthDateLayout)), nil
		}
		return s.stepText(l, domain.ProfileDraftStepBirthDate, l.message(msgBirthDateInvalid, birthDateLayout)), nil
	}

	age := s.ageFromBirthDate(birthDate, s.now(msg.ReceivedAt))
	if age < minAge || age > maxAge {
		if isEdit {
			return s.editText(l, domain.ProfileDraftStepBirthDate, l.message(msgAgeInvalid, minAge, maxAge)), nil
		}
		return s.stepText(l, domain.ProfileDraftStepBirthDate, l.message(msgAgeInvalid, minAge, maxAge)), nil
	}

	draft.BirthDate = birthDate
	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if isEdit {
		if missingStep := s.nextRequiredStep(draft); missingStep != "" {
			draft.Step = missingStep
			if err := s.drafts.Save(ctx, draft); err != nil {
				return l.message(msgProfileEditSaveFailed), err
			}
			return s.editPrompt(l, missingStep), nil
		}

		if err := s.drafts.Save(ctx, draft); err != nil {
			return l.message(msgProfileEditSaveFailed), err
		}

		return s.saveProfile(ctx, draft, l)
	}

	draft.Step = domain.ProfileDraftStepCountry
	if err := s.drafts.Save(ctx, draft); err != nil {
		return l.message(msgProfileSetupSaveFailed), err
	}

	return s.countryPrompt(l), nil
}

func (s *service) applyCountry(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft, l localizer) (string, error) {
	value := strings.TrimSpace(msg.Text)
	draft.Mode = s.draftMode(draft)
	isEdit := draft.Mode == domain.ProfileDraftModeEdit
	country, ok := s.normalizeCountry(value)
	if !ok {
		if isEdit {
			return s.editText(l, domain.ProfileDraftStepCountry, l.message(msgCountryInvalid)), nil
		}
		return s.stepText(l, domain.ProfileDraftStepCountry, l.message(msgCountryInvalid)), nil
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
				return l.message(msgProfileEditSaveFailed), err
			}
			return s.editPrompt(l, missingStep), nil
		}

		if err := s.drafts.Save(ctx, draft); err != nil {
			return l.message(msgProfileEditSaveFailed), err
		}

		return s.saveProfile(ctx, draft, l)
	}

	draft.Step = domain.ProfileDraftStepCity
	if err := s.drafts.Save(ctx, draft); err != nil {
		return l.message(msgProfileSetupSaveFailed), err
	}

	return s.cityPrompt(l), nil
}

func (s *service) applyCity(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft, l localizer) (string, error) {
	draft.Mode = s.draftMode(draft)
	isEdit := draft.Mode == domain.ProfileDraftModeEdit
	city, ok := s.normalizeCity(draft.Country, msg.Text)
	if !ok {
		if isEdit {
			return s.editText(l, domain.ProfileDraftStepCity, l.message(msgCityInvalid)), nil
		}
		return s.stepText(l, domain.ProfileDraftStepCity, l.message(msgCityInvalid)), nil
	}

	draft.City = city
	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if isEdit {
		if missingStep := s.nextRequiredStep(draft); missingStep != "" {
			draft.Step = missingStep
			if err := s.drafts.Save(ctx, draft); err != nil {
				return l.message(msgProfileEditSaveFailed), err
			}
			return s.editPrompt(l, missingStep), nil
		}

		if err := s.drafts.Save(ctx, draft); err != nil {
			return l.message(msgProfileEditSaveFailed), err
		}

		return s.saveProfile(ctx, draft, l)
	}

	draft.Step = domain.ProfileDraftStepDescription
	if err := s.drafts.Save(ctx, draft); err != nil {
		return l.message(msgProfileSetupSaveFailed), err
	}

	return s.descriptionPrompt(l), nil
}

func (s *service) applyDescription(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft, l localizer) (string, error) {
	description := strings.TrimSpace(msg.Text)
	draft.Mode = s.draftMode(draft)
	isEdit := draft.Mode == domain.ProfileDraftModeEdit
	if description == "" {
		if isEdit {
			return s.editText(l, domain.ProfileDraftStepDescription, l.message(msgDescriptionEmpty)), nil
		}
		return s.stepText(l, domain.ProfileDraftStepDescription, l.message(msgDescriptionEmpty)), nil
	}
	if len(description) > maxDescriptionLength {
		if isEdit {
			return s.editText(l, domain.ProfileDraftStepDescription, l.message(msgDescriptionTooLong, maxDescriptionLength)), nil
		}
		return s.stepText(l, domain.ProfileDraftStepDescription, l.message(msgDescriptionTooLong, maxDescriptionLength)), nil
	}

	draft.Description = description
	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if err := s.drafts.Save(ctx, draft); err != nil {
		if isEdit {
			return l.message(msgProfileEditSaveFailed), err
		}
		return l.message(msgProfileSetupSaveFailed), err
	}

	if isEdit {
		return s.saveProfile(ctx, draft, l)
	}

	draft.Step = domain.ProfileDraftStepEmoji
	if err := s.drafts.Save(ctx, draft); err != nil {
		return l.message(msgProfileSetupSaveFailed), err
	}

	return s.emojiPrompt(l), nil
}

func (s *service) applyEmoji(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft, l localizer) (string, error) {
	value := strings.TrimSpace(msg.Text)
	draft.Mode = s.draftMode(draft)
	isEdit := draft.Mode == domain.ProfileDraftModeEdit
	code, ok := s.normalizeEmojiCode(value)
	if !ok {
		if isEdit {
			return s.editText(l, domain.ProfileDraftStepEmoji, l.message(msgEmojiInvalid)), nil
		}
		return s.stepText(l, domain.ProfileDraftStepEmoji, l.message(msgEmojiInvalid)), nil
	}

	draft.EmojiCode = code
	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if isEdit {
		if missingStep := s.nextRequiredStep(draft); missingStep != "" {
			draft.Step = missingStep
			if err := s.drafts.Save(ctx, draft); err != nil {
				return l.message(msgProfileEditSaveFailed), err
			}
			return s.editPrompt(l, missingStep), nil
		}

		if err := s.drafts.Save(ctx, draft); err != nil {
			return l.message(msgProfileEditSaveFailed), err
		}

		return s.saveProfile(ctx, draft, l)
	}

	draft.Step = domain.ProfileDraftStepPhotos
	if err := s.drafts.Save(ctx, draft); err != nil {
		return l.message(msgProfileSetupSaveFailed), err
	}

	return s.photosPrompt(l), nil
}

func (s *service) saveProfile(ctx context.Context, draft domain.ProfileDraft, l localizer) (string, error) {
	if s.profiles == nil {
		return l.message(msgProfileServiceUnavailable), errors.New("profile service is not configured")
	}

	if err := s.profiles.CreateProfile(ctx, domain.Profile{
		UserID:      draft.UserID,
		Username:    s.profileUsername(ctx, draft.UserID),
		Name:        draft.Name,
		Gender:      draft.Gender,
		BirthDate:   draft.BirthDate,
		Country:     draft.Country,
		City:        draft.City,
		Description: draft.Description,
		EmojiCode:   draft.EmojiCode,
		Photos:      draft.Photos,
		IsHidden:    draft.IsHidden,
	}); err != nil {
		if isBannedError(err) {
			return s.userBannedText(l), err
		}
		return l.message(msgProfileSaveFailed), err
	}

	success := s.profileCreated(l)
	if s.draftMode(draft) == domain.ProfileDraftModeEdit {
		success = s.profileUpdated(l)
	}

	if s.draftMode(draft) == domain.ProfileDraftModeCreate {
		s.registerProfileSetupCleanup(draft)
	}

	if err := s.drafts.Delete(ctx, draft.UserID); err != nil {
		return success + "\n" + l.message(msgProfileDeletedDraftWarningOnly), err
	}

	return success, nil
}

func (s *service) applyPhotos(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft, l localizer) (string, error) {
	draft.Mode = s.draftMode(draft)
	isEdit := draft.Mode == domain.ProfileDraftModeEdit

	addedPhotos := false
	if len(msg.PhotoIDs) > 0 {
		draft.Photos, _ = s.mergePhotoIDs(draft.Photos, msg.PhotoIDs, maxPhotos)
		draft.UpdatedAt = s.now(msg.ReceivedAt)
		if err := s.drafts.Save(ctx, draft); err != nil {
			return l.message(msgProfileSetupSaveFailed), err
		}
		addedPhotos = true
	}

	if s.isAlbumDone(msg.Text) {
		if len(draft.Photos) < minPhotos {
			return s.photosPromptText(l, isEdit, l.message(msgPhotosNotEnough, minPhotos)), nil
		}

		if !addedPhotos {
			draft.UpdatedAt = s.now(msg.ReceivedAt)
			if err := s.drafts.Save(ctx, draft); err != nil {
				return l.message(msgProfileSetupSaveFailed), err
			}
		}

		return s.saveProfile(ctx, draft, l)
	}

	if !addedPhotos {
		return s.photosPromptText(l, isEdit, l.message(msgPhotosPromptRepeat)), nil
	}

	if len(draft.Photos) >= maxPhotos {
		return s.photosPromptText(l, isEdit, l.message(msgPhotosLimitReached, maxPhotos)), nil
	}

	return s.photosPromptText(l, isEdit, l.message(msgPhotosProgress, len(draft.Photos), maxPhotos)), nil
}

func (s *service) profileSetupPrompt(l localizer, step domain.ProfileDraftStep, draft domain.ProfileDraft, user domain.User) string {
	switch step {
	case domain.ProfileDraftStepBotCheck:
		return s.botCheckPrompt(l, draft.BotCheckQuestion)
	case domain.ProfileDraftStepName:
		return s.namePrompt(l, user)
	case domain.ProfileDraftStepGender:
		return s.genderPrompt(l)
	case domain.ProfileDraftStepBirthDate:
		return s.birthDatePrompt(l)
	case domain.ProfileDraftStepCountry:
		return s.countryPrompt(l)
	case domain.ProfileDraftStepCity:
		return s.cityPrompt(l)
	case domain.ProfileDraftStepDescription:
		return s.descriptionPrompt(l)
	case domain.ProfileDraftStepEmoji:
		return s.emojiPrompt(l)
	case domain.ProfileDraftStepPhotos:
		return s.photosPrompt(l)
	default:
		return s.stepText(l, step, "")
	}
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

func (s *service) ensureProfileSetupStart(ctx context.Context, msg domain.IncomingMessage, draft domain.ProfileDraft) domain.ProfileDraft {
	if s == nil || s.drafts == nil {
		return draft
	}

	if s.draftMode(draft) != domain.ProfileDraftModeCreate {
		return draft
	}

	if draft.SetupStartMessageID != 0 || msg.MessageID == 0 {
		return draft
	}

	draft.SetupStartMessageID = msg.MessageID
	draft.UpdatedAt = s.now(msg.ReceivedAt)
	if err := s.drafts.Save(ctx, draft); err != nil {
		return draft
	}

	return draft
}

func (s *service) previousProfileSetupStep(step domain.ProfileDraftStep) domain.ProfileDraftStep {
	switch step {
	case domain.ProfileDraftStepName:
		return domain.ProfileDraftStepBotCheck
	case domain.ProfileDraftStepGender:
		return domain.ProfileDraftStepName
	case domain.ProfileDraftStepBirthDate:
		return domain.ProfileDraftStepGender
	case domain.ProfileDraftStepCountry:
		return domain.ProfileDraftStepBirthDate
	case domain.ProfileDraftStepCity:
		return domain.ProfileDraftStepCountry
	case domain.ProfileDraftStepDescription:
		return domain.ProfileDraftStepCity
	case domain.ProfileDraftStepEmoji:
		return domain.ProfileDraftStepDescription
	case domain.ProfileDraftStepPhotos:
		return domain.ProfileDraftStepEmoji
	default:
		return ""
	}
}
