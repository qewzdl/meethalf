package bot

import (
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

const cityKeyboardColumns = 2

var countryCities = map[domain.Country][]string{
	domain.CountryRussia: {
		"Moscow",
		"Saint Petersburg",
		"Novosibirsk",
		"Krasnodar",
		"Omsk",
		"Rostov-on-Don",
		"Perm",
		"Krasnoyarsk",
		"Yekaterinburg",
		"Kazan",
		"Nizhny Novgorod",
		"Ufa",
		"Chelyabinsk",
		"Samara",
		"Voronezh",
		"Volgograd",
	},
	domain.CountryKazakhstan: {
		"Astana",
		"Almaty",
		"Semey",
		"Pavlodar",
		"Shymkent",
		"Aktobe",
		"Karaganda",
		"Taraz",
		"Ust-Kamenogorsk",
		"Atyrau",
	},
	domain.CountryBelarus: {
		"Minsk",
		"Gomel",
		"Mogilev",
		"Vitebsk",
		"Grodno",
		"Brest",
		"Bobruisk",
		"Baranovichi",
		"Borisov",
	},
}

func (s *service) cityOptions(country domain.Country) []string {
	return countryCities[country]
}

func (s *service) cityOptionsLocalized(l localizer, country domain.Country) []localizedOption {
	options := countryCities[country]
	if len(options) == 0 {
		return nil
	}

	localized := make([]localizedOption, 0, len(options))
	for _, city := range options {
		label := l.cityLabel(city)
		if strings.TrimSpace(label) == "" {
			label = city
		}
		localized = append(localized, localizedOption{
			Label: label,
			Value: city,
		})
	}

	return localized
}

func (s *service) normalizeCity(country domain.Country, value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxCityLength {
		return "", false
	}

	options := countryCities[country]
	for _, option := range options {
		if strings.EqualFold(option, value) {
			return option, true
		}
	}

	ruLabels := cityLabels[domain.LanguageRussian]
	if len(ruLabels) == 0 {
		return "", false
	}
	for _, option := range options {
		if label, ok := ruLabels[option]; ok && strings.EqualFold(label, value) {
			return option, true
		}
	}

	return "", false
}
