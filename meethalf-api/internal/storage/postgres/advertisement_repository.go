package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"meethalf-api/internal/domain"
)

type AdvertisementRepository struct {
	db     *sql.DB
	schema string
	table  string
}

func NewAdvertisementRepository(db *sql.DB, schema string) (*AdvertisementRepository, error) {
	if db == nil {
		return nil, errors.New("postgres db is not initialized")
	}

	normalizedSchema := normalizeSchema(schema)
	repo := &AdvertisementRepository{
		db:     db,
		schema: normalizedSchema,
		table:  advertisementTable(normalizedSchema),
	}

	if err := repo.ensureSchema(context.Background()); err != nil {
		return nil, err
	}
	if err := repo.ensureTable(context.Background()); err != nil {
		return nil, err
	}
	if err := repo.ensureColumns(context.Background()); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *AdvertisementRepository) Create(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error) {
	if r == nil || r.db == nil {
		return domain.Advertisement{}, errors.New("postgres advertisement repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return domain.Advertisement{}, err
	}

	now := time.Now().UTC()
	buttonsPayload := []byte("[]")
	if len(ad.Buttons) > 0 {
		encoded, err := json.Marshal(ad.Buttons)
		if err != nil {
			return domain.Advertisement{}, err
		}
		buttonsPayload = encoded
	}
	query := fmt.Sprintf(
		`INSERT INTO %s (text, photo_id, buttons, created_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		r.table,
	)

	if err := r.db.QueryRowContext(ctx, query, ad.Text, ad.PhotoID, buttonsPayload, now).Scan(&ad.ID, &ad.CreatedAt); err != nil {
		return domain.Advertisement{}, err
	}

	return ad, nil
}

func (r *AdvertisementRepository) ensureSchema(ctx context.Context) error {
	if r == nil || r.db == nil {
		return errors.New("postgres advertisement repository is not configured")
	}
	if r.schema == "" || r.schema == "public" {
		return nil
	}

	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", quoteIdentifier(r.schema))
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *AdvertisementRepository) ensureTable(ctx context.Context) error {
	if r == nil || r.db == nil {
		return errors.New("postgres advertisement repository is not configured")
	}

	query := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (
			id BIGSERIAL PRIMARY KEY,
			text TEXT NOT NULL DEFAULT '',
			photo_id TEXT NOT NULL DEFAULT '',
			buttons JSONB NOT NULL DEFAULT '[]'::jsonb,
			created_at TIMESTAMPTZ NOT NULL
		)`,
		r.table,
	)

	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *AdvertisementRepository) ensureColumns(ctx context.Context) error {
	if r == nil || r.db == nil {
		return errors.New("postgres advertisement repository is not configured")
	}

	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS buttons JSONB NOT NULL DEFAULT '[]'::jsonb", r.table)
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func advertisementTable(schema string) string {
	schema = normalizeSchema(schema)
	return quoteIdentifier(schema) + "." + quoteIdentifier("ads")
}
