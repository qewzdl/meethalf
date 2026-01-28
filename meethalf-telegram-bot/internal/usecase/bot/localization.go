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
	return formatTemplate(value, args...)
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

func (l localizer) adminStatusLabel(isBanned, isModerator, isHidden bool) string {
	parts := make([]string, 0, 3)
	if isBanned {
		parts = append(parts, l.label(labelStatusBanned))
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
	msgDefaultHelp                   messageKey = "default_help"
	msgUnknownCommand                messageKey = "unknown_command"
	msgAdminBadge                    messageKey = "admin_badge"
	msgModeratorBadge                messageKey = "moderator_badge"
	msgAdminMenu                     messageKey = "admin_menu"
	msgModeratorMenu                 messageKey = "moderator_menu"
	msgAdminAccessDenied             messageKey = "admin_access_denied"
	msgAdminModerationRestricted     messageKey = "admin_moderation_restricted"
	msgAdminUsersEmpty               messageKey = "admin_users_empty"
	msgAdminUsersLoadFailed          messageKey = "admin_users_load_failed"
	msgAdminUsersPage                messageKey = "admin_users_page"
	msgAdminUsersEmptyPage           messageKey = "admin_users_empty_page"
	msgAdminBannedUsersEmpty         messageKey = "admin_banned_users_empty"
	msgAdminBannedUsersLoadFailed    messageKey = "admin_banned_users_load_failed"
	msgAdminBannedUsersPage          messageKey = "admin_banned_users_page"
	msgAdminBannedUsersEmptyPage     messageKey = "admin_banned_users_empty_page"
	msgAdminModeratorsEmpty          messageKey = "admin_moderators_empty"
	msgAdminModeratorsLoadFailed     messageKey = "admin_moderators_load_failed"
	msgAdminModeratorsPage           messageKey = "admin_moderators_page"
	msgAdminModeratorsEmptyPage      messageKey = "admin_moderators_empty_page"
	msgAdminReportsEmpty             messageKey = "admin_reports_empty"
	msgAdminReportsLoadFailed        messageKey = "admin_reports_load_failed"
	msgAdminReportsPage              messageKey = "admin_reports_page"
	msgAdminReportsEmptyPage         messageKey = "admin_reports_empty_page"
	msgAdminBanUsage                 messageKey = "admin_ban_usage"
	msgAdminBanFailed                messageKey = "admin_ban_failed"
	msgAdminUserNotFound             messageKey = "admin_user_not_found"
	msgAdminBanSuccess               messageKey = "admin_ban_success"
	msgAdminUnbanUsage               messageKey = "admin_unban_usage"
	msgAdminUnbanFailed              messageKey = "admin_unban_failed"
	msgAdminUnbanSuccess             messageKey = "admin_unban_success"
	msgAdminActionFailed             messageKey = "admin_action_failed"
	msgAdminModeratorUsage           messageKey = "admin_moderator_usage"
	msgAdminModeratorFailed          messageKey = "admin_moderator_failed"
	msgAdminModeratorSuccess         messageKey = "admin_moderator_success"
	msgAdminUnmoderatorUsage         messageKey = "admin_unmoderator_usage"
	msgAdminUnmoderatorFailed        messageKey = "admin_unmoderator_failed"
	msgAdminUnmoderatorSuccess       messageKey = "admin_unmoderator_success"
	msgAdminResetChoicesUsage        messageKey = "admin_reset_choices_usage"
	msgAdminResetChoicesFailed       messageKey = "admin_reset_choices_failed"
	msgAdminResetChoicesSuccess      messageKey = "admin_reset_choices_success"
	msgAdminClearReportsUsage        messageKey = "admin_clear_reports_usage"
	msgAdminClearReportsFailed       messageKey = "admin_clear_reports_failed"
	msgAdminClearReportsSuccess      messageKey = "admin_clear_reports_success"
	msgUserBanned                    messageKey = "user_banned"
	msgProfileCreated                messageKey = "profile_created"
	msgProfileUpdated                messageKey = "profile_updated"
	msgLoadingStart                  messageKey = "loading_start"
	msgLoadingProfileView            messageKey = "loading_profile_view"
	msgLoadingProfilePreview         messageKey = "loading_profile_preview"
	msgLoadingEditName               messageKey = "loading_edit_name"
	msgLoadingEditGender             messageKey = "loading_edit_gender"
	msgLoadingEditBirthDate          messageKey = "loading_edit_birth_date"
	msgLoadingEditCountry            messageKey = "loading_edit_country"
	msgLoadingEditCity               messageKey = "loading_edit_city"
	msgLoadingEditDesc               messageKey = "loading_edit_desc"
	msgLoadingEditEmoji              messageKey = "loading_edit_emoji"
	msgLoadingEditPhotos             messageKey = "loading_edit_photos"
	msgLoadingProfileVisibility      messageKey = "loading_profile_visibility"
	msgLoadingSearchStart            messageKey = "loading_search_start"
	msgLoadingSearchNext             messageKey = "loading_search_next"
	msgLoadingSearchPrev             messageKey = "loading_search_prev"
	msgLoadingSearchHistory          messageKey = "loading_search_history"
	msgLoadingSearchHistoryProfile   messageKey = "loading_search_history_profile"
	msgLoadingSearchHistoryAction    messageKey = "loading_search_history_action"
	msgLoadingAdminUsers             messageKey = "loading_admin_users"
	msgLoadingAdminBan               messageKey = "loading_admin_ban"
	msgLoadingAdminUnban             messageKey = "loading_admin_unban"
	msgLoadingAdminModerator         messageKey = "loading_admin_moderator"
	msgLoadingAdminUnmoderator       messageKey = "loading_admin_unmoderator"
	msgLoadingAdminResetChoices      messageKey = "loading_admin_reset_choices"
	msgLoadingAdminClearReports      messageKey = "loading_admin_clear_reports"
	msgLoadingAdminBannedUsers       messageKey = "loading_admin_banned_users"
	msgLoadingAdminModerators        messageKey = "loading_admin_moderators"
	msgLoadingAdminReports           messageKey = "loading_admin_reports"
	msgCreatingProfile               messageKey = "creating_profile"
	msgUpdatingProfile               messageKey = "updating_profile"
	msgDeletingProfile               messageKey = "deleting_profile"
	msgProfileDeleteConfirm           messageKey = "profile_delete_confirm"
	msgProfileDeleteCanceled          messageKey = "profile_delete_canceled"
	msgProfileDeleteExpired           messageKey = "profile_delete_expired"
	msgProfileSetupCanceled           messageKey = "profile_setup_canceled"
	msgProfileEditCanceled            messageKey = "profile_edit_canceled"
	msgActionCanceled                 messageKey = "action_canceled"
	msgProfileHidden                  messageKey = "profile_hidden"
	msgProfileVisible                 messageKey = "profile_visible"
	msgProfileVisibilityUpdateFailed  messageKey = "profile_visibility_update_failed"
	msgSearchGenderPrompt             messageKey = "search_gender_prompt"
	msgSearchAccuracyPrompt           messageKey = "search_accuracy_prompt"
	msgSearchNoCandidates             messageKey = "search_no_candidates"
	msgSearchNoPrevious               messageKey = "search_no_previous"
	msgSearchHistoryEmpty             messageKey = "search_history_empty"
	msgSearchStartRequired            messageKey = "search_start_required"
	msgSearchProfileMissing           messageKey = "search_profile_missing"
	msgSearchUnavailable              messageKey = "search_unavailable"
	msgSearchActionFailed             messageKey = "search_action_failed"
	msgSearchHistoryPage              messageKey = "search_history_page"
	msgSearchHistoryEmptyPage         messageKey = "search_history_empty_page"
	msgSearchHistoryAction            messageKey = "search_history_action"
	msgMatchActions                   messageKey = "match_actions"
	msgMatchActionSaved               messageKey = "match_action_saved"
	msgMatchProfileNotFound           messageKey = "match_profile_not_found"
	msgProfileViewRequiresProfile     messageKey = "profile_view_requires_profile"
	msgMatchSuccess                   messageKey = "match_success"
	msgMatchNickname                  messageKey = "match_nickname"
	msgNamePromptNoTelegram           messageKey = "name_prompt_no_telegram"
	msgNamePromptWithTelegram         messageKey = "name_prompt_with_telegram"
	msgBotCheckPrompt                 messageKey = "bot_check_prompt"
	msgBirthDatePrompt                messageKey = "birth_date_prompt"
	msgGenderPrompt                   messageKey = "gender_prompt"
	msgCountryPrompt                  messageKey = "country_prompt"
	msgCityPrompt                     messageKey = "city_prompt"
	msgDescriptionPrompt              messageKey = "description_prompt"
	msgEmojiPrompt                    messageKey = "emoji_prompt"
	msgPhotosPrompt                   messageKey = "photos_prompt"
	msgProfileEditMenu                messageKey = "profile_edit_menu"
	msgProfileSettings                messageKey = "profile_settings"
	msgProfileSettingsWithVisibility  messageKey = "profile_settings_with_visibility"
	msgProfileActions                 messageKey = "profile_actions"
	msgProfilePreviewActions          messageKey = "profile_preview_actions"
	msgEditNamePrompt                 messageKey = "edit_name_prompt"
	msgEditBirthDatePrompt            messageKey = "edit_birth_date_prompt"
	msgEditGenderPrompt               messageKey = "edit_gender_prompt"
	msgEditCountryPrompt              messageKey = "edit_country_prompt"
	msgEditCityPrompt                 messageKey = "edit_city_prompt"
	msgEditDescriptionPrompt          messageKey = "edit_description_prompt"
	msgEditEmojiPrompt                messageKey = "edit_emoji_prompt"
	msgEditPhotosPrompt               messageKey = "edit_photos_prompt"
	msgEditDefaultPrompt              messageKey = "edit_default_prompt"
	msgProfileDetailsHeader           messageKey = "profile_details_header"
	msgProfilePreviewFallbackName     messageKey = "profile_preview_fallback_name"
	msgProfileEditHeader              messageKey = "profile_edit_header"
	msgProfileSetupHeader             messageKey = "profile_setup_header"
	msgProfileSetupEstimate           messageKey = "profile_setup_estimate"
	msgStartGreeting                  messageKey = "start_greeting"
	msgStartGreetingNamed             messageKey = "start_greeting_named"
	msgProfileNotFoundCreateButton    messageKey = "profile_not_found_create_button"
	msgProfileNotFoundButtonBelow     messageKey = "profile_not_found_button_below"
	msgProfileNotFoundUseProfile      messageKey = "profile_not_found_use_profile"
	msgProfileDeleted                 messageKey = "profile_deleted"
	msgProfileDeletedDraftWarning     messageKey = "profile_deleted_draft_warning"
	msgProfileDeletedDraftWarningOnly messageKey = "profile_deleted_draft_warning_only"
	msgProfileSetupUnavailable        messageKey = "profile_setup_unavailable"
	msgProfileSetupUnavailableChat    messageKey = "profile_setup_unavailable_chat"
	msgProfileSetupStartFailed        messageKey = "profile_setup_start_failed"
	msgProfileSetupLoadFailed         messageKey = "profile_setup_load_failed"
	msgProfileSetupNotActive          messageKey = "profile_setup_not_active"
	msgProfileSetupSaveFailed         messageKey = "profile_setup_save_failed"
	msgProfileEditUnavailable         messageKey = "profile_edit_unavailable"
	msgProfileEditUnavailableChat     messageKey = "profile_edit_unavailable_chat"
	msgProfileServiceUnavailable      messageKey = "profile_service_unavailable"
	msgProfileLoadFailed              messageKey = "profile_load_failed"
	msgProfileDeleteUnavailable       messageKey = "profile_delete_unavailable"
	msgProfileDeleteUnavailableChat   messageKey = "profile_delete_unavailable_chat"
	msgProfileDeletePrepareFailed     messageKey = "profile_delete_prepare_failed"
	msgProfileDeleteFailed            messageKey = "profile_delete_failed"
	msgProfileDeleteCancelFailed      messageKey = "profile_delete_cancel_failed"
	msgProfileEditStartFailed         messageKey = "profile_edit_start_failed"
	msgProfileEditSaveFailed          messageKey = "profile_edit_save_failed"
	msgProfileSaveFailed              messageKey = "profile_save_failed"
	msgBotCheckTooManyAttempts        messageKey = "bot_check_too_many_attempts"
	msgBotCheckIncorrect              messageKey = "bot_check_incorrect"
	msgNamePromptEmpty                messageKey = "name_prompt_empty"
	msgNamePromptEmptyCreate          messageKey = "name_prompt_empty_create"
	msgNamePromptTelegramMissing      messageKey = "name_prompt_telegram_missing"
	msgNameTooLong                    messageKey = "name_too_long"
	msgGenderInvalid                  messageKey = "gender_invalid"
	msgBirthDateInvalid               messageKey = "birth_date_invalid"
	msgAgeInvalid                     messageKey = "age_invalid"
	msgCountryInvalid                 messageKey = "country_invalid"
	msgCityInvalid                    messageKey = "city_invalid"
	msgDescriptionEmpty               messageKey = "description_empty"
	msgDescriptionTooLong             messageKey = "description_too_long"
	msgEmojiInvalid                   messageKey = "emoji_invalid"
	msgPhotosNotEnough                messageKey = "photos_not_enough"
	msgPhotosPromptRepeat             messageKey = "photos_prompt_repeat"
	msgPhotosLimitReached             messageKey = "photos_limit_reached"
	msgPhotosProgress                 messageKey = "photos_progress"
	msgProfileViewUnavailable         messageKey = "profile_view_unavailable"
	msgProfileViewUnavailableChat     messageKey = "profile_view_unavailable_chat"
	msgProfileDeleteNoteDraft         messageKey = "profile_delete_note_draft"
	msgAdminUserListNoName            messageKey = "admin_user_list_no_name"
	msgAdminUserListNA                messageKey = "admin_user_list_na"
	msgAdminUserListLine              messageKey = "admin_user_list_line"
	msgAdminReportedUserListLine      messageKey = "admin_reported_user_list_line"
	msgSearchHistoryLine              messageKey = "search_history_line"
	msgSearchHistoryLineShort         messageKey = "search_history_line_short"
	msgSearchHistoryUserFallback      messageKey = "search_history_user_fallback"
	msgSearchHistoryLabel             messageKey = "search_history_label"
	msgAgeShort                       messageKey = "age_short"
	msgProfileDetailsLine             messageKey = "profile_details_line"
	msgProfileDetailsPhotos           messageKey = "profile_details_photos"
	msgProfileDetailsCreated          messageKey = "profile_details_created"
	msgProfileDetailsUpdated          messageKey = "profile_details_updated"
	msgProfileDetailsDescriptionLabel messageKey = "profile_details_description_label"
	msgLikeNotification               messageKey = "like_notification"
	msgLikeNotificationFrom           messageKey = "like_notification_from"
	msgLanguagePrompt                 messageKey = "language_prompt"
	msgLanguageUpdated                messageKey = "language_updated"
	msgLanguageUnsupported            messageKey = "language_unsupported"
	msgProfileSettingsLanguageHint    messageKey = "profile_settings_language_hint"
)

const (
	btnCancel                 buttonKey = "cancel"
	btnProfileSetupBack       buttonKey = "profile_setup_back"
	btnProfileDeleteCancel    buttonKey = "profile_delete_cancel"
	btnBackToProfile          buttonKey = "back_to_profile"
	btnEditCancel             buttonKey = "edit_cancel"
	btnSearchAccuracyCancel   buttonKey = "search_accuracy_cancel"
	btnSearchRefresh          buttonKey = "search_refresh"
	btnSearchHistory          buttonKey = "search_history"
	btnSearchHistoryPrev      buttonKey = "search_history_prev"
	btnSearchHistoryNext      buttonKey = "search_history_next"
	btnSearchHistoryBack      buttonKey = "search_history_back"
	btnAdminMenu              buttonKey = "admin_menu"
	btnModeratorMenu          buttonKey = "moderator_menu"
	btnAdminUsers             buttonKey = "admin_users"
	btnAdminBan               buttonKey = "admin_ban"
	btnAdminUnban             buttonKey = "admin_unban"
	btnAdminModerator         buttonKey = "admin_moderator"
	btnAdminUnmoderator       buttonKey = "admin_unmoderator"
	btnAdminResetChoices      buttonKey = "admin_reset_choices"
	btnAdminBannedUsers       buttonKey = "admin_banned_users"
	btnAdminModerators        buttonKey = "admin_moderators"
	btnAdminReports           buttonKey = "admin_reports"
	btnAdminClearReports      buttonKey = "admin_clear_reports"
	btnAdminUsersPrev         buttonKey = "admin_users_prev"
	btnAdminUsersNext         buttonKey = "admin_users_next"
	btnAdminBackToMenu        buttonKey = "admin_back_to_menu"
	btnStartSearch            buttonKey = "start_search"
	btnProfile                buttonKey = "profile"
	btnCreateProfile          buttonKey = "create_profile"
	btnSettings               buttonKey = "settings"
	btnPreviewProfile         buttonKey = "preview_profile"
	btnEditProfile            buttonKey = "edit_profile"
	btnDeleteProfile          buttonKey = "delete_profile"
	btnDeleteConfirm          buttonKey = "delete_confirm"
	btnProfileEditName        buttonKey = "profile_edit_name"
	btnProfileEditGender      buttonKey = "profile_edit_gender"
	btnProfileEditBirthDate   buttonKey = "profile_edit_birth_date"
	btnProfileEditCountry     buttonKey = "profile_edit_country"
	btnProfileEditCity        buttonKey = "profile_edit_city"
	btnProfileEditDescription buttonKey = "profile_edit_description"
	btnProfileEditEmoji       buttonKey = "profile_edit_emoji"
	btnProfileEditPhotos      buttonKey = "profile_edit_photos"
	btnTelegramName           buttonKey = "telegram_name"
	btnAlbumDone              buttonKey = "album_done"
	btnGenderMale             buttonKey = "gender_male"
	btnGenderFemale           buttonKey = "gender_female"
	btnGenderOther            buttonKey = "gender_other"
	btnCountryRussia          buttonKey = "country_russia"
	btnCountryKazakhstan      buttonKey = "country_kazakhstan"
	btnCountryBelarus         buttonKey = "country_belarus"
	btnSearchMen              buttonKey = "search_men"
	btnSearchWomen            buttonKey = "search_women"
	btnSearchOther            buttonKey = "search_other"
	btnSearchAny              buttonKey = "search_any"
	btnReport                 buttonKey = "report"
	btnBackToPreviousProfile  buttonKey = "back_to_previous_profile"
	btnViewProfile            buttonKey = "view_profile"
	btnHideFromSearch         buttonKey = "hide_from_search"
	btnShowInSearch           buttonKey = "show_in_search"
	btnLanguage               buttonKey = "language"
	btnBackToSettings         buttonKey = "back_to_settings"
	btnLanguageEnglish        buttonKey = "language_english"
	btnLanguageRussian        buttonKey = "language_russian"
)

const (
	labelNotSet            labelKey = "not_set"
	labelUnknown           labelKey = "unknown"
	labelProfile           labelKey = "profile"
	labelStepVerification  labelKey = "step_verification"
	labelStepName          labelKey = "step_name"
	labelStepGender        labelKey = "step_gender"
	labelStepBirthDate     labelKey = "step_birth_date"
	labelStepCountry       labelKey = "step_country"
	labelStepCity          labelKey = "step_city"
	labelStepDescription   labelKey = "step_description"
	labelStepEmoji         labelKey = "step_emoji"
	labelStepPhotos        labelKey = "step_photos"
	labelGenderMale        labelKey = "gender_male"
	labelGenderFemale      labelKey = "gender_female"
	labelGenderOther       labelKey = "gender_other"
	labelCountryRussia     labelKey = "country_russia"
	labelCountryKazakhstan labelKey = "country_kazakhstan"
	labelCountryBelarus    labelKey = "country_belarus"
	labelActionLike        labelKey = "action_like"
	labelActionDislike     labelKey = "action_dislike"
	labelActionReport      labelKey = "action_report"
	labelActionNone        labelKey = "action_none"
	labelStatusBanned      labelKey = "status_banned"
	labelStatusModerator   labelKey = "status_moderator"
	labelStatusHidden      labelKey = "status_hidden"
	labelStatusVisible     labelKey = "status_visible"
	labelName              labelKey = "label_name"
	labelEmoji             labelKey = "label_emoji"
	labelGender            labelKey = "label_gender"
	labelAge               labelKey = "label_age"
	labelCountry           labelKey = "label_country"
	labelCity              labelKey = "label_city"
	labelSearchVisibility  labelKey = "label_search_visibility"
	labelDescription       labelKey = "label_description"
	labelPhotos            labelKey = "label_photos"
	labelCreated           labelKey = "label_created"
	labelUpdated           labelKey = "label_updated"
	labelLanguageEnglish   labelKey = "language_english"
	labelLanguageRussian   labelKey = "language_russian"
	labelThisUser          labelKey = "this_user"
)

var messageCatalog = map[domain.Language]map[messageKey]string{
	domain.LanguageEnglish: {
		msgDefaultHelp:                "Use the buttons below to start searching, view or create your profile, or open settings.",
		msgUnknownCommand:             "Unknown command.",
		msgAdminBadge:                 "Admin access enabled.",
		msgModeratorBadge:             "Moderator access enabled.",
		msgAdminMenu:                  "Admin panel. Choose an action.",
		msgModeratorMenu:              "Moderator panel. Choose an action.",
		msgAdminAccessDenied:          "Admin access required.",
		msgAdminModerationRestricted:  "Moderators can only manage bans for regular users.",
		msgAdminUsersEmpty:            "No users found.",
		msgAdminUsersLoadFailed:       "Unable to load users. Please try again later.",
		msgAdminUsersPage:             "Users: %d total. Showing %d-%d.",
		msgAdminUsersEmptyPage:        "Users: %d total. No users found on this page.",
		msgAdminBannedUsersEmpty:      "No banned users found.",
		msgAdminBannedUsersLoadFailed: "Unable to load banned users. Please try again later.",
		msgAdminBannedUsersPage:       "Banned users: %d total. Showing %d-%d.",
		msgAdminBannedUsersEmptyPage:  "Banned users: %d total. No users found on this page.",
		msgAdminModeratorsEmpty:       "No moderators found.",
		msgAdminModeratorsLoadFailed:  "Unable to load moderators. Please try again later.",
		msgAdminModeratorsPage:        "Moderators: %d total. Showing %d-%d.",
		msgAdminModeratorsEmptyPage:   "Moderators: %d total. No users found on this page.",
		msgAdminReportsEmpty:          "No reported users found.",
		msgAdminReportsLoadFailed:     "Unable to load reported users. Please try again later.",
		msgAdminReportsPage:           "Reported users: %d total. Showing %d-%d.",
		msgAdminReportsEmptyPage:      "Reported users: %d total. No users found on this page.",
		msgAdminBanUsage:              "Send the user ID or @username to ban.",
		msgAdminBanFailed:             "Unable to ban user. Please try again later.",
		msgAdminUserNotFound:          "User not found.",
		msgAdminBanSuccess:            "User %s has been banned.",
		msgAdminUnbanUsage:            "Send the user ID or @username to unban.",
		msgAdminUnbanFailed:           "Unable to unban user. Please try again later.",
		msgAdminUnbanSuccess:          "User %s has been unbanned.",
		msgAdminActionFailed:          "Unable to complete the admin action. Please try again later.",
		msgAdminModeratorUsage:        "Send the user ID or @username to assign the moderator role.",
		msgAdminModeratorFailed:       "Unable to assign the moderator role. Please try again later.",
		msgAdminModeratorSuccess:      "User %s is now a moderator.",
		msgAdminUnmoderatorUsage:      "Send the user ID or @username to remove the moderator role.",
		msgAdminUnmoderatorFailed:     "Unable to remove the moderator role. Please try again later.",
		msgAdminUnmoderatorSuccess:    "User %s is no longer a moderator.",
		msgAdminResetChoicesUsage:     "Send the user ID or @username to reset match choices.",
		msgAdminResetChoicesFailed:    "Unable to reset match choices. Please try again later.",
		msgAdminResetChoicesSuccess:   "Match choices were reset for %s.",
		msgAdminClearReportsUsage:     "Send the user ID or @username to clear reports.",
		msgAdminClearReportsFailed:    "Unable to clear reports. Please try again later.",
		msgAdminClearReportsSuccess:   "Reports were cleared for %s.",
		msgUserBanned:                 "Your account is banned. Contact support.",
		msgProfileCreated:             "Profile created.",
		msgProfileUpdated:             "Profile updated.",
		msgLoadingStart:               "Checking your profile...",
		msgLoadingProfileView:         "Loading your profile...",
		msgLoadingProfilePreview:      "Loading profile preview...",
		msgLoadingEditName:            "Preparing name update...",
		msgLoadingEditGender:          "Preparing gender update...",
		msgLoadingEditBirthDate:       "Preparing birth date update...",
		msgLoadingEditCountry:         "Preparing country update...",
		msgLoadingEditCity:            "Preparing city update...",
		msgLoadingEditDesc:            "Preparing description update...",
		msgLoadingEditEmoji:           "Preparing emoji update...",
		msgLoadingEditPhotos:          "Preparing photo update...",
		msgLoadingProfileVisibility:   "Updating search visibility...",
		msgLoadingSearchStart:         "Finding profiles...",
		msgLoadingSearchNext:          "Searching for the next profile...",
		msgLoadingSearchPrev:          "Opening the previous profile...",
		msgLoadingSearchHistory:       "Loading search history...",
		msgLoadingSearchHistoryProfile:"Opening history profile...",
		msgLoadingSearchHistoryAction: "Updating decision...",
		msgLoadingAdminUsers:          "Loading users...",
		msgLoadingAdminBan:            "Banning user...",
		msgLoadingAdminUnban:          "Unbanning user...",
		msgLoadingAdminModerator:      "Assigning moderator role...",
		msgLoadingAdminUnmoderator:    "Removing moderator role...",
		msgLoadingAdminResetChoices:   "Resetting match choices...",
		msgLoadingAdminClearReports:   "Clearing reports...",
		msgLoadingAdminBannedUsers:    "Loading banned users...",
		msgLoadingAdminModerators:     "Loading moderators...",
		msgLoadingAdminReports:        "Loading reported users...",
		msgCreatingProfile:            "Creating your profile...",
		msgUpdatingProfile:            "Updating your profile...",
		msgDeletingProfile:            "Deleting your profile...",
		msgProfileDeleteConfirm:       "Are you sure you want to delete your profile? This action cannot be undone.",
		msgProfileDeleteCanceled:      "Profile deletion canceled.",
		msgProfileDeleteExpired:       "Profile deletion confirmation expired. Use Settings to start again.",
		msgProfileSetupCanceled:       "Profile setup canceled.",
		msgProfileEditCanceled:        "Profile edit canceled.",
		msgActionCanceled:             "Action canceled.",
		msgProfileHidden:              "Profile is now hidden from search.",
		msgProfileVisible:             "Profile is now visible in search.",
		msgProfileVisibilityUpdateFailed: "Unable to update search visibility. Please try again later.",
		msgSearchGenderPrompt:            "Select the gender to search for.",
		msgSearchAccuracyPrompt:          "Match accuracy (0-4): 0 wide/random → 4 strict/precise.",
		msgSearchNoCandidates:            "No matching profiles yet. Try again later.",
		msgSearchNoPrevious:              "No previous profile.",
		msgSearchHistoryEmpty:            "History is empty.",
		msgSearchStartRequired:           "Press \"Start search\" first.",
		msgSearchProfileMissing:          "Create a profile first to start searching and view profiles.",
		msgSearchUnavailable:             "Search is currently unavailable.",
		msgSearchActionFailed:            "Unable to process the action. Try again later.",
		msgSearchHistoryPage:             "History: %d total. Showing %d-%d.",
		msgSearchHistoryEmptyPage:        "History: %d total. No profiles found on this page.",
		msgSearchHistoryAction:           "Current decision: %s.\nChoose an action.",
		msgMatchActions:                  "Choose an action.",
		msgMatchActionSaved:              "Decision saved.",
		msgMatchProfileNotFound:          "Profile not found.",
		msgProfileViewRequiresProfile:    "Create a profile first to view other profiles.",
		msgMatchSuccess:                  "It's a match! You and %s liked each other.",
		msgMatchNickname:                 "Nickname: %s",
		msgNamePromptNoTelegram:          "Your Telegram profile has no name set. Please type the name you want to use.",
		msgNamePromptWithTelegram:        "Current Telegram name: %s\nUse the button below to use it, or send the name you prefer.",
		msgBotCheckPrompt:                "To protect from bots, solve: %s\nChoose the correct answer below.",
		msgBirthDatePrompt:               "Enter your birth date in %s format (for example, 1990-04-23).",
		msgGenderPrompt:                  "Select your gender using the buttons below.",
		msgCountryPrompt:                 "Select your country using the buttons below.",
		msgCityPrompt:                    "Select your city using the buttons below.",
		msgDescriptionPrompt:             "Write a short description about yourself.",
		msgEmojiPrompt:                   "Select the emoji that describes you using the buttons below.",
		msgPhotosPrompt:                  "Send 1-4 photos for your album. Use the Done button when finished.",
		msgProfileEditMenu:               "Choose what you want to update in your profile.",
		msgProfileSettings:               "Profile settings. Use the buttons below to manage search visibility or delete your profile.",
		msgProfileSettingsWithVisibility: "Profile settings.\nSearch visibility: %s.\nUse the buttons below to manage search visibility or delete your profile.",
		msgProfileActions:                "Use the buttons below to preview or edit your profile.",
		msgProfilePreviewActions:         "Use the buttons below to return to your profile or edit it.",
		msgEditNamePrompt:                "Enter the new name.",
		msgEditBirthDatePrompt:           "Enter the new birth date in %s format (for example, 1990-04-23).",
		msgEditGenderPrompt:              "Select the new gender using the buttons below.",
		msgEditCountryPrompt:             "Select the new country using the buttons below.",
		msgEditCityPrompt:                "Select the new city using the buttons below.",
		msgEditDescriptionPrompt:         "Write the new description.",
		msgEditEmojiPrompt:               "Choose the new emoji using the buttons below.",
		msgEditPhotosPrompt:              "Send 1-4 photos to replace your album. Use the Done button when finished.",
		msgEditDefaultPrompt:             "Enter the updated value.",
		msgProfileDetailsHeader:          "Your profile:",
		msgProfilePreviewFallbackName:    "Profile",
		msgProfileEditHeader:             "Profile edit: %s",
		msgProfileSetupHeader:            "Profile setup (step %d/%d): %s\n%s",
		msgProfileSetupEstimate:          "Estimated total time: ~%d min",
		msgStartGreeting:                 "Welcome to Meethalf bot.",
		msgStartGreetingNamed:            "Welcome to Meethalf bot, %s.",
		msgProfileNotFoundCreateButton:   "Profile not found. Use the Create Profile button to create it.",
		msgProfileNotFoundButtonBelow:    "Profile not found. Use the button below to create it.",
		msgProfileNotFoundUseProfile:     "Profile not found. Use /profile to create it.",
		msgProfileDeleted:                "Profile deleted. Use the Create Profile button to create a new one.",
		msgProfileDeletedDraftWarning:    "Profile deleted. Note: could not clear the profile draft.",
		msgProfileDeletedDraftWarningOnly:"Note: could not clear the profile draft.",
		msgProfileSetupUnavailable:       "Profile setup is not available right now.",
		msgProfileSetupUnavailableChat:   "Profile setup is not available for this chat.",
		msgProfileSetupStartFailed:       "Failed to start profile setup. Please try again later.",
		msgProfileSetupLoadFailed:        "Unable to load profile setup. Please try again later.",
		msgProfileSetupNotActive:         "Profile setup is not active. Use /profile to start.",
		msgProfileSetupSaveFailed:        "Failed to save profile setup. Please try again later.",
		msgProfileEditUnavailable:        "Profile edit is not available right now.",
		msgProfileEditUnavailableChat:    "Profile edit is not available for this chat.",
		msgProfileServiceUnavailable:     "Profile service is not available right now.",
		msgProfileLoadFailed:             "Unable to load profile. Please try again later.",
		msgProfileDeleteUnavailable:      "Profile delete is not available right now.",
		msgProfileDeleteUnavailableChat:  "Profile delete is not available for this chat.",
		msgProfileDeletePrepareFailed:    "Unable to prepare profile deletion. Please try again later.",
		msgProfileDeleteFailed:           "Unable to delete profile. Please try again later.",
		msgProfileDeleteCancelFailed:     "Unable to cancel profile deletion. Please try again later.",
		msgProfileEditStartFailed:        "Failed to start profile edit. Please try again later.",
		msgProfileEditSaveFailed:         "Failed to save profile edit. Please try again later.",
		msgProfileSaveFailed:             "Failed to save profile. Please try again later.",
		msgBotCheckTooManyAttempts:       "Too many attempts. Let's try a new check.",
		msgBotCheckIncorrect:             "Incorrect answer. Try again.",
		msgNamePromptEmpty:               "Please enter a name.",
		msgNamePromptEmptyCreate:         "Please enter a name or use the button below to use your Telegram name.",
		msgNamePromptTelegramMissing:     "Your Telegram profile has no name set. Please type the name you want to use.",
		msgNameTooLong:                   "Name is too long (max %d characters).",
		msgGenderInvalid:                 "Gender must be one of: male, female, or other.",
		msgBirthDateInvalid:              "Birth date must be in %s format (for example, 1990-04-23). Try again.",
		msgAgeInvalid:                    "Age must be between %d and %d years. Please check the birth date.",
		msgCountryInvalid:                "Country must be one of: Russia, Kazakhstan, or Belarus.",
		msgCityInvalid:                   "City must be selected from the list for your country.",
		msgDescriptionEmpty:              "Description cannot be empty. Please try again.",
		msgDescriptionTooLong:            "Description is too long (max %d characters).",
		msgEmojiInvalid:                  "Emoji must be selected from the list.",
		msgPhotosNotEnough:               "Please send at least %d photo before finishing.",
		msgPhotosPromptRepeat:            "Send 1-4 photos for your album. Use the Done button when finished.",
		msgPhotosLimitReached:            "You reached the limit of %d photos. Use the Done button to finish.",
		msgPhotosProgress:                "Photos in album: %d/%d. Send more or use the Done button.",
		msgProfileViewUnavailable:        "Profile service is not available right now.",
		msgProfileViewUnavailableChat:    "Profile is not available for this chat.",
		msgProfileDeleteNoteDraft:        "Profile deleted. Note: could not clear the profile draft.",
		msgAdminUserListNoName:           "No name",
		msgAdminUserListNA:               "n/a",
		msgAdminUserListLine:             "%d. %s (ID: %d, username: %s, %s)",
		msgAdminReportedUserListLine:     "%d. %s (ID: %d, username: %s, reports: %d, %s)",
		msgSearchHistoryLine:             "%d. %s - %s",
		msgSearchHistoryLineShort:        "%d. %s",
		msgSearchHistoryUserFallback:     "User %d",
		msgSearchHistoryLabel:            "%s (%s)",
		msgAgeShort:                      "%d y.o.",
		msgProfileDetailsLine:            "%s: %s",
		msgProfileDetailsPhotos:          "Photos: %d",
		msgProfileDetailsCreated:         "Created: %s",
		msgProfileDetailsUpdated:         "Updated: %s",
		msgProfileDetailsDescriptionLabel:"Description: \n%s",
		msgLikeNotification:              "You received a ❤️. View the profile?",
		msgLikeNotificationFrom:          "You received a ❤️ from %s. View the profile?",
		msgLanguagePrompt:                "Select your language.",
		msgLanguageUpdated:               "Language updated.",
		msgLanguageUnsupported:           "Unsupported language. Please choose one of the available options.",
		msgProfileSettingsLanguageHint:   "Profile settings. Use the buttons below to manage search visibility, update language, or delete your profile.",
	},
	domain.LanguageRussian: {
		msgDefaultHelp:                "Используйте кнопки ниже, чтобы начать поиск, посмотреть или создать профиль, либо открыть настройки.",
		msgUnknownCommand:             "Неизвестная команда.",
		msgAdminBadge:                 "Доступ администратора включен.",
		msgModeratorBadge:             "Доступ модератора включен.",
		msgAdminMenu:                  "Админ-панель. Выберите действие.",
		msgModeratorMenu:              "Панель модератора. Выберите действие.",
		msgAdminAccessDenied:          "Требуется доступ администратора.",
		msgAdminModerationRestricted:  "Модераторы могут управлять банами только обычных пользователей.",
		msgAdminUsersEmpty:            "Пользователи не найдены.",
		msgAdminUsersLoadFailed:       "Не удалось загрузить пользователей. Попробуйте позже.",
		msgAdminUsersPage:             "Пользователи: всего %d. Показано %d-%d.",
		msgAdminUsersEmptyPage:        "Пользователи: всего %d. На этой странице никого нет.",
		msgAdminBannedUsersEmpty:      "Забаненные пользователи не найдены.",
		msgAdminBannedUsersLoadFailed: "Не удалось загрузить забаненных пользователей. Попробуйте позже.",
		msgAdminBannedUsersPage:       "Забаненные пользователи: всего %d. Показано %d-%d.",
		msgAdminBannedUsersEmptyPage:  "Забаненные пользователи: всего %d. На этой странице никого нет.",
		msgAdminModeratorsEmpty:       "Модераторы не найдены.",
		msgAdminModeratorsLoadFailed:  "Не удалось загрузить модераторов. Попробуйте позже.",
		msgAdminModeratorsPage:        "Модераторы: всего %d. Показано %d-%d.",
		msgAdminModeratorsEmptyPage:   "Модераторы: всего %d. На этой странице никого нет.",
		msgAdminReportsEmpty:          "Пользователи с жалобами не найдены.",
		msgAdminReportsLoadFailed:     "Не удалось загрузить жалобы. Попробуйте позже.",
		msgAdminReportsPage:           "Пользователи с жалобами: всего %d. Показано %d-%d.",
		msgAdminReportsEmptyPage:      "Пользователи с жалобами: всего %d. На этой странице никого нет.",
		msgAdminBanUsage:              "Отправьте ID пользователя или @username для бана.",
		msgAdminBanFailed:             "Не удалось забанить пользователя. Попробуйте позже.",
		msgAdminUserNotFound:          "Пользователь не найден.",
		msgAdminBanSuccess:            "Пользователь %s забанен.",
		msgAdminUnbanUsage:            "Отправьте ID пользователя или @username для разбана.",
		msgAdminUnbanFailed:           "Не удалось разбанить пользователя. Попробуйте позже.",
		msgAdminUnbanSuccess:          "Пользователь %s разбанен.",
		msgAdminActionFailed:          "Не удалось выполнить действие. Попробуйте позже.",
		msgAdminModeratorUsage:        "Отправьте ID пользователя или @username, чтобы назначить модератора.",
		msgAdminModeratorFailed:       "Не удалось назначить модератора. Попробуйте позже.",
		msgAdminModeratorSuccess:      "Пользователь %s теперь модератор.",
		msgAdminUnmoderatorUsage:      "Отправьте ID пользователя или @username, чтобы снять модератора.",
		msgAdminUnmoderatorFailed:     "Не удалось снять модератора. Попробуйте позже.",
		msgAdminUnmoderatorSuccess:    "Пользователь %s больше не модератор.",
		msgAdminResetChoicesUsage:     "Отправьте ID пользователя или @username, чтобы сбросить решения.",
		msgAdminResetChoicesFailed:    "Не удалось сбросить решения. Попробуйте позже.",
		msgAdminResetChoicesSuccess:   "Решения по совпадениям сброшены для %s.",
		msgAdminClearReportsUsage:     "Отправьте ID пользователя или @username, чтобы очистить жалобы.",
		msgAdminClearReportsFailed:    "Не удалось очистить жалобы. Попробуйте позже.",
		msgAdminClearReportsSuccess:   "Жалобы очищены для %s.",
		msgUserBanned:                 "Ваш аккаунт заблокирован. Свяжитесь с поддержкой.",
		msgProfileCreated:             "Профиль создан.",
		msgProfileUpdated:             "Профиль обновлен.",
		msgLoadingStart:               "Проверяем ваш профиль...",
		msgLoadingProfileView:         "Загружаем ваш профиль...",
		msgLoadingProfilePreview:      "Загружаем предпросмотр профиля...",
		msgLoadingEditName:            "Готовим обновление имени...",
		msgLoadingEditGender:          "Готовим обновление пола...",
		msgLoadingEditBirthDate:       "Готовим обновление даты рождения...",
		msgLoadingEditCountry:         "Готовим обновление страны...",
		msgLoadingEditCity:            "Готовим обновление города...",
		msgLoadingEditDesc:            "Готовим обновление описания...",
		msgLoadingEditEmoji:           "Готовим обновление эмодзи...",
		msgLoadingEditPhotos:          "Готовим обновление фото...",
		msgLoadingProfileVisibility:   "Обновляем видимость в поиске...",
		msgLoadingSearchStart:         "Ищем профили...",
		msgLoadingSearchNext:          "Ищем следующий профиль...",
		msgLoadingSearchPrev:          "Открываем предыдущий профиль...",
		msgLoadingSearchHistory:       "Загружаем историю поиска...",
		msgLoadingSearchHistoryProfile:"Открываем профиль из истории...",
		msgLoadingSearchHistoryAction: "Обновляем решение...",
		msgLoadingAdminUsers:          "Загружаем пользователей...",
		msgLoadingAdminBan:            "Баним пользователя...",
		msgLoadingAdminUnban:          "Разбаниваем пользователя...",
		msgLoadingAdminModerator:      "Назначаем модератора...",
		msgLoadingAdminUnmoderator:    "Снимаем модератора...",
		msgLoadingAdminResetChoices:   "Сбрасываем решения...",
		msgLoadingAdminClearReports:   "Очищаем жалобы...",
		msgLoadingAdminBannedUsers:    "Загружаем забаненных пользователей...",
		msgLoadingAdminModerators:     "Загружаем модераторов...",
		msgLoadingAdminReports:        "Загружаем жалобы...",
		msgCreatingProfile:            "Создаем ваш профиль...",
		msgUpdatingProfile:            "Обновляем ваш профиль...",
		msgDeletingProfile:            "Удаляем ваш профиль...",
		msgProfileDeleteConfirm:       "Вы уверены, что хотите удалить профиль? Это действие нельзя отменить.",
		msgProfileDeleteCanceled:      "Удаление профиля отменено.",
		msgProfileDeleteExpired:       "Подтверждение удаления профиля истекло. Используйте «Настройки», чтобы начать заново.",
		msgProfileSetupCanceled:       "Создание профиля отменено.",
		msgProfileEditCanceled:        "Редактирование профиля отменено.",
		msgActionCanceled:             "Действие отменено.",
		msgProfileHidden:              "Профиль скрыт из поиска.",
		msgProfileVisible:             "Профиль теперь видим в поиске.",
		msgProfileVisibilityUpdateFailed: "Не удалось обновить видимость в поиске. Попробуйте позже.",
		msgSearchGenderPrompt:            "Выберите пол для поиска.",
		msgSearchAccuracyPrompt:          "Точность совпадения (0-4): 0 — широкий/случайный → 4 — строгий/точный.",
		msgSearchNoCandidates:            "Пока нет подходящих профилей. Попробуйте позже.",
		msgSearchNoPrevious:              "Предыдущего профиля нет.",
		msgSearchHistoryEmpty:            "История пуста.",
		msgSearchStartRequired:           "Сначала нажмите «Начать поиск».",
		msgSearchProfileMissing:          "Сначала создайте профиль, чтобы начать поиск и смотреть анкеты.",
		msgSearchUnavailable:             "Поиск сейчас недоступен.",
		msgSearchActionFailed:            "Не удалось выполнить действие. Попробуйте позже.",
		msgSearchHistoryPage:             "История: всего %d. Показано %d-%d.",
		msgSearchHistoryEmptyPage:        "История: всего %d. На этой странице анкет нет.",
		msgSearchHistoryAction:           "Текущее решение: %s.\nВыберите действие.",
		msgMatchActions:                  "Выберите действие.",
		msgMatchActionSaved:              "Решение сохранено.",
		msgMatchProfileNotFound:          "Профиль не найден.",
		msgProfileViewRequiresProfile:    "Сначала создайте профиль, чтобы смотреть анкеты.",
		msgMatchSuccess:                  "Совпадение! Вы и %s понравились друг другу.",
		msgMatchNickname:                 "Никнейм: %s",
		msgNamePromptNoTelegram:          "В вашем профиле Telegram не указано имя. Пожалуйста, введите имя, которое хотите использовать.",
		msgNamePromptWithTelegram:        "Текущее имя в Telegram: %s\nИспользуйте кнопку ниже, чтобы взять его, или отправьте другое имя.",
		msgBotCheckPrompt:                "Чтобы защититься от ботов, решите: %s\nВыберите правильный ответ ниже.",
		msgBirthDatePrompt:               "Введите дату рождения в формате %s (например, 1990-04-23).",
		msgGenderPrompt:                  "Выберите свой пол с помощью кнопок ниже.",
		msgCountryPrompt:                 "Выберите страну с помощью кнопок ниже.",
		msgCityPrompt:                    "Выберите город с помощью кнопок ниже.",
		msgDescriptionPrompt:             "Напишите короткое описание о себе.",
		msgEmojiPrompt:                   "Выберите эмодзи, которое вас описывает, с помощью кнопок ниже.",
		msgPhotosPrompt:                  "Отправьте 1-4 фото для альбома. Используйте кнопку «Готово», когда закончите.",
		msgProfileEditMenu:               "Выберите, что хотите обновить в профиле.",
		msgProfileSettings:               "Настройки профиля. Используйте кнопки ниже, чтобы управлять видимостью в поиске или удалить профиль.",
		msgProfileSettingsWithVisibility: "Настройки профиля.\nВидимость в поиске: %s.\nИспользуйте кнопки ниже, чтобы управлять видимостью в поиске или удалить профиль.",
		msgProfileActions:                "Используйте кнопки ниже, чтобы посмотреть или отредактировать профиль.",
		msgProfilePreviewActions:         "Используйте кнопки ниже, чтобы вернуться к профилю или отредактировать его.",
		msgEditNamePrompt:                "Введите новое имя.",
		msgEditBirthDatePrompt:           "Введите новую дату рождения в формате %s (например, 1990-04-23).",
		msgEditGenderPrompt:              "Выберите новый пол с помощью кнопок ниже.",
		msgEditCountryPrompt:             "Выберите новую страну с помощью кнопок ниже.",
		msgEditCityPrompt:                "Выберите новый город с помощью кнопок ниже.",
		msgEditDescriptionPrompt:         "Введите новое описание.",
		msgEditEmojiPrompt:               "Выберите новое эмодзи с помощью кнопок ниже.",
		msgEditPhotosPrompt:              "Отправьте 1-4 фото, чтобы заменить альбом. Используйте кнопку «Готово», когда закончите.",
		msgEditDefaultPrompt:             "Введите обновленное значение.",
		msgProfileDetailsHeader:          "Ваш профиль:",
		msgProfilePreviewFallbackName:    "Профиль",
		msgProfileEditHeader:             "Редактирование профиля: %s",
		msgProfileSetupHeader:            "Создание профиля (шаг %d/%d): %s\n%s",
		msgProfileSetupEstimate:          "Примерное время: ~%d мин",
		msgStartGreeting:                 "Добро пожаловать в Meethalf bot.",
		msgStartGreetingNamed:            "Добро пожаловать в Meethalf bot, %s.",
		msgProfileNotFoundCreateButton:   "Профиль не найден. Используйте кнопку «Создать профиль», чтобы создать его.",
		msgProfileNotFoundButtonBelow:    "Профиль не найден. Используйте кнопку ниже, чтобы создать его.",
		msgProfileNotFoundUseProfile:     "Профиль не найден. Используйте /profile, чтобы создать его.",
		msgProfileDeleted:                "Профиль удален. Используйте кнопку «Создать профиль», чтобы создать новый.",
		msgProfileDeletedDraftWarning:    "Профиль удален. Примечание: не удалось очистить черновик профиля.",
		msgProfileDeletedDraftWarningOnly:"Примечание: не удалось очистить черновик профиля.",
		msgProfileSetupUnavailable:       "Создание профиля сейчас недоступно.",
		msgProfileSetupUnavailableChat:   "Создание профиля недоступно в этом чате.",
		msgProfileSetupStartFailed:       "Не удалось начать создание профиля. Попробуйте позже.",
		msgProfileSetupLoadFailed:        "Не удалось загрузить создание профиля. Попробуйте позже.",
		msgProfileSetupNotActive:         "Создание профиля не активно. Используйте /profile, чтобы начать.",
		msgProfileSetupSaveFailed:        "Не удалось сохранить создание профиля. Попробуйте позже.",
		msgProfileEditUnavailable:        "Редактирование профиля сейчас недоступно.",
		msgProfileEditUnavailableChat:    "Редактирование профиля недоступно в этом чате.",
		msgProfileServiceUnavailable:     "Сервис профиля сейчас недоступен.",
		msgProfileLoadFailed:             "Не удалось загрузить профиль. Попробуйте позже.",
		msgProfileDeleteUnavailable:      "Удаление профиля сейчас недоступно.",
		msgProfileDeleteUnavailableChat:  "Удаление профиля недоступно в этом чате.",
		msgProfileDeletePrepareFailed:    "Не удалось подготовить удаление профиля. Попробуйте позже.",
		msgProfileDeleteFailed:           "Не удалось удалить профиль. Попробуйте позже.",
		msgProfileDeleteCancelFailed:     "Не удалось отменить удаление профиля. Попробуйте позже.",
		msgProfileEditStartFailed:        "Не удалось начать редактирование профиля. Попробуйте позже.",
		msgProfileEditSaveFailed:         "Не удалось сохранить редактирование профиля. Попробуйте позже.",
		msgProfileSaveFailed:             "Не удалось сохранить профиль. Попробуйте позже.",
		msgBotCheckTooManyAttempts:       "Слишком много попыток. Попробуем другую проверку.",
		msgBotCheckIncorrect:             "Неверный ответ. Попробуйте снова.",
		msgNamePromptEmpty:               "Пожалуйста, введите имя.",
		msgNamePromptEmptyCreate:         "Пожалуйста, введите имя или используйте кнопку ниже, чтобы взять имя из Telegram.",
		msgNamePromptTelegramMissing:     "В вашем профиле Telegram не указано имя. Пожалуйста, введите имя, которое хотите использовать.",
		msgNameTooLong:                   "Имя слишком длинное (максимум %d символов).",
		msgGenderInvalid:                 "Пол должен быть одним из: мужской, женский или другое.",
		msgBirthDateInvalid:              "Дата рождения должна быть в формате %s (например, 1990-04-23). Попробуйте снова.",
		msgAgeInvalid:                    "Возраст должен быть от %d до %d лет. Проверьте дату рождения.",
		msgCountryInvalid:                "Страна должна быть одной из: Россия, Казахстан или Беларусь.",
		msgCityInvalid:                   "Город нужно выбрать из списка для вашей страны.",
		msgDescriptionEmpty:              "Описание не может быть пустым. Попробуйте снова.",
		msgDescriptionTooLong:            "Описание слишком длинное (максимум %d символов).",
		msgEmojiInvalid:                  "Эмодзи нужно выбрать из списка.",
		msgPhotosNotEnough:               "Пожалуйста, отправьте минимум %d фото перед завершением.",
		msgPhotosPromptRepeat:            "Отправьте 1-4 фото для альбома. Используйте кнопку «Готово», когда закончите.",
		msgPhotosLimitReached:            "Вы достигли лимита в %d фото. Используйте кнопку «Готово», чтобы завершить.",
		msgPhotosProgress:                "Фото в альбоме: %d/%d. Отправьте еще или используйте кнопку «Готово».",
		msgProfileViewUnavailable:        "Сервис профиля сейчас недоступен.",
		msgProfileViewUnavailableChat:    "Профиль недоступен в этом чате.",
		msgProfileDeleteNoteDraft:        "Профиль удален. Примечание: не удалось очистить черновик профиля.",
		msgAdminUserListNoName:           "Без имени",
		msgAdminUserListNA:               "н/д",
		msgAdminUserListLine:             "%d. %s (ID: %d, username: %s, %s)",
		msgAdminReportedUserListLine:     "%d. %s (ID: %d, username: %s, жалобы: %d, %s)",
		msgSearchHistoryLine:             "%d. %s — %s",
		msgSearchHistoryLineShort:        "%d. %s",
		msgSearchHistoryUserFallback:     "Пользователь %d",
		msgSearchHistoryLabel:            "%s (%s)",
		msgAgeShort:                      "%d лет",
		msgProfileDetailsLine:            "%s: %s",
		msgProfileDetailsPhotos:          "Фото: %d",
		msgProfileDetailsCreated:         "Создан: %s",
		msgProfileDetailsUpdated:         "Обновлен: %s",
		msgProfileDetailsDescriptionLabel:"Описание: \n%s",
		msgLikeNotification:              "Вы получили ❤️. Посмотреть профиль?",
		msgLikeNotificationFrom:          "Вы получили ❤️ от %s. Посмотреть профиль?",
		msgLanguagePrompt:                "Выберите язык.",
		msgLanguageUpdated:               "Язык обновлен.",
		msgLanguageUnsupported:           "Неподдерживаемый язык. Пожалуйста, выберите один из доступных.",
		msgProfileSettingsLanguageHint:   "Настройки профиля. Используйте кнопки ниже, чтобы управлять видимостью, изменить язык или удалить профиль.",
	},
}

var buttonCatalog = map[domain.Language]map[buttonKey]string{
	domain.LanguageEnglish: {
		btnCancel:                 "Back to menu",
		btnProfileSetupBack:       "Previous step",
		btnProfileDeleteCancel:    "Keep profile",
		btnBackToProfile:          "Back to profile",
		btnEditCancel:             "Cancel",
		btnSearchAccuracyCancel:   "Cancel",
		btnSearchRefresh:          "Refresh feed",
		btnSearchHistory:          "History",
		btnSearchHistoryPrev:      "Previous",
		btnSearchHistoryNext:      "Next",
		btnSearchHistoryBack:      "Back to history",
		btnAdminMenu:              "Admin panel",
		btnModeratorMenu:          "Moderator panel",
		btnAdminUsers:             "User list",
		btnAdminBan:               "Ban user",
		btnAdminUnban:             "Unban user",
		btnAdminModerator:         "Make moderator",
		btnAdminUnmoderator:       "Remove moderator",
		btnAdminResetChoices:      "Reset choices",
		btnAdminBannedUsers:       "Banned users",
		btnAdminModerators:        "Moderators",
		btnAdminReports:           "Reported users",
		btnAdminClearReports:      "Clear reports",
		btnAdminUsersPrev:         "Previous",
		btnAdminUsersNext:         "Next",
		btnAdminBackToMenu:        "Back to admin",
		btnStartSearch:            "Start search",
		btnProfile:                "Profile",
		btnCreateProfile:          "Create Profile",
		btnSettings:               "Settings",
		btnPreviewProfile:         "Preview profile",
		btnEditProfile:            "Edit profile",
		btnDeleteProfile:          "Delete profile",
		btnDeleteConfirm:          "Yes, delete",
		btnProfileEditName:        "Name",
		btnProfileEditGender:      "Gender",
		btnProfileEditBirthDate:   "Birth date",
		btnProfileEditCountry:     "Country",
		btnProfileEditCity:        "City",
		btnProfileEditDescription: "Description",
		btnProfileEditEmoji:       "Emoji",
		btnProfileEditPhotos:      "Photos",
		btnTelegramName:           "Use Telegram name",
		btnAlbumDone:              "Done",
		btnGenderMale:             "Male",
		btnGenderFemale:           "Female",
		btnGenderOther:            "Other",
		btnCountryRussia:          "Russia",
		btnCountryKazakhstan:      "Kazakhstan",
		btnCountryBelarus:         "Belarus",
		btnSearchMen:              "Men",
		btnSearchWomen:            "Women",
		btnSearchOther:            "Other",
		btnSearchAny:              "Any",
		btnReport:                 "Report",
		btnBackToPreviousProfile:  "Back to previous profile",
		btnViewProfile:            "View profile",
		btnHideFromSearch:         "Hide from search",
		btnShowInSearch:           "Show in search",
		btnLanguage:               "Language",
		btnBackToSettings:         "Back to settings",
		btnLanguageEnglish:        "English",
		btnLanguageRussian:        "Русский",
	},
	domain.LanguageRussian: {
		btnCancel:                 "Назад в меню",
		btnProfileSetupBack:       "Предыдущий шаг",
		btnProfileDeleteCancel:    "Оставить профиль",
		btnBackToProfile:          "Назад к профилю",
		btnEditCancel:             "Отмена",
		btnSearchAccuracyCancel:   "Отмена",
		btnSearchRefresh:          "Обновить ленту",
		btnSearchHistory:          "История",
		btnSearchHistoryPrev:      "Назад",
		btnSearchHistoryNext:      "Далее",
		btnSearchHistoryBack:      "Назад к истории",
		btnAdminMenu:              "Админ-панель",
		btnModeratorMenu:          "Панель модератора",
		btnAdminUsers:             "Список пользователей",
		btnAdminBan:               "Забанить пользователя",
		btnAdminUnban:             "Разбанить пользователя",
		btnAdminModerator:         "Сделать модератором",
		btnAdminUnmoderator:       "Снять модератора",
		btnAdminResetChoices:      "Сбросить решения",
		btnAdminBannedUsers:       "Забаненные",
		btnAdminModerators:        "Модераторы",
		btnAdminReports:           "Жалобы",
		btnAdminClearReports:      "Очистить жалобы",
		btnAdminUsersPrev:         "Назад",
		btnAdminUsersNext:         "Далее",
		btnAdminBackToMenu:        "Назад в админ",
		btnStartSearch:            "Начать поиск",
		btnProfile:                "Профиль",
		btnCreateProfile:          "Создать профиль",
		btnSettings:               "Настройки",
		btnPreviewProfile:         "Предпросмотр профиля",
		btnEditProfile:            "Редактировать профиль",
		btnDeleteProfile:          "Удалить профиль",
		btnDeleteConfirm:          "Да, удалить",
		btnProfileEditName:        "Имя",
		btnProfileEditGender:      "Пол",
		btnProfileEditBirthDate:   "Дата рождения",
		btnProfileEditCountry:     "Страна",
		btnProfileEditCity:        "Город",
		btnProfileEditDescription: "Описание",
		btnProfileEditEmoji:       "Эмодзи",
		btnProfileEditPhotos:      "Фото",
		btnTelegramName:           "Использовать имя из Telegram",
		btnAlbumDone:              "Готово",
		btnGenderMale:             "Мужской",
		btnGenderFemale:           "Женский",
		btnGenderOther:            "Другое",
		btnCountryRussia:          "Россия",
		btnCountryKazakhstan:      "Казахстан",
		btnCountryBelarus:         "Беларусь",
		btnSearchMen:              "Мужчины",
		btnSearchWomen:            "Женщины",
		btnSearchOther:            "Другое",
		btnSearchAny:              "Любой",
		btnReport:                 "Пожаловаться",
		btnBackToPreviousProfile:  "Назад к предыдущему профилю",
		btnViewProfile:            "Посмотреть профиль",
		btnHideFromSearch:         "Скрыть из поиска",
		btnShowInSearch:           "Показывать в поиске",
		btnLanguage:               "Язык",
		btnBackToSettings:         "Назад к настройкам",
		btnLanguageEnglish:        "English",
		btnLanguageRussian:        "Русский",
	},
}

var labelCatalog = map[domain.Language]map[labelKey]string{
	domain.LanguageEnglish: {
		labelNotSet:            "Not set",
		labelUnknown:           "Unknown",
		labelProfile:           "Profile",
		labelStepVerification:  "Verification",
		labelStepName:          "Name",
		labelStepGender:        "Gender",
		labelStepBirthDate:     "Birth date",
		labelStepCountry:       "Country",
		labelStepCity:          "City",
		labelStepDescription:   "Description",
		labelStepEmoji:         "Emoji",
		labelStepPhotos:        "Photos",
		labelGenderMale:        "Male",
		labelGenderFemale:      "Female",
		labelGenderOther:       "Other",
		labelCountryRussia:     "Russia",
		labelCountryKazakhstan: "Kazakhstan",
		labelCountryBelarus:    "Belarus",
		labelActionLike:        "Like",
		labelActionDislike:     "Dislike",
		labelActionReport:      "Report",
		labelActionNone:        "No decision",
		labelStatusBanned:      "banned",
		labelStatusModerator:   "moderator",
		labelStatusHidden:      "hidden",
		labelStatusVisible:     "visible",
		labelName:              "Name",
		labelEmoji:             "Emoji",
		labelGender:            "Gender",
		labelAge:               "Age",
		labelCountry:           "Country",
		labelCity:              "City",
		labelSearchVisibility:  "Search visibility",
		labelDescription:       "Description",
		labelPhotos:            "Photos",
		labelCreated:           "Created",
		labelUpdated:           "Updated",
		labelLanguageEnglish:   "English",
		labelLanguageRussian:   "Russian",
		labelThisUser:          "this user",
	},
	domain.LanguageRussian: {
		labelNotSet:            "Не указано",
		labelUnknown:           "Неизвестно",
		labelProfile:           "Профиль",
		labelStepVerification:  "Проверка",
		labelStepName:          "Имя",
		labelStepGender:        "Пол",
		labelStepBirthDate:     "Дата рождения",
		labelStepCountry:       "Страна",
		labelStepCity:          "Город",
		labelStepDescription:   "Описание",
		labelStepEmoji:         "Эмодзи",
		labelStepPhotos:        "Фото",
		labelGenderMale:        "Мужской",
		labelGenderFemale:      "Женский",
		labelGenderOther:       "Другое",
		labelCountryRussia:     "Россия",
		labelCountryKazakhstan: "Казахстан",
		labelCountryBelarus:    "Беларусь",
		labelActionLike:        "Лайк",
		labelActionDislike:     "Дизлайк",
		labelActionReport:      "Жалоба",
		labelActionNone:        "Нет решения",
		labelStatusBanned:      "забанен",
		labelStatusModerator:   "модератор",
		labelStatusHidden:      "скрыт",
		labelStatusVisible:     "видим",
		labelName:              "Имя",
		labelEmoji:             "Эмодзи",
		labelGender:            "Пол",
		labelAge:               "Возраст",
		labelCountry:           "Страна",
		labelCity:              "Город",
		labelSearchVisibility:  "Видимость в поиске",
		labelDescription:       "Описание",
		labelPhotos:            "Фото",
		labelCreated:           "Создан",
		labelUpdated:           "Обновлен",
		labelLanguageEnglish:   "Английский",
		labelLanguageRussian:   "Русский",
		labelThisUser:          "этот пользователь",
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
	},
}

func normalizeLanguageValue(lang domain.Language) domain.Language {
	switch lang {
	case domain.LanguageRussian:
		return domain.LanguageRussian
	default:
		return domain.LanguageEnglish
	}
}
