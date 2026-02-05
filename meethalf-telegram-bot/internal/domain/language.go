package domain

type Language string

const (
	LanguageEnglish   Language = "en"
	LanguageRussian   Language = "ru"
	LanguageUkrainian Language = "uk"
	DefaultLanguage   Language = LanguageEnglish
)
