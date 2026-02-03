package backup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"meethalf-telegram-bot/internal/domain"
)

const CurrentVersion = 1

type Metadata struct {
	Version     int
	GeneratedAt time.Time
	Entries     int
}

type Repository interface {
	Dump(ctx context.Context) ([]domain.BackupEntry, error)
	Restore(ctx context.Context, entries []domain.BackupEntry) error
}

type Usecase interface {
	Export(ctx context.Context, w io.Writer) (Metadata, error)
	Import(ctx context.Context, r io.Reader) error
}

type service struct {
	repo Repository
	now  func() time.Time
}

func New(repo Repository) Usecase {
	return &service{
		repo: repo,
		now:  time.Now,
	}
}

func (s *service) Export(ctx context.Context, w io.Writer) (Metadata, error) {
	if s == nil || s.repo == nil {
		return Metadata{}, errors.New("backup repository is not configured")
	}
	if w == nil {
		return Metadata{}, errors.New("backup writer is nil")
	}

	entries, err := s.repo.Dump(ctx)
	if err != nil {
		return Metadata{}, fmt.Errorf("dump backup: %w", err)
	}

	snapshot := domain.BackupSnapshot{
		Version:     CurrentVersion,
		GeneratedAt: s.now().UTC(),
		Entries:     entries,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(snapshot); err != nil {
		return Metadata{}, fmt.Errorf("encode backup: %w", err)
	}

	return Metadata{
		Version:     snapshot.Version,
		GeneratedAt: snapshot.GeneratedAt,
		Entries:     len(snapshot.Entries),
	}, nil
}

func (s *service) Import(ctx context.Context, r io.Reader) error {
	if s == nil || s.repo == nil {
		return errors.New("backup repository is not configured")
	}
	if r == nil {
		return errors.New("backup reader is nil")
	}

	var snapshot domain.BackupSnapshot
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&snapshot); err != nil {
		return fmt.Errorf("decode backup: %w", err)
	}

	if snapshot.Version == 0 {
		snapshot.Version = CurrentVersion
	}
	if snapshot.Version != CurrentVersion {
		return fmt.Errorf("unsupported backup version: %d", snapshot.Version)
	}

	if err := s.repo.Restore(ctx, snapshot.Entries); err != nil {
		return fmt.Errorf("restore backup: %w", err)
	}

	return nil
}
