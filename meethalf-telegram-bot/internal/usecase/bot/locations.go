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

	return "", false
}
