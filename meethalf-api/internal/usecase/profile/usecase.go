package profile

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"meethalf-api/internal/domain"
)

const (
	minAge            = 1
	maxAge            = 120
	maxNameLength     = 64
	maxCityLength     = 64
	maxDescriptionLen = 500
	minPhotos         = 1
	maxPhotos         = 4
)

var (
	ErrInvalidUserID      = errors.New("user id is required")
	ErrInvalidName        = errors.New("profile name is required")
	ErrInvalidGender      = errors.New("profile gender is invalid")
	ErrInvalidBirthDate   = errors.New("profile birth date must result in age between 1 and 120")
	ErrInvalidCountry     = errors.New("profile country is invalid")
	ErrInvalidCity        = errors.New("profile city is invalid")
	ErrInvalidDescription = errors.New("profile description is required")
	ErrInvalidEmojiCode   = errors.New("profile emoji code is invalid")
	ErrInvalidPhotos      = errors.New("profile photos must include 1 to 4 entries")
	ErrProfileNotFound    = errors.New("profile not found")
)

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

type Usecase interface {
	Upsert(ctx context.Context, profile domain.Profile) (domain.Profile, error)
	GetByUserID(ctx context.Context, userID int64) (domain.Profile, error)
	DeleteByUserID(ctx context.Context, userID int64) error
	UpdateVisibility(ctx context.Context, userID int64, isHidden bool) error
}

type Repository interface {
	Upsert(ctx context.Context, profile domain.Profile) (domain.Profile, error)
	GetByUserID(ctx context.Context, userID int64) (domain.Profile, error)
	DeleteByUserID(ctx context.Context, userID int64) error
	UpdateVisibility(ctx context.Context, userID int64, isHidden bool) error
}

type service struct {
	repo Repository
}

func New(repo Repository) Usecase {
	return &service{repo: repo}
}

func (s *service) Upsert(ctx context.Context, profile domain.Profile) (domain.Profile, error) {
	if s == nil || s.repo == nil {
		return domain.Profile{}, errors.New("profile repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return domain.Profile{}, err
	}

	normalized, err := normalize(profile)
	if err != nil {
		return domain.Profile{}, err
	}

	return s.repo.Upsert(ctx, normalized)
}

func (s *service) GetByUserID(ctx context.Context, userID int64) (domain.Profile, error) {
	if s == nil || s.repo == nil {
		return domain.Profile{}, errors.New("profile repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return domain.Profile{}, err
	}

	if userID <= 0 {
		return domain.Profile{}, ErrInvalidUserID
	}

	stored, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Profile{}, ErrProfileNotFound
		}
		return domain.Profile{}, err
	}

	stored.BirthDate = normalizeBirthDateValue(stored.BirthDate)
	if !stored.BirthDate.IsZero() {
		stored.Age = ageFromBirthDate(stored.BirthDate, time.Now().UTC())
	}

	return stored, nil
}

func (s *service) DeleteByUserID(ctx context.Context, userID int64) error {
	if s == nil || s.repo == nil {
		return errors.New("profile repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if userID <= 0 {
		return ErrInvalidUserID
	}

	if err := s.repo.DeleteByUserID(ctx, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrProfileNotFound
		}
		return err
	}

	return nil
}

func (s *service) UpdateVisibility(ctx context.Context, userID int64, isHidden bool) error {
	if s == nil || s.repo == nil {
		return errors.New("profile repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if userID <= 0 {
		return ErrInvalidUserID
	}

	if err := s.repo.UpdateVisibility(ctx, userID, isHidden); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrProfileNotFound
		}
		return err
	}

	return nil
}

func normalize(profile domain.Profile) (domain.Profile, error) {
	if profile.UserID <= 0 {
		return domain.Profile{}, ErrInvalidUserID
	}

	name := strings.TrimSpace(profile.Name)
	if name == "" {
		return domain.Profile{}, ErrInvalidName
	}
	if len(name) > maxNameLength {
		return domain.Profile{}, ErrInvalidName
	}

	gender, err := normalizeGender(profile.Gender)
	if err != nil {
		return domain.Profile{}, err
	}

	birthDate, age, err := normalizeBirthDate(profile.BirthDate)
	if err != nil {
		return domain.Profile{}, err
	}

	country, err := normalizeCountry(profile.Country)
	if err != nil {
		return domain.Profile{}, err
	}

	city, err := normalizeCity(country, profile.City)
	if err != nil {
		return domain.Profile{}, err
	}

	description := strings.TrimSpace(profile.Description)
	if description == "" {
		return domain.Profile{}, ErrInvalidDescription
	}
	if len(description) > maxDescriptionLen {
		return domain.Profile{}, ErrInvalidDescription
	}

	emojiCode, err := normalizeEmojiCode(profile.EmojiCode)
	if err != nil {
		return domain.Profile{}, err
	}

	profile.Name = name
	profile.Gender = gender
	profile.BirthDate = birthDate
	profile.Age = age
	profile.Country = country
	profile.City = city
	profile.Description = description
	profile.EmojiCode = emojiCode

	photos := make([]string, 0, len(profile.Photos))
	for _, photo := range profile.Photos {
		photo = strings.TrimSpace(photo)
		if photo == "" {
			return domain.Profile{}, ErrInvalidPhotos
		}
		photos = append(photos, photo)
	}
	if len(photos) < minPhotos || len(photos) > maxPhotos {
		return domain.Profile{}, ErrInvalidPhotos
	}

	profile.Photos = photos
	return profile, nil
}

func normalizeEmojiCode(code domain.ProfileEmojiCode) (domain.ProfileEmojiCode, error) {
	normalized := strings.ToUpper(strings.TrimSpace(string(code)))
	if normalized == "" {
		return "", ErrInvalidEmojiCode
	}

	switch domain.ProfileEmojiCode(normalized) {
	case domain.ProfileEmojiLeader,
		domain.ProfileEmojiStrategist,
		domain.ProfileEmojiAnalyst,
		domain.ProfileEmojiCreator,
		domain.ProfileEmojiCommunicator,
		domain.ProfileEmojiEmpath,
		domain.ProfileEmojiMediator,
		domain.ProfileEmojiPerfectionist,
		domain.ProfileEmojiResearcher,
		domain.ProfileEmojiInnovator,
		domain.ProfileEmojiExecutor,
		domain.ProfileEmojiAdventurer,
		domain.ProfileEmojiContemplator,
		domain.ProfileEmojiRealist,
		domain.ProfileEmojiMotivator,
		domain.ProfileEmojiSkeptic:
		return domain.ProfileEmojiCode(normalized), nil
	default:
		return "", ErrInvalidEmojiCode
	}
}

func normalizeGender(gender domain.Gender) (domain.Gender, error) {
	normalized := strings.ToLower(strings.TrimSpace(string(gender)))
	if normalized == "" {
		return domain.GenderUnspecified, nil
	}

	switch domain.Gender(normalized) {
	case domain.GenderMale, domain.GenderFemale, domain.GenderOther, domain.GenderUnspecified:
		return domain.Gender(normalized), nil
	default:
		return "", ErrInvalidGender
	}
}

func normalizeCountry(country domain.Country) (domain.Country, error) {
	normalized := strings.ToLower(strings.TrimSpace(string(country)))
	if normalized == "" {
		return "", ErrInvalidCountry
	}

	switch domain.Country(normalized) {
	case domain.CountryRussia, domain.CountryKazakhstan, domain.CountryBelarus:
		return domain.Country(normalized), nil
	default:
		return "", ErrInvalidCountry
	}
}

func normalizeCity(country domain.Country, city string) (string, error) {
	city = strings.TrimSpace(city)
	if city == "" || len(city) > maxCityLength {
		return "", ErrInvalidCity
	}

	options := countryCities[country]
	for _, option := range options {
		if strings.EqualFold(option, city) {
			return option, nil
		}
	}

	return "", ErrInvalidCity
}

func normalizeBirthDate(value time.Time) (time.Time, int, error) {
	if value.IsZero() {
		return time.Time{}, 0, ErrInvalidBirthDate
	}

	normalized := normalizeBirthDateValue(value)
	now := time.Now().UTC()
	if normalized.After(now) {
		return time.Time{}, 0, ErrInvalidBirthDate
	}

	age := ageFromBirthDate(normalized, now)
	if age < minAge || age > maxAge {
		return time.Time{}, 0, ErrInvalidBirthDate
	}

	return normalized, age, nil
}

func normalizeBirthDateValue(value time.Time) time.Time {
	if value.IsZero() {
		return value
	}

	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func ageFromBirthDate(birthDate time.Time, now time.Time) int {
	if birthDate.IsZero() {
		return 0
	}

	birthDate = birthDate.UTC()
	now = now.UTC()

	age := now.Year() - birthDate.Year()
	if now.Month() < birthDate.Month() || (now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		age--
	}

	return age
}
