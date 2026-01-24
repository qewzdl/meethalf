package bot

import (
	"context"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) reply(ctx context.Context, msg domain.IncomingMessage) (string, error) {
	if msg.Command != "" {
		return s.handleCommand(ctx, msg)
	}

	if s != nil && s.drafts != nil && msg.User.ID != 0 {
		draft, found, err := s.drafts.Get(ctx, msg.User.ID)
		if err != nil {
			return "Unable to load profile setup. Please try again later.", err
		}
		if found {
			return s.handleDraft(ctx, msg, draft)
		}
	}

	if s.isProfileShortcut(msg.Text) {
		return s.startProfileSetup(ctx, msg)
	}

	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	return s.helpTextFor(msg.User, role), roleErr
}

func (s *service) handleCommand(ctx context.Context, msg domain.IncomingMessage) (string, error) {
	switch msg.Command {
	case domain.CommandStart:
		text, _, _, err := s.startText(ctx, msg)
		return text, err
	case domain.CommandCancel:
		return s.cancelAction(ctx, msg)
	case domain.CommandProfileView:
		return s.showProfile(ctx, msg)
	case domain.CommandProfilePreview:
		return s.showProfilePreview(ctx, msg)
	case domain.CommandProfileEdit:
		return s.profileEditMenu(ctx, msg)
	case domain.CommandProfileEditName:
		return s.startProfileEdit(ctx, msg, domain.ProfileDraftStepName)
	case domain.CommandProfileEditGender:
		return s.startProfileEdit(ctx, msg, domain.ProfileDraftStepGender)
	case domain.CommandProfileEditBirthDate:
		return s.startProfileEdit(ctx, msg, domain.ProfileDraftStepBirthDate)
	case domain.CommandProfileEditCountry:
		return s.startProfileEdit(ctx, msg, domain.ProfileDraftStepCountry)
	case domain.CommandProfileEditCity:
		return s.startProfileEdit(ctx, msg, domain.ProfileDraftStepCity)
	case domain.CommandProfileEditDesc:
		return s.startProfileEdit(ctx, msg, domain.ProfileDraftStepDescription)
	case domain.CommandProfileEditEmoji:
		return s.startProfileEdit(ctx, msg, domain.ProfileDraftStepEmoji)
	case domain.CommandProfileEditPhotos:
		return s.startProfileEdit(ctx, msg, domain.ProfileDraftStepPhotos)
	case domain.CommandProfile:
		return s.startProfileSetup(ctx, msg)
	case domain.CommandProfileSetupBack:
		return s.profileSetupBack(ctx, msg)
	case domain.CommandProfileSettings:
		return s.profileSettingsText(), nil
	case domain.CommandProfileVisibility:
		return s.updateProfileVisibility(ctx, msg)
	case domain.CommandProfileDelete:
		return s.requestProfileDelete(ctx, msg)
	case domain.CommandProfileDeleteConfirm:
		return s.confirmProfileDelete(ctx, msg)
	case domain.CommandProfileDeleteCancel:
		return s.cancelProfileDelete(ctx, msg)
	case "":
		role, roleErr := s.resolveAdminRole(ctx, msg.User)
		return s.helpTextFor(msg.User, role), roleErr
	default:
		role, roleErr := s.resolveAdminRole(ctx, msg.User)
		return "Unknown command.\n" + s.helpTextFor(msg.User, role), roleErr
	}
}

func (s *service) startMessage(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	text, status, role, err := s.startText(ctx, msg)
	if isBannedError(err) {
		return text, nil, err
	}
	return text, s.startInlineKeyboardByStatus(status, role), err
}

func (s *service) startText(ctx context.Context, msg domain.IncomingMessage) (string, profileStatus, adminRole, error) {
	profile, status, err := s.resolveProfileStatus(ctx, msg.User.ID)
	if isBannedError(err) {
		return s.userBannedText(), profileStatusUnknown, adminRoleNone, err
	}
	role := s.adminRoleForProfile(msg.User, profile, status)
	greeting := s.startGreeting(msg.User, profile, status)
	return greeting + "\n" + s.helpTextFor(msg.User, role), status, role, err
}

func (s *service) resolveProfileStatus(ctx context.Context, userID int64) (domain.Profile, profileStatus, error) {
	if s == nil || s.profiles == nil || userID == 0 {
		return domain.Profile{}, profileStatusUnknown, nil
	}

	profile, found, err := s.profiles.GetProfile(ctx, userID)
	if err != nil {
		return domain.Profile{}, profileStatusUnknown, err
	}
	if !found {
		return domain.Profile{}, profileStatusMissing, nil
	}

	return profile, profileStatusPresent, nil
}
