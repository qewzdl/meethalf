package analytics

import (
	"context"
	"errors"
)

var ErrNotImplemented = errors.New("analytics usecase is not implemented")

type Usecase interface {
	Ready(ctx context.Context) error
}

type service struct{}

func New() Usecase {
	return &service{}
}

func (s *service) Ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return ErrNotImplemented
}
