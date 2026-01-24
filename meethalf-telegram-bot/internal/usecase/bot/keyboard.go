package bot

import (
	"strconv"

	"meethalf-telegram-bot/internal/domain"
)

const (
	cancelButtonText               = "Back to menu"
	profileSetupBackButtonText     = "Previous step"
	profileDeleteCancelButtonText  = "Keep profile"
	backToProfileButtonText        = "Back to profile"
	editCancelButtonText           = "Cancel"
	searchAccuracyCancelButtonText = "Cancel"
	searchRefreshButtonText        = "Refresh feed"
	adminMenuButtonText            = "Admin panel"
	adminUsersButtonText           = "User list"
	adminBanButtonText             = "Ban user"
	adminUnbanButtonText           = "Unban user"
	adminModeratorButtonText       = "Make moderator"
	adminUnmoderatorButtonText     = "Remove moderator"
	adminBannedUsersButtonText     = "Banned users"
	adminModeratorsButtonText      = "Moderators"
	adminUsersPrevButtonText       = "Previous"
	adminUsersNextButtonText       = "Next"
	adminBackToMenuButtonText      = "Back to admin"
)

func (s *service) startInlineKeyboardByStatus(status profileStatus, user domain.User) *domain.InlineKeyboard {
	if status == profileStatusMissing {
		return s.withAdminMenuInlineKeyboard(s.profileCreateInlineKeyboard(), user)
	}

	return s.withAdminMenuInlineKeyboard(s.profileInlineKeyboard(), user)
}

func (s *service) profileInlineKeyboard() *domain.InlineKeyboard {
	return s.profileStartInlineKeyboard("Profile", domain.CommandProfileView)
}

func (s *service) profileCreateInlineKeyboard() *domain.InlineKeyboard {
	return s.profileStartInlineKeyboard("Create Profile", domain.CommandProfile)
}

func (s *service) profileStartInlineKeyboard(text, callbackData string) *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         "Start search",
					CallbackData: domain.CommandSearchStart,
				},
			},
			{
				{
					Text:         text,
					CallbackData: callbackData,
				},
			},
			{
				{
					Text:         "Settings",
					CallbackData: domain.CommandProfileSettings,
				},
			},
		},
	}
}

func (s *service) adminMenuInlineKeyboard() *domain.InlineKeyboard {
	return withCancelInlineKeyboard(&domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         adminUsersButtonText,
					CallbackData: domain.CommandAdminUsers,
				},
			},
			{
				{
					Text:         adminBannedUsersButtonText,
					CallbackData: domain.CommandAdminBannedUsers,
				},
			},
			{
				{
					Text:         adminModeratorsButtonText,
					CallbackData: domain.CommandAdminModerators,
				},
			},
			{
				{
					Text:         adminBanButtonText,
					CallbackData: domain.CommandAdminBan,
				},
			},
			{
				{
					Text:         adminUnbanButtonText,
					CallbackData: domain.CommandAdminUnban,
				},
			},
			{
				{
					Text:         adminModeratorButtonText,
					CallbackData: domain.CommandAdminModerator,
				},
			},
			{
				{
					Text:         adminUnmoderatorButtonText,
					CallbackData: domain.CommandAdminUnmoderator,
				},
			},
		},
	})
}

func (s *service) adminBanInlineKeyboard() *domain.InlineKeyboard {
	return withCancelInlineKeyboard(&domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         adminBackToMenuButtonText,
					CallbackData: domain.CommandAdminMenu,
				},
			},
		},
	})
}

func (s *service) adminModeratorInlineKeyboard() *domain.InlineKeyboard {
	return s.adminBanInlineKeyboard()
}

func (s *service) adminUsersInlineKeyboard(offset, limit, total int) *domain.InlineKeyboard {
	rows := [][]domain.InlineButton{}

	hasPrev := offset > 0
	hasNext := limit > 0 && (offset+limit) < total
	if hasPrev || hasNext {
		row := []domain.InlineButton{}
		if hasPrev {
			prevOffset := offset - limit
			if prevOffset < 0 {
				prevOffset = 0
			}
			row = append(row, domain.InlineButton{
				Text:         adminUsersPrevButtonText,
				CallbackData: domain.CommandAdminUsers + ":" + strconv.Itoa(prevOffset),
			})
		}
		if hasNext {
			nextOffset := offset + limit
			row = append(row, domain.InlineButton{
				Text:         adminUsersNextButtonText,
				CallbackData: domain.CommandAdminUsers + ":" + strconv.Itoa(nextOffset),
			})
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	rows = append(rows, []domain.InlineButton{
		{
			Text:         adminBackToMenuButtonText,
			CallbackData: domain.CommandAdminMenu,
		},
	})

	return withCancelInlineKeyboard(&domain.InlineKeyboard{Buttons: rows})
}

func (s *service) adminBannedUsersInlineKeyboard(offset, limit, total int) *domain.InlineKeyboard {
	rows := [][]domain.InlineButton{}

	hasPrev := offset > 0
	hasNext := limit > 0 && (offset+limit) < total
	if hasPrev || hasNext {
		row := []domain.InlineButton{}
		if hasPrev {
			prevOffset := offset - limit
			if prevOffset < 0 {
				prevOffset = 0
			}
			row = append(row, domain.InlineButton{
				Text:         adminUsersPrevButtonText,
				CallbackData: domain.CommandAdminBannedUsers + ":" + strconv.Itoa(prevOffset),
			})
		}
		if hasNext {
			nextOffset := offset + limit
			row = append(row, domain.InlineButton{
				Text:         adminUsersNextButtonText,
				CallbackData: domain.CommandAdminBannedUsers + ":" + strconv.Itoa(nextOffset),
			})
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	rows = append(rows, []domain.InlineButton{
		{
			Text:         adminBackToMenuButtonText,
			CallbackData: domain.CommandAdminMenu,
		},
	})

	return withCancelInlineKeyboard(&domain.InlineKeyboard{Buttons: rows})
}

func (s *service) adminModeratorsInlineKeyboard(offset, limit, total int) *domain.InlineKeyboard {
	rows := [][]domain.InlineButton{}

	hasPrev := offset > 0
	hasNext := limit > 0 && (offset+limit) < total
	if hasPrev || hasNext {
		row := []domain.InlineButton{}
		if hasPrev {
			prevOffset := offset - limit
			if prevOffset < 0 {
				prevOffset = 0
			}
			row = append(row, domain.InlineButton{
				Text:         adminUsersPrevButtonText,
				CallbackData: domain.CommandAdminModerators + ":" + strconv.Itoa(prevOffset),
			})
		}
		if hasNext {
			nextOffset := offset + limit
			row = append(row, domain.InlineButton{
				Text:         adminUsersNextButtonText,
				CallbackData: domain.CommandAdminModerators + ":" + strconv.Itoa(nextOffset),
			})
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	rows = append(rows, []domain.InlineButton{
		{
			Text:         adminBackToMenuButtonText,
			CallbackData: domain.CommandAdminMenu,
		},
	})

	return withCancelInlineKeyboard(&domain.InlineKeyboard{Buttons: rows})
}

func (s *service) adminUnbanInlineKeyboard() *domain.InlineKeyboard {
	return s.adminBanInlineKeyboard()
}

func (s *service) adminUnmoderatorInlineKeyboard() *domain.InlineKeyboard {
	return s.adminBanInlineKeyboard()
}

func (s *service) withAdminMenuInlineKeyboard(keyboard *domain.InlineKeyboard, user domain.User) *domain.InlineKeyboard {
	if s == nil || !s.isAdminUser(user) {
		return keyboard
	}
	if keyboard == nil {
		keyboard = &domain.InlineKeyboard{}
	}
	if inlineKeyboardHasCallback(keyboard, domain.CommandAdminMenu) || inlineKeyboardHasText(keyboard, adminMenuButtonText) {
		return keyboard
	}

	keyboard.Buttons = append(keyboard.Buttons, []domain.InlineButton{
		{
			Text:         adminMenuButtonText,
			CallbackData: domain.CommandAdminMenu,
		},
	})

	return keyboard
}

func (s *service) profileViewInlineKeyboard() *domain.InlineKeyboard {
	return withCancelInlineKeyboard(&domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         "Preview profile",
					CallbackData: domain.CommandProfilePreview,
				},
			},
			{
				{
					Text:         "Edit profile",
					CallbackData: domain.CommandProfileEdit,
				},
			},
		},
	})
}

func (s *service) profilePreviewInlineKeyboard() *domain.InlineKeyboard {
	return withCancelInlineKeyboard(&domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         backToProfileButtonText,
					CallbackData: domain.CommandProfileView,
				},
			},
			{
				{
					Text:         "Edit profile",
					CallbackData: domain.CommandProfileEdit,
				},
			},
		},
	})
}

func (s *service) profileSettingsInlineKeyboard(isHidden bool) *domain.InlineKeyboard {
	visibilityText := "Hide from search"
	visibilityAction := profileVisibilityHideAction
	if isHidden {
		visibilityText = "Show in search"
		visibilityAction = profileVisibilityShowAction
	}

	return withCancelInlineKeyboard(&domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         visibilityText,
					CallbackData: domain.CommandProfileVisibility + ":" + visibilityAction,
				},
			},
			{
				{
					Text:         "Delete profile",
					CallbackData: domain.CommandProfileDelete,
				},
			},
		},
	})
}

func (s *service) profileDeleteConfirmInlineKeyboard() *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         "Yes, delete",
					CallbackData: domain.CommandProfileDeleteConfirm,
				},
			},
			{
				{
					Text:         profileDeleteCancelButtonText,
					CallbackData: domain.CommandProfileDeleteCancel,
				},
			},
		},
	}
}

func (s *service) profileEditMenuKeyboard() *domain.InlineKeyboard {
	return withBackToProfileInlineKeyboard(&domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         "Name",
					CallbackData: domain.CommandProfileEditName,
				},
				{
					Text:         "Gender",
					CallbackData: domain.CommandProfileEditGender,
				},
			},
			{
				{
					Text:         "Birth date",
					CallbackData: domain.CommandProfileEditBirthDate,
				},
				{
					Text:         "Country",
					CallbackData: domain.CommandProfileEditCountry,
				},
			},
			{
				{
					Text:         "City",
					CallbackData: domain.CommandProfileEditCity,
				},
				{
					Text:         "Description",
					CallbackData: domain.CommandProfileEditDesc,
				},
			},
			{
				{
					Text:         "Emoji",
					CallbackData: domain.CommandProfileEditEmoji,
				},
				{
					Text:         "Photos",
					CallbackData: domain.CommandProfileEditPhotos,
				},
			},
		},
	})
}

func (s *service) telegramNameInlineKeyboard() *domain.InlineKeyboard {
	return withCancelInlineKeyboard(&domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         telegramNameButtonText,
					CallbackData: telegramNameCallbackData,
				},
			},
		},
	})
}

func (s *service) photosDoneInlineKeyboard() *domain.InlineKeyboard {
	return withCancelInlineKeyboard(&domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         albumDoneButtonText,
					CallbackData: albumDoneCallbackData,
				},
			},
		},
	})
}

func (s *service) botCheckInlineKeyboard(answer int) *domain.InlineKeyboard {
	options := s.botCheckOptions(answer)
	if len(options) == 0 {
		return nil
	}

	rows := make([][]domain.InlineButton, 0, (len(options)+botCheckOptionsColumns-1)/botCheckOptionsColumns)
	for i, option := range options {
		if i%botCheckOptionsColumns == 0 {
			rows = append(rows, []domain.InlineButton{})
		}
		rowIndex := len(rows) - 1
		label := strconv.Itoa(option)
		rows[rowIndex] = append(rows[rowIndex], domain.InlineButton{
			Text:         label,
			CallbackData: label,
		})
	}

	return withCancelInlineKeyboard(&domain.InlineKeyboard{Buttons: rows})
}

func (s *service) genderInlineKeyboard() *domain.InlineKeyboard {
	return withCancelInlineKeyboard(&domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         "Male",
					CallbackData: string(domain.GenderMale),
				},
				{
					Text:         "Female",
					CallbackData: string(domain.GenderFemale),
				},
				{
					Text:         "Other",
					CallbackData: string(domain.GenderOther),
				},
			},
		},
	})
}

func (s *service) countryInlineKeyboard() *domain.InlineKeyboard {
	return withCancelInlineKeyboard(&domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         "Russia",
					CallbackData: string(domain.CountryRussia),
				},
				{
					Text:         "Kazakhstan",
					CallbackData: string(domain.CountryKazakhstan),
				},
				{
					Text:         "Belarus",
					CallbackData: string(domain.CountryBelarus),
				},
			},
		},
	})
}

func (s *service) cityInlineKeyboard(country domain.Country) *domain.InlineKeyboard {
	options := s.cityOptions(country)
	if len(options) == 0 {
		return nil
	}

	return withCancelInlineKeyboard(listInlineKeyboard(options, cityKeyboardColumns))
}

func listInlineKeyboard(options []string, columns int) *domain.InlineKeyboard {
	if len(options) == 0 {
		return nil
	}
	if columns <= 0 {
		columns = 1
	}

	rows := make([][]domain.InlineButton, 0, (len(options)+columns-1)/columns)
	for i, option := range options {
		if i%columns == 0 {
			rows = append(rows, []domain.InlineButton{})
		}
		rowIndex := len(rows) - 1
		rows[rowIndex] = append(rows[rowIndex], domain.InlineButton{
			Text:         option,
			CallbackData: option,
		})
	}

	return &domain.InlineKeyboard{Buttons: rows}
}

func (s *service) searchGenderInlineKeyboard() *domain.InlineKeyboard {
	return withCancelInlineKeyboard(&domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         "Men",
					CallbackData: domain.CommandSearchGender + ":" + string(domain.GenderMale),
				},
				{
					Text:         "Women",
					CallbackData: domain.CommandSearchGender + ":" + string(domain.GenderFemale),
				},
			},
			{
				{
					Text:         "Other",
					CallbackData: domain.CommandSearchGender + ":" + string(domain.GenderOther),
				},
				{
					Text:         "Any",
					CallbackData: domain.CommandSearchGender + ":" + string(domain.GenderUnspecified),
				},
			},
		},
	})
}

func (s *service) searchAccuracyInlineKeyboard(gender domain.Gender) *domain.InlineKeyboard {
	prefix := domain.CommandSearchAccuracy + ":" + string(gender) + ":"
	return withSearchGenderCancelInlineKeyboard(&domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         "0",
					CallbackData: prefix + "0",
				},
				{
					Text:         "1",
					CallbackData: prefix + "1",
				},
				{
					Text:         "2",
					CallbackData: prefix + "2",
				},
			},
			{
				{
					Text:         "3",
					CallbackData: prefix + "3",
				},
				{
					Text:         "4",
					CallbackData: prefix + "4",
				},
			},
		},
	})
}

func (s *service) searchNoCandidatesInlineKeyboard() *domain.InlineKeyboard {
	return withCancelInlineKeyboard(&domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         searchRefreshButtonText,
					CallbackData: domain.CommandSearchRefresh,
				},
			},
		},
	})
}

func (s *service) matchActionsInlineKeyboard(targetID int64, hasPrevious bool) *domain.InlineKeyboard {
	target := strconv.FormatInt(targetID, 10)
	buttons := [][]domain.InlineButton{
		{
			{
				Text:         "👎",
				CallbackData: domain.CommandMatchDislike + ":" + target,
			},
			{
				Text:         "❤️",
				CallbackData: domain.CommandMatchLike + ":" + target,
			},
		},
	}
	row := []domain.InlineButton{
		{
			Text:         "Report",
			CallbackData: domain.CommandMatchReport + ":" + target,
		},
	}
	if hasPrevious {
		row = append(row, domain.InlineButton{
			Text:         "Back to previous profile",
			CallbackData: domain.CommandMatchPrevious,
		})
	}
	buttons = append(buttons, row)

	return withCancelInlineKeyboard(&domain.InlineKeyboard{Buttons: buttons})
}

func (s *service) matchViewInlineKeyboard(targetID int64) *domain.InlineKeyboard {
	target := strconv.FormatInt(targetID, 10)
	return withCancelInlineKeyboard(&domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         "View profile",
					CallbackData: domain.CommandMatchViewProfile + ":" + target,
				},
			},
		},
	})
}

func withDraftCancelInlineKeyboard(keyboard *domain.InlineKeyboard, mode domain.ProfileDraftMode) *domain.InlineKeyboard {
	if mode == domain.ProfileDraftModeEdit {
		return withEditCancelInlineKeyboard(withoutCancelInlineKeyboard(keyboard))
	}
	return withCancelInlineKeyboard(keyboard)
}

func withProfileSetupBackInlineKeyboard(keyboard *domain.InlineKeyboard) *domain.InlineKeyboard {
	if keyboard == nil {
		return profileSetupBackInlineKeyboard()
	}
	if inlineKeyboardHasCallback(keyboard, domain.CommandProfileSetupBack) || inlineKeyboardHasText(keyboard, profileSetupBackButtonText) {
		return keyboard
	}

	backRow := []domain.InlineButton{
		{
			Text:         profileSetupBackButtonText,
			CallbackData: domain.CommandProfileSetupBack,
		},
	}

	insertAt := len(keyboard.Buttons)
	for i, row := range keyboard.Buttons {
		if inlineKeyboardRowHasCancel(row) {
			insertAt = i
			break
		}
	}

	keyboard.Buttons = append(keyboard.Buttons[:insertAt], append([][]domain.InlineButton{backRow}, keyboard.Buttons[insertAt:]...)...)
	return keyboard
}

func withSearchGenderCancelInlineKeyboard(keyboard *domain.InlineKeyboard) *domain.InlineKeyboard {
	keyboard = withoutCancelInlineKeyboard(keyboard)
	if keyboard == nil {
		return searchGenderCancelInlineKeyboard()
	}
	if inlineKeyboardHasCallback(keyboard, domain.CommandSearchGender) || inlineKeyboardHasText(keyboard, searchAccuracyCancelButtonText) {
		return keyboard
	}

	keyboard.Buttons = append(keyboard.Buttons, []domain.InlineButton{
		{
			Text:         searchAccuracyCancelButtonText,
			CallbackData: domain.CommandSearchGender,
		},
	})

	return keyboard
}

func withCancelInlineKeyboard(keyboard *domain.InlineKeyboard) *domain.InlineKeyboard {
	if keyboard == nil {
		return cancelInlineKeyboard()
	}
	if inlineKeyboardHasCallback(keyboard, domain.CommandCancel) || inlineKeyboardHasText(keyboard, cancelButtonText) {
		return keyboard
	}

	keyboard.Buttons = append(keyboard.Buttons, []domain.InlineButton{
		{
			Text:         cancelButtonText,
			CallbackData: domain.CommandCancel,
		},
	})

	return keyboard
}

func withEditCancelInlineKeyboard(keyboard *domain.InlineKeyboard) *domain.InlineKeyboard {
	if keyboard == nil {
		return editCancelInlineKeyboard()
	}
	if inlineKeyboardHasCallback(keyboard, domain.CommandProfileEdit) || inlineKeyboardHasText(keyboard, editCancelButtonText) {
		return keyboard
	}

	keyboard.Buttons = append(keyboard.Buttons, []domain.InlineButton{
		{
			Text:         editCancelButtonText,
			CallbackData: domain.CommandProfileEdit,
		},
	})

	return keyboard
}

func withBackToProfileInlineKeyboard(keyboard *domain.InlineKeyboard) *domain.InlineKeyboard {
	if keyboard == nil {
		return backToProfileInlineKeyboard()
	}
	if inlineKeyboardHasCallback(keyboard, domain.CommandProfileView) || inlineKeyboardHasText(keyboard, backToProfileButtonText) {
		return keyboard
	}

	keyboard.Buttons = append(keyboard.Buttons, []domain.InlineButton{
		{
			Text:         backToProfileButtonText,
			CallbackData: domain.CommandProfileView,
		},
	})

	return keyboard
}

func cancelInlineKeyboard() *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         cancelButtonText,
					CallbackData: domain.CommandCancel,
				},
			},
		},
	}
}

func editCancelInlineKeyboard() *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         editCancelButtonText,
					CallbackData: domain.CommandProfileEdit,
				},
			},
		},
	}
}

func searchGenderCancelInlineKeyboard() *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         searchAccuracyCancelButtonText,
					CallbackData: domain.CommandSearchGender,
				},
			},
		},
	}
}

func backToProfileInlineKeyboard() *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         backToProfileButtonText,
					CallbackData: domain.CommandProfileView,
				},
			},
		},
	}
}

func profileSetupBackInlineKeyboard() *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         profileSetupBackButtonText,
					CallbackData: domain.CommandProfileSetupBack,
				},
			},
		},
	}
}

func withoutCancelInlineKeyboard(keyboard *domain.InlineKeyboard) *domain.InlineKeyboard {
	if keyboard == nil {
		return nil
	}

	rows := make([][]domain.InlineButton, 0, len(keyboard.Buttons))
	for _, row := range keyboard.Buttons {
		filtered := make([]domain.InlineButton, 0, len(row))
		for _, button := range row {
			if button.CallbackData == domain.CommandCancel || button.Text == cancelButtonText {
				continue
			}
			filtered = append(filtered, button)
		}
		if len(filtered) > 0 {
			rows = append(rows, filtered)
		}
	}

	keyboard.Buttons = rows
	return keyboard
}

func inlineKeyboardHasCallback(keyboard *domain.InlineKeyboard, callbackData string) bool {
	if keyboard == nil || callbackData == "" {
		return false
	}

	for _, row := range keyboard.Buttons {
		for _, button := range row {
			if button.CallbackData == callbackData {
				return true
			}
		}
	}

	return false
}

func inlineKeyboardHasText(keyboard *domain.InlineKeyboard, text string) bool {
	if keyboard == nil || text == "" {
		return false
	}

	for _, row := range keyboard.Buttons {
		for _, button := range row {
			if button.Text == text {
				return true
			}
		}
	}

	return false
}

func inlineKeyboardRowHasCancel(row []domain.InlineButton) bool {
	for _, button := range row {
		if button.CallbackData == domain.CommandCancel || button.Text == cancelButtonText {
			return true
		}
	}

	return false
}
