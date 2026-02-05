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
	domain.CountryUkraine: {
		"Kyiv",
		"Kharkiv",
		"Odesa",
		"Dnipro",
		"Donetsk",
		"Lviv",
		"Zaporizhzhia",
		"Kryvyi Rih",
		"Mykolaiv",
		"Luhansk",
		"Mariupol",
		"Kherson",
		"Vinnytsia",
		"Poltava",
		"Chernihiv",
		"Cherkasy",
		"Zhytomyr",
		"Sumy",
		"Khmelnytskyi",
		"Rivne",
		"Ternopil",
		"Ivano-Frankivsk",
		"Lutsk",
		"Chernivtsi",
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

	labelSets := []map[string]string{
		cityLabels[domain.LanguageRussian],
		cityLabels[domain.LanguageUkrainian],
	}
	for _, labels := range labelSets {
		if len(labels) == 0 {
			continue
		}
		for _, option := range options {
			if label, ok := labels[option]; ok && strings.EqualFold(label, value) {
				return option, true
			}
		}
	}

	return "", false
}
