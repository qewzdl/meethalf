package bot

import (
	"context"
	"errors"

	"meethalf-telegram-bot/internal/domain"
)

type Usecase interface {
	Handle(ctx context.Context, msg domain.IncomingMessage) ([]domain.OutgoingMessage, error)
}

type SessionRepository interface {
	Touch(ctx context.Context, session domain.Session) error
}

type ProfileDraftRepository interface {
	Get(ctx context.Context, userID int64) (domain.ProfileDraft, bool, error)
	Save(ctx context.Context, draft domain.ProfileDraft) error
	Delete(ctx context.Context, userID int64) error
}

type ProfileDeletionConfirmationRepository interface {
	Get(ctx context.Context, userID int64) (domain.ProfileDeletionConfirmation, bool, error)
	Save(ctx context.Context, confirmation domain.ProfileDeletionConfirmation) error
	Delete(ctx context.Context, userID int64) error
}

type ProfileService interface {
	CreateProfile(ctx context.Context, profile domain.Profile) error
	GetProfile(ctx context.Context, userID int64) (domain.Profile, bool, error)
	DeleteProfile(ctx context.Context, userID int64) (bool, error)
}

type service struct {
	sessions            SessionRepository
	drafts              ProfileDraftRepository
	deleteConfirmations ProfileDeletionConfirmationRepository
	profiles            ProfileService
	helpText            string
}

func New(sessions SessionRepository, drafts ProfileDraftRepository, deleteConfirmations ProfileDeletionConfirmationRepository, profiles ProfileService) Usecase {
	return &service{
		sessions:            sessions,
		drafts:              drafts,
		deleteConfirmations: deleteConfirmations,
		profiles:            profiles,
		helpText:            defaultHelpText,
	}
}

func (s *service) Handle(ctx context.Context, msg domain.IncomingMessage) ([]domain.OutgoingMessage, error) {
	var touchErr error
	if s != nil && s.sessions != nil {
		touchErr = s.sessions.Touch(ctx, domain.Session{
			UserID:   msg.User.ID,
			ChatID:   msg.ChatID,
			LastSeen: s.now(msg.ReceivedAt),
		})
	}

	response := domain.OutgoingMessage{ChatID: msg.ChatID}
	var replyErr error
	if msg.Command == domain.CommandStart {
		response.Text, response.InlineKeyboard, replyErr = s.startMessage(ctx, msg)
	} else {
		response.Text, replyErr = s.reply(ctx, msg)
		if msg.Command == domain.CommandProfileEdit && replyErr == nil {
			response.InlineKeyboard = s.profileEditMenuKeyboard()
		}
		if msg.Command == domain.CommandProfileDelete && replyErr == nil {
			response.InlineKeyboard = s.profileDeleteConfirmInlineKeyboard()
		}
		if msg.Command == domain.CommandProfileDeleteConfirm && replyErr == nil {
			if response.Text == profileDeleteExpiredText {
				response.InlineKeyboard = s.profileSettingsInlineKeyboard()
			} else {
				response.InlineKeyboard = s.profileCreateInlineKeyboard()
			}
		}
		if msg.Command == domain.CommandProfileDeleteCancel && replyErr == nil {
			response.InlineKeyboard = s.profileSettingsInlineKeyboard()
		}
	}

	if replyErr == nil {
		s.attachTelegramNameKeyboard(ctx, msg, &response)
		s.attachGenderKeyboard(ctx, msg, &response)
		s.attachCountryKeyboard(ctx, msg, &response)
		s.attachCityKeyboard(ctx, msg, &response)
		s.attachEmojiKeyboard(ctx, msg, &response)
		s.attachPhotosDoneKeyboard(ctx, msg, &response)
	}

	messages := []domain.OutgoingMessage{response}
	if replyErr == nil && s != nil && s.profiles != nil && msg.User.ID != 0 {
		switch msg.Command {
		case domain.CommandProfileView:
			profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
			if err != nil {
				replyErr = errors.Join(replyErr, err)
				break
			}
			if found {
				response.InlineKeyboard = s.profileViewInlineKeyboard()
				if len(profile.Photos) > 0 {
					messages = s.profileAlbumMessages(msg.ChatID, profile, s.profileViewInlineKeyboard())
				} else {
					messages[0] = response
				}
			} else {
				response.InlineKeyboard = s.profileCreateInlineKeyboard()
				messages[0] = response
			}
		case domain.CommandProfilePreview:
			profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
			if err != nil {
				replyErr = errors.Join(replyErr, err)
				break
			}
			if found {
				response.InlineKeyboard = s.profilePreviewInlineKeyboard()
				if len(profile.Photos) > 0 {
					messages = s.profilePreviewAlbumMessages(msg.ChatID, profile, s.profilePreviewInlineKeyboard())
				} else {
					messages[0] = response
				}
			} else {
				response.InlineKeyboard = s.profileCreateInlineKeyboard()
				messages[0] = response
			}
		case domain.CommandProfileSettings:
			_, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
			if err != nil {
				replyErr = errors.Join(replyErr, err)
				break
			}
			if found {
				response.InlineKeyboard = s.profileSettingsInlineKeyboard()
			} else {
				response.Text = "Profile not found. Use the Create Profile button to create it."
				response.InlineKeyboard = s.profileCreateInlineKeyboard()
			}
			messages[0] = response
		}
	}

	followUps, followUpErr := s.profileDetailsFollowUp(ctx, msg, response.Text, replyErr)
	if len(followUps) > 0 {
		messages = append(messages, followUps...)
	}

	return messages, errors.Join(touchErr, replyErr, followUpErr)
}

func (s *service) attachGenderKeyboard(ctx context.Context, msg domain.IncomingMessage, response *domain.OutgoingMessage) {
	if response == nil || response.InlineKeyboard != nil || s == nil || s.drafts == nil || msg.User.ID == 0 {
		return
	}

	if msg.Command != "" && msg.Command != domain.CommandProfileEditGender {
		return
	}

	draft, found, err := s.drafts.Get(ctx, msg.User.ID)
	if err != nil || !found {
		return
	}

	if draft.Step != domain.ProfileDraftStepGender {
		return
	}

	if msg.Command == domain.CommandProfileEditGender && s.draftMode(draft) != domain.ProfileDraftModeEdit {
		return
	}

	response.InlineKeyboard = s.genderInlineKeyboard()
}

func (s *service) attachTelegramNameKeyboard(ctx context.Context, msg domain.IncomingMessage, response *domain.OutgoingMessage) {
	if response == nil || response.InlineKeyboard != nil || s == nil || s.drafts == nil || msg.User.ID == 0 {
		return
	}

	if msg.Command != "" && msg.Command != domain.CommandProfile && msg.Command != domain.CommandProfileEditName {
		return
	}

	draft, found, err := s.drafts.Get(ctx, msg.User.ID)
	if err != nil || !found {
		return
	}

	if draft.Step != domain.ProfileDraftStepName {
		return
	}

	if s.userFullName(msg.User) == "" {
		return
	}

	response.InlineKeyboard = s.telegramNameInlineKeyboard()
}

func (s *service) attachCountryKeyboard(ctx context.Context, msg domain.IncomingMessage, response *domain.OutgoingMessage) {
	if response == nil || response.InlineKeyboard != nil || s == nil || s.drafts == nil || msg.User.ID == 0 {
		return
	}

	if msg.Command != "" && !s.isProfileEditAction(msg.Command) {
		return
	}

	draft, found, err := s.drafts.Get(ctx, msg.User.ID)
	if err != nil || !found {
		return
	}

	if draft.Step != domain.ProfileDraftStepCountry {
		return
	}

	if msg.Command == domain.CommandProfileEditCountry && s.draftMode(draft) != domain.ProfileDraftModeEdit {
		return
	}

	response.InlineKeyboard = s.countryInlineKeyboard()
}

func (s *service) attachCityKeyboard(ctx context.Context, msg domain.IncomingMessage, response *domain.OutgoingMessage) {
	if response == nil || response.InlineKeyboard != nil || s == nil || s.drafts == nil || msg.User.ID == 0 {
		return
	}

	if msg.Command != "" && !s.isProfileEditAction(msg.Command) {
		return
	}

	draft, found, err := s.drafts.Get(ctx, msg.User.ID)
	if err != nil || !found {
		return
	}

	if draft.Step != domain.ProfileDraftStepCity {
		return
	}

	if msg.Command == domain.CommandProfileEditCity && s.draftMode(draft) != domain.ProfileDraftModeEdit {
		return
	}

	keyboard := s.cityInlineKeyboard(draft.Country)
	if keyboard == nil {
		return
	}

	response.InlineKeyboard = keyboard
}

func (s *service) attachEmojiKeyboard(ctx context.Context, msg domain.IncomingMessage, response *domain.OutgoingMessage) {
	if response == nil || response.InlineKeyboard != nil || s == nil || s.drafts == nil || msg.User.ID == 0 {
		return
	}

	if msg.Command != "" && !s.isProfileEditAction(msg.Command) {
		return
	}

	draft, found, err := s.drafts.Get(ctx, msg.User.ID)
	if err != nil || !found {
		return
	}

	if draft.Step != domain.ProfileDraftStepEmoji {
		return
	}

	if msg.Command == domain.CommandProfileEditEmoji && s.draftMode(draft) != domain.ProfileDraftModeEdit {
		return
	}

	response.InlineKeyboard = s.emojiInlineKeyboard()
}

func (s *service) attachPhotosDoneKeyboard(ctx context.Context, msg domain.IncomingMessage, response *domain.OutgoingMessage) {
	if response == nil || response.InlineKeyboard != nil || s == nil || s.drafts == nil || msg.User.ID == 0 {
		return
	}

	if msg.Command != "" && !s.isProfileEditAction(msg.Command) {
		return
	}

	draft, found, err := s.drafts.Get(ctx, msg.User.ID)
	if err != nil || !found {
		return
	}

	if draft.Step != domain.ProfileDraftStepPhotos {
		return
	}

	if len(draft.Photos) < minPhotos {
		return
	}

	if msg.Command == domain.CommandProfileEditPhotos && s.draftMode(draft) != domain.ProfileDraftModeEdit {
		return
	}

	response.InlineKeyboard = s.photosDoneInlineKeyboard()
}
