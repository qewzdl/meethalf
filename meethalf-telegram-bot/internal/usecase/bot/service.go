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

type AgeConfirmationRepository interface {
	Get(ctx context.Context, userID int64) (domain.AgeConfirmation, bool, error)
	Save(ctx context.Context, confirmation domain.AgeConfirmation) error
	Delete(ctx context.Context, userID int64) error
	FindUserIDByUsername(ctx context.Context, username string) (int64, bool, error)
}

type ProfileService interface {
	CreateProfile(ctx context.Context, profile domain.Profile) error
	GetProfile(ctx context.Context, userID int64) (domain.Profile, bool, error)
	DeleteProfile(ctx context.Context, userID int64) (bool, error)
	SetProfileVisibility(ctx context.Context, userID int64, isHidden bool) (bool, error)
	SetProfileLikeNotifications(ctx context.Context, userID int64, enabled bool) (bool, error)
}

type SearchService interface {
	StartSearch(ctx context.Context, userID int64, gender domain.Gender, accuracy int) (domain.MatchCandidate, bool, error)
	SearchWithAI(ctx context.Context, userID int64, message string) (domain.MatchCandidate, bool, error)
	AIAvailable(ctx context.Context, userID int64) (bool, error)
	NextCandidate(ctx context.Context, userID int64) (domain.MatchCandidate, bool, error)
	PreviousCandidate(ctx context.Context, userID int64) (domain.MatchCandidate, bool, error)
	History(ctx context.Context, userID int64, limit, offset int) (domain.MatchHistoryList, error)
	ReceivedLikes(ctx context.Context, userID int64, limit, offset int) (domain.MatchLikesList, error)
	RecordAction(ctx context.Context, userID, targetID int64, action domain.MatchAction) (domain.MatchActionResult, error)
	PendingLikes(ctx context.Context, userID int64) ([]domain.Profile, error)
}

type AdminService interface {
	ListUsers(ctx context.Context, limit, offset int, onlyBanned, onlyModerators, onlyHidden, onlyShadowBanned bool) (domain.UserList, error)
	ListReportedUsers(ctx context.Context, limit, offset int) (domain.ReportedUserList, error)
	GetUser(ctx context.Context, userID int64) (domain.UserSummary, error)
	GetUserByUsername(ctx context.Context, username string) (domain.UserSummary, error)
	BanUser(ctx context.Context, userID int64) error
	BanUserByUsername(ctx context.Context, username string) error
	UnbanUser(ctx context.Context, userID int64) error
	UnbanUserByUsername(ctx context.Context, username string) error
	ShadowBanUser(ctx context.Context, userID int64) error
	ShadowBanUserByUsername(ctx context.Context, username string) error
	UnshadowBanUser(ctx context.Context, userID int64) error
	UnshadowBanUserByUsername(ctx context.Context, username string) error
	HideUser(ctx context.Context, userID int64) error
	HideUserByUsername(ctx context.Context, username string) error
	ShowUser(ctx context.Context, userID int64) error
	ShowUserByUsername(ctx context.Context, username string) error
	DeleteProfile(ctx context.Context, userID int64) error
	DeleteProfileByUsername(ctx context.Context, username string) error
	MakeModerator(ctx context.Context, userID int64) error
	MakeModeratorByUsername(ctx context.Context, username string) error
	RemoveModerator(ctx context.Context, userID int64) error
	RemoveModeratorByUsername(ctx context.Context, username string) error
	ResetUserChoices(ctx context.Context, userID int64) error
	ResetUserChoicesByUsername(ctx context.Context, username string) error
	ClearUserReports(ctx context.Context, userID int64) error
	ClearUserReportsByUsername(ctx context.Context, username string) error
	CreateAd(ctx context.Context, text, photoID string, buttons []domain.AdButton) (domain.Advertisement, error)
}

type BroadcastSender interface {
	Send(ctx context.Context, msg domain.OutgoingMessage) (int, error)
}

type service struct {
	sessions            SessionRepository
	drafts              ProfileDraftRepository
	deleteConfirmations ProfileDeletionConfirmationRepository
	adminActions        AdminActionRepository
	ageConfirmations    AgeConfirmationRepository
	profiles            ProfileService
	search              SearchService
	admin               AdminService
	broadcastSender     BroadcastSender
	adminUsernames      map[string]struct{}
	setupCleanupMu      sync.Mutex
	setupCleanup        map[int64]setupCleanupInfo
	adminCleanupMu      sync.Mutex
	adminCleanup        map[int64]adminCleanupInfo
	adBroadcastQueue    chan adBroadcastJob
}

type setupCleanupInfo struct {
	chatID         int64
	startMessageID int
}

type adminCleanupInfo struct {
	chatID     int64
	messageIDs []int
}

func New(sessions SessionRepository, drafts ProfileDraftRepository, deleteConfirmations ProfileDeletionConfirmationRepository, adminActions AdminActionRepository, ageConfirmations AgeConfirmationRepository, profiles ProfileService, search SearchService, admin AdminService, broadcastSender BroadcastSender, adminUsernames []string) Usecase {
	svc := &service{
		sessions:            sessions,
		drafts:              drafts,
		deleteConfirmations: deleteConfirmations,
		adminActions:        adminActions,
		ageConfirmations:    ageConfirmations,
		profiles:            profiles,
		search:              search,
		admin:               admin,
		broadcastSender:     broadcastSender,
		adminUsernames:      normalizeAdminUsernames(adminUsernames),
		setupCleanup:        make(map[int64]setupCleanupInfo),
	}
	svc.startAdBroadcastWorker()
	return svc
}

func (s *service) localizerForMessage(ctx context.Context, msg domain.IncomingMessage) localizer {
	lang := msg.Language
	if lang == "" {
		lang = s.resolveLanguage(ctx, msg)
	}
	return newLocalizer(lang)
}

func (s *service) localizerForUser(ctx context.Context, userID int64, fallback domain.Language) localizer {
	return newLocalizer(s.resolveLanguageForUser(ctx, userID, fallback))
}

func (s *service) Handle(ctx context.Context, msg domain.IncomingMessage) ([]domain.OutgoingMessage, error) {
	if msg.Language == "" {
		msg.Language = s.resolveLanguage(ctx, msg)
	} else {
		msg.Language = normalizeLanguageValue(msg.Language)
	}
	l := newLocalizer(msg.Language)

	var touchErr error
	searchAccuracyEnabled := false
	pendingAISearch := false
	if s != nil && s.sessions != nil && msg.User.ID != 0 {
		if session, found, err := s.sessions.Get(ctx, msg.User.ID); err == nil && found {
			searchAccuracyEnabled = session.SearchAccuracyEnabled
			pendingAISearch = session.PendingAISearch
		}
	}
	if s != nil && s.sessions != nil {
		touchErr = s.sessions.Touch(ctx, domain.Session{
			UserID:                msg.User.ID,
			ChatID:                msg.ChatID,
			Username:              msg.User.Username,
			Language:              msg.Language,
			LastSeen:              s.now(msg.ReceivedAt),
			SearchAccuracyEnabled: searchAccuracyEnabled,
			PendingAISearch:       pendingAISearch,
		})
	}

	if msg.Command != "" &&
		msg.Command != domain.CommandAdminBan &&
		msg.Command != domain.CommandAdminUnban &&
		msg.Command != domain.CommandAdminShadowBan &&
		msg.Command != domain.CommandAdminShadowUnban &&
		msg.Command != domain.CommandAdminHideProfile &&
		msg.Command != domain.CommandAdminShowProfile &&
		msg.Command != domain.CommandAdminDeleteProfile &&
		msg.Command != domain.CommandAdminModerator &&
		msg.Command != domain.CommandAdminUnmoderator &&
		msg.Command != domain.CommandAdminResetChoices &&
		msg.Command != domain.CommandAdminResetStart &&
		msg.Command != domain.CommandAdminClearReports &&
		msg.Command != domain.CommandAdminPostAd &&
		msg.Command != domain.CommandAdminAdPreview &&
		msg.Command != domain.CommandAdminAdSend &&
		msg.Command != domain.CommandAdminAdAddButton &&
		msg.Command != domain.CommandAdminAdClearButtons &&
		msg.Command != domain.CommandAdminAdRemovePhoto {
		_ = s.clearAdminAction(ctx, msg.User.ID)
	}
	if msg.Command != "" && msg.Command != domain.CommandSearchAI {
		_ = s.setSessionAISearchPending(ctx, msg, false)
	}

	if accessResponse, handled, accessErr := s.enforceAgeAccess(ctx, msg, l); handled {
		return []domain.OutgoingMessage{*accessResponse}, errors.Join(touchErr, accessErr)
	}

	if msg.Command == "" {
		updatedMsg, response, handled, err := s.applyAdminAction(ctx, msg, l)
		if handled {
			if response == nil {
				return nil, errors.Join(touchErr, err)
			}
			return []domain.OutgoingMessage{*response}, errors.Join(touchErr, err)
		}
		msg = updatedMsg
		if msg.Command == "" && s.sessionAISearchPending(ctx, msg.User.ID) && !s.hasActiveDraft(ctx, msg.User.ID) {
			msg.Command = domain.CommandSearchAI
		}
		if msg.Language == "" {
			msg.Language = l.lang
		}
		l = newLocalizer(msg.Language)
	}

	if s.isSearchCommand(msg.Command) {
		messages, replyErr := s.handleSearch(ctx, msg, l)
		likesFollowUps, likesErr := s.pendingLikesFollowUp(ctx, msg, l)
		if len(likesFollowUps) > 0 {
			messages = append(messages, likesFollowUps...)
		}
		return messages, errors.Join(touchErr, replyErr, likesErr)
	}

	response := domain.OutgoingMessage{ChatID: msg.ChatID}
	var replyErr error
	switch msg.Command {
	case domain.CommandStart:
		response.Text, response.InlineKeyboard, replyErr = s.startMessage(ctx, msg, l)
	case domain.CommandCancel:
		response.Text, response.InlineKeyboard, replyErr = s.cancelMessage(ctx, msg, l)
	case domain.CommandAdminMenu:
		response.Text, response.InlineKeyboard, replyErr = s.adminMenuMessage(ctx, msg, l)
	case domain.CommandAdminUsers:
		response.Text, response.InlineKeyboard, replyErr = s.adminUsersMessage(ctx, msg, l)
	case domain.CommandAdminBannedUsers:
		response.Text, response.InlineKeyboard, replyErr = s.adminBannedUsersMessage(ctx, msg, l)
	case domain.CommandAdminShadowBannedUsers:
		response.Text, response.InlineKeyboard, replyErr = s.adminShadowBannedUsersMessage(ctx, msg, l)
	case domain.CommandAdminHiddenUsers:
		response.Text, response.InlineKeyboard, replyErr = s.adminHiddenUsersMessage(ctx, msg, l)
	case domain.CommandAdminModerators:
		response.Text, response.InlineKeyboard, replyErr = s.adminModeratorsMessage(ctx, msg, l)
	case domain.CommandAdminReports:
		response.Text, response.InlineKeyboard, replyErr = s.adminReportsMessage(ctx, msg, l)
	case domain.CommandAdminBan:
		response.Text, response.InlineKeyboard, replyErr = s.adminBanMessage(ctx, msg, l)
	case domain.CommandAdminUnban:
		response.Text, response.InlineKeyboard, replyErr = s.adminUnbanMessage(ctx, msg, l)
	case domain.CommandAdminShadowBan:
		response.Text, response.InlineKeyboard, replyErr = s.adminShadowBanMessage(ctx, msg, l)
	case domain.CommandAdminShadowUnban:
		response.Text, response.InlineKeyboard, replyErr = s.adminShadowUnbanMessage(ctx, msg, l)
	case domain.CommandAdminHideProfile:
		response.Text, response.InlineKeyboard, replyErr = s.adminHideProfileMessage(ctx, msg, l)
	case domain.CommandAdminShowProfile:
		response.Text, response.InlineKeyboard, replyErr = s.adminShowProfileMessage(ctx, msg, l)
	case domain.CommandAdminDeleteProfile:
		response.Text, response.InlineKeyboard, replyErr = s.adminDeleteProfileMessage(ctx, msg, l)
	case domain.CommandAdminModerator:
		response.Text, response.InlineKeyboard, replyErr = s.adminModeratorMessage(ctx, msg, l)
	case domain.CommandAdminUnmoderator:
		response.Text, response.InlineKeyboard, replyErr = s.adminUnmoderatorMessage(ctx, msg, l)
	case domain.CommandAdminResetChoices:
		response.Text, response.InlineKeyboard, replyErr = s.adminResetChoicesMessage(ctx, msg, l)
	case domain.CommandAdminResetStart:
		response.Text, response.InlineKeyboard, replyErr = s.adminResetStartMessage(ctx, msg, l)
	case domain.CommandAdminClearReports:
		response.Text, response.InlineKeyboard, replyErr = s.adminClearReportsMessage(ctx, msg, l)
	case domain.CommandAdminPostAd:
		response.Text, response.InlineKeyboard, replyErr = s.adminAdMessage(ctx, msg, l)
	case domain.CommandAdminAdPreview:
		response.Text, response.InlineKeyboard, replyErr = s.adminAdPreviewMessage(ctx, msg, l)
	case domain.CommandAdminAdSend:
		response.Text, response.InlineKeyboard, replyErr = s.adminAdSendMessage(ctx, msg, l)
	case domain.CommandAdminAdAddButton:
		response.Text, response.InlineKeyboard, replyErr = s.adminAdAddButtonMessage(ctx, msg, l)
	case domain.CommandAdminAdClearButtons:
		response.Text, response.InlineKeyboard, replyErr = s.adminAdClearButtonsMessage(ctx, msg, l)
	case domain.CommandAdminAdRemovePhoto:
		response.Text, response.InlineKeyboard, replyErr = s.adminAdRemovePhotoMessage(ctx, msg, l)
	case domain.CommandLanguageOnboarding:
		response.Text, response.InlineKeyboard, replyErr = s.languageOnboardingSelectionMessage(ctx, msg, l)
	case domain.CommandProfileLanguage, domain.CommandLanguage:
		response.Text, response.InlineKeyboard, replyErr = s.profileLanguageMessage(ctx, msg, l)
	default:
		response.Text, replyErr = s.reply(ctx, msg, l)
		if msg.Command == domain.CommandProfileEdit && replyErr == nil {
			response.InlineKeyboard = s.profileEditMenuKeyboard(l)
		}
		if msg.Command == domain.CommandProfileDelete && replyErr == nil {
			response.InlineKeyboard = s.profileDeleteConfirmInlineKeyboard(l)
		}
		if msg.Command == domain.CommandProfileDeleteConfirm && replyErr == nil {
			if response.Text != s.profileDeleteExpiredText(l) {
				response.InlineKeyboard = s.profileCreateInlineKeyboard(l)
			}
		}
	}

	if replyErr == nil {
		s.attachBotCheckKeyboard(ctx, msg, l, &response)
		s.attachTelegramNameKeyboard(ctx, msg, l, &response)
		s.attachGenderKeyboard(ctx, msg, l, &response)
		s.attachCountryKeyboard(ctx, msg, l, &response)
		s.attachCityKeyboard(ctx, msg, l, &response)
		s.attachEmojiKeyboard(ctx, msg, l, &response)
		s.attachPhotosDoneKeyboard(ctx, msg, l, &response)
		s.attachCancelKeyboard(ctx, msg, l, &response)
	}

	if replyErr == nil && response.Text == s.profileCreated(l) {
		response.CleanupFromMessageID = s.takeProfileSetupCleanupStart(msg.User.ID, msg.ChatID)
	}

	messages := []domain.OutgoingMessage{response}
	if replyErr == nil && s != nil && s.profiles != nil && msg.User.ID != 0 {
		switch msg.Command {
		case domain.CommandProfileView:
			profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
			if err != nil {
				if isBannedError(err) {
					response.Text = s.userBannedText(l)
					response.InlineKeyboard = nil
					messages[0] = response
				}
				replyErr = errors.Join(replyErr, err)
				break
			}
			if found {
				response.InlineKeyboard = s.profileViewInlineKeyboard(l)
				if len(profile.Photos) > 0 {
					messages = s.profileAlbumMessages(msg.ChatID, profile, s.profileViewInlineKeyboard(l), l)
				} else {
					messages[0] = response
				}
			} else {
				response.InlineKeyboard = s.profileCreateInlineKeyboard(l)
				messages[0] = response
			}
		case domain.CommandProfilePreview:
			profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
			if err != nil {
				if isBannedError(err) {
					response.Text = s.userBannedText(l)
					response.InlineKeyboard = nil
					messages[0] = response
				}
				replyErr = errors.Join(replyErr, err)
				break
			}
			if found {
				response.InlineKeyboard = s.profilePreviewInlineKeyboard(l)
				if len(profile.Photos) > 0 {
					messages = s.profilePreviewAlbumMessages(msg.ChatID, profile, s.profilePreviewInlineKeyboard(l), l)
				} else {
					messages[0] = response
				}
			} else {
				response.InlineKeyboard = s.profileCreateInlineKeyboard(l)
				messages[0] = response
			}
		case domain.CommandProfileSettings:
			profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
			if err != nil {
				if isBannedError(err) {
					response.Text = s.userBannedText(l)
					response.InlineKeyboard = nil
					messages[0] = response
				}
				replyErr = errors.Join(replyErr, err)
				break
			}
			if found {
				searchAccuracyEnabled := s.sessionSearchAccuracyEnabled(ctx, msg.User.ID)
				response.Text = s.profileSettingsTextWithVisibility(l, profile.IsHidden, searchAccuracyEnabled, profile.LikesNotificationsEnabled)
				response.InlineKeyboard = s.profileSettingsInlineKeyboard(l, profile.IsHidden, searchAccuracyEnabled, profile.LikesNotificationsEnabled)
			} else {
				response.Text = s.profileSettingsText(l)
				response.InlineKeyboard = s.profileSettingsGuestInlineKeyboard(l)
			}
			messages[0] = response
		case domain.CommandProfileVisibility:
			profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
			if err != nil {
				if isBannedError(err) {
					response.Text = s.userBannedText(l)
					response.InlineKeyboard = nil
					messages[0] = response
				}
				replyErr = errors.Join(replyErr, err)
				break
			}
			if found {
				searchAccuracyEnabled := s.sessionSearchAccuracyEnabled(ctx, msg.User.ID)
				response.InlineKeyboard = s.profileSettingsInlineKeyboard(l, profile.IsHidden, searchAccuracyEnabled, profile.LikesNotificationsEnabled)
			} else {
				response.Text = l.message(msgProfileNotFoundCreateButton)
				response.InlineKeyboard = s.profileCreateInlineKeyboard(l)
			}
			messages[0] = response
		case domain.CommandProfileSearchAccuracy:
			profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
			if err != nil {
				if isBannedError(err) {
					response.Text = s.userBannedText(l)
					response.InlineKeyboard = nil
					messages[0] = response
				}
				replyErr = errors.Join(replyErr, err)
				break
			}
			if found {
				searchAccuracyEnabled := s.sessionSearchAccuracyEnabled(ctx, msg.User.ID)
				response.InlineKeyboard = s.profileSettingsInlineKeyboard(l, profile.IsHidden, searchAccuracyEnabled, profile.LikesNotificationsEnabled)
			} else {
				response.Text = l.message(msgProfileNotFoundCreateButton)
				response.InlineKeyboard = s.profileCreateInlineKeyboard(l)
			}
			messages[0] = response
		case domain.CommandProfileLikeNotifications:
			profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
			if err != nil {
				if isBannedError(err) {
					response.Text = s.userBannedText(l)
					response.InlineKeyboard = nil
					messages[0] = response
				}
				replyErr = errors.Join(replyErr, err)
				break
			}
			if found {
				searchAccuracyEnabled := s.sessionSearchAccuracyEnabled(ctx, msg.User.ID)
				response.InlineKeyboard = s.profileSettingsInlineKeyboard(l, profile.IsHidden, searchAccuracyEnabled, profile.LikesNotificationsEnabled)
			} else {
				response.Text = l.message(msgProfileNotFoundCreateButton)
				response.InlineKeyboard = s.profileCreateInlineKeyboard(l)
			}
			messages[0] = response
		case domain.CommandProfileDeleteCancel:
			profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
			if err != nil {
				if isBannedError(err) {
					response.Text = s.userBannedText(l)
					response.InlineKeyboard = nil
					messages[0] = response
				}
				replyErr = errors.Join(replyErr, err)
				break
			}
			if found {
				searchAccuracyEnabled := s.sessionSearchAccuracyEnabled(ctx, msg.User.ID)
				response.InlineKeyboard = s.profileSettingsInlineKeyboard(l, profile.IsHidden, searchAccuracyEnabled, profile.LikesNotificationsEnabled)
			} else {
				response.Text = l.message(msgProfileNotFoundCreateButton)
				response.InlineKeyboard = s.profileCreateInlineKeyboard(l)
			}
			messages[0] = response
		case domain.CommandProfileDeleteConfirm:
			if response.Text != s.profileDeleteExpiredText(l) {
				break
			}
			profile, found, err := s.profiles.GetProfile(ctx, msg.User.ID)
			if err != nil {
				if isBannedError(err) {
					response.Text = s.userBannedText(l)
					response.InlineKeyboard = nil
					messages[0] = response
				}
				replyErr = errors.Join(replyErr, err)
				break
			}
			if found {
				searchAccuracyEnabled := s.sessionSearchAccuracyEnabled(ctx, msg.User.ID)
				response.InlineKeyboard = s.profileSettingsInlineKeyboard(l, profile.IsHidden, searchAccuracyEnabled, profile.LikesNotificationsEnabled)
			} else {
				response.Text = l.message(msgProfileNotFoundCreateButton)
				response.InlineKeyboard = s.profileCreateInlineKeyboard(l)
			}
			messages[0] = response
		}
	}

	followUps, followUpErr := s.profileDetailsFollowUp(ctx, msg, l, response.Text, replyErr)
	if len(followUps) > 0 {
		messages = append(messages, followUps...)
	}

	likesFollowUps, likesErr := s.pendingLikesFollowUp(ctx, msg, l)
	if len(likesFollowUps) > 0 {
		messages = append(messages, likesFollowUps...)
	}

	if cleanup := s.takeAdminActionCleanup(msg.User.ID, msg.ChatID); len(cleanup) > 0 && len(messages) > 0 {
		messages[0].DeleteMessageIDs = append(messages[0].DeleteMessageIDs, cleanup...)
	}

	return messages, errors.Join(touchErr, replyErr, followUpErr, likesErr)
}

func (s *service) attachBotCheckKeyboard(ctx context.Context, msg domain.IncomingMessage, l localizer, response *domain.OutgoingMessage) {
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

	keyboard := s.botCheckInlineKeyboard(l, draft.BotCheckAnswer)
	if keyboard == nil {
		return
	}

	response.InlineKeyboard = keyboard
}

func (s *service) attachGenderKeyboard(ctx context.Context, msg domain.IncomingMessage, l localizer, response *domain.OutgoingMessage) {
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

	response.InlineKeyboard = s.genderInlineKeyboard(l)
}

func (s *service) attachTelegramNameKeyboard(ctx context.Context, msg domain.IncomingMessage, l localizer, response *domain.OutgoingMessage) {
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

	response.InlineKeyboard = s.telegramNameInlineKeyboard(l)
}

func (s *service) attachCountryKeyboard(ctx context.Context, msg domain.IncomingMessage, l localizer, response *domain.OutgoingMessage) {
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

	response.InlineKeyboard = s.countryInlineKeyboard(l)
}

func (s *service) attachCityKeyboard(ctx context.Context, msg domain.IncomingMessage, l localizer, response *domain.OutgoingMessage) {
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

	keyboard := s.cityInlineKeyboard(l, draft.Country)
	if keyboard == nil {
		return
	}

	response.InlineKeyboard = keyboard
}

func (s *service) attachEmojiKeyboard(ctx context.Context, msg domain.IncomingMessage, l localizer, response *domain.OutgoingMessage) {
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

	response.InlineKeyboard = s.emojiInlineKeyboard(l)
}

func (s *service) attachPhotosDoneKeyboard(ctx context.Context, msg domain.IncomingMessage, l localizer, response *domain.OutgoingMessage) {
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

	response.InlineKeyboard = s.photosDoneInlineKeyboard(l)
}

func (s *service) attachCancelKeyboard(ctx context.Context, msg domain.IncomingMessage, l localizer, response *domain.OutgoingMessage) {
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
		response.InlineKeyboard = withProfileSetupBackInlineKeyboard(l, response.InlineKeyboard)
	}

	response.InlineKeyboard = withDraftCancelInlineKeyboard(l, response.InlineKeyboard, s.draftMode(draft))
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

func (s *service) registerAdminActionCleanup(msg domain.IncomingMessage) {
	if s == nil || msg.User.ID == 0 || msg.ChatID == 0 {
		return
	}

	messageIDs := adminActionCleanupIDs(msg)
	if len(messageIDs) == 0 {
		return
	}

	s.adminCleanupMu.Lock()
	if s.adminCleanup == nil {
		s.adminCleanup = make(map[int64]adminCleanupInfo)
	}
	ids := make([]int, len(messageIDs))
	copy(ids, messageIDs)
	s.adminCleanup[msg.User.ID] = adminCleanupInfo{
		chatID:     msg.ChatID,
		messageIDs: ids,
	}
	s.adminCleanupMu.Unlock()
}

func (s *service) takeAdminActionCleanup(userID, chatID int64) []int {
	if s == nil || userID == 0 {
		return nil
	}

	s.adminCleanupMu.Lock()
	info, ok := s.adminCleanup[userID]
	if ok {
		delete(s.adminCleanup, userID)
	}
	s.adminCleanupMu.Unlock()
	if !ok {
		return nil
	}
	if info.chatID != 0 && chatID != 0 && info.chatID != chatID {
		return nil
	}

	return info.messageIDs
}
