package bot

import (
	"fmt"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

type messageKey string
type buttonKey string
type labelKey string

type localizer struct {
	lang domain.Language
}

func newLocalizer(lang domain.Language) localizer {
	return localizer{lang: normalizeLanguageValue(lang)}
}

func (l localizer) message(key messageKey, args ...any) string {
	value := lookupMessage(l.lang, key)
	if value == "" {
		value = lookupMessage(domain.LanguageEnglish, key)
	}
	return formatTemplate(value, args...)
}

func (l localizer) button(key buttonKey, args ...any) string {
	value := lookupButton(l.lang, key)
	if value == "" {
		value = lookupButton(domain.LanguageEnglish, key)
	}
	text := formatTemplate(value, args...)
	return decorateButtonText(key, text)
}

func decorateButtonText(key buttonKey, text string) string {
	emoji := buttonEmoji(key)
	if emoji == "" {
		return text
	}
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return text
	}
	if strings.HasPrefix(trimmed, emoji+" ") || trimmed == emoji {
		return text
	}
	return emoji + " " + text
}

func (l localizer) label(key labelKey, args ...any) string {
	value := lookupLabel(l.lang, key)
	if value == "" {
		value = lookupLabel(domain.LanguageEnglish, key)
	}
	return formatTemplate(value, args...)
}

func (l localizer) genderLabel(gender domain.Gender) string {
	switch gender {
	case domain.GenderMale:
		return l.label(labelGenderMale)
	case domain.GenderFemale:
		return l.label(labelGenderFemale)
	case domain.GenderOther:
		return l.label(labelGenderOther)
	default:
		return l.label(labelNotSet)
	}
}

func (l localizer) countryLabel(country domain.Country) string {
	switch country {
	case domain.CountryRussia:
		return l.label(labelCountryRussia)
	case domain.CountryKazakhstan:
		return l.label(labelCountryKazakhstan)
	case domain.CountryBelarus:
		return l.label(labelCountryBelarus)
	case domain.CountryUkraine:
		return l.label(labelCountryUkraine)
	default:
		return l.label(labelNotSet)
	}
}

func (l localizer) stepLabel(step domain.ProfileDraftStep) string {
	switch step {
	case domain.ProfileDraftStepBotCheck:
		return l.label(labelStepVerification)
	case domain.ProfileDraftStepName:
		return l.label(labelStepName)
	case domain.ProfileDraftStepGender:
		return l.label(labelStepGender)
	case domain.ProfileDraftStepBirthDate:
		return l.label(labelStepBirthDate)
	case domain.ProfileDraftStepCountry:
		return l.label(labelStepCountry)
	case domain.ProfileDraftStepCity:
		return l.label(labelStepCity)
	case domain.ProfileDraftStepDescription:
		return l.label(labelStepDescription)
	case domain.ProfileDraftStepEmoji:
		return l.label(labelStepEmoji)
	case domain.ProfileDraftStepPhotos:
		return l.label(labelStepPhotos)
	default:
		return l.label(labelProfile)
	}
}

func (l localizer) historyActionLabel(action domain.MatchAction) string {
	switch action {
	case domain.MatchActionLike:
		return l.label(labelActionLike)
	case domain.MatchActionDislike:
		return l.label(labelActionDislike)
	case domain.MatchActionReport:
		return l.label(labelActionReport)
	default:
		return l.label(labelActionNone)
	}
}

func (l localizer) adminStatusLabel(isBanned, isShadowBanned, isModerator, isHidden bool) string {
	parts := make([]string, 0, 4)
	if isBanned {
		parts = append(parts, l.label(labelStatusBanned))
	}
	if isShadowBanned {
		parts = append(parts, l.label(labelStatusShadowBanned))
	}
	if isModerator {
		parts = append(parts, l.label(labelStatusModerator))
	}
	if isHidden {
		parts = append(parts, l.label(labelStatusHidden))
	}
	if len(parts) == 0 {
		return l.label(labelStatusVisible)
	}
	return joinWithComma(parts)
}

func (l localizer) cityLabel(city string) string {
	if city == "" {
		return city
	}
	if labels, ok := cityLabels[l.lang]; ok {
		if value, ok := labels[city]; ok {
			return value
		}
	}
	return city
}

func (l localizer) languageLabel(lang domain.Language) string {
	switch lang {
	case domain.LanguageUkrainian:
		return l.label(labelLanguageUkrainian)
	case domain.LanguageRussian:
		return l.label(labelLanguageRussian)
	default:
		return l.label(labelLanguageEnglish)
	}
}

func formatTemplate(template string, args ...any) string {
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}

func joinWithComma(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.Join(values, ", ")
}

func lookupMessage(lang domain.Language, key messageKey) string {
	if catalog, ok := messageCatalog[lang]; ok {
		if value, ok := catalog[key]; ok {
			return value
		}
	}
	return ""
}

func lookupButton(lang domain.Language, key buttonKey) string {
	if catalog, ok := buttonCatalog[lang]; ok {
		if value, ok := catalog[key]; ok {
			return value
		}
	}
	return ""
}

func lookupLabel(lang domain.Language, key labelKey) string {
	if catalog, ok := labelCatalog[lang]; ok {
		if value, ok := catalog[key]; ok {
			return value
		}
	}
	return ""
}

const (
	msgDefaultHelp                      messageKey = "default_help"
	msgUnknownCommand                   messageKey = "unknown_command"
	msgAdminBadge                       messageKey = "admin_badge"
	msgModeratorBadge                   messageKey = "moderator_badge"
	msgAdminMenu                        messageKey = "admin_menu"
	msgModeratorMenu                    messageKey = "moderator_menu"
	msgAdminAccessDenied                messageKey = "admin_access_denied"
	msgAdminModerationRestricted        messageKey = "admin_moderation_restricted"
	msgAdminUsersEmpty                  messageKey = "admin_users_empty"
	msgAdminUsersLoadFailed             messageKey = "admin_users_load_failed"
	msgAdminUsersPage                   messageKey = "admin_users_page"
	msgAdminUsersEmptyPage              messageKey = "admin_users_empty_page"
	msgAdminBannedUsersEmpty            messageKey = "admin_banned_users_empty"
	msgAdminBannedUsersLoadFailed       messageKey = "admin_banned_users_load_failed"
	msgAdminBannedUsersPage             messageKey = "admin_banned_users_page"
	msgAdminBannedUsersEmptyPage        messageKey = "admin_banned_users_empty_page"
	msgAdminShadowBannedUsersEmpty      messageKey = "admin_shadow_banned_users_empty"
	msgAdminShadowBannedUsersLoadFailed messageKey = "admin_shadow_banned_users_load_failed"
	msgAdminShadowBannedUsersPage       messageKey = "admin_shadow_banned_users_page"
	msgAdminShadowBannedUsersEmptyPage  messageKey = "admin_shadow_banned_users_empty_page"
	msgAdminHiddenUsersEmpty            messageKey = "admin_hidden_users_empty"
	msgAdminHiddenUsersLoadFailed       messageKey = "admin_hidden_users_load_failed"
	msgAdminHiddenUsersPage             messageKey = "admin_hidden_users_page"
	msgAdminHiddenUsersEmptyPage        messageKey = "admin_hidden_users_empty_page"
	msgAdminModeratorsEmpty             messageKey = "admin_moderators_empty"
	msgAdminModeratorsLoadFailed        messageKey = "admin_moderators_load_failed"
	msgAdminModeratorsPage              messageKey = "admin_moderators_page"
	msgAdminModeratorsEmptyPage         messageKey = "admin_moderators_empty_page"
	msgAdminReportsEmpty                messageKey = "admin_reports_empty"
	msgAdminReportsLoadFailed           messageKey = "admin_reports_load_failed"
	msgAdminReportsPage                 messageKey = "admin_reports_page"
	msgAdminReportsEmptyPage            messageKey = "admin_reports_empty_page"
	msgAdminBanUsage                    messageKey = "admin_ban_usage"
	msgAdminBanFailed                   messageKey = "admin_ban_failed"
	msgAdminUserNotFound                messageKey = "admin_user_not_found"
	msgAdminBanSuccess                  messageKey = "admin_ban_success"
	msgAdminUnbanUsage                  messageKey = "admin_unban_usage"
	msgAdminUnbanFailed                 messageKey = "admin_unban_failed"
	msgAdminUnbanSuccess                messageKey = "admin_unban_success"
	msgAdminShadowBanUsage              messageKey = "admin_shadow_ban_usage"
	msgAdminShadowBanFailed             messageKey = "admin_shadow_ban_failed"
	msgAdminShadowBanSuccess            messageKey = "admin_shadow_ban_success"
	msgAdminShadowUnbanUsage            messageKey = "admin_shadow_unban_usage"
	msgAdminShadowUnbanFailed           messageKey = "admin_shadow_unban_failed"
	msgAdminShadowUnbanSuccess          messageKey = "admin_shadow_unban_success"
	msgAdminHideProfileUsage            messageKey = "admin_hide_profile_usage"
	msgAdminHideProfileFailed           messageKey = "admin_hide_profile_failed"
	msgAdminHideProfileSuccess          messageKey = "admin_hide_profile_success"
	msgAdminShowProfileUsage            messageKey = "admin_show_profile_usage"
	msgAdminShowProfileFailed           messageKey = "admin_show_profile_failed"
	msgAdminShowProfileSuccess          messageKey = "admin_show_profile_success"
	msgAdminViewProfileUsage            messageKey = "admin_view_profile_usage"
	msgAdminViewProfileFailed           messageKey = "admin_view_profile_failed"
	msgAdminDeleteProfileUsage          messageKey = "admin_delete_profile_usage"
	msgAdminDeleteProfileFailed         messageKey = "admin_delete_profile_failed"
	msgAdminDeleteProfileSuccess        messageKey = "admin_delete_profile_success"
	msgAdminActionFailed                messageKey = "admin_action_failed"
	msgAdminModeratorUsage              messageKey = "admin_moderator_usage"
	msgAdminModeratorFailed             messageKey = "admin_moderator_failed"
	msgAdminModeratorSuccess            messageKey = "admin_moderator_success"
	msgAdminUnmoderatorUsage            messageKey = "admin_unmoderator_usage"
	msgAdminUnmoderatorFailed           messageKey = "admin_unmoderator_failed"
	msgAdminUnmoderatorSuccess          messageKey = "admin_unmoderator_success"
	msgAdminResetChoicesUsage           messageKey = "admin_reset_choices_usage"
	msgAdminResetChoicesFailed          messageKey = "admin_reset_choices_failed"
	msgAdminResetChoicesSuccess         messageKey = "admin_reset_choices_success"
	msgAdminResetStartUsage             messageKey = "admin_reset_start_usage"
	msgAdminResetStartFailed            messageKey = "admin_reset_start_failed"
	msgAdminResetStartSuccess           messageKey = "admin_reset_start_success"
	msgAdminClearReportsUsage           messageKey = "admin_clear_reports_usage"
	msgAdminClearReportsFailed          messageKey = "admin_clear_reports_failed"
	msgAdminClearReportsSuccess         messageKey = "admin_clear_reports_success"
	msgAdminAdUsage                     messageKey = "admin_ad_usage"
	msgAdminAdFailed                    messageKey = "admin_ad_failed"
	msgAdminAdQueued                    messageKey = "admin_ad_queued"
	msgAdminAdSummary                   messageKey = "admin_ad_summary"
	msgAdminAdSummaryFailed             messageKey = "admin_ad_summary_failed"
	msgAdminAdTooLong                   messageKey = "admin_ad_too_long"
	msgAdminAdDraftStatus               messageKey = "admin_ad_draft_status"
	msgAdminAdButtonUsage               messageKey = "admin_ad_button_usage"
	msgAdminAdButtonAdded               messageKey = "admin_ad_button_added"
	msgAdminAdPreviewSent               messageKey = "admin_ad_preview_sent"
	msgUserBanned                       messageKey = "user_banned"
	msgProfileCreated                   messageKey = "profile_created"
	msgProfileUpdated                   messageKey = "profile_updated"
	msgLoadingStart                     messageKey = "loading_start"
	msgLoadingProfileView               messageKey = "loading_profile_view"
	msgLoadingProfilePreview            messageKey = "loading_profile_preview"
	msgLoadingEditName                  messageKey = "loading_edit_name"
	msgLoadingEditGender                messageKey = "loading_edit_gender"
	msgLoadingEditBirthDate             messageKey = "loading_edit_birth_date"
	msgLoadingEditCountry               messageKey = "loading_edit_country"
	msgLoadingEditCity                  messageKey = "loading_edit_city"
	msgLoadingEditDesc                  messageKey = "loading_edit_desc"
	msgLoadingEditEmoji                 messageKey = "loading_edit_emoji"
	msgLoadingEditPhotos                messageKey = "loading_edit_photos"
	msgLoadingProfileVisibility         messageKey = "loading_profile_visibility"
	msgLoadingProfileLikeNotifications  messageKey = "loading_profile_like_notifications"
	msgLoadingSearchStart               messageKey = "loading_search_start"
	msgLoadingSearchNext                messageKey = "loading_search_next"
	msgLoadingSearchPrev                messageKey = "loading_search_prev"
	msgLoadingSearchAI                  messageKey = "loading_search_ai"
	msgLoadingSearchHistory             messageKey = "loading_search_history"
	msgLoadingSearchHistoryProfile      messageKey = "loading_search_history_profile"
	msgLoadingSearchHistoryAction       messageKey = "loading_search_history_action"
	msgLoadingLikesList                 messageKey = "loading_likes_list"
	msgLoadingLikesProfile              messageKey = "loading_likes_profile"
	msgLoadingAdminUsers                messageKey = "loading_admin_users"
	msgLoadingAdminBan                  messageKey = "loading_admin_ban"
	msgLoadingAdminUnban                messageKey = "loading_admin_unban"
	msgLoadingAdminShadowBan            messageKey = "loading_admin_shadow_ban"
	msgLoadingAdminShadowUnban          messageKey = "loading_admin_shadow_unban"
	msgLoadingAdminHiddenUsers          messageKey = "loading_admin_hidden_users"
	msgLoadingAdminHideProfile          messageKey = "loading_admin_hide_profile"
	msgLoadingAdminShowProfile          messageKey = "loading_admin_show_profile"
	msgLoadingAdminViewProfile          messageKey = "loading_admin_view_profile"
	msgLoadingAdminDeleteProfile        messageKey = "loading_admin_delete_profile"
	msgLoadingAdminModerator            messageKey = "loading_admin_moderator"
	msgLoadingAdminUnmoderator          messageKey = "loading_admin_unmoderator"
	msgLoadingAdminResetChoices         messageKey = "loading_admin_reset_choices"
	msgLoadingAdminResetStart           messageKey = "loading_admin_reset_start"
	msgLoadingAdminClearReports         messageKey = "loading_admin_clear_reports"
	msgLoadingAdminAd                   messageKey = "loading_admin_ad"
	msgLoadingAdminBannedUsers          messageKey = "loading_admin_banned_users"
	msgLoadingAdminShadowBannedUsers    messageKey = "loading_admin_shadow_banned_users"
	msgLoadingAdminModerators           messageKey = "loading_admin_moderators"
	msgLoadingAdminReports              messageKey = "loading_admin_reports"
	msgCreatingProfile                  messageKey = "creating_profile"
	msgUpdatingProfile                  messageKey = "updating_profile"
	msgDeletingProfile                  messageKey = "deleting_profile"
	msgProfileDeleteConfirm             messageKey = "profile_delete_confirm"
	msgProfileDeleteCanceled            messageKey = "profile_delete_canceled"
	msgProfileDeleteExpired             messageKey = "profile_delete_expired"
	msgProfileSetupCanceled             messageKey = "profile_setup_canceled"
	msgProfileEditCanceled              messageKey = "profile_edit_canceled"
	msgActionCanceled                   messageKey = "action_canceled"
	msgProfileHidden                    messageKey = "profile_hidden"
	msgProfileVisible                   messageKey = "profile_visible"
	msgProfileVisibilityUpdateFailed    messageKey = "profile_visibility_update_failed"
	msgLikesNotificationsEnabled        messageKey = "likes_notifications_enabled"
	msgLikesNotificationsDisabled       messageKey = "likes_notifications_disabled"
	msgLikesNotificationsUpdateFailed   messageKey = "likes_notifications_update_failed"
	msgSearchGenderPrompt               messageKey = "search_gender_prompt"
	msgSearchAccuracyPrompt             messageKey = "search_accuracy_prompt"
	msgSearchAccuracyEnabled            messageKey = "search_accuracy_enabled"
	msgSearchAccuracyDisabled           messageKey = "search_accuracy_disabled"
	msgSearchAccuracyUpdateFailed       messageKey = "search_accuracy_update_failed"
	msgSearchAIPrompt                   messageKey = "search_ai_prompt"
	msgSearchAITooShort                 messageKey = "search_ai_too_short"
	msgSearchAIUnavailable              messageKey = "search_ai_unavailable"
	msgSearchNoCandidates               messageKey = "search_no_candidates"
	msgSearchNoPrevious                 messageKey = "search_no_previous"
	msgSearchHistoryEmpty               messageKey = "search_history_empty"
	msgSearchStartRequired              messageKey = "search_start_required"
	msgSearchProfileMissing             messageKey = "search_profile_missing"
	msgSearchUnavailable                messageKey = "search_unavailable"
	msgSearchActionFailed               messageKey = "search_action_failed"
	msgSearchHistoryPage                messageKey = "search_history_page"
	msgSearchHistoryEmptyPage           messageKey = "search_history_empty_page"
	msgSearchHistoryAction              messageKey = "search_history_action"
	msgLikesListEmpty                   messageKey = "likes_list_empty"
	msgLikesListEmptyPage               messageKey = "likes_list_empty_page"
	msgLikesListPage                    messageKey = "likes_list_page"
	msgLikesListLine                    messageKey = "likes_list_line"
	msgMatchActions                     messageKey = "match_actions"
	msgMatchActionSaved                 messageKey = "match_action_saved"
	msgMatchProfileNotFound             messageKey = "match_profile_not_found"
	msgProfileViewRequiresProfile       messageKey = "profile_view_requires_profile"
	msgMatchSuccess                     messageKey = "match_success"
	msgMatchNickname                    messageKey = "match_nickname"
	msgNamePromptNoTelegram             messageKey = "name_prompt_no_telegram"
	msgNamePromptWithTelegram           messageKey = "name_prompt_with_telegram"
	msgBotCheckPrompt                   messageKey = "bot_check_prompt"
	msgBirthDatePrompt                  messageKey = "birth_date_prompt"
	msgGenderPrompt                     messageKey = "gender_prompt"
	msgCountryPrompt                    messageKey = "country_prompt"
	msgCityPrompt                       messageKey = "city_prompt"
	msgDescriptionPrompt                messageKey = "description_prompt"
	msgEmojiPrompt                      messageKey = "emoji_prompt"
	msgPhotosPrompt                     messageKey = "photos_prompt"
	msgProfileEditMenu                  messageKey = "profile_edit_menu"
	msgProfileSettings                  messageKey = "profile_settings"
	msgProfileSettingsWithVisibility    messageKey = "profile_settings_with_visibility"
	msgProfileActions                   messageKey = "profile_actions"
	msgProfilePreviewActions            messageKey = "profile_preview_actions"
	msgEditNamePrompt                   messageKey = "edit_name_prompt"
	msgEditBirthDatePrompt              messageKey = "edit_birth_date_prompt"
	msgEditGenderPrompt                 messageKey = "edit_gender_prompt"
	msgEditCountryPrompt                messageKey = "edit_country_prompt"
	msgEditCityPrompt                   messageKey = "edit_city_prompt"
	msgEditDescriptionPrompt            messageKey = "edit_description_prompt"
	msgEditEmojiPrompt                  messageKey = "edit_emoji_prompt"
	msgEditPhotosPrompt                 messageKey = "edit_photos_prompt"
	msgEditDefaultPrompt                messageKey = "edit_default_prompt"
	msgProfileDetailsHeader             messageKey = "profile_details_header"
	msgProfilePreviewFallbackName       messageKey = "profile_preview_fallback_name"
	msgProfileEditHeader                messageKey = "profile_edit_header"
	msgProfileSetupHeader               messageKey = "profile_setup_header"
	msgProfileSetupEstimate             messageKey = "profile_setup_estimate"
	msgStartGreeting                    messageKey = "start_greeting"
	msgStartGreetingNamed               messageKey = "start_greeting_named"
	msgProfileNotFoundCreateButton      messageKey = "profile_not_found_create_button"
	msgProfileNotFoundButtonBelow       messageKey = "profile_not_found_button_below"
	msgProfileNotFoundUseProfile        messageKey = "profile_not_found_use_profile"
	msgProfileDeleted                   messageKey = "profile_deleted"
	msgProfileDeletedDraftWarning       messageKey = "profile_deleted_draft_warning"
	msgProfileDeletedDraftWarningOnly   messageKey = "profile_deleted_draft_warning_only"
	msgProfileSetupUnavailable          messageKey = "profile_setup_unavailable"
	msgProfileSetupUnavailableChat      messageKey = "profile_setup_unavailable_chat"
	msgProfileSetupStartFailed          messageKey = "profile_setup_start_failed"
	msgProfileSetupLoadFailed           messageKey = "profile_setup_load_failed"
	msgProfileSetupNotActive            messageKey = "profile_setup_not_active"
	msgProfileSetupSaveFailed           messageKey = "profile_setup_save_failed"
	msgProfileEditUnavailable           messageKey = "profile_edit_unavailable"
	msgProfileEditUnavailableChat       messageKey = "profile_edit_unavailable_chat"
	msgProfileServiceUnavailable        messageKey = "profile_service_unavailable"
	msgProfileLoadFailed                messageKey = "profile_load_failed"
	msgProfileDeleteUnavailable         messageKey = "profile_delete_unavailable"
	msgProfileDeleteUnavailableChat     messageKey = "profile_delete_unavailable_chat"
	msgProfileDeletePrepareFailed       messageKey = "profile_delete_prepare_failed"
	msgProfileDeleteFailed              messageKey = "profile_delete_failed"
	msgProfileDeleteCancelFailed        messageKey = "profile_delete_cancel_failed"
	msgProfileEditStartFailed           messageKey = "profile_edit_start_failed"
	msgProfileEditSaveFailed            messageKey = "profile_edit_save_failed"
	msgProfileSaveFailed                messageKey = "profile_save_failed"
	msgBotCheckTooManyAttempts          messageKey = "bot_check_too_many_attempts"
	msgBotCheckIncorrect                messageKey = "bot_check_incorrect"
	msgNamePromptEmpty                  messageKey = "name_prompt_empty"
	msgNamePromptEmptyCreate            messageKey = "name_prompt_empty_create"
	msgNamePromptTelegramMissing        messageKey = "name_prompt_telegram_missing"
	msgNameTooLong                      messageKey = "name_too_long"
	msgGenderInvalid                    messageKey = "gender_invalid"
	msgBirthDateInvalid                 messageKey = "birth_date_invalid"
	msgAgeConfirmPrompt                 messageKey = "age_confirm_prompt"
	msgAgeConfirmDeclined               messageKey = "age_confirm_declined"
	msgAgeConfirmSaveFailed             messageKey = "age_confirm_save_failed"
	msgAgeTooYoung                      messageKey = "age_too_young"
	msgAgeTooOld                        messageKey = "age_too_old"
	msgAgeAccessDenied                  messageKey = "age_access_denied"
	msgCountryInvalid                   messageKey = "country_invalid"
	msgCityInvalid                      messageKey = "city_invalid"
	msgDescriptionEmpty                 messageKey = "description_empty"
	msgDescriptionTooLong               messageKey = "description_too_long"
	msgEmojiInvalid                     messageKey = "emoji_invalid"
	msgPhotosNotEnough                  messageKey = "photos_not_enough"
	msgPhotosPromptRepeat               messageKey = "photos_prompt_repeat"
	msgPhotosLimitReached               messageKey = "photos_limit_reached"
	msgPhotosProgress                   messageKey = "photos_progress"
	msgProfileViewUnavailable           messageKey = "profile_view_unavailable"
	msgProfileViewUnavailableChat       messageKey = "profile_view_unavailable_chat"
	msgProfileDeleteNoteDraft           messageKey = "profile_delete_note_draft"
	msgAdminUserListNoName              messageKey = "admin_user_list_no_name"
	msgAdminUserListNA                  messageKey = "admin_user_list_na"
	msgAdminUserListLine                messageKey = "admin_user_list_line"
	msgAdminReportedUserListLine        messageKey = "admin_reported_user_list_line"
	msgSearchHistoryLine                messageKey = "search_history_line"
	msgSearchHistoryLineShort           messageKey = "search_history_line_short"
	msgSearchHistoryUserFallback        messageKey = "search_history_user_fallback"
	msgSearchHistoryLabel               messageKey = "search_history_label"
	msgAgeShort                         messageKey = "age_short"
	msgProfileDetailsLine               messageKey = "profile_details_line"
	msgProfileDetailsPhotos             messageKey = "profile_details_photos"
	msgProfileDetailsCreated            messageKey = "profile_details_created"
	msgProfileDetailsUpdated            messageKey = "profile_details_updated"
	msgProfileDetailsDescriptionLabel   messageKey = "profile_details_description_label"
	msgLikeNotification                 messageKey = "like_notification"
	msgLikeNotificationFrom             messageKey = "like_notification_from"
	msgLanguagePrompt                   messageKey = "language_prompt"
	msgLanguageUpdated                  messageKey = "language_updated"
	msgLanguageUnsupported              messageKey = "language_unsupported"
	msgProfileSettingsLanguageHint      messageKey = "profile_settings_language_hint"
)

const (
	btnCancel                    buttonKey = "cancel"
	btnAgeConfirmYes             buttonKey = "age_confirm_yes"
	btnAgeConfirmNo              buttonKey = "age_confirm_no"
	btnProfileSetupBack          buttonKey = "profile_setup_back"
	btnProfileDeleteCancel       buttonKey = "profile_delete_cancel"
	btnBackToProfile             buttonKey = "back_to_profile"
	btnEditCancel                buttonKey = "edit_cancel"
	btnSearchAccuracyCancel      buttonKey = "search_accuracy_cancel"
	btnSearchRefresh             buttonKey = "search_refresh"
	btnSearchHistory             buttonKey = "search_history"
	btnSearchHistoryPrev         buttonKey = "search_history_prev"
	btnSearchHistoryNext         buttonKey = "search_history_next"
	btnSearchHistoryBack         buttonKey = "search_history_back"
	btnLikesInbox                buttonKey = "likes_inbox"
	btnLikesBack                 buttonKey = "likes_back"
	btnAdminMenu                 buttonKey = "admin_menu"
	btnModeratorMenu             buttonKey = "moderator_menu"
	btnAdminUsers                buttonKey = "admin_users"
	btnAdminBan                  buttonKey = "admin_ban"
	btnAdminUnban                buttonKey = "admin_unban"
	btnAdminShadowBan            buttonKey = "admin_shadow_ban"
	btnAdminShadowUnban          buttonKey = "admin_shadow_unban"
	btnAdminHiddenUsers          buttonKey = "admin_hidden_users"
	btnAdminHideProfile          buttonKey = "admin_hide_profile"
	btnAdminShowProfile          buttonKey = "admin_show_profile"
	btnAdminViewProfile          buttonKey = "admin_view_profile"
	btnAdminDeleteProfile        buttonKey = "admin_delete_profile"
	btnAdminModerator            buttonKey = "admin_moderator"
	btnAdminUnmoderator          buttonKey = "admin_unmoderator"
	btnAdminResetChoices         buttonKey = "admin_reset_choices"
	btnAdminResetStart           buttonKey = "admin_reset_start"
	btnAdminPostAd               buttonKey = "admin_post_ad"
	btnAdminBannedUsers          buttonKey = "admin_banned_users"
	btnAdminShadowBannedUsers    buttonKey = "admin_shadow_banned_users"
	btnAdminModerators           buttonKey = "admin_moderators"
	btnAdminReports              buttonKey = "admin_reports"
	btnAdminClearReports         buttonKey = "admin_clear_reports"
	btnAdminAdAddButton          buttonKey = "admin_ad_add_button"
	btnAdminAdClearButtons       buttonKey = "admin_ad_clear_buttons"
	btnAdminAdRemovePhoto        buttonKey = "admin_ad_remove_photo"
	btnAdminAdPreview            buttonKey = "admin_ad_preview"
	btnAdminAdSend               buttonKey = "admin_ad_send"
	btnAdminUsersPrev            buttonKey = "admin_users_prev"
	btnAdminUsersNext            buttonKey = "admin_users_next"
	btnAdminBackToMenu           buttonKey = "admin_back_to_menu"
	btnStartSearch               buttonKey = "start_search"
	btnStartSearchAI             buttonKey = "start_search_ai"
	btnSearchAIRepeat            buttonKey = "search_ai_repeat"
	btnProfile                   buttonKey = "profile"
	btnCreateProfile             buttonKey = "create_profile"
	btnSettings                  buttonKey = "settings"
	btnPreviewProfile            buttonKey = "preview_profile"
	btnEditProfile               buttonKey = "edit_profile"
	btnDeleteProfile             buttonKey = "delete_profile"
	btnDeleteConfirm             buttonKey = "delete_confirm"
	btnProfileEditName           buttonKey = "profile_edit_name"
	btnProfileEditGender         buttonKey = "profile_edit_gender"
	btnProfileEditBirthDate      buttonKey = "profile_edit_birth_date"
	btnProfileEditCountry        buttonKey = "profile_edit_country"
	btnProfileEditCity           buttonKey = "profile_edit_city"
	btnProfileEditDescription    buttonKey = "profile_edit_description"
	btnProfileEditEmoji          buttonKey = "profile_edit_emoji"
	btnProfileEditPhotos         buttonKey = "profile_edit_photos"
	btnTelegramName              buttonKey = "telegram_name"
	btnAlbumDone                 buttonKey = "album_done"
	btnGenderMale                buttonKey = "gender_male"
	btnGenderFemale              buttonKey = "gender_female"
	btnGenderOther               buttonKey = "gender_other"
	btnCountryRussia             buttonKey = "country_russia"
	btnCountryKazakhstan         buttonKey = "country_kazakhstan"
	btnCountryBelarus            buttonKey = "country_belarus"
	btnCountryUkraine            buttonKey = "country_ukraine"
	btnSearchMen                 buttonKey = "search_men"
	btnSearchWomen               buttonKey = "search_women"
	btnSearchOther               buttonKey = "search_other"
	btnSearchAny                 buttonKey = "search_any"
	btnReport                    buttonKey = "report"
	btnBackToPreviousProfile     buttonKey = "back_to_previous_profile"
	btnViewProfile               buttonKey = "view_profile"
	btnHideFromSearch            buttonKey = "hide_from_search"
	btnShowInSearch              buttonKey = "show_in_search"
	btnLanguage                  buttonKey = "language"
	btnAdvancedSearchEnable      buttonKey = "advanced_search_enable"
	btnAdvancedSearchDisable     buttonKey = "advanced_search_disable"
	btnLikesNotificationsEnable  buttonKey = "likes_notifications_enable"
	btnLikesNotificationsDisable buttonKey = "likes_notifications_disable"
	btnAdvancedSearchStatus      buttonKey = "advanced_search_status"
	btnLikesNotificationsStatus  buttonKey = "likes_notifications_status"
	btnBackToSettings            buttonKey = "back_to_settings"
	btnLanguageEnglish           buttonKey = "language_english"
	btnLanguageRussian           buttonKey = "language_russian"
	btnLanguageUkrainian         buttonKey = "language_ukrainian"
)

const (
	labelNotSet             labelKey = "not_set"
	labelUnknown            labelKey = "unknown"
	labelProfile            labelKey = "profile"
	labelStepVerification   labelKey = "step_verification"
	labelStepName           labelKey = "step_name"
	labelStepGender         labelKey = "step_gender"
	labelStepBirthDate      labelKey = "step_birth_date"
	labelStepCountry        labelKey = "step_country"
	labelStepCity           labelKey = "step_city"
	labelStepDescription    labelKey = "step_description"
	labelStepEmoji          labelKey = "step_emoji"
	labelStepPhotos         labelKey = "step_photos"
	labelGenderMale         labelKey = "gender_male"
	labelGenderFemale       labelKey = "gender_female"
	labelGenderOther        labelKey = "gender_other"
	labelCountryRussia      labelKey = "country_russia"
	labelCountryKazakhstan  labelKey = "country_kazakhstan"
	labelCountryBelarus     labelKey = "country_belarus"
	labelCountryUkraine     labelKey = "country_ukraine"
	labelActionLike         labelKey = "action_like"
	labelActionDislike      labelKey = "action_dislike"
	labelActionReport       labelKey = "action_report"
	labelActionNone         labelKey = "action_none"
	labelStatusBanned       labelKey = "status_banned"
	labelStatusShadowBanned labelKey = "status_shadow_banned"
	labelStatusModerator    labelKey = "status_moderator"
	labelStatusHidden       labelKey = "status_hidden"
	labelStatusVisible      labelKey = "status_visible"
	labelStatusEnabled      labelKey = "status_enabled"
	labelStatusDisabled     labelKey = "status_disabled"
	labelYes                labelKey = "yes"
	labelNo                 labelKey = "no"
	labelName               labelKey = "label_name"
	labelEmoji              labelKey = "label_emoji"
	labelGender             labelKey = "label_gender"
	labelAge                labelKey = "label_age"
	labelCountry            labelKey = "label_country"
	labelCity               labelKey = "label_city"
	labelSearchVisibility   labelKey = "label_search_visibility"
	labelDescription        labelKey = "label_description"
	labelPhotos             labelKey = "label_photos"
	labelCreated            labelKey = "label_created"
	labelUpdated            labelKey = "label_updated"
	labelLanguageEnglish    labelKey = "language_english"
	labelLanguageRussian    labelKey = "language_russian"
	labelLanguageUkrainian  labelKey = "language_ukrainian"
	labelThisUser           labelKey = "this_user"
)

var messageCatalog = map[domain.Language]map[messageKey]string{
	domain.LanguageEnglish: {
		msgDefaultHelp:                      "Meethalf helps you find people by interests and vibe. Tell us who you want to meet and we will do the rest.",
		msgUnknownCommand:                   "I didn't recognize that command. Try the buttons below.",
		msgAdminBadge:                       "Admin access enabled. You're in control.",
		msgModeratorBadge:                   "Moderator access enabled. Thanks for keeping it safe.",
		msgAdminMenu:                        "Admin dashboard — choose an action.",
		msgModeratorMenu:                    "Moderator dashboard — choose an action.",
		msgAdminAccessDenied:                "Admin access is required for this action.",
		msgAdminModerationRestricted:        "Moderators can act only on regular users.",
		msgAdminUsersEmpty:                  "No users yet.",
		msgAdminUsersLoadFailed:             "Couldn't load users. Please try again soon.",
		msgAdminUsersPage:                   "Users: %d total. Showing %d–%d.",
		msgAdminUsersEmptyPage:              "Users: %d total. Nothing on this page.",
		msgAdminBannedUsersEmpty:            "No banned users.",
		msgAdminBannedUsersLoadFailed:       "Couldn't load banned users. Please try again soon.",
		msgAdminBannedUsersPage:             "Banned users: %d total. Showing %d–%d.",
		msgAdminBannedUsersEmptyPage:        "Banned users: %d total. Nothing on this page.",
		msgAdminShadowBannedUsersEmpty:      "No shadow banned users.",
		msgAdminShadowBannedUsersLoadFailed: "Couldn't load shadow banned users. Please try again soon.",
		msgAdminShadowBannedUsersPage:       "Shadow banned users: %d total. Showing %d–%d.",
		msgAdminShadowBannedUsersEmptyPage:  "Shadow banned users: %d total. Nothing on this page.",
		msgAdminHiddenUsersEmpty:            "No hidden profiles.",
		msgAdminHiddenUsersLoadFailed:       "Couldn't load hidden profiles. Please try again soon.",
		msgAdminHiddenUsersPage:             "Hidden profiles: %d total. Showing %d–%d.",
		msgAdminHiddenUsersEmptyPage:        "Hidden profiles: %d total. Nothing on this page.",
		msgAdminModeratorsEmpty:             "No moderators yet.",
		msgAdminModeratorsLoadFailed:        "Couldn't load moderators. Please try again soon.",
		msgAdminModeratorsPage:              "Moderators: %d total. Showing %d–%d.",
		msgAdminModeratorsEmptyPage:         "Moderators: %d total. Nothing on this page.",
		msgAdminReportsEmpty:                "No reports yet.",
		msgAdminReportsLoadFailed:           "Couldn't load reports. Please try again soon.",
		msgAdminReportsPage:                 "Reported users: %d total. Showing %d–%d.",
		msgAdminReportsEmptyPage:            "Reported users: %d total. Nothing on this page.",
		msgAdminBanUsage:                    "Send a user ID or @username to ban.",
		msgAdminBanFailed:                   "Couldn't ban the user. Please try again soon.",
		msgAdminUserNotFound:                "User not found. Check the ID or @username.",
		msgAdminBanSuccess:                  "User %s has been banned.",
		msgAdminUnbanUsage:                  "Send a user ID or @username to unban.",
		msgAdminUnbanFailed:                 "Couldn't unban the user. Please try again soon.",
		msgAdminUnbanSuccess:                "User %s has been unbanned.",
		msgAdminShadowBanUsage:              "Send a user ID or @username to shadow ban.",
		msgAdminShadowBanFailed:             "Couldn't shadow ban the user. Please try again soon.",
		msgAdminShadowBanSuccess:            "User %s has been shadow banned.",
		msgAdminShadowUnbanUsage:            "Send a user ID or @username to remove shadow ban.",
		msgAdminShadowUnbanFailed:           "Couldn't remove shadow ban. Please try again soon.",
		msgAdminShadowUnbanSuccess:          "Shadow ban removed for %s.",
		msgAdminHideProfileUsage:            "Send a user ID or @username to hide the profile.",
		msgAdminHideProfileFailed:           "Couldn't hide the profile. Please try again soon.",
		msgAdminHideProfileSuccess:          "Profile hidden for %s.",
		msgAdminShowProfileUsage:            "Send a user ID or @username to show the profile.",
		msgAdminShowProfileFailed:           "Couldn't show the profile. Please try again soon.",
		msgAdminShowProfileSuccess:          "Profile visible again for %s.",
		msgAdminViewProfileUsage:            "Send a user ID or @username to view the profile.",
		msgAdminViewProfileFailed:           "Couldn't load the profile. Please try again soon.",
		msgAdminDeleteProfileUsage:          "Send a user ID or @username to delete the profile.",
		msgAdminDeleteProfileFailed:         "Couldn't delete the profile. Please try again soon.",
		msgAdminDeleteProfileSuccess:        "Profile deleted for %s.",
		msgAdminActionFailed:                "Couldn't complete the admin action. Please try again soon.",
		msgAdminModeratorUsage:              "Send a user ID or @username to grant moderator access.",
		msgAdminModeratorFailed:             "Couldn't grant moderator access. Please try again soon.",
		msgAdminModeratorSuccess:            "User %s is now a moderator.",
		msgAdminUnmoderatorUsage:            "Send a user ID or @username to remove moderator access.",
		msgAdminUnmoderatorFailed:           "Couldn't remove moderator access. Please try again soon.",
		msgAdminUnmoderatorSuccess:          "User %s is no longer a moderator.",
		msgAdminResetChoicesUsage:           "Send a user ID or @username to reset match choices.",
		msgAdminResetChoicesFailed:          "Couldn't reset match choices. Please try again soon.",
		msgAdminResetChoicesSuccess:         "Match choices reset for %s.",
		msgAdminResetStartUsage:             "Send a user ID or @username to reset the 16+ check.",
		msgAdminResetStartFailed:            "Couldn't reset the 16+ check. Please try again soon.",
		msgAdminResetStartSuccess:           "16+ check reset for %s.",
		msgAdminClearReportsUsage:           "Send a user ID or @username to clear reports.",
		msgAdminClearReportsFailed:          "Couldn't clear reports. Please try again soon.",
		msgAdminClearReportsSuccess:         "Reports cleared for %s.",
		msgAdminAdUsage:                     "Send the ad text and optional photo in any order. Use the buttons below to add links, preview, or send the broadcast.",
		msgAdminAdFailed:                    "Couldn't post the ad. Please try again soon.",
		msgAdminAdQueued:                    "Ad broadcast started. I'll send a summary here when it's done.",
		msgAdminAdSummary:                   "Ad broadcast complete. Recipients: %d. Sent: %d. Failed: %d. Skipped: %d.",
		msgAdminAdSummaryFailed:             "Ad broadcast stopped with errors. Recipients: %d. Sent: %d. Failed: %d. Skipped: %d.",
		msgAdminAdTooLong:                   "Ad text is too long. Max %d characters.",
		msgAdminAdDraftStatus:               "Ad draft updated.\nText: %s\nPhoto: %s\nButtons: %d.",
		msgAdminAdButtonUsage:               "Send the button as: Label | https://example.com",
		msgAdminAdButtonAdded:               "Button added. Total buttons: %d.",
		msgAdminAdPreviewSent:               "Preview sent below.",
		msgUserBanned:                       "Your account is banned. Contact support if this is a mistake.",
		msgProfileCreated:                   "Your profile is ready! 🎉",
		msgProfileUpdated:                   "Profile updated — looking good!",
		msgLoadingStart:                     "Getting your profile ready...",
		msgLoadingProfileView:               "Loading your profile...",
		msgLoadingProfilePreview:            "Preparing your profile preview...",
		msgLoadingEditName:                  "Opening name editor...",
		msgLoadingEditGender:                "Opening gender editor...",
		msgLoadingEditBirthDate:             "Opening birth date editor...",
		msgLoadingEditCountry:               "Opening country selector...",
		msgLoadingEditCity:                  "Opening city selector...",
		msgLoadingEditDesc:                  "Opening description editor...",
		msgLoadingEditEmoji:                 "Opening emoji selector...",
		msgLoadingEditPhotos:                "Opening photo editor...",
		msgLoadingProfileVisibility:         "Updating search visibility...",
		msgLoadingProfileLikeNotifications:  "Updating like notifications...",
		msgLoadingSearchStart:               "Looking for matches...",
		msgLoadingSearchNext:                "Finding the next profile...",
		msgLoadingSearchPrev:                "Opening the previous profile...",
		msgLoadingSearchAI:                  "Analyzing your request...",
		msgLoadingSearchHistory:             "Loading your history...",
		msgLoadingSearchHistoryProfile:      "Opening a history profile...",
		msgLoadingSearchHistoryAction:       "Saving your decision...",
		msgLoadingLikesList:                 "Loading your likes...",
		msgLoadingLikesProfile:              "Opening the profile...",
		msgLoadingAdminUsers:                "Loading users...",
		msgLoadingAdminBan:                  "Applying ban...",
		msgLoadingAdminUnban:                "Lifting ban...",
		msgLoadingAdminShadowBan:            "Applying shadow ban...",
		msgLoadingAdminShadowUnban:          "Lifting shadow ban...",
		msgLoadingAdminHiddenUsers:          "Loading hidden profiles...",
		msgLoadingAdminHideProfile:          "Hiding profile...",
		msgLoadingAdminShowProfile:          "Showing profile...",
		msgLoadingAdminViewProfile:          "Loading profile...",
		msgLoadingAdminDeleteProfile:        "Deleting profile...",
		msgLoadingAdminModerator:            "Granting moderator access...",
		msgLoadingAdminUnmoderator:          "Revoking moderator access...",
		msgLoadingAdminResetChoices:         "Resetting match choices...",
		msgLoadingAdminResetStart:           "Resetting the 16+ check...",
		msgLoadingAdminClearReports:         "Clearing reports...",
		msgLoadingAdminAd:                   "Starting ad broadcast...",
		msgLoadingAdminBannedUsers:          "Loading banned users...",
		msgLoadingAdminShadowBannedUsers:    "Loading shadow banned users...",
		msgLoadingAdminModerators:           "Loading moderators...",
		msgLoadingAdminReports:              "Loading reports...",
		msgCreatingProfile:                  "Building your profile...",
		msgUpdatingProfile:                  "Saving your updates...",
		msgDeletingProfile:                  "Removing your profile...",
		msgProfileDeleteConfirm:             "Delete your profile? This can't be undone.",
		msgProfileDeleteCanceled:            "Deletion canceled. Your profile stays.",
		msgProfileDeleteExpired:             "Deletion request expired. Open Preferences to try again.",
		msgProfileSetupCanceled:             "Profile setup canceled. You're back in the menu.",
		msgProfileEditCanceled:              "Profile edit canceled. You're back in the menu.",
		msgActionCanceled:                   "Action canceled.",
		msgProfileHidden:                    "Your profile is now hidden from search.",
		msgProfileVisible:                   "Your profile is visible in search again.",
		msgProfileVisibilityUpdateFailed:    "Couldn't update search visibility. Please try again soon.",
		msgLikesNotificationsEnabled:        "Like notifications are on.",
		msgLikesNotificationsDisabled:       "Like notifications are off.",
		msgLikesNotificationsUpdateFailed:   "Couldn't update like notifications. Please try again soon.",
		msgSearchGenderPrompt:               "Who are you looking for? Choose a gender.",
		msgSearchAccuracyPrompt:             "Choose match accuracy (0-4).\n0 — wider search with more variety.\n4 — stricter search with closer matches.",
		msgSearchAccuracyEnabled:            "Advanced search is on.",
		msgSearchAccuracyDisabled:           "Advanced search is off.",
		msgSearchAccuracyUpdateFailed:       "Couldn't update advanced search. Please try again soon.",
		msgSearchAIPrompt:                   "Tell us what matters: gender, age range, city, and a few traits or interests.\nExample: \"Woman 25–30, Moscow, loves art, travel, and hiking.\"",
		msgSearchAITooShort:                 "Please add a bit more detail so the AI can help.",
		msgSearchAIUnavailable:              "AI search is unavailable right now. Please try again later.",
		msgSearchNoCandidates:               "No matches yet. Try again a bit later.",
		msgSearchNoPrevious:                 "No previous profile to return to.",
		msgSearchHistoryEmpty:               "Your history is empty.",
		msgSearchStartRequired:              "Tap \"Start matching\" to begin.",
		msgSearchProfileMissing:             "Create your profile first to start matching and view profiles.",
		msgSearchUnavailable:                "Search is unavailable right now.",
		msgSearchActionFailed:               "Couldn't process that action. Try again soon.",
		msgSearchHistoryPage:                "History: %d total. Showing %d–%d.",
		msgSearchHistoryEmptyPage:           "History: %d total. Nothing on this page.",
		msgSearchHistoryAction:              "Current decision: %s.\nWhat would you like to do?",
		msgLikesListEmpty:                   "No likes yet.",
		msgLikesListEmptyPage:               "Likes: %d total. Nothing on this page.",
		msgLikesListPage:                    "Likes: %d total. Showing %d–%d.",
		msgLikesListLine:                    "%d. %s",
		msgMatchActions:                     "What would you like to do?",
		msgMatchActionSaved:                 "Decision saved.",
		msgMatchProfileNotFound:             "Profile not found.",
		msgProfileViewRequiresProfile:       "Create your profile first to view others.",
		msgMatchSuccess:                     "It's a match! You and %s liked each other. 🎉",
		msgMatchNickname:                    "Nickname: %s",
		msgNamePromptNoTelegram:             "Your Telegram profile has no name yet. Type the name you'd like shown (e.g., Alex).",
		msgNamePromptWithTelegram:           "Telegram name: %s\nTap below to use it, or type a different one (e.g., Alex).",
		msgBotCheckPrompt:                   "Quick bot check: %s\nPick the correct answer below.",
		msgBirthDatePrompt:                  "Enter your birth date in %s format (e.g., %s).",
		msgGenderPrompt:                     "Select your gender.",
		msgCountryPrompt:                    "Select your country.",
		msgCityPrompt:                       "Select your city.",
		msgDescriptionPrompt:                "Tell people a bit about yourself.",
		msgEmojiPrompt:                      "Pick the emoji that best describes you.",
		msgPhotosPrompt:                     "Send 1–4 photos for your album. Tap Done when you're finished.",
		msgProfileEditMenu:                  "What would you like to update?",
		msgProfileSettings:                  "Profile settings. Use the buttons below to manage visibility, advanced search, like notifications, language, or delete your profile.",
		msgProfileSettingsWithVisibility:    "Profile settings.\nSearch visibility: %s.\nAdvanced search: %s.\nLike notifications: %s.\nUse the buttons below to manage visibility, advanced search, like notifications, language, or delete your profile.",
		msgProfileActions:                   "Tap below to preview or edit your profile.",
		msgProfilePreviewActions:            "Tap below to go back or edit your profile.",
		msgEditNamePrompt:                   "Enter the new name.",
		msgEditBirthDatePrompt:              "Enter the new birth date in %s format (e.g., %s).",
		msgEditGenderPrompt:                 "Choose the new gender.",
		msgEditCountryPrompt:                "Choose the new country.",
		msgEditCityPrompt:                   "Choose the new city.",
		msgEditDescriptionPrompt:            "Write the new description.",
		msgEditEmojiPrompt:                  "Pick the new emoji.",
		msgEditPhotosPrompt:                 "Send 1–4 photos to replace the album. Tap Done when finished.",
		msgEditDefaultPrompt:                "Enter the new value.",
		msgProfileDetailsHeader:             "Your profile:",
		msgProfilePreviewFallbackName:       "Profile card",
		msgProfileEditHeader:                "Editing profile: %s",
		msgProfileSetupHeader:               "Let's set up your profile (step %d/%d): %s\n%s",
		msgProfileSetupEstimate:             "Estimated time: ~%d min",
		msgStartGreeting:                    "Welcome to Meethalf! Ready to meet someone?",
		msgStartGreetingNamed:               "Welcome to Meethalf, %s! Ready to meet someone?",
		msgProfileNotFoundCreateButton:      "Profile not found. Tap \"Create profile\" to get started.",
		msgProfileNotFoundButtonBelow:       "Profile not found. Tap the button below to create one.",
		msgProfileNotFoundUseProfile:        "Profile not found. Send /profile to create one.",
		msgProfileDeleted:                   "Profile deleted. Tap \"Create profile\" to make a new one.",
		msgProfileDeletedDraftWarning:       "Profile deleted. Note: couldn't clear the profile draft.",
		msgProfileDeletedDraftWarningOnly:   "Note: couldn't clear the profile draft.",
		msgProfileSetupUnavailable:          "Profile setup is unavailable right now.",
		msgProfileSetupUnavailableChat:      "Profile setup isn't available in this chat.",
		msgProfileSetupStartFailed:          "Couldn't start profile setup. Please try again soon.",
		msgProfileSetupLoadFailed:           "Couldn't load profile setup. Please try again soon.",
		msgProfileSetupNotActive:            "Profile setup isn't active. Send /profile to start.",
		msgProfileSetupSaveFailed:           "Couldn't save profile setup. Please try again soon.",
		msgProfileEditUnavailable:           "Profile editing is unavailable right now.",
		msgProfileEditUnavailableChat:       "Profile editing isn't available in this chat.",
		msgProfileServiceUnavailable:        "Profile service is unavailable right now.",
		msgProfileLoadFailed:                "Couldn't load your profile. Please try again soon.",
		msgProfileDeleteUnavailable:         "Profile deletion is unavailable right now.",
		msgProfileDeleteUnavailableChat:     "Profile deletion isn't available in this chat.",
		msgProfileDeletePrepareFailed:       "Couldn't prepare profile deletion. Please try again soon.",
		msgProfileDeleteFailed:              "Couldn't delete the profile. Please try again soon.",
		msgProfileDeleteCancelFailed:        "Couldn't cancel profile deletion. Please try again soon.",
		msgProfileEditStartFailed:           "Couldn't start profile edit. Please try again soon.",
		msgProfileEditSaveFailed:            "Couldn't save profile changes. Please try again soon.",
		msgProfileSaveFailed:                "Couldn't save the profile. Please try again soon.",
		msgBotCheckTooManyAttempts:          "Too many attempts — let's try a new check.",
		msgBotCheckIncorrect:                "Not quite. Try again.",
		msgNamePromptEmpty:                  "Please enter a name (e.g., Alex).",
		msgNamePromptEmptyCreate:            "Please enter a name (e.g., Alex), or use your Telegram name below.",
		msgNamePromptTelegramMissing:        "Your Telegram profile has no name yet. Type the name you'd like shown (e.g., Alex).",
		msgNameTooLong:                      "Name is too long — max %d characters.",
		msgGenderInvalid:                    "Please choose: male, female, or other.",
		msgBirthDateInvalid:                 "Birth date must be in %s format (e.g., %s). Try again.",
		msgAgeConfirmPrompt:                 "To access Meethalf, please confirm that you are at least %d years old. By confirming, you attest that you meet the minimum age requirement.",
		msgAgeConfirmDeclined:               "Only %d+ can use the bot. If that was a mistake, confirm your age below.",
		msgAgeConfirmSaveFailed:             "Couldn't confirm your age. Please try again soon.",
		msgAgeTooYoung:                      "Age must be at least %d. Check the birth date.",
		msgAgeTooOld:                        "Age must be %d or younger. Check the birth date.",
		msgAgeAccessDenied:                  "Access is limited to users %d+.",
		msgCountryInvalid:                   "Please choose a country: Russia, Kazakhstan, Belarus, or Ukraine.",
		msgCityInvalid:                      "Please pick a city from your country's list.",
		msgDescriptionEmpty:                 "Description can't be empty. Try again.",
		msgDescriptionTooLong:               "Description is too long — max %d characters.",
		msgEmojiInvalid:                     "Please choose an emoji from the list.",
		msgPhotosNotEnough:                  "Please send at least %d photos before finishing.",
		msgPhotosPromptRepeat:               "Send 1–4 photos for your album. Tap Done when finished.",
		msgPhotosLimitReached:               "You've reached the %d-photo limit. Tap Done to finish.",
		msgPhotosProgress:                   "Album photos: %d/%d. Send more or tap Done.",
		msgProfileViewUnavailable:           "Profile view is unavailable right now.",
		msgProfileViewUnavailableChat:       "Profile view isn't available in this chat.",
		msgProfileDeleteNoteDraft:           "Profile deleted. Note: couldn't clear the profile draft.",
		msgAdminUserListNoName:              "Name missing",
		msgAdminUserListNA:                  "n/a",
		msgAdminUserListLine:                "%d. %s — ID %d, username %s, %s",
		msgAdminReportedUserListLine:        "%d. %s — ID %d, username %s, reports: %d, %s",
		msgSearchHistoryLine:                "%d. %s — %s",
		msgSearchHistoryLineShort:           "%d) %s",
		msgSearchHistoryUserFallback:        "User #%d",
		msgSearchHistoryLabel:               "%s • %s",
		msgAgeShort:                         "%d years",
		msgProfileDetailsLine:               "%s — %s",
		msgProfileDetailsPhotos:             "Photos saved: %d",
		msgProfileDetailsCreated:            "Created on: %s",
		msgProfileDetailsUpdated:            "Updated on: %s",
		msgProfileDetailsDescriptionLabel:   "About me:\n%s",
		msgLikeNotification:                 "You got a like ❤️. View the profile?",
		msgLikeNotificationFrom:             "You got a like ❤️ from %s. View the profile?",
		msgLanguagePrompt:                   "Choose your language.",
		msgLanguageUpdated:                  "Language updated. Enjoy!",
		msgLanguageUnsupported:              "That language isn't supported. Please choose an available option.",
		msgProfileSettingsLanguageHint:      "Profile settings. Without a profile you can only change the language.",
	},
	domain.LanguageRussian: {
		msgDefaultHelp:                      "Meethalf помогает находить людей по интересам и настроению. Расскажите, кого хотите встретить, и мы сделаем остальное.",
		msgUnknownCommand:                   "Не понимаю эту команду. Воспользуйтесь кнопками ниже.",
		msgAdminBadge:                       "Режим администратора включён. Всё под контролем.",
		msgModeratorBadge:                   "Режим модератора включён. Спасибо, что следите за порядком.",
		msgAdminMenu:                        "Админ-меню — выберите действие.",
		msgModeratorMenu:                    "Меню модератора — выберите действие.",
		msgAdminAccessDenied:                "Для этого нужна админская роль.",
		msgAdminModerationRestricted:        "Модераторы могут модерировать только обычных пользователей.",
		msgAdminUsersEmpty:                  "Пользователей пока нет.",
		msgAdminUsersLoadFailed:             "Не удалось загрузить пользователей. Попробуйте чуть позже.",
		msgAdminUsersPage:                   "Пользователи: всего %d. Показаны %d–%d.",
		msgAdminUsersEmptyPage:              "Пользователи: всего %d. На этой странице пусто.",
		msgAdminBannedUsersEmpty:            "Забаненных пока нет.",
		msgAdminBannedUsersLoadFailed:       "Не удалось загрузить список забаненных. Попробуйте позже.",
		msgAdminBannedUsersPage:             "Забаненные: всего %d. Показаны %d–%d.",
		msgAdminBannedUsersEmptyPage:        "Забаненные: всего %d. На этой странице пусто.",
		msgAdminShadowBannedUsersEmpty:      "Шадоу-банов пока нет.",
		msgAdminShadowBannedUsersLoadFailed: "Не удалось загрузить список шадоу-банов. Попробуйте позже.",
		msgAdminShadowBannedUsersPage:       "Шадоу-бан: всего %d. Показаны %d–%d.",
		msgAdminShadowBannedUsersEmptyPage:  "Шадоу-бан: всего %d. На этой странице пусто.",
		msgAdminHiddenUsersEmpty:            "Скрытых профилей пока нет.",
		msgAdminHiddenUsersLoadFailed:       "Не удалось загрузить скрытые профили. Попробуйте позже.",
		msgAdminHiddenUsersPage:             "Скрытые профили: всего %d. Показаны %d–%d.",
		msgAdminHiddenUsersEmptyPage:        "Скрытые профили: всего %d. На этой странице пусто.",
		msgAdminModeratorsEmpty:             "Модераторов пока нет.",
		msgAdminModeratorsLoadFailed:        "Не удалось загрузить модераторов. Попробуйте позже.",
		msgAdminModeratorsPage:              "Модераторы: всего %d. Показаны %d–%d.",
		msgAdminModeratorsEmptyPage:         "Модераторы: всего %d. На этой странице пусто.",
		msgAdminReportsEmpty:                "Жалоб пока нет.",
		msgAdminReportsLoadFailed:           "Не удалось загрузить жалобы. Попробуйте позже.",
		msgAdminReportsPage:                 "Жалобы: всего %d. Показаны %d–%d.",
		msgAdminReportsEmptyPage:            "Жалобы: всего %d. На этой странице пусто.",
		msgAdminBanUsage:                    "Отправьте ID пользователя или @username, чтобы забанить.",
		msgAdminBanFailed:                   "Не удалось забанить пользователя. Попробуйте позже.",
		msgAdminUserNotFound:                "Пользователь не найден. Проверьте ID или @username.",
		msgAdminBanSuccess:                  "Пользователь %s забанен.",
		msgAdminUnbanUsage:                  "Отправьте ID пользователя или @username, чтобы разбанить.",
		msgAdminUnbanFailed:                 "Не удалось разбанить пользователя. Попробуйте позже.",
		msgAdminUnbanSuccess:                "Пользователь %s разбанен.",
		msgAdminShadowBanUsage:              "Отправьте ID пользователя или @username, чтобы выдать шадоу-бан.",
		msgAdminShadowBanFailed:             "Не удалось выдать шадоу-бан. Попробуйте позже.",
		msgAdminShadowBanSuccess:            "Пользователь %s получил шадоу-бан.",
		msgAdminShadowUnbanUsage:            "Отправьте ID пользователя или @username, чтобы снять шадоу-бан.",
		msgAdminShadowUnbanFailed:           "Не удалось снять шадоу-бан. Попробуйте позже.",
		msgAdminShadowUnbanSuccess:          "С пользователя %s снят шадоу-бан.",
		msgAdminHideProfileUsage:            "Отправьте ID пользователя или @username, чтобы скрыть профиль.",
		msgAdminHideProfileFailed:           "Не удалось скрыть профиль. Попробуйте позже.",
		msgAdminHideProfileSuccess:          "Профиль для %s скрыт.",
		msgAdminShowProfileUsage:            "Отправьте ID пользователя или @username, чтобы открыть профиль.",
		msgAdminShowProfileFailed:           "Не удалось открыть профиль. Попробуйте позже.",
		msgAdminShowProfileSuccess:          "Профиль для %s снова открыт.",
		msgAdminViewProfileUsage:            "Отправьте ID пользователя или @username, чтобы посмотреть профиль.",
		msgAdminViewProfileFailed:           "Не удалось загрузить профиль. Попробуйте позже.",
		msgAdminDeleteProfileUsage:          "Отправьте ID пользователя или @username, чтобы удалить профиль.",
		msgAdminDeleteProfileFailed:         "Не удалось удалить профиль. Попробуйте позже.",
		msgAdminDeleteProfileSuccess:        "Профиль удалён для %s.",
		msgAdminActionFailed:                "Не удалось выполнить админ-действие. Попробуйте позже.",
		msgAdminModeratorUsage:              "Отправьте ID пользователя или @username, чтобы выдать роль модератора.",
		msgAdminModeratorFailed:             "Не удалось выдать роль модератора. Попробуйте позже.",
		msgAdminModeratorSuccess:            "Пользователь %s теперь модератор.",
		msgAdminUnmoderatorUsage:            "Отправьте ID пользователя или @username, чтобы снять роль модератора.",
		msgAdminUnmoderatorFailed:           "Не удалось снять роль модератора. Попробуйте позже.",
		msgAdminUnmoderatorSuccess:          "Пользователь %s больше не модератор.",
		msgAdminResetChoicesUsage:           "Отправьте ID пользователя или @username, чтобы сбросить решения по матчам.",
		msgAdminResetChoicesFailed:          "Не удалось сбросить решения. Попробуйте позже.",
		msgAdminResetChoicesSuccess:         "Решения для %s сброшены.",
		msgAdminResetStartUsage:             "Отправьте ID пользователя или @username, чтобы сбросить проверку 16+.",
		msgAdminResetStartFailed:            "Не удалось сбросить проверку 16+. Попробуйте позже.",
		msgAdminResetStartSuccess:           "Проверка 16+ для %s сброшена.",
		msgAdminClearReportsUsage:           "Отправьте ID пользователя или @username, чтобы очистить жалобы.",
		msgAdminClearReportsFailed:          "Не удалось очистить жалобы. Попробуйте позже.",
		msgAdminClearReportsSuccess:         "Жалобы для %s очищены.",
		msgAdminAdUsage:                     "Отправьте текст и/или фото в любом порядке. Кнопками ниже можно добавить ссылки, посмотреть предпросмотр и отправить рассылку.",
		msgAdminAdFailed:                    "Не удалось отправить рекламу. Попробуйте позже.",
		msgAdminAdQueued:                    "Рассылка рекламы запущена. Я пришлю сводку, когда закончу.",
		msgAdminAdSummary:                   "Рассылка завершена. Получателей: %d. Отправлено: %d. Ошибок: %d. Пропущено: %d.",
		msgAdminAdSummaryFailed:             "Рассылка завершена с ошибками. Получателей: %d. Отправлено: %d. Ошибок: %d. Пропущено: %d.",
		msgAdminAdTooLong:                   "Текст рекламы слишком длинный. Максимум %d символов.",
		msgAdminAdDraftStatus:               "Черновик рекламы обновлён.\nТекст: %s\nФото: %s\nКнопок: %d.",
		msgAdminAdButtonUsage:               "Отправьте кнопку в формате: Текст | https://example.com",
		msgAdminAdButtonAdded:               "Кнопка добавлена. Всего кнопок: %d.",
		msgAdminAdPreviewSent:               "Предпросмотр отправлен ниже.",
		msgUserBanned:                       "Ваш аккаунт заблокирован. Если это ошибка, свяжитесь с поддержкой.",
		msgProfileCreated:                   "Профиль готов! 🎉",
		msgProfileUpdated:                   "Профиль обновлён — отлично выглядит!",
		msgLoadingStart:                     "Готовим ваш профиль...",
		msgLoadingProfileView:               "Загружаем ваш профиль...",
		msgLoadingProfilePreview:            "Готовим предпросмотр профиля...",
		msgLoadingEditName:                  "Открываем изменение имени...",
		msgLoadingEditGender:                "Открываем выбор пола...",
		msgLoadingEditBirthDate:             "Открываем изменение даты рождения...",
		msgLoadingEditCountry:               "Открываем выбор страны...",
		msgLoadingEditCity:                  "Открываем выбор города...",
		msgLoadingEditDesc:                  "Открываем изменение описания...",
		msgLoadingEditEmoji:                 "Открываем выбор эмодзи...",
		msgLoadingEditPhotos:                "Открываем управление фото...",
		msgLoadingProfileVisibility:         "Обновляем видимость в поиске...",
		msgLoadingProfileLikeNotifications:  "Обновляем уведомления о лайках...",
		msgLoadingSearchStart:               "Ищем интересные профили...",
		msgLoadingSearchNext:                "Ищем следующий профиль...",
		msgLoadingSearchPrev:                "Открываем предыдущий профиль...",
		msgLoadingSearchAI:                  "Анализируем ваш запрос...",
		msgLoadingSearchHistory:             "Загружаем историю...",
		msgLoadingSearchHistoryProfile:      "Открываем профиль из истории...",
		msgLoadingSearchHistoryAction:       "Сохраняем ваше решение...",
		msgLoadingLikesList:                 "Загружаем лайки...",
		msgLoadingLikesProfile:              "Открываем профиль...",
		msgLoadingAdminUsers:                "Загружаем пользователей...",
		msgLoadingAdminBan:                  "Баним пользователя...",
		msgLoadingAdminUnban:                "Снимаем бан...",
		msgLoadingAdminShadowBan:            "Выдаем шадоу-бан...",
		msgLoadingAdminShadowUnban:          "Снимаем шадоу-бан...",
		msgLoadingAdminHiddenUsers:          "Загружаем скрытые профили...",
		msgLoadingAdminHideProfile:          "Скрываем профиль...",
		msgLoadingAdminShowProfile:          "Открываем профиль...",
		msgLoadingAdminViewProfile:          "Загружаем профиль...",
		msgLoadingAdminDeleteProfile:        "Удаляем профиль...",
		msgLoadingAdminModerator:            "Выдаём роль модератора...",
		msgLoadingAdminUnmoderator:          "Снимаем роль модератора...",
		msgLoadingAdminResetChoices:         "Сбрасываем решения...",
		msgLoadingAdminResetStart:           "Сбрасываем проверку 16+...",
		msgLoadingAdminClearReports:         "Очищаем жалобы...",
		msgLoadingAdminAd:                   "Запускаем рассылку рекламы...",
		msgLoadingAdminBannedUsers:          "Загружаем бан-лист...",
		msgLoadingAdminShadowBannedUsers:    "Загружаем список шадоу-банов...",
		msgLoadingAdminModerators:           "Загружаем модераторов...",
		msgLoadingAdminReports:              "Загружаем жалобы...",
		msgCreatingProfile:                  "Создаём профиль...",
		msgUpdatingProfile:                  "Сохраняем обновления...",
		msgDeletingProfile:                  "Удаляем профиль...",
		msgProfileDeleteConfirm:             "Удалить профиль? Это действие нельзя отменить.",
		msgProfileDeleteCanceled:            "Удаление отменено. Профиль сохранён.",
		msgProfileDeleteExpired:             "Запрос на удаление устарел. Откройте параметры и попробуйте снова.",
		msgProfileSetupCanceled:             "Создание профиля отменено. Вы в меню.",
		msgProfileEditCanceled:              "Редактирование отменено. Вы в меню.",
		msgActionCanceled:                   "Действие отменено.",
		msgProfileHidden:                    "Профиль скрыт из поиска.",
		msgProfileVisible:                   "Профиль снова виден в поиске.",
		msgProfileVisibilityUpdateFailed:    "Не удалось обновить видимость в поиске. Попробуйте позже.",
		msgLikesNotificationsEnabled:        "Уведомления о лайках включены.",
		msgLikesNotificationsDisabled:       "Уведомления о лайках выключены.",
		msgLikesNotificationsUpdateFailed:   "Не удалось обновить уведомления о лайках. Попробуйте позже.",
		msgSearchGenderPrompt:               "Кого вы ищете? Выберите пол.",
		msgSearchAccuracyPrompt:             "Выберите точность подбора (0-4).\n0 — шире и больше случайных профилей.\n4 — строже и ближе по совпадению.",
		msgSearchAccuracyEnabled:            "Расширенный поиск включён.",
		msgSearchAccuracyDisabled:           "Расширенный поиск выключен.",
		msgSearchAccuracyUpdateFailed:       "Не удалось обновить расширенный поиск. Попробуйте позже.",
		msgSearchAIPrompt:                   "Кого вы хотите найти? Укажите пол, возраст, город и пару важных черт/интересов.\nПример: \"Девушка 25–30, Москва, любит искусство, путешествия, прогулки.\"",
		msgSearchAITooShort:                 "Добавьте чуть больше деталей, чтобы ИИ смог подобрать профиль.",
		msgSearchAIUnavailable:              "Поиск с ИИ сейчас недоступен. Попробуйте позже.",
		msgSearchNoCandidates:               "Пока нет подходящих профилей. Загляните чуть позже.",
		msgSearchNoPrevious:                 "Предыдущего профиля нет.",
		msgSearchHistoryEmpty:               "История пока пуста.",
		msgSearchStartRequired:              "Нажмите «Начать знакомиться», чтобы начать.",
		msgSearchProfileMissing:             "Сначала создайте профиль, чтобы искать и смотреть других.",
		msgSearchUnavailable:                "Поиск сейчас недоступен.",
		msgSearchActionFailed:               "Не удалось обработать действие. Попробуйте позже.",
		msgSearchHistoryPage:                "История: всего %d. Показаны %d–%d.",
		msgSearchHistoryEmptyPage:           "История: всего %d. На этой странице пусто.",
		msgSearchHistoryAction:              "Текущее решение: %s.\nЧто хотите сделать?",
		msgLikesListEmpty:                   "Пока нет лайков.",
		msgLikesListEmptyPage:               "Лайки: всего %d. На этой странице пусто.",
		msgLikesListPage:                    "Лайки: всего %d. Показаны %d–%d.",
		msgLikesListLine:                    "%d. %s",
		msgMatchActions:                     "Что хотите сделать?",
		msgMatchActionSaved:                 "Решение сохранено.",
		msgMatchProfileNotFound:             "Профиль не найден.",
		msgProfileViewRequiresProfile:       "Сначала создайте профиль, чтобы смотреть других.",
		msgMatchSuccess:                     "Это мэтч! Вы и %s понравились друг другу. 🎉",
		msgMatchNickname:                    "Никнейм: %s",
		msgNamePromptNoTelegram:             "В профиле Telegram нет имени. Напишите имя, которое будет видно другим (например, Алексей).",
		msgNamePromptWithTelegram:           "Имя в Telegram: %s\nНажмите кнопку ниже, чтобы использовать его, или напишите другое (например, Алексей).",
		msgBotCheckPrompt:                   "Быстрая проверка на бота: %s\nВыберите правильный ответ ниже.",
		msgBirthDatePrompt:                  "Введите дату рождения в формате %s (например, %s).",
		msgGenderPrompt:                     "Выберите свой пол.",
		msgCountryPrompt:                    "Выберите страну.",
		msgCityPrompt:                       "Выберите город.",
		msgDescriptionPrompt:                "Пару слов о себе — так вас легче найти.",
		msgEmojiPrompt:                      "Выберите эмодзи, который вас описывает.",
		msgPhotosPrompt:                     "Пришлите 1–4 фото для альбома. Когда закончите — «Готово».",
		msgProfileEditMenu:                  "Что хотите обновить в профиле?",
		msgProfileSettings:                  "Параметры профиля. Кнопками ниже можно управлять видимостью, расширенным поиском, уведомлениями о лайках, языком или удалить профиль.",
		msgProfileSettingsWithVisibility:    "Параметры профиля.\nВидимость в поиске: %s.\nРасширенный поиск: %s.\nУведомления о лайках: %s.\nКнопками ниже можно управлять видимостью, расширенным поиском, уведомлениями о лайках, языком или удалить профиль.",
		msgProfileActions:                   "Кнопками ниже можно посмотреть предпросмотр или отредактировать профиль.",
		msgProfilePreviewActions:            "Кнопками ниже можно вернуться к профилю или отредактировать его.",
		msgEditNamePrompt:                   "Введите новое имя.",
		msgEditBirthDatePrompt:              "Введите новую дату рождения в формате %s (например, %s).",
		msgEditGenderPrompt:                 "Выберите новый пол.",
		msgEditCountryPrompt:                "Выберите новую страну.",
		msgEditCityPrompt:                   "Выберите новый город.",
		msgEditDescriptionPrompt:            "Напишите новое описание.",
		msgEditEmojiPrompt:                  "Выберите новый эмодзи.",
		msgEditPhotosPrompt:                 "Пришлите 1–4 фото, чтобы заменить альбом. Когда закончите — «Готово».",
		msgEditDefaultPrompt:                "Введите новое значение.",
		msgProfileDetailsHeader:             "Ваш профиль:",
		msgProfilePreviewFallbackName:       "Анкета",
		msgProfileEditHeader:                "Редактирование профиля: %s",
		msgProfileSetupHeader:               "Создаём профиль (шаг %d/%d): %s\n%s",
		msgProfileSetupEstimate:             "Примерное время: ~%d мин",
		msgStartGreeting:                    "Добро пожаловать в Meethalf! Готовы познакомиться?",
		msgStartGreetingNamed:               "Добро пожаловать в Meethalf, %s! Готовы познакомиться?",
		msgProfileNotFoundCreateButton:      "Профиль не найден. Нажмите «Создать профиль», чтобы начать.",
		msgProfileNotFoundButtonBelow:       "Профиль не найден. Нажмите кнопку ниже, чтобы создать его.",
		msgProfileNotFoundUseProfile:        "Профиль не найден. Отправьте /profile, чтобы создать его.",
		msgProfileDeleted:                   "Профиль удалён. Нажмите «Создать профиль», чтобы создать новый.",
		msgProfileDeletedDraftWarning:       "Профиль удалён. Примечание: не удалось очистить черновик профиля.",
		msgProfileDeletedDraftWarningOnly:   "Примечание: не удалось очистить черновик профиля.",
		msgProfileSetupUnavailable:          "Создание профиля сейчас недоступно.",
		msgProfileSetupUnavailableChat:      "Создание профиля недоступно в этом чате.",
		msgProfileSetupStartFailed:          "Не удалось запустить создание профиля. Попробуйте позже.",
		msgProfileSetupLoadFailed:           "Не удалось загрузить создание профиля. Попробуйте позже.",
		msgProfileSetupNotActive:            "Создание профиля не активно. Отправьте /profile, чтобы начать.",
		msgProfileSetupSaveFailed:           "Не удалось сохранить создание профиля. Попробуйте позже.",
		msgProfileEditUnavailable:           "Редактирование профиля сейчас недоступно.",
		msgProfileEditUnavailableChat:       "Редактирование профиля недоступно в этом чате.",
		msgProfileServiceUnavailable:        "Сервис профиля сейчас недоступен.",
		msgProfileLoadFailed:                "Не удалось загрузить профиль. Попробуйте позже.",
		msgProfileDeleteUnavailable:         "Удаление профиля сейчас недоступно.",
		msgProfileDeleteUnavailableChat:     "Удаление профиля недоступно в этом чате.",
		msgProfileDeletePrepareFailed:       "Не удалось подготовить удаление профиля. Попробуйте позже.",
		msgProfileDeleteFailed:              "Не удалось удалить профиль. Попробуйте позже.",
		msgProfileDeleteCancelFailed:        "Не удалось отменить удаление профиля. Попробуйте позже.",
		msgProfileEditStartFailed:           "Не удалось начать редактирование профиля. Попробуйте позже.",
		msgProfileEditSaveFailed:            "Не удалось сохранить изменения профиля. Попробуйте позже.",
		msgProfileSaveFailed:                "Не удалось сохранить профиль. Попробуйте позже.",
		msgBotCheckTooManyAttempts:          "Слишком много попыток — давайте попробуем другую проверку.",
		msgBotCheckIncorrect:                "Ответ неверный. Попробуйте ещё раз.",
		msgNamePromptEmpty:                  "Введите имя (например, Алексей).",
		msgNamePromptEmptyCreate:            "Введите имя (например, Алексей) или воспользуйтесь кнопкой ниже, чтобы взять имя из Telegram.",
		msgNamePromptTelegramMissing:        "В профиле Telegram нет имени. Напишите имя, которое будет видно другим (например, Алексей).",
		msgNameTooLong:                      "Имя слишком длинное (максимум %d символов).",
		msgGenderInvalid:                    "Выберите пол: мужской, женский или другое.",
		msgBirthDateInvalid:                 "Дата рождения должна быть в формате %s (например, %s). Попробуйте снова.",
		msgAgeConfirmPrompt:                 "Для доступа к сервису Meethalf подтвердите, что вам исполнилось не менее %d лет. Нажимая кнопку подтверждения, вы заявляете о соответствии возрастному требованию.",
		msgAgeConfirmDeclined:               "Доступ только для %d+. Если вы нажали по ошибке, подтвердите возраст ниже.",
		msgAgeConfirmSaveFailed:             "Не удалось подтвердить возраст. Попробуйте позже.",
		msgAgeTooYoung:                      "Возраст должен быть не меньше %d. Проверьте дату рождения.",
		msgAgeTooOld:                        "Возраст должен быть не больше %d. Проверьте дату рождения.",
		msgAgeAccessDenied:                  "Доступ к боту только для пользователей %d+.",
		msgCountryInvalid:                   "Выберите страну: Россия, Казахстан, Беларусь или Украина.",
		msgCityInvalid:                      "Выберите город из списка вашей страны.",
		msgDescriptionEmpty:                 "Описание не может быть пустым. Попробуйте снова.",
		msgDescriptionTooLong:               "Описание слишком длинное (максимум %d символов).",
		msgEmojiInvalid:                     "Выберите эмодзи из списка.",
		msgPhotosNotEnough:                  "Отправьте минимум %d фото перед завершением.",
		msgPhotosPromptRepeat:               "Пришлите 1–4 фото для альбома. Когда закончите — «Готово».",
		msgPhotosLimitReached:               "Вы достигли лимита в %d фото. Нажмите «Готово», чтобы завершить.",
		msgPhotosProgress:                   "Фото в альбоме: %d/%d. Отправьте ещё или нажмите «Готово».",
		msgProfileViewUnavailable:           "Просмотр профиля сейчас недоступен.",
		msgProfileViewUnavailableChat:       "Просмотр профиля недоступен в этом чате.",
		msgProfileDeleteNoteDraft:           "Профиль удалён. Примечание: не удалось очистить черновик профиля.",
		msgAdminUserListNoName:              "Имя не указано",
		msgAdminUserListNA:                  "н/д",
		msgAdminUserListLine:                "%d. %s — ID %d, username %s, %s",
		msgAdminReportedUserListLine:        "%d. %s — ID %d, username %s, жалобы: %d, %s",
		msgSearchHistoryLine:                "%d. %s — %s",
		msgSearchHistoryLineShort:           "%d) %s",
		msgSearchHistoryUserFallback:        "Пользователь №%d",
		msgSearchHistoryLabel:               "%s • %s",
		msgAgeShort:                         "возраст %d",
		msgProfileDetailsLine:               "%s — %s",
		msgProfileDetailsPhotos:             "Фото в альбоме: %d",
		msgProfileDetailsCreated:            "Дата создания: %s",
		msgProfileDetailsUpdated:            "Последнее обновление: %s",
		msgProfileDetailsDescriptionLabel:   "О себе:\n%s",
		msgLikeNotification:                 "Вам поставили лайк ❤️. Посмотреть профиль?",
		msgLikeNotificationFrom:             "Вам поставили лайк ❤️ от %s. Посмотреть профиль?",
		msgLanguagePrompt:                   "Выберите язык.",
		msgLanguageUpdated:                  "Язык обновлён. Готово!",
		msgLanguageUnsupported:              "Этот язык не поддерживается. Выберите один из доступных.",
		msgProfileSettingsLanguageHint:      "Параметры профиля. Пока доступна только смена языка.",
	},
}

var buttonCatalog = map[domain.Language]map[buttonKey]string{
	domain.LanguageEnglish: {
		btnCancel:                    "Back to main menu",
		btnAgeConfirmYes:             "Yes, I'm 16+",
		btnAgeConfirmNo:              "No, I'm under 16",
		btnProfileSetupBack:          "Back a step",
		btnProfileDeleteCancel:       "Keep my profile",
		btnBackToProfile:             "Back to my profile",
		btnEditCancel:                "Cancel edit",
		btnSearchAccuracyCancel:      "Back to gender",
		btnSearchRefresh:             "Refresh matches",
		btnSearchHistory:             "My history",
		btnSearchHistoryPrev:         "Prev",
		btnSearchHistoryNext:         "Next",
		btnSearchHistoryBack:         "Back to list",
		btnLikesInbox:                "People who liked me",
		btnLikesBack:                 "Back to likes",
		btnAdminMenu:                 "Admin dashboard",
		btnModeratorMenu:             "Moderator dashboard",
		btnAdminUsers:                "All users",
		btnAdminBan:                  "Ban a user",
		btnAdminUnban:                "Lift ban",
		btnAdminShadowBan:            "Shadow ban",
		btnAdminShadowUnban:          "Lift shadow ban",
		btnAdminHiddenUsers:          "Hidden profiles",
		btnAdminHideProfile:          "Hide profile",
		btnAdminShowProfile:          "Show profile",
		btnAdminViewProfile:          "View profile",
		btnAdminDeleteProfile:        "Delete profile",
		btnAdminModerator:            "Grant moderator",
		btnAdminUnmoderator:          "Revoke moderator",
		btnAdminResetChoices:         "Reset matches",
		btnAdminResetStart:           "Reset 16+ check",
		btnAdminPostAd:               "Post ad",
		btnAdminBannedUsers:          "Banned list",
		btnAdminShadowBannedUsers:    "Shadow banned list",
		btnAdminModerators:           "Moderator list",
		btnAdminReports:              "Reports",
		btnAdminClearReports:         "Clear reports",
		btnAdminAdAddButton:          "Add button",
		btnAdminAdClearButtons:       "Clear buttons",
		btnAdminAdRemovePhoto:        "Remove photo",
		btnAdminAdPreview:            "Preview",
		btnAdminAdSend:               "Send ad",
		btnAdminUsersPrev:            "Prev",
		btnAdminUsersNext:            "Next",
		btnAdminBackToMenu:           "Back to admin menu",
		btnStartSearch:               "Start matching",
		btnStartSearchAI:             "Search with AI",
		btnSearchAIRepeat:            "Try again",
		btnProfile:                   "My profile",
		btnCreateProfile:             "Create profile",
		btnSettings:                  "Preferences",
		btnPreviewProfile:            "Preview profile",
		btnEditProfile:               "Edit profile",
		btnDeleteProfile:             "Delete profile",
		btnDeleteConfirm:             "Yes, delete it",
		btnProfileEditName:           "Display name",
		btnProfileEditGender:         "My gender",
		btnProfileEditBirthDate:      "Date of birth",
		btnProfileEditCountry:        "My country",
		btnProfileEditCity:           "My city",
		btnProfileEditDescription:    "About me",
		btnProfileEditEmoji:          "My emoji",
		btnProfileEditPhotos:         "My photos",
		btnTelegramName:              "Use my Telegram name",
		btnAlbumDone:                 "All done",
		btnGenderMale:                "Man",
		btnGenderFemale:              "Woman",
		btnGenderOther:               "Other",
		btnCountryRussia:             "Russia 🇷🇺",
		btnCountryKazakhstan:         "Kazakhstan 🇰🇿",
		btnCountryBelarus:            "Belarus 🇧🇾",
		btnCountryUkraine:            "Ukraine 🇺🇦",
		btnSearchMen:                 "Men 👨",
		btnSearchWomen:               "Women 👩",
		btnSearchOther:               "Other",
		btnSearchAny:                 "Anyone",
		btnReport:                    "Report",
		btnBackToPreviousProfile:     "Previous profile",
		btnViewProfile:               "Open profile",
		btnHideFromSearch:            "Hide my profile",
		btnShowInSearch:              "Show my profile",
		btnLanguage:                  "Language",
		btnAdvancedSearchEnable:      "Advanced search: on",
		btnAdvancedSearchDisable:     "Advanced search: off",
		btnLikesNotificationsEnable:  "Like notifications: on",
		btnLikesNotificationsDisable: "Like notifications: off",
		btnAdvancedSearchStatus:      "Advanced search: %s",
		btnLikesNotificationsStatus:  "Like notifications: %s",
		btnBackToSettings:            "Back to preferences",
		btnLanguageEnglish:           "English 🇺🇸",
		btnLanguageRussian:           "Русский 🇷🇺",
		btnLanguageUkrainian:         "Українська 🇺🇦",
	},
	domain.LanguageRussian: {
		btnCancel:                    "В главное меню",
		btnAgeConfirmYes:             "Да, мне 16+",
		btnAgeConfirmNo:              "Нет, мне нет 16",
		btnProfileSetupBack:          "Шаг назад",
		btnProfileDeleteCancel:       "Оставить мой профиль",
		btnBackToProfile:             "К моему профилю",
		btnEditCancel:                "Отменить редактирование",
		btnSearchAccuracyCancel:      "К выбору пола",
		btnSearchRefresh:             "Обновить подборку",
		btnSearchHistory:             "Просмотренные анкеты",
		btnSearchHistoryPrev:         "Предыдущие",
		btnSearchHistoryNext:         "Следующие",
		btnSearchHistoryBack:         "К списку истории",
		btnLikesInbox:                "Кто лайкнул меня",
		btnLikesBack:                 "Назад к лайкам",
		btnAdminMenu:                 "Админ-меню",
		btnModeratorMenu:             "Меню модератора",
		btnAdminUsers:                "Все пользователи",
		btnAdminBan:                  "Забанить",
		btnAdminUnban:                "Разбанить",
		btnAdminShadowBan:            "Шадоу-бан",
		btnAdminShadowUnban:          "Снять шадоу-бан",
		btnAdminHiddenUsers:          "Скрытые профили",
		btnAdminHideProfile:          "Скрыть профиль",
		btnAdminShowProfile:          "Открыть профиль",
		btnAdminViewProfile:          "Посмотреть профиль",
		btnAdminDeleteProfile:        "Удалить профиль",
		btnAdminModerator:            "Назначить модератором",
		btnAdminUnmoderator:          "Снять роль модератора",
		btnAdminResetChoices:         "Сбросить решения по матчам",
		btnAdminResetStart:           "Сбросить проверку 16+",
		btnAdminPostAd:               "Разослать рекламу",
		btnAdminBannedUsers:          "Бан-лист",
		btnAdminShadowBannedUsers:    "Список шадоу-банов",
		btnAdminModerators:           "Список модераторов",
		btnAdminReports:              "Репорты",
		btnAdminClearReports:         "Очистить репорты",
		btnAdminAdAddButton:          "Добавить кнопку",
		btnAdminAdClearButtons:       "Очистить кнопки",
		btnAdminAdRemovePhoto:        "Убрать фото",
		btnAdminAdPreview:            "Предпросмотр",
		btnAdminAdSend:               "Отправить",
		btnAdminUsersPrev:            "Предыдущие",
		btnAdminUsersNext:            "Следующие",
		btnAdminBackToMenu:           "Назад в админ-меню",
		btnStartSearch:               "Начать знакомиться",
		btnStartSearchAI:             "Искать с ИИ",
		btnSearchAIRepeat:            "Попробовать ещё раз",
		btnProfile:                   "Мой профиль",
		btnCreateProfile:             "Создать профиль",
		btnSettings:                  "Параметры",
		btnPreviewProfile:            "Предпросмотр",
		btnEditProfile:               "Редактировать мой профиль",
		btnDeleteProfile:             "Удалить мой профиль",
		btnDeleteConfirm:             "Да, удалить профиль",
		btnProfileEditName:           "Имя в профиле",
		btnProfileEditGender:         "Пол/гендер",
		btnProfileEditBirthDate:      "День рождения",
		btnProfileEditCountry:        "Моя страна",
		btnProfileEditCity:           "Мой город",
		btnProfileEditDescription:    "О себе",
		btnProfileEditEmoji:          "Мой эмодзи",
		btnProfileEditPhotos:         "Мои фото",
		btnTelegramName:              "Взять имя из Telegram",
		btnAlbumDone:                 "Готово",
		btnGenderMale:                "Мужчина",
		btnGenderFemale:              "Женщина",
		btnGenderOther:               "Другое",
		btnCountryRussia:             "Россия 🇷🇺",
		btnCountryKazakhstan:         "Казахстан 🇰🇿",
		btnCountryBelarus:            "Беларусь 🇧🇾",
		btnCountryUkraine:            "Украина 🇺🇦",
		btnSearchMen:                 "Парни 👨",
		btnSearchWomen:               "Девушки 👩",
		btnSearchOther:               "Другое",
		btnSearchAny:                 "Любые",
		btnReport:                    "Пожаловаться",
		btnBackToPreviousProfile:     "Предыдущий профиль",
		btnViewProfile:               "Открыть профиль",
		btnHideFromSearch:            "Скрыть профиль",
		btnShowInSearch:              "Показывать профиль",
		btnLanguage:                  "Язык",
		btnAdvancedSearchEnable:      "Расширенный поиск: активно",
		btnAdvancedSearchDisable:     "Расширенный поиск: неактивно",
		btnLikesNotificationsEnable:  "Уведомления о лайках: активно",
		btnLikesNotificationsDisable: "Уведомления о лайках: неактивно",
		btnAdvancedSearchStatus:      "Расширенный поиск: %s",
		btnLikesNotificationsStatus:  "Уведомления о лайках: %s",
		btnBackToSettings:            "Назад к параметрам",
		btnLanguageEnglish:           "English 🇺🇸",
		btnLanguageRussian:           "Русский 🇷🇺",
		btnLanguageUkrainian:         "Українська 🇺🇦",
	},
}

var labelCatalog = map[domain.Language]map[labelKey]string{
	domain.LanguageEnglish: {
		labelNotSet:             "Not set yet",
		labelUnknown:            "Unknown value",
		labelProfile:            "Profile card",
		labelStepVerification:   "Quick check",
		labelStepName:           "Display name",
		labelStepGender:         "Gender choice",
		labelStepBirthDate:      "Date of birth",
		labelStepCountry:        "Country/region",
		labelStepCity:           "City/town",
		labelStepDescription:    "About me",
		labelStepEmoji:          "Emoji vibe",
		labelStepPhotos:         "Photo album",
		labelGenderMale:         "Man",
		labelGenderFemale:       "Woman",
		labelGenderOther:        "Other",
		labelCountryRussia:      "Russia 🇷🇺",
		labelCountryKazakhstan:  "Kazakhstan 🇰🇿",
		labelCountryBelarus:     "Belarus 🇧🇾",
		labelCountryUkraine:     "Ukraine 🇺🇦",
		labelActionLike:         "Liked",
		labelActionDislike:      "Passed",
		labelActionReport:       "Reported",
		labelActionNone:         "No action",
		labelStatusBanned:       "blocked",
		labelStatusShadowBanned: "shadow banned",
		labelStatusModerator:    "Moderator",
		labelStatusHidden:       "invisible",
		labelStatusVisible:      "shown",
		labelStatusEnabled:      "active",
		labelStatusDisabled:     "inactive",
		labelYes:                "Yes",
		labelNo:                 "No",
		labelName:               "Display name",
		labelEmoji:              "Emoji vibe",
		labelGender:             "Gender identity",
		labelAge:                "Age (years)",
		labelCountry:            "Country/region",
		labelCity:               "City/town",
		labelSearchVisibility:   "Visibility in search",
		labelDescription:        "About me",
		labelPhotos:             "Photo album",
		labelCreated:            "Created on",
		labelUpdated:            "Updated on",
		labelLanguageEnglish:    "English 🇺🇸",
		labelLanguageRussian:    "Russian 🇷🇺",
		labelLanguageUkrainian:  "Ukrainian 🇺🇦",
		labelThisUser:           "this account",
	},
	domain.LanguageRussian: {
		labelNotSet:             "Не указано пока",
		labelUnknown:            "Неизвестно пока",
		labelProfile:            "Анкета",
		labelStepVerification:   "Быстрая проверка",
		labelStepName:           "Имя в профиле",
		labelStepGender:         "Пол/гендер",
		labelStepBirthDate:      "День рождения",
		labelStepCountry:        "Страна/регион",
		labelStepCity:           "Город/посёлок",
		labelStepDescription:    "О себе",
		labelStepEmoji:          "Эмодзи-настроение",
		labelStepPhotos:         "Фотоальбом",
		labelGenderMale:         "Мужчина",
		labelGenderFemale:       "Женщина",
		labelGenderOther:        "Другое",
		labelCountryRussia:      "Россия",
		labelCountryKazakhstan:  "Казахстан",
		labelCountryBelarus:     "Беларусь",
		labelCountryUkraine:     "Украина",
		labelActionLike:         "Понравилось",
		labelActionDislike:      "Пропуск",
		labelActionReport:       "Репорт",
		labelActionNone:         "Без решения",
		labelStatusBanned:       "заблокирован",
		labelStatusShadowBanned: "в шадоу-бане",
		labelStatusModerator:    "Модератор",
		labelStatusHidden:       "невидим",
		labelStatusVisible:      "видимый",
		labelStatusEnabled:      "активно",
		labelStatusDisabled:     "неактивно",
		labelYes:                "Да",
		labelNo:                 "Нет",
		labelName:               "Имя/ник",
		labelEmoji:              "Эмодзи-настроение",
		labelGender:             "Пол/гендер",
		labelAge:                "Возраст (лет)",
		labelCountry:            "Страна/регион",
		labelCity:               "Город/посёлок",
		labelSearchVisibility:   "Показ в поиске",
		labelDescription:        "О себе",
		labelPhotos:             "Фотоальбом",
		labelCreated:            "Дата создания",
		labelUpdated:            "Последнее обновление",
		labelLanguageEnglish:    "Английский 🇺🇸",
		labelLanguageRussian:    "Русский 🇷🇺",
		labelLanguageUkrainian:  "Украинский 🇺🇦",
		labelThisUser:           "этот аккаунт",
	},
}

var cityLabels = map[domain.Language]map[string]string{
	domain.LanguageRussian: {
		"Moscow":           "Москва",
		"Saint Petersburg": "Санкт-Петербург",
		"Novosibirsk":      "Новосибирск",
		"Krasnodar":        "Краснодар",
		"Omsk":             "Омск",
		"Rostov-on-Don":    "Ростов-на-Дону",
		"Perm":             "Пермь",
		"Krasnoyarsk":      "Красноярск",
		"Yekaterinburg":    "Екатеринбург",
		"Kazan":            "Казань",
		"Nizhny Novgorod":  "Нижний Новгород",
		"Ufa":              "Уфа",
		"Chelyabinsk":      "Челябинск",
		"Samara":           "Самара",
		"Voronezh":         "Воронеж",
		"Volgograd":        "Волгоград",
		"Astana":           "Астана",
		"Almaty":           "Алматы",
		"Semey":            "Семей",
		"Pavlodar":         "Павлодар",
		"Shymkent":         "Шымкент",
		"Aktobe":           "Актобе",
		"Karaganda":        "Караганда",
		"Taraz":            "Тараз",
		"Ust-Kamenogorsk":  "Усть-Каменогорск",
		"Atyrau":           "Атырау",
		"Minsk":            "Минск",
		"Gomel":            "Гомель",
		"Mogilev":          "Могилев",
		"Vitebsk":          "Витебск",
		"Grodno":           "Гродно",
		"Brest":            "Брест",
		"Bobruisk":         "Бобруйск",
		"Baranovichi":      "Барановичи",
		"Borisov":          "Борисов",
		"Kyiv":            "Киев",
		"Kharkiv":         "Харьков",
		"Odesa":           "Одесса",
		"Dnipro":          "Днепр",
		"Donetsk":         "Донецк",
		"Lviv":            "Львов",
		"Zaporizhzhia":    "Запорожье",
		"Kryvyi Rih":      "Кривой Рог",
		"Mykolaiv":        "Николаев",
		"Luhansk":         "Луганск",
		"Mariupol":        "Мариуполь",
		"Kherson":         "Херсон",
		"Vinnytsia":       "Винница",
		"Poltava":         "Полтава",
		"Chernihiv":       "Чернигов",
		"Cherkasy":        "Черкассы",
		"Zhytomyr":        "Житомир",
		"Sumy":            "Сумы",
		"Khmelnytskyi":    "Хмельницкий",
		"Rivne":           "Ровно",
		"Ternopil":        "Тернополь",
		"Ivano-Frankivsk": "Ивано-Франковск",
		"Lutsk":           "Луцк",
		"Chernivtsi":      "Черновцы",
	},
}

const defaultButtonEmoji = ""

var buttonEmojiOverrides = map[buttonKey]string{
	btnAgeConfirmYes:             "✅",
	btnAgeConfirmNo:              "🚫",
	btnCancel:                    "↩️",
	btnStartSearch:               "🔥",
	btnStartSearchAI:             "⚡️",
	btnSearchAIRepeat:            "🔁",
	btnSearchHistory:             "🕘",
	btnSearchHistoryBack:         "⬅️",
	btnSearchHistoryPrev:         "◀️",
	btnSearchHistoryNext:         "▶️",
	btnSearchRefresh:             "🔄",
	btnSearchAccuracyCancel:      "🚫",
	btnAdminUsersPrev:            "◀️",
	btnAdminUsersNext:            "▶️",
	btnAdminBackToMenu:           "⬅️",
	btnAdminAdAddButton:          "📢",
	btnAdminAdClearButtons:       "📢",
	btnAdminAdRemovePhoto:        "📢",
	btnAdminAdPreview:            "📢",
	btnAdminAdSend:               "📢",
	btnBackToSettings:            "⬅️",
	btnBackToProfile:             "⬅️",
	btnBackToPreviousProfile:     "◀️",
	btnProfileSetupBack:          "↩️",
	btnProfile:                   "👤",
	btnCreateProfile:             "📝",
	btnSettings:                  "⚙️",
	btnPreviewProfile:            "👁️",
	btnEditProfile:               "✏️",
	btnDeleteProfile:             "🗑️",
	btnDeleteConfirm:             "🗑️",
	btnProfileEditName:           "",
	btnProfileEditGender:         "",
	btnProfileEditBirthDate:      "",
	btnProfileEditCountry:        "",
	btnProfileEditCity:           "",
	btnProfileEditDescription:    "",
	btnProfileEditEmoji:          "",
	btnProfileEditPhotos:         "",
	btnProfileDeleteCancel:       "❌",
	btnEditCancel:                "❌",
	btnAlbumDone:                 "✅",
	btnTelegramName:              "📲",
	btnReport:                    "🚩",
	btnLikesInbox:                "❤️",
	btnLikesBack:                 "⬅️",
	btnLikesNotificationsEnable:  "🔔",
	btnLikesNotificationsDisable: "🔔",
	btnLikesNotificationsStatus:  "🔔",
	btnHideFromSearch:            "🙈",
	btnShowInSearch:              "👀",
	btnViewProfile:               "👁️",
	btnGenderMale:                "",
	btnGenderFemale:              "",
	btnGenderOther:               "",
	btnCountryRussia:             "🇷🇺",
	btnCountryKazakhstan:         "🇰🇿",
	btnCountryBelarus:            "🇧🇾",
	btnCountryUkraine:            "🇺🇦",
	btnLanguage:                  "🌐",
	btnLanguageEnglish:           "",
	btnLanguageRussian:           "",
	btnLanguageUkrainian:         "",
	btnAdvancedSearchEnable:      "⚙️",
	btnAdvancedSearchDisable:     "⚙️",
	btnAdvancedSearchStatus:      "⚙️",
	btnSearchMen:                 "",
	btnSearchWomen:               "",
	btnSearchOther:               "",
	btnSearchAny:                 "",
	btnAdminMenu:                 "🗂️",
	btnAdminUsers:                "👥",
	btnAdminHiddenUsers:          "🙈",
	btnAdminBan:                  "⛔",
	btnAdminUnban:                "✅",
	btnAdminShadowBan:            "🕵️",
	btnAdminShadowUnban:          "🕵️",
	btnAdminHideProfile:          "🙈",
	btnAdminShowProfile:          "👀",
	btnAdminViewProfile:          "👁️",
	btnAdminDeleteProfile:        "🗑️",
	btnAdminModerator:            "🛡️",
	btnAdminUnmoderator:          "🚫",
	btnAdminResetChoices:         "🔄",
	btnAdminResetStart:           "🧭",
	btnAdminPostAd:               "📣",
	btnAdminReports:              "📋",
	btnAdminClearReports:         "🧹",
	btnAdminBannedUsers:          "🚫",
	btnAdminShadowBannedUsers:    "👻",
	btnAdminModerators:           "🛡️",
	btnModeratorMenu:             "🛡️",
}

func buttonEmoji(key buttonKey) string {
	if emoji, ok := buttonEmojiOverrides[key]; ok {
		return emoji
	}
	return defaultButtonEmoji
}

func normalizeLanguageValue(lang domain.Language) domain.Language {
	switch lang {
	case domain.LanguageUkrainian:
		return domain.LanguageUkrainian
	case domain.LanguageRussian:
		return domain.LanguageRussian
	default:
		return domain.LanguageEnglish
	}
}
