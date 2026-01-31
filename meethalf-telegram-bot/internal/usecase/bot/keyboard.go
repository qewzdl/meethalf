package bot

import (
	"strconv"
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

const (
	historyLabelMaxRunes   = 20
	matchDislikeButtonText = "\U0001F44E"
	matchLikeButtonText    = "\u2764\ufe0f"
	matchUnlikeButtonText  = "\U0001F494"
)

type localizedOption struct {
	Label string
	Value string
}

func (s *service) startInlineKeyboardByStatus(l localizer, status profileStatus, role adminRole) *domain.InlineKeyboard {
	if status == profileStatusMissing {
		return s.withAdminMenuInlineKeyboard(l, s.profileCreateInlineKeyboard(l), role)
	}

	return s.withAdminMenuInlineKeyboard(l, s.profileInlineKeyboard(l), role)
}

func (s *service) ageConfirmationInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnAgeConfirmYes),
					CallbackData: domain.CommandAgeConfirmYes,
				},
				{
					Text:         l.button(btnAgeConfirmNo),
					CallbackData: domain.CommandAgeConfirmNo,
				},
			},
		},
	}
}

func (s *service) profileInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return s.profileStartInlineKeyboard(l, l.button(btnProfile), domain.CommandProfileView, true)
}

func (s *service) profileCreateInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return s.profileStartInlineKeyboard(l, l.button(btnCreateProfile), domain.CommandProfile, false)
}

func (s *service) profileStartInlineKeyboard(l localizer, text, callbackData string, showSettings bool) *domain.InlineKeyboard {
	rows := [][]domain.InlineButton{
		{
			{
				Text:         l.button(btnStartSearch),
				CallbackData: domain.CommandSearchStart,
			},
		},
		{
			{
				Text:         l.button(btnStartSearchAI),
				CallbackData: domain.CommandSearchAI,
			},
		},
		{
			{
				Text:         l.button(btnSearchHistory),
				CallbackData: domain.CommandMatchHistory,
			},
		},
		{
			{
				Text:         l.button(btnLikesInbox),
				CallbackData: domain.CommandMatchLikes,
			},
		},
		{
			{
				Text:         text,
				CallbackData: callbackData,
			},
		},
	}

	if showSettings {
		rows = append(rows, []domain.InlineButton{
			{
				Text:         l.button(btnSettings),
				CallbackData: domain.CommandProfileSettings,
			},
		})
	}

	return &domain.InlineKeyboard{
		Buttons: rows,
	}
}

func (s *service) adminMenuInlineKeyboard(l localizer, role adminRole) *domain.InlineKeyboard {
	rows := [][]domain.InlineButton{
		{
			{
				Text:         l.button(btnAdminUsers),
				CallbackData: domain.CommandAdminUsers,
			},
		},
		{
			{
				Text:         l.button(btnAdminBannedUsers),
				CallbackData: domain.CommandAdminBannedUsers,
			},
		},
		{
			{
				Text:         l.button(btnAdminShadowBannedUsers),
				CallbackData: domain.CommandAdminShadowBannedUsers,
			},
		},
		{
			{
				Text:         l.button(btnAdminHiddenUsers),
				CallbackData: domain.CommandAdminHiddenUsers,
			},
		},
	}

	if role.canManageModerators() {
		rows = append(rows, []domain.InlineButton{
			{
				Text:         l.button(btnAdminModerators),
				CallbackData: domain.CommandAdminModerators,
			},
		})
	}

	rows = append(rows,
		[]domain.InlineButton{
			{
				Text:         l.button(btnAdminReports),
				CallbackData: domain.CommandAdminReports,
			},
		},
		[]domain.InlineButton{
			{
				Text:         l.button(btnAdminClearReports),
				CallbackData: domain.CommandAdminClearReports,
			},
		},
		[]domain.InlineButton{
			{
				Text:         l.button(btnAdminBan),
				CallbackData: domain.CommandAdminBan,
			},
		},
		[]domain.InlineButton{
			{
				Text:         l.button(btnAdminUnban),
				CallbackData: domain.CommandAdminUnban,
			},
		},
		[]domain.InlineButton{
			{
				Text:         l.button(btnAdminShadowBan),
				CallbackData: domain.CommandAdminShadowBan,
			},
		},
		[]domain.InlineButton{
			{
				Text:         l.button(btnAdminShadowUnban),
				CallbackData: domain.CommandAdminShadowUnban,
			},
		},
		[]domain.InlineButton{
			{
				Text:         l.button(btnAdminHideProfile),
				CallbackData: domain.CommandAdminHideProfile,
			},
		},
		[]domain.InlineButton{
			{
				Text:         l.button(btnAdminShowProfile),
				CallbackData: domain.CommandAdminShowProfile,
			},
		},
	)

	rows = append(rows, []domain.InlineButton{
		{
			Text:         l.button(btnAdminResetChoices),
			CallbackData: domain.CommandAdminResetChoices,
		},
	})
	rows = append(rows, []domain.InlineButton{
		{
			Text:         l.button(btnAdminResetStart),
			CallbackData: domain.CommandAdminResetStart,
		},
	})

	if role.canManageModerators() {
		rows = append(rows, []domain.InlineButton{
			{
				Text:         l.button(btnAdminPostAd),
				CallbackData: domain.CommandAdminPostAd,
			},
		})
		rows = append(rows,
			[]domain.InlineButton{
				{
					Text:         l.button(btnAdminModerator),
					CallbackData: domain.CommandAdminModerator,
				},
			},
			[]domain.InlineButton{
				{
					Text:         l.button(btnAdminUnmoderator),
					CallbackData: domain.CommandAdminUnmoderator,
				},
			},
		)
	}

	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: rows,
	})
}

func (s *service) adminAdInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return s.adminBanInlineKeyboard(l)
}

func (s *service) adminBanInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnAdminBackToMenu),
					CallbackData: domain.CommandAdminMenu,
				},
			},
		},
	})
}

func (s *service) adminModeratorInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return s.adminBanInlineKeyboard(l)
}

func (s *service) adminResetChoicesInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return s.adminBanInlineKeyboard(l)
}

func (s *service) adminResetStartInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return s.adminBanInlineKeyboard(l)
}

func (s *service) adminClearReportsInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return s.adminBanInlineKeyboard(l)
}

func (s *service) adminUsersInlineKeyboard(l localizer, offset, limit, total int) *domain.InlineKeyboard {
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
				Text:         l.button(btnAdminUsersPrev),
				CallbackData: domain.CommandAdminUsers + ":" + strconv.Itoa(prevOffset),
			})
		}
		if hasNext {
			nextOffset := offset + limit
			row = append(row, domain.InlineButton{
				Text:         l.button(btnAdminUsersNext),
				CallbackData: domain.CommandAdminUsers + ":" + strconv.Itoa(nextOffset),
			})
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	rows = append(rows, []domain.InlineButton{
		{
			Text:         l.button(btnAdminBackToMenu),
			CallbackData: domain.CommandAdminMenu,
		},
	})

	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{Buttons: rows})
}

func (s *service) adminBannedUsersInlineKeyboard(l localizer, offset, limit, total int) *domain.InlineKeyboard {
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
				Text:         l.button(btnAdminUsersPrev),
				CallbackData: domain.CommandAdminBannedUsers + ":" + strconv.Itoa(prevOffset),
			})
		}
		if hasNext {
			nextOffset := offset + limit
			row = append(row, domain.InlineButton{
				Text:         l.button(btnAdminUsersNext),
				CallbackData: domain.CommandAdminBannedUsers + ":" + strconv.Itoa(nextOffset),
			})
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	rows = append(rows, []domain.InlineButton{
		{
			Text:         l.button(btnAdminBackToMenu),
			CallbackData: domain.CommandAdminMenu,
		},
	})

	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{Buttons: rows})
}

func (s *service) adminShadowBannedUsersInlineKeyboard(l localizer, offset, limit, total int) *domain.InlineKeyboard {
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
				Text:         l.button(btnAdminUsersPrev),
				CallbackData: domain.CommandAdminShadowBannedUsers + ":" + strconv.Itoa(prevOffset),
			})
		}
		if hasNext {
			nextOffset := offset + limit
			row = append(row, domain.InlineButton{
				Text:         l.button(btnAdminUsersNext),
				CallbackData: domain.CommandAdminShadowBannedUsers + ":" + strconv.Itoa(nextOffset),
			})
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	rows = append(rows, []domain.InlineButton{
		{
			Text:         l.button(btnAdminBackToMenu),
			CallbackData: domain.CommandAdminMenu,
		},
	})

	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{Buttons: rows})
}

func (s *service) adminHiddenUsersInlineKeyboard(l localizer, offset, limit, total int) *domain.InlineKeyboard {
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
				Text:         l.button(btnAdminUsersPrev),
				CallbackData: domain.CommandAdminHiddenUsers + ":" + strconv.Itoa(prevOffset),
			})
		}
		if hasNext {
			nextOffset := offset + limit
			row = append(row, domain.InlineButton{
				Text:         l.button(btnAdminUsersNext),
				CallbackData: domain.CommandAdminHiddenUsers + ":" + strconv.Itoa(nextOffset),
			})
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	rows = append(rows, []domain.InlineButton{
		{
			Text:         l.button(btnAdminBackToMenu),
			CallbackData: domain.CommandAdminMenu,
		},
	})

	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{Buttons: rows})
}

func (s *service) adminModeratorsInlineKeyboard(l localizer, offset, limit, total int) *domain.InlineKeyboard {
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
				Text:         l.button(btnAdminUsersPrev),
				CallbackData: domain.CommandAdminModerators + ":" + strconv.Itoa(prevOffset),
			})
		}
		if hasNext {
			nextOffset := offset + limit
			row = append(row, domain.InlineButton{
				Text:         l.button(btnAdminUsersNext),
				CallbackData: domain.CommandAdminModerators + ":" + strconv.Itoa(nextOffset),
			})
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	rows = append(rows, []domain.InlineButton{
		{
			Text:         l.button(btnAdminBackToMenu),
			CallbackData: domain.CommandAdminMenu,
		},
	})

	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{Buttons: rows})
}

func (s *service) adminReportsInlineKeyboard(l localizer, offset, limit, total int) *domain.InlineKeyboard {
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
				Text:         l.button(btnAdminUsersPrev),
				CallbackData: domain.CommandAdminReports + ":" + strconv.Itoa(prevOffset),
			})
		}
		if hasNext {
			nextOffset := offset + limit
			row = append(row, domain.InlineButton{
				Text:         l.button(btnAdminUsersNext),
				CallbackData: domain.CommandAdminReports + ":" + strconv.Itoa(nextOffset),
			})
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	rows = append(rows, []domain.InlineButton{
		{
			Text:         l.button(btnAdminBackToMenu),
			CallbackData: domain.CommandAdminMenu,
		},
	})

	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{Buttons: rows})
}

func (s *service) adminUnbanInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return s.adminBanInlineKeyboard(l)
}

func (s *service) adminShadowBanInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return s.adminBanInlineKeyboard(l)
}

func (s *service) adminShadowUnbanInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return s.adminBanInlineKeyboard(l)
}

func (s *service) adminHideProfileInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return s.adminBanInlineKeyboard(l)
}

func (s *service) adminShowProfileInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return s.adminBanInlineKeyboard(l)
}

func (s *service) adminUnmoderatorInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return s.adminBanInlineKeyboard(l)
}

func (s *service) withAdminMenuInlineKeyboard(l localizer, keyboard *domain.InlineKeyboard, role adminRole) *domain.InlineKeyboard {
	if s == nil || !role.canAccessPanel() {
		return keyboard
	}
	if keyboard == nil {
		keyboard = &domain.InlineKeyboard{}
	}
	if inlineKeyboardHasCallback(keyboard, domain.CommandAdminMenu) ||
		inlineKeyboardHasText(keyboard, l.button(btnAdminMenu)) ||
		inlineKeyboardHasText(keyboard, l.button(btnModeratorMenu)) {
		return keyboard
	}

	keyboard.Buttons = append(keyboard.Buttons, []domain.InlineButton{
		{
			Text:         adminMenuButtonTextForRole(l, role),
			CallbackData: domain.CommandAdminMenu,
		},
	})

	return keyboard
}

func adminMenuButtonTextForRole(l localizer, role adminRole) string {
	if role == adminRoleModerator {
		return l.button(btnModeratorMenu)
	}
	return l.button(btnAdminMenu)
}

func (s *service) profileViewInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnPreviewProfile),
					CallbackData: domain.CommandProfilePreview,
				},
			},
			{
				{
					Text:         l.button(btnEditProfile),
					CallbackData: domain.CommandProfileEdit,
				},
			},
		},
	})
}

func (s *service) profilePreviewInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnBackToProfile),
					CallbackData: domain.CommandProfileView,
				},
			},
			{
				{
					Text:         l.button(btnEditProfile),
					CallbackData: domain.CommandProfileEdit,
				},
			},
		},
	})
}

func (s *service) profileSettingsInlineKeyboard(l localizer, isHidden bool, searchAccuracyEnabled bool, likesNotificationsEnabled bool) *domain.InlineKeyboard {
	visibilityText := l.button(btnHideFromSearch)
	visibilityAction := profileVisibilityHideAction
	if isHidden {
		visibilityText = l.button(btnShowInSearch)
		visibilityAction = profileVisibilityShowAction
	}

	searchAccuracyText := l.button(btnAdvancedSearchStatus, s.searchAccuracyStatus(l, searchAccuracyEnabled))
	searchAccuracyAction := searchAccuracyEnableAction
	if searchAccuracyEnabled {
		searchAccuracyAction = searchAccuracyDisableAction
	}

	notificationsText := l.button(btnLikesNotificationsStatus, s.likeNotificationsStatus(l, likesNotificationsEnabled))
	notificationsAction := likeNotificationsEnableAction
	if likesNotificationsEnabled {
		notificationsAction = likeNotificationsDisableAction
	}

	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         visibilityText,
					CallbackData: domain.CommandProfileVisibility + ":" + visibilityAction,
				},
			},
			{
				{
					Text:         searchAccuracyText,
					CallbackData: domain.CommandProfileSearchAccuracy + ":" + searchAccuracyAction,
				},
			},
			{
				{
					Text:         notificationsText,
					CallbackData: domain.CommandProfileLikeNotifications + ":" + notificationsAction,
				},
			},
			{
				{
					Text:         l.button(btnLanguage),
					CallbackData: domain.CommandProfileLanguage,
				},
			},
			{
				{
					Text:         l.button(btnDeleteProfile),
					CallbackData: domain.CommandProfileDelete,
				},
			},
		},
	})
}

func (s *service) languageInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnLanguageEnglish),
					CallbackData: domain.CommandProfileLanguage + ":" + string(domain.LanguageEnglish),
				},
				{
					Text:         l.button(btnLanguageRussian),
					CallbackData: domain.CommandProfileLanguage + ":" + string(domain.LanguageRussian),
				},
			},
			{
				{
					Text:         l.button(btnBackToSettings),
					CallbackData: domain.CommandProfileSettings,
				},
			},
		},
	})
}

func (s *service) profileDeleteConfirmInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnDeleteConfirm),
					CallbackData: domain.CommandProfileDeleteConfirm,
				},
			},
			{
				{
					Text:         l.button(btnProfileDeleteCancel),
					CallbackData: domain.CommandProfileDeleteCancel,
				},
			},
		},
	}
}

func (s *service) profileEditMenuKeyboard(l localizer) *domain.InlineKeyboard {
	return withBackToProfileInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnProfileEditName),
					CallbackData: domain.CommandProfileEditName,
				},
				{
					Text:         l.button(btnProfileEditGender),
					CallbackData: domain.CommandProfileEditGender,
				},
			},
			{
				{
					Text:         l.button(btnProfileEditBirthDate),
					CallbackData: domain.CommandProfileEditBirthDate,
				},
				{
					Text:         l.button(btnProfileEditCountry),
					CallbackData: domain.CommandProfileEditCountry,
				},
			},
			{
				{
					Text:         l.button(btnProfileEditCity),
					CallbackData: domain.CommandProfileEditCity,
				},
				{
					Text:         l.button(btnProfileEditDescription),
					CallbackData: domain.CommandProfileEditDesc,
				},
			},
			{
				{
					Text:         l.button(btnProfileEditEmoji),
					CallbackData: domain.CommandProfileEditEmoji,
				},
				{
					Text:         l.button(btnProfileEditPhotos),
					CallbackData: domain.CommandProfileEditPhotos,
				},
			},
		},
	})
}

func (s *service) telegramNameInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnTelegramName),
					CallbackData: telegramNameCallbackData,
				},
			},
		},
	})
}

func (s *service) photosDoneInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnAlbumDone),
					CallbackData: albumDoneCallbackData,
				},
			},
		},
	})
}

func (s *service) botCheckInlineKeyboard(l localizer, answer int) *domain.InlineKeyboard {
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

	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{Buttons: rows})
}

func (s *service) genderInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnGenderMale),
					CallbackData: string(domain.GenderMale),
				},
				{
					Text:         l.button(btnGenderFemale),
					CallbackData: string(domain.GenderFemale),
				},
				{
					Text:         l.button(btnGenderOther),
					CallbackData: string(domain.GenderOther),
				},
			},
		},
	})
}

func (s *service) countryInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnCountryRussia),
					CallbackData: string(domain.CountryRussia),
				},
				{
					Text:         l.button(btnCountryKazakhstan),
					CallbackData: string(domain.CountryKazakhstan),
				},
				{
					Text:         l.button(btnCountryBelarus),
					CallbackData: string(domain.CountryBelarus),
				},
			},
		},
	})
}

func (s *service) cityInlineKeyboard(l localizer, country domain.Country) *domain.InlineKeyboard {
	options := s.cityOptionsLocalized(l, country)
	if len(options) == 0 {
		return nil
	}

	return withCancelInlineKeyboard(l, listInlineKeyboardWithValues(options, cityKeyboardColumns))
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

func (s *service) searchGenderInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnSearchMen),
					CallbackData: domain.CommandSearchGender + ":" + string(domain.GenderMale),
				},
				{
					Text:         l.button(btnSearchWomen),
					CallbackData: domain.CommandSearchGender + ":" + string(domain.GenderFemale),
				},
			},
			{
				{
					Text:         l.button(btnSearchOther),
					CallbackData: domain.CommandSearchGender + ":" + string(domain.GenderOther),
				},
				{
					Text:         l.button(btnSearchAny),
					CallbackData: domain.CommandSearchGender + ":" + string(domain.GenderUnspecified),
				},
			},
		},
	})
}

func listInlineKeyboardWithValues(options []localizedOption, columns int) *domain.InlineKeyboard {
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
		label := strings.TrimSpace(option.Label)
		if label == "" {
			label = option.Value
		}
		rows[rowIndex] = append(rows[rowIndex], domain.InlineButton{
			Text:         label,
			CallbackData: option.Value,
		})
	}

	return &domain.InlineKeyboard{Buttons: rows}
}

func (s *service) searchAccuracyInlineKeyboard(l localizer, gender domain.Gender) *domain.InlineKeyboard {
	prefix := domain.CommandSearchAccuracy + ":" + string(gender) + ":"
	total := searchAccuracyMax - searchAccuracyMin + 1
	if total <= 0 {
		return withSearchGenderCancelInlineKeyboard(l, nil)
	}

	columns := searchAccuracyColumns
	if columns <= 0 {
		columns = 1
	}

	rows := make([][]domain.InlineButton, 0, (total+columns-1)/columns)
	index := 0
	for accuracy := searchAccuracyMax; accuracy >= searchAccuracyMin; accuracy-- {
		if index%columns == 0 {
			rows = append(rows, []domain.InlineButton{})
		}
		label := strconv.Itoa(accuracy)
		rowIndex := len(rows) - 1
		rows[rowIndex] = append(rows[rowIndex], domain.InlineButton{
			Text:         label,
			CallbackData: prefix + label,
		})
		index++
	}

	return withSearchGenderCancelInlineKeyboard(l, &domain.InlineKeyboard{Buttons: rows})
}

func (s *service) searchNoCandidatesInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnSearchRefresh),
					CallbackData: domain.CommandSearchRefresh,
				},
			},
		},
	})
}

func (s *service) searchAIUnavailableInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnSearchAIRepeat),
					CallbackData: domain.CommandSearchAI,
				},
			},
		},
	})
}

func (s *service) matchActionsInlineKeyboard(l localizer, targetID int64, hasPrevious bool) *domain.InlineKeyboard {
	target := strconv.FormatInt(targetID, 10)
	buttons := [][]domain.InlineButton{
		{
			{
				Text:         matchDislikeButtonText,
				CallbackData: domain.CommandMatchDislike + ":" + target,
			},
			{
				Text:         matchLikeButtonText,
				CallbackData: domain.CommandMatchLike + ":" + target,
			},
		},
	}
	buttons = append(buttons, []domain.InlineButton{
		{
			Text:         l.button(btnReport),
			CallbackData: domain.CommandMatchReport + ":" + target,
		},
		{
			Text:         l.button(btnSearchHistory),
			CallbackData: domain.CommandMatchHistory,
		},
	})
	if hasPrevious {
		buttons = append(buttons, []domain.InlineButton{
			{
				Text:         l.button(btnBackToPreviousProfile),
				CallbackData: domain.CommandMatchPrevious,
			},
		})
	}

	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{Buttons: buttons})
}

func (s *service) matchViewActionsInlineKeyboard(l localizer, targetID int64) *domain.InlineKeyboard {
	target := strconv.FormatInt(targetID, 10)
	buttons := [][]domain.InlineButton{
		{
			{
				Text:         matchDislikeButtonText,
				CallbackData: domain.CommandMatchViewDislike + ":" + target,
			},
			{
				Text:         matchLikeButtonText,
				CallbackData: domain.CommandMatchViewLike + ":" + target,
			},
		},
		{
			{
				Text:         l.button(btnReport),
				CallbackData: domain.CommandMatchViewReport + ":" + target,
			},
		},
	}

	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{Buttons: buttons})
}

func (s *service) matchViewInlineKeyboard(l localizer, targetID int64) *domain.InlineKeyboard {
	target := strconv.FormatInt(targetID, 10)
	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnViewProfile),
					CallbackData: domain.CommandMatchViewProfile + ":" + target,
				},
			},
		},
	})
}

func (s *service) historyInlineKeyboard(l localizer, list domain.MatchHistoryList) *domain.InlineKeyboard {
	rows := make([][]domain.InlineButton, 0, len(list.Items)+2)
	for i, item := range list.Items {
		index := list.Offset + i + 1
		target := strconv.FormatInt(item.Profile.UserID, 10)
		args := target + ":" + strconv.Itoa(list.Offset)
		rows = append(rows, []domain.InlineButton{
			{
				Text:         s.historyItemButtonLabel(l, item.Profile, index),
				CallbackData: domain.CommandMatchHistoryView + ":" + args,
			},
		})
	}

	hasPrev := list.Offset > 0
	hasNext := list.Limit > 0 && (list.Offset+list.Limit) < list.Total
	if hasPrev || hasNext {
		row := []domain.InlineButton{}
		if hasPrev {
			prevOffset := list.Offset - list.Limit
			if prevOffset < 0 {
				prevOffset = 0
			}
			row = append(row, domain.InlineButton{
				Text:         l.button(btnSearchHistoryPrev),
				CallbackData: domain.CommandMatchHistory + ":" + strconv.Itoa(prevOffset),
			})
		}
		if hasNext {
			nextOffset := list.Offset + list.Limit
			row = append(row, domain.InlineButton{
				Text:         l.button(btnSearchHistoryNext),
				CallbackData: domain.CommandMatchHistory + ":" + strconv.Itoa(nextOffset),
			})
		}
		rows = append(rows, row)
	}

	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{Buttons: rows})
}

func (s *service) likesInlineKeyboard(l localizer, list domain.MatchLikesList) *domain.InlineKeyboard {
	rows := make([][]domain.InlineButton, 0, len(list.Items)+2)
	for i, profile := range list.Items {
		index := list.Offset + i + 1
		target := strconv.FormatInt(profile.UserID, 10)
		args := target + ":" + strconv.Itoa(list.Offset)
		rows = append(rows, []domain.InlineButton{
			{
				Text:         s.historyItemButtonLabel(l, profile, index),
				CallbackData: domain.CommandMatchLikesView + ":" + args,
			},
		})
	}

	hasPrev := list.Offset > 0
	hasNext := list.Limit > 0 && (list.Offset+list.Limit) < list.Total
	if hasPrev || hasNext {
		row := []domain.InlineButton{}
		if hasPrev {
			prevOffset := list.Offset - list.Limit
			if prevOffset < 0 {
				prevOffset = 0
			}
			row = append(row, domain.InlineButton{
				Text:         l.button(btnSearchHistoryPrev),
				CallbackData: domain.CommandMatchLikes + ":" + strconv.Itoa(prevOffset),
			})
		}
		if hasNext {
			nextOffset := list.Offset + list.Limit
			row = append(row, domain.InlineButton{
				Text:         l.button(btnSearchHistoryNext),
				CallbackData: domain.CommandMatchLikes + ":" + strconv.Itoa(nextOffset),
			})
		}
		rows = append(rows, row)
	}

	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{Buttons: rows})
}

func (s *service) historyActionsInlineKeyboard(l localizer, targetID int64, offset int, action domain.MatchAction) *domain.InlineKeyboard {
	target := strconv.FormatInt(targetID, 10)
	offsetValue := strconv.Itoa(offset)
	likeText := matchLikeButtonText
	likeCommand := domain.CommandMatchHistoryLike
	if action == domain.MatchActionLike {
		likeText = matchUnlikeButtonText
		likeCommand = domain.CommandMatchHistoryDislike
	}
	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         matchDislikeButtonText,
					CallbackData: domain.CommandMatchHistoryDislike + ":" + target + ":" + offsetValue,
				},
				{
					Text:         likeText,
					CallbackData: likeCommand + ":" + target + ":" + offsetValue,
				},
			},
			{
				{
					Text:         l.button(btnReport),
					CallbackData: domain.CommandMatchHistoryReport + ":" + target + ":" + offsetValue,
				},
				{
					Text:         l.button(btnSearchHistoryBack),
					CallbackData: domain.CommandMatchHistory + ":" + offsetValue,
				},
			},
		},
	})
}

func (s *service) likesActionsInlineKeyboard(l localizer, targetID int64, offset int) *domain.InlineKeyboard {
	target := strconv.FormatInt(targetID, 10)
	offsetValue := strconv.Itoa(offset)
	return withCancelInlineKeyboard(l, &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         matchDislikeButtonText,
					CallbackData: domain.CommandMatchLikesDislike + ":" + target + ":" + offsetValue,
				},
				{
					Text:         matchLikeButtonText,
					CallbackData: domain.CommandMatchLikesLike + ":" + target + ":" + offsetValue,
				},
			},
			{
				{
					Text:         l.button(btnReport),
					CallbackData: domain.CommandMatchLikesReport + ":" + target + ":" + offsetValue,
				},
				{
					Text:         l.button(btnLikesBack),
					CallbackData: domain.CommandMatchLikes + ":" + offsetValue,
				},
			},
		},
	})
}

func withDraftCancelInlineKeyboard(l localizer, keyboard *domain.InlineKeyboard, mode domain.ProfileDraftMode) *domain.InlineKeyboard {
	if mode == domain.ProfileDraftModeEdit {
		return withEditCancelInlineKeyboard(l, withoutCancelInlineKeyboard(l, keyboard))
	}
	return withCancelInlineKeyboard(l, keyboard)
}

func withProfileSetupBackInlineKeyboard(l localizer, keyboard *domain.InlineKeyboard) *domain.InlineKeyboard {
	if keyboard == nil {
		return profileSetupBackInlineKeyboard(l)
	}
	if inlineKeyboardHasCallback(keyboard, domain.CommandProfileSetupBack) || inlineKeyboardHasText(keyboard, l.button(btnProfileSetupBack)) {
		return keyboard
	}

	backRow := []domain.InlineButton{
		{
			Text:         l.button(btnProfileSetupBack),
			CallbackData: domain.CommandProfileSetupBack,
		},
	}

	insertAt := len(keyboard.Buttons)
	for i, row := range keyboard.Buttons {
		if inlineKeyboardRowHasCancel(l, row) {
			insertAt = i
			break
		}
	}

	keyboard.Buttons = append(keyboard.Buttons[:insertAt], append([][]domain.InlineButton{backRow}, keyboard.Buttons[insertAt:]...)...)
	return keyboard
}

func withSearchGenderCancelInlineKeyboard(l localizer, keyboard *domain.InlineKeyboard) *domain.InlineKeyboard {
	keyboard = withoutCancelInlineKeyboard(l, keyboard)
	if keyboard == nil {
		return searchGenderCancelInlineKeyboard(l)
	}
	if inlineKeyboardHasCallback(keyboard, domain.CommandSearchGender) || inlineKeyboardHasText(keyboard, l.button(btnSearchAccuracyCancel)) {
		return keyboard
	}

	keyboard.Buttons = append(keyboard.Buttons, []domain.InlineButton{
		{
			Text:         l.button(btnSearchAccuracyCancel),
			CallbackData: domain.CommandSearchGender,
		},
	})

	return keyboard
}

func withCancelInlineKeyboard(l localizer, keyboard *domain.InlineKeyboard) *domain.InlineKeyboard {
	if keyboard == nil {
		return cancelInlineKeyboard(l)
	}
	if inlineKeyboardHasCallback(keyboard, domain.CommandCancel) || inlineKeyboardHasText(keyboard, l.button(btnCancel)) {
		return keyboard
	}

	keyboard.Buttons = append(keyboard.Buttons, []domain.InlineButton{
		{
			Text:         l.button(btnCancel),
			CallbackData: domain.CommandCancel,
		},
	})

	return keyboard
}

func withEditCancelInlineKeyboard(l localizer, keyboard *domain.InlineKeyboard) *domain.InlineKeyboard {
	if keyboard == nil {
		return editCancelInlineKeyboard(l)
	}
	if inlineKeyboardHasCallback(keyboard, domain.CommandProfileEdit) || inlineKeyboardHasText(keyboard, l.button(btnEditCancel)) {
		return keyboard
	}

	keyboard.Buttons = append(keyboard.Buttons, []domain.InlineButton{
		{
			Text:         l.button(btnEditCancel),
			CallbackData: domain.CommandProfileEdit,
		},
	})

	return keyboard
}

func withBackToProfileInlineKeyboard(l localizer, keyboard *domain.InlineKeyboard) *domain.InlineKeyboard {
	if keyboard == nil {
		return backToProfileInlineKeyboard(l)
	}
	if inlineKeyboardHasCallback(keyboard, domain.CommandProfileView) || inlineKeyboardHasText(keyboard, l.button(btnBackToProfile)) {
		return keyboard
	}

	keyboard.Buttons = append(keyboard.Buttons, []domain.InlineButton{
		{
			Text:         l.button(btnBackToProfile),
			CallbackData: domain.CommandProfileView,
		},
	})

	return keyboard
}

func cancelInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnCancel),
					CallbackData: domain.CommandCancel,
				},
			},
		},
	}
}

func editCancelInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnEditCancel),
					CallbackData: domain.CommandProfileEdit,
				},
			},
		},
	}
}

func searchGenderCancelInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnSearchAccuracyCancel),
					CallbackData: domain.CommandSearchGender,
				},
			},
		},
	}
}

func backToProfileInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnBackToProfile),
					CallbackData: domain.CommandProfileView,
				},
			},
		},
	}
}

func profileSetupBackInlineKeyboard(l localizer) *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         l.button(btnProfileSetupBack),
					CallbackData: domain.CommandProfileSetupBack,
				},
			},
		},
	}
}

func withoutCancelInlineKeyboard(l localizer, keyboard *domain.InlineKeyboard) *domain.InlineKeyboard {
	if keyboard == nil {
		return nil
	}

	rows := make([][]domain.InlineButton, 0, len(keyboard.Buttons))
	for _, row := range keyboard.Buttons {
		filtered := make([]domain.InlineButton, 0, len(row))
		for _, button := range row {
			if button.CallbackData == domain.CommandCancel || button.Text == l.button(btnCancel) {
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

func (s *service) historyItemButtonLabel(l localizer, profile domain.Profile, index int) string {
	name := strings.TrimSpace(profile.Name)
	if name == "" {
		name = l.message(msgSearchHistoryUserFallback, profile.UserID)
	}
	if historyLabelMaxRunes > 0 {
		name = truncateRunes(name, historyLabelMaxRunes)
	}
	if index <= 0 {
		return name
	}
	return l.message(msgSearchHistoryLineShort, index, name)
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "..."
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

func inlineKeyboardRowHasCancel(l localizer, row []domain.InlineButton) bool {
	for _, button := range row {
		if button.CallbackData == domain.CommandCancel || button.Text == l.button(btnCancel) {
			return true
		}
	}

	return false
}
