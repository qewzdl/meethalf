package ads

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"meethalf-api/internal/domain"
)

const (
	maxTextLength       = 4096
	maxButtons          = 6
	maxButtonTextLength = 64
	adButtonEmoji       = "🔗"
)

var (
	ErrInvalidAdContent   = errors.New("ad text or photo is required")
	ErrInvalidAdText      = errors.New("ad text is too long")
	ErrInvalidAdButtons   = errors.New("ad buttons are invalid")
	ErrTooManyAdButtons   = errors.New("too many ad buttons")
	ErrInvalidAdButton    = errors.New("ad button text is required")
	ErrInvalidAdButtonURL = errors.New("ad button url is invalid")
)

type Usecase interface {
	Create(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error)
}

type Repository interface {
	Create(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error)
}

type service struct {
	repo Repository
}

func New(repo Repository) Usecase {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error) {
	if s == nil || s.repo == nil {
		return domain.Advertisement{}, errors.New("ads repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return domain.Advertisement{}, err
	}

	ad.Text = strings.TrimSpace(ad.Text)
	ad.PhotoID = strings.TrimSpace(ad.PhotoID)
	if ad.Text == "" && ad.PhotoID == "" {
		return domain.Advertisement{}, ErrInvalidAdContent
	}
	if len([]rune(ad.Text)) > maxTextLength {
		return domain.Advertisement{}, ErrInvalidAdText
	}
	buttons, err := normalizeButtons(ad.Buttons)
	if err != nil {
		return domain.Advertisement{}, err
	}
	ad.Buttons = buttons

	return s.repo.Create(ctx, ad)
}

func normalizeButtons(buttons []domain.AdButton) ([]domain.AdButton, error) {
	if len(buttons) == 0 {
		return nil, nil
	}
	if maxButtons > 0 && len(buttons) > maxButtons {
		return nil, ErrTooManyAdButtons
	}

	out := make([]domain.AdButton, 0, len(buttons))
	for _, button := range buttons {
		text := strings.TrimSpace(button.Text)
		link := strings.TrimSpace(button.URL)
		if text == "" {
			return nil, ErrInvalidAdButton
		}
		decorated := decorateAdButtonText(text)
		if maxButtonTextLength > 0 && len([]rune(decorated)) > maxButtonTextLength {
			return nil, ErrInvalidAdButtons
		}
		if !isValidAdURL(link) {
			return nil, ErrInvalidAdButtonURL
		}
		out = append(out, domain.AdButton{
			Text: decorated,
			URL:  link,
		})
	}

	return out, nil
}

func decorateAdButtonText(text string) string {
	if text == "" {
		return text
	}
	if strings.HasPrefix(text, adButtonEmoji+" ") || text == adButtonEmoji {
		return text
	}
	return adButtonEmoji + " " + text
}

func isValidAdURL(value string) bool {
	if value == "" {
		return false
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed == nil {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
	default:
		return false
	}
	return parsed.Host != ""
}
