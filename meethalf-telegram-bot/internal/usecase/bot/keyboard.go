package bot

import "meethalf-telegram-bot/internal/domain"

func (s *service) startInlineKeyboardByStatus(status profileStatus) *domain.InlineKeyboard {
	if status == profileStatusMissing {
		return s.profileCreateInlineKeyboard()
	}

	return s.profileInlineKeyboard()
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

func (s *service) profileViewInlineKeyboard() *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
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
	}
}

func (s *service) profilePreviewInlineKeyboard() *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         "Back to profile",
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
	}
}

func (s *service) profileSettingsInlineKeyboard() *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         "Delete profile",
					CallbackData: domain.CommandProfileDelete,
				},
			},
		},
	}
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
					Text:         "Cancel",
					CallbackData: domain.CommandProfileDeleteCancel,
				},
			},
		},
	}
}

func (s *service) profileEditMenuKeyboard() *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
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
	}
}

func (s *service) telegramNameInlineKeyboard() *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         telegramNameButtonText,
					CallbackData: telegramNameCallbackData,
				},
			},
		},
	}
}

func (s *service) photosDoneInlineKeyboard() *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
		Buttons: [][]domain.InlineButton{
			{
				{
					Text:         albumDoneButtonText,
					CallbackData: albumDoneCallbackData,
				},
			},
		},
	}
}

func (s *service) genderInlineKeyboard() *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
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
	}
}

func (s *service) countryInlineKeyboard() *domain.InlineKeyboard {
	return &domain.InlineKeyboard{
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
	}
}

func (s *service) cityInlineKeyboard(country domain.Country) *domain.InlineKeyboard {
	options := s.cityOptions(country)
	if len(options) == 0 {
		return nil
	}

	return listInlineKeyboard(options, cityKeyboardColumns)
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
