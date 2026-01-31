package ads

import (
	"context"
	"errors"
	"strings"

	"meethalf-api/internal/domain"
)

const maxTextLength = 4096

var (
	ErrInvalidAdContent = errors.New("ad text or photo is required")
	ErrInvalidAdText    = errors.New("ad text is too long")
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

	return s.repo.Create(ctx, ad)
}
