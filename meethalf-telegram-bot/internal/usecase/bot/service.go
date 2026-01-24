package bot

import (
	"context"
	"errors"
	"sync"

	"meethalf-telegram-bot/internal/domain"
)

type Usecase interface {
	Handle(ctx context.Context, msg domain.IncomingMessage) ([]domain.OutgoingMessage, error)
}

type SessionRepository interface {
	Touch(ctx context.Context, session domain.Session) error
	Get(ctx context.Context, userID int64) (domain.Session, bool, error)
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

type AdminActionRepository interface {
	Get(ctx context.Context, userID int64) (domain.AdminActionState, bool, error)
	Save(ctx context.Context, action domain.AdminActionState) error
	Delete(ctx context.Context, userID int64) error
}

type ProfileService interface {
	CreateProfile(ctx context.Context, profile domain.Profile) error
	GetProfile(ctx context.Context, userID int64) (domain.Profile, bool, error)
	DeleteProfile(ctx context.Context, userID int64) (bool, error)
	SetProfileVisibility(ctx context.Context, userID int64, isHidden bool) (bool, error)
}

type SearchService interface {
	StartSearch(ctx context.Context, userID int64, gender domain.Gender, accuracy int) (domain.MatchCandidate, bool, error)
	NextCandidate(ctx context.Context, userID int64) (domain.MatchCandidate, bool, error)
	PreviousCandidate(ctx context.Context, userID int64) (domain.MatchCandidate, bool, error)
	RecordAction(ctx context.Context, userID, targetID int64, action domain.MatchAction) (domain.MatchActionResult, error)
	PendingLikes(ctx context.Context, userID int64) ([]domain.Profile, error)
}

type AdminService interface {
	ListUsers(ctx context.Context, limit, offset int, onlyBanned, onlyModerators bool) (domain.UserList, error)
	ListReportedUsers(ctx context.Context, limit, offset int) (domain.ReportedUserList, error)
	GetUser(ctx context.Context, userID int64) (domain.UserSummary, error)
	GetUserByUsername(ctx context.Context, username string) (domain.UserSummary, error)
	BanUser(ctx context.Context, userID int64) error
	BanUserByUsername(ctx context.Context, username string) error
	UnbanUser(ctx context.Context, userID int64) error
	UnbanUserByUsername(ctx context.Context, username string) error
	MakeModerator(ctx context.Context, userID int64) error
	MakeModeratorByUsername(ctx context.Context, username string) error
	RemoveModerator(ctx context.Context, userID int64) error
	RemoveModeratorByUsername(ctx context.Context, username string) error
}

type service struct {
	sessions            SessionRepository
	drafts              ProfileDraftRepository
	deleteConfirmations ProfileDeletionConfirmationRepository
	adminActions        AdminActionRepository
	profiles            ProfileService
	search              SearchService
	admin               AdminService
	helpText            string
	adminUsernames      map[string]struct{}
	setupCleanupMu      sync.Mutex
	setupCleanup        map[int64]setupCleanupInfo
}

type setupCleanupInfo struct {
	chatID         int64
	startMessageID int
}

func New(sessions SessionRepository, drafts ProfileDraftRepository, deleteConfirmations ProfileDeletionConfirmationRepository, adminActions AdminActionRepository, profiles ProfileService, search SearchService, admin AdminService, adminUsernames []string) Usecase {
	return &service{
		sessions:            sessions,
		drafts:              drafts,
		deleteConfirmations: deleteConfirmations,
		adminActions:        adminActions,
		profiles:            profiles,
		search:              search,
		admin:               admin,
		helpText:            defaultHelpText,
		adminUsernames:      normalizeAdminUsernames(adminUsernames),
		setupCleanup:        make(map[int64]setupCleanupInfo),
	}
}

func (s *service) Handle(ctx context.Context, msg domain.IncomingMessage) ([]domain.OutgoingMessage, error) {
	var touchErr error
	if s != nil && s.sessions != nil {
		touchErr = s.sessions.Touch(ctx, domain.Session{
			UserID:   msg.User.ID,
			ChatID:   msg.ChatID,
			Username: msg.User.Username,
			LastSeen: s.now(msg.ReceivedAt),
		})
	}

	if msg.Command != "" &&
		msg.Command != domain.CommandAdminBan &&
		msg.Command != domain.CommandAdminUnban &&
		msg.Command != domain.CommandAdminModerator &&
		msg.Command != domain.CommandAdminUnmoderator {
		_ = s.clearAdminAction(ctx, msg.User.ID)
	}

	if msg.Command == "" {
		updatedMsg, response, handled, err := s.applyAdminAction(ctx, msg)
		if handled {
			if response == nil {
				return nil, errors.Join(touchErr, err)
			}
			return []domain.OutgoingMessage{*response}, errors.Join(touchErr, err)
		}
		msg = updatedMsg
	}

	if s.isSearchCommand(msg.Command) {
		messages, replyErr := s.handleSearch(ctx, msg)
		likesFollowUps, likesErr := s.pendingLikesFollowUp(ctx, msg)
		if len(likesFollowUps) > 0 {
			messages = append(messages, likesFollowUps...)
		}
		return messages, errors.Join(touchErr, replyErr, likesErr)
	}

	response := domain.OutgoingMessage{ChatID: msg.ChatID}
	var replyErr error
	switch msg.Command {
	case domain.CommandStart:
		response.Text, response.InlineKeyboard, replyErr = s.startMessage(ctx, msg)
	case domain.CommandCancel:
		response.Text, response.InlineKeyboard, replyErr = s.cancelMessage(ctx, msg)
	case domain.CommandAdminMenu:
		response.Text, response.InlineKeyboard, replyErr = s.adminMenuMessage(ctx, msg)
	case domain.CommandAdminUsers:
		response.Text, response.InlineKeyboard, replyErr = s.adminUsersMessage(ctx, msg)
	case domain.CommandAdminBannedUsers:
		response.Text, response.InlineKeyboard, replyErr = s.adminBannedUsersMessage(ctx, msg)
	case domain.CommandAdminModerators:
		response.Text, response.InlineKeyboard, replyErr = s.adminModeratorsMessage(ctx, msg)
	case domain.CommandAdminReports:
		response.Text, response.InlineKeyboard, replyErr = s.adminReportsMessage(ctx, msg)
	case domain.CommandAdminBan:
		response.Text, response.InlineKeyboard, replyErr = s.adminBanMessage(ctx, msg)
	case domain.CommandAdminUnban:
		response.Text, response.InlineKeyboard, replyErr = s.adminUnbanMessage(ctx, msg)
	case domain.CommandAdminModerator:
		response.Text, response.InlineKeyboard, replyErr = s.adminModeratorMessage(ctx, msg)
	case domain.CommandAdminUnmoderator:
		response.Text, response.InlineKeyboard, replyErr = s.adminUnmoderatorMessage(ctx, msg)
	default:
		response.Text, replyErr = s.reply(ctx, msg)
		if msg.Command == domain.CommandProfileEdit && replyErr == nil {
			response.InlineKeyboard = s.profileEditMenuKeyboard()
		}
		if msg.Command == domain.CommandProfileDelete && replyErr == nil {
			response.InlineKeyboard = s.profileDeleteConfirmInlineKeyboard()
		}
		if msg.Command == domain.CommandProfileDeleteConfirm && replyErr == nil {
			if response.Text != profileDeleteExpiredText {
				response.InlineKeyboard = s.profileCreateInlineKeyboard()
			}
		}
	}

	if replyErr == nil {
		s.attachBotCheckKeyboard(ctx, msg, &response)
		s.attachTelegramNameKeyboard(ctx, msg, &response)
		s.attachGenderKeyboard(ctx, msg, &response)
		s.attachCountryKeyboard(ctx, msg, &response)
		s.attachCityKeyboard(ctx, msg, &response)
		s.attachEmojiKeyboard(ctx, msg, &response)
		s.attachPhotosDoneKeyboard(ctx, msg, &response)
		s.attachCancelKeyboard(ctx, msg, &response)
	}

	if replyErr == nil && response.Text == profileCreatedText {
		response.CleanupFromMessageID = s.takeProfileSetupCleanupStart(msg.User.ID, msg.ChatID)
	}

	messages := []domain.OutgoingMessage{response}
	if replyErr == nil && s != nil && s.profiles != nil && msg.User.ID != 0 {
		switch msg.Command {
		case domain.CommandProfileView:
			profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
			if err != nil {
				if isBannedError(err) {
					response.Text = s.userBannedText()
					response.InlineKeyboard = nil
					messages[0] = response
				}
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
				if isBannedError(err) {
					response.Text = s.userBannedText()
					response.InlineKeyboard = nil
					messages[0] = response
				}
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
			profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
			if err != nil {
				if isBannedError(err) {
					response.Text = s.userBannedText()
					response.InlineKeyboard = nil
					messages[0] = response
				}
				replyErr = errors.Join(replyErr, err)
				break
			}
			if found {
				response.Text = s.profileSettingsTextWithVisibility(profile.IsHidden)
				response.InlineKeyboard = s.profileSettingsInlineKeyboard(profile.IsHidden)
			} else {
				response.Text = "Profile not found. Use the Create Profile button to create it."
				response.InlineKeyboard = s.profileCreateInlineKeyboard()
			}
			messages[0] = response
		case domain.CommandProfileVisibility:
			profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
			if err != nil {
				if isBannedError(err) {
					response.Text = s.userBannedText()
					response.InlineKeyboard = nil
					messages[0] = response
				}
				replyErr = errors.Join(replyErr, err)
				break
			}
			if found {
				response.InlineKeyboard = s.profileSettingsInlineKeyboard(profile.IsHidden)
			} else {
				response.Text = "Profile not found. Use the Create Profile button to create it."
				response.InlineKeyboard = s.profileCreateInlineKeyboard()
			}
			messages[0] = response
		case domain.CommandProfileDeleteCancel:
			profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
			if err != nil {
				if isBannedError(err) {
					response.Text = s.userBannedText()
					response.InlineKeyboard = nil
					messages[0] = response
				}
				replyErr = errors.Join(replyErr, err)
				break
			}
			if found {
				response.InlineKeyboard = s.profileSettingsInlineKeyboard(profile.IsHidden)
			} else {
				response.Text = "Profile not found. Use the Create Profile button to create it."
				response.InlineKeyboard = s.profileCreateInlineKeyboard()
			}
			messages[0] = response
		case domain.CommandProfileDeleteConfirm:
			if response.Text != profileDeleteExpiredText {
				break
			}
			profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
			if err != nil {
				if isBannedError(err) {
					response.Text = s.userBannedText()
					response.InlineKeyboard = nil
					messages[0] = response
				}
				replyErr = errors.Join(replyErr, err)
				break
			}
			if found {
				response.InlineKeyboard = s.profileSettingsInlineKeyboard(profile.IsHidden)
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

	likesFollowUps, likesErr := s.pendingLikesFollowUp(ctx, msg)
	if len(likesFollowUps) > 0 {
		messages = append(messages, likesFollowUps...)
	}

	return messages, errors.Join(touchErr, replyErr, followUpErr, likesErr)
}

func (s *service) attachBotCheckKeyboard(ctx context.Context, msg domain.IncomingMessage, response *domain.OutgoingMessage) {
	if response == nil || response.InlineKeyboard != nil || s == nil || s.drafts == nil || msg.User.ID == 0 {
		return
	}

	if !s.isDraftCommand(msg.Command) {
		return
	}

	draft, found, err := s.drafts.Get(ctx, msg.User.ID)
	if err != nil || !found {
		return
	}

	if draft.Step != domain.ProfileDraftStepBotCheck {
		return
	}

	if s.draftMode(draft) != domain.ProfileDraftModeCreate {
		return
	}

	keyboard := s.botCheckInlineKeyboard(draft.BotCheckAnswer)
	if keyboard == nil {
		return
	}

	response.InlineKeyboard = keyboard
}

func (s *service) attachGenderKeyboard(ctx context.Context, msg domain.IncomingMessage, response *domain.OutgoingMessage) {
	if response == nil || response.InlineKeyboard != nil || s == nil || s.drafts == nil || msg.User.ID == 0 {
		return
	}

	if !s.isDraftCommand(msg.Command) {
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

	if !s.isDraftCommand(msg.Command) {
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

	if !s.isDraftCommand(msg.Command) {
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

	if !s.isDraftCommand(msg.Command) {
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

	if !s.isDraftCommand(msg.Command) {
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

	if !s.isDraftCommand(msg.Command) {
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

func (s *service) attachCancelKeyboard(ctx context.Context, msg domain.IncomingMessage, response *domain.OutgoingMessage) {
	if response == nil || s == nil || s.drafts == nil || msg.User.ID == 0 {
		return
	}

	if !s.isDraftCommand(msg.Command) {
		return
	}

	draft, found, err := s.drafts.Get(ctx, msg.User.ID)
	if err != nil || !found {
		return
	}

	if s.draftMode(draft) == domain.ProfileDraftModeCreate && s.previousProfileSetupStep(draft.Step) != "" {
		response.InlineKeyboard = withProfileSetupBackInlineKeyboard(response.InlineKeyboard)
	}

	response.InlineKeyboard = withDraftCancelInlineKeyboard(response.InlineKeyboard, s.draftMode(draft))
}

func (s *service) registerProfileSetupCleanup(draft domain.ProfileDraft) {
	if s == nil {
		return
	}
	if draft.UserID == 0 || draft.ChatID == 0 || draft.SetupStartMessageID == 0 {
		return
	}

	s.setupCleanupMu.Lock()
	if s.setupCleanup == nil {
		s.setupCleanup = make(map[int64]setupCleanupInfo)
	}
	s.setupCleanup[draft.UserID] = setupCleanupInfo{
		chatID:         draft.ChatID,
		startMessageID: draft.SetupStartMessageID,
	}
	s.setupCleanupMu.Unlock()
}

func (s *service) takeProfileSetupCleanupStart(userID, chatID int64) int {
	if s == nil || userID == 0 {
		return 0
	}

	s.setupCleanupMu.Lock()
	defer s.setupCleanupMu.Unlock()

	info, ok := s.setupCleanup[userID]
	if !ok {
		return 0
	}
	delete(s.setupCleanup, userID)
	if info.chatID != 0 && chatID != 0 && info.chatID != chatID {
		return 0
	}

	return info.startMessageID
}
