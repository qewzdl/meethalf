package bot

import (
	"context"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) reply(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, error) {
	if msg.Command != "" {
		return s.handleCommand(ctx, msg, l)
	}

	if s != nil && s.drafts != nil && msg.User.ID != 0 {
		draft, found, err := s.drafts.Get(ctx, msg.User.ID)
		if err != nil {
			return l.message(msgProfileSetupLoadFailed), err
		}
		if found {
			return s.handleDraft(ctx, msg, draft, l)
		}
	}

	if s.isProfileShortcut(msg.Text) {
		return s.startProfileSetup(ctx, msg, l)
	}

	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	return s.helpTextFor(l, role), roleErr
}

func (s *service) handleCommand(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, error) {
	switch msg.Command {
	case domain.CommandStart:
		text, _, _, err := s.startText(ctx, msg, l)
		return text, err
	case domain.CommandCancel:
		return s.cancelAction(ctx, msg, l)
	case domain.CommandProfileView:
		return s.showProfile(ctx, msg, l)
	case domain.CommandProfilePreview:
		return s.showProfilePreview(ctx, msg, l)
	case domain.CommandProfileEdit:
		return s.profileEditMenu(ctx, msg, l)
	case domain.CommandProfileEditName:
		return s.startProfileEdit(ctx, msg, domain.ProfileDraftStepName, l)
	case domain.CommandProfileEditGender:
		return s.startProfileEdit(ctx, msg, domain.ProfileDraftStepGender, l)
	case domain.CommandProfileEditBirthDate:
		return s.startProfileEdit(ctx, msg, domain.ProfileDraftStepBirthDate, l)
	case domain.CommandProfileEditCountry:
		return s.startProfileEdit(ctx, msg, domain.ProfileDraftStepCountry, l)
	case domain.CommandProfileEditCity:
		return s.startProfileEdit(ctx, msg, domain.ProfileDraftStepCity, l)
	case domain.CommandProfileEditDesc:
		return s.startProfileEdit(ctx, msg, domain.ProfileDraftStepDescription, l)
	case domain.CommandProfileEditEmoji:
		return s.startProfileEdit(ctx, msg, domain.ProfileDraftStepEmoji, l)
	case domain.CommandProfileEditPhotos:
		return s.startProfileEdit(ctx, msg, domain.ProfileDraftStepPhotos, l)
	case domain.CommandProfile:
		return s.startProfileSetup(ctx, msg, l)
	case domain.CommandProfileSetupBack:
		return s.profileSetupBack(ctx, msg, l)
	case domain.CommandProfileSettings:
		return s.profileSettingsText(l), nil
	case domain.CommandProfileVisibility:
		return s.updateProfileVisibility(ctx, msg, l)
	case domain.CommandProfileDelete:
		return s.requestProfileDelete(ctx, msg, l)
	case domain.CommandProfileDeleteConfirm:
		return s.confirmProfileDelete(ctx, msg, l)
	case domain.CommandProfileDeleteCancel:
		return s.cancelProfileDelete(ctx, msg, l)
	case "":
		role, roleErr := s.resolveAdminRole(ctx, msg.User)
		return s.helpTextFor(l, role), roleErr
	default:
		role, roleErr := s.resolveAdminRole(ctx, msg.User)
		return l.message(msgUnknownCommand) + "\n" + s.helpTextFor(l, role), roleErr
	}
}

func (s *service) startMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	text, status, role, err := s.startText(ctx, msg, l)
	if isBannedError(err) {
		return text, nil, err
	}
	return text, s.startInlineKeyboardByStatus(l, status, role), err
}

func (s *service) startText(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, profileStatus, adminRole, error) {
	profile, status, err := s.resolveProfileStatus(ctx, msg.User.ID)
	if isBannedError(err) {
		return s.userBannedText(l), profileStatusUnknown, adminRoleNone, err
	}
	role := s.adminRoleForProfile(msg.User, profile, status)
	greeting := s.startGreeting(l, msg.User, profile, status)
	return greeting + "\n" + s.helpTextFor(l, role), status, role, err
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
