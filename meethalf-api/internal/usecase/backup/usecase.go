package backup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

const (
	defaultFilenamePrefix = "meethalf-api-backup"
	defaultExtension      = ".dump"
	defaultContentType    = "application/octet-stream"
)

type Backup struct {
	Filename    string
	ContentType string
	Reader      io.ReadCloser
}

type Usecase interface {
	Create(ctx context.Context) (Backup, error)
	Restore(ctx context.Context, source io.Reader) error
}

type Repository interface {
	Dump(ctx context.Context) (io.ReadCloser, error)
	Restore(ctx context.Context, source io.Reader) error
}

type service struct {
	repo           Repository
	now            func() time.Time
	filenamePrefix string
}

func New(repo Repository) Usecase {
	return &service{
		repo:           repo,
		now:            time.Now,
		filenamePrefix: defaultFilenamePrefix,
	}
}

func (s *service) Create(ctx context.Context) (Backup, error) {
	if s == nil || s.repo == nil {
		return Backup{}, errors.New("backup repository is not configured")
	}

	reader, err := s.repo.Dump(ctx)
	if err != nil {
		return Backup{}, fmt.Errorf("create backup: %w", err)
	}

	timestamp := s.now().UTC().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s%s", s.filenamePrefix, timestamp, defaultExtension)

	return Backup{
		Filename:    filename,
		ContentType: defaultContentType,
		Reader:      reader,
	}, nil
}

func (s *service) Restore(ctx context.Context, source io.Reader) error {
	if s == nil || s.repo == nil {
		return errors.New("backup repository is not configured")
	}
	if source == nil {
		return errors.New("backup source is nil")
	}

	if err := s.repo.Restore(ctx, source); err != nil {
		return fmt.Errorf("restore backup: %w", err)
	}

	return nil
}
