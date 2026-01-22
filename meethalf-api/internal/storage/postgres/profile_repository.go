package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"meethalf-api/internal/domain"

	"github.com/jackc/pgx/v5/pgtype"
)

type ProfileRepository struct {
	db      *sql.DB
	schema  string
	table   string
	typeMap *pgtype.Map
}

func NewProfileRepository(db *sql.DB, schema string) (*ProfileRepository, error) {
	if db == nil {
		return nil, errors.New("postgres db is not initialized")
	}

	normalizedSchema := normalizeSchema(schema)
	table := profileTable(normalizedSchema)
	repo := &ProfileRepository{
		db:      db,
		schema:  normalizedSchema,
		table:   table,
		typeMap: pgtype.NewMap(),
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

func (r *ProfileRepository) Upsert(ctx context.Context, profile domain.Profile) (domain.Profile, error) {
	if r == nil || r.db == nil {
		return domain.Profile{}, errors.New("postgres profile repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return domain.Profile{}, err
	}

	now := time.Now().UTC()
	query := fmt.Sprintf(
		`INSERT INTO %s (user_id, name, gender, birth_date, age, country, city, description, emoji_code, photos, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 ON CONFLICT (user_id) DO UPDATE SET
			name = EXCLUDED.name,
			gender = EXCLUDED.gender,
			birth_date = EXCLUDED.birth_date,
			age = EXCLUDED.age,
			country = EXCLUDED.country,
			city = EXCLUDED.city,
			description = EXCLUDED.description,
			emoji_code = EXCLUDED.emoji_code,
			photos = EXCLUDED.photos,
			updated_at = EXCLUDED.updated_at
		 RETURNING created_at, updated_at`,
		r.table,
	)

	var createdAt time.Time
	var updatedAt time.Time
	if err := r.db.QueryRowContext(
		ctx,
		query,
		profile.UserID,
		profile.Name,
		profile.Gender,
		profile.BirthDate,
		profile.Age,
		profile.Country,
		profile.City,
		profile.Description,
		profile.EmojiCode,
		profile.Photos,
		now,
		now,
	).Scan(&createdAt, &updatedAt); err != nil {
		return domain.Profile{}, err
	}

	profile.CreatedAt = createdAt
	profile.UpdatedAt = updatedAt
	return profile, nil
}

func (r *ProfileRepository) GetByUserID(ctx context.Context, userID int64) (domain.Profile, error) {
	if r == nil || r.db == nil {
		return domain.Profile{}, errors.New("postgres profile repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return domain.Profile{}, err
	}

	query := fmt.Sprintf(
		`SELECT user_id, name, gender, birth_date, age, country, city, description, emoji_code, photos, created_at, updated_at
		 FROM %s
		 WHERE user_id = $1`,
		r.table,
	)

	typeMap := r.typeMap
	if typeMap == nil {
		typeMap = pgtype.NewMap()
	}
	var photos []string
	photosScanner := typeMap.SQLScanner(&photos)

	var stored domain.Profile
	var birthDate sql.NullTime
	var age sql.NullInt32
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stored.UserID,
		&stored.Name,
		&stored.Gender,
		&birthDate,
		&age,
		&stored.Country,
		&stored.City,
		&stored.Description,
		&stored.EmojiCode,
		photosScanner,
		&stored.CreatedAt,
		&stored.UpdatedAt,
	); err != nil {
		return domain.Profile{}, err
	}

	if birthDate.Valid {
		stored.BirthDate = birthDate.Time
	}
	if age.Valid {
		stored.Age = int(age.Int32)
	}

	stored.Photos = photos
	return stored, nil
}

func (r *ProfileRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	if r == nil || r.db == nil {
		return errors.New("postgres profile repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	query := fmt.Sprintf(
		`DELETE FROM %s
		 WHERE user_id = $1`,
		r.table,
	)

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *ProfileRepository) ensureTable(ctx context.Context) error {
	query := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (
			user_id BIGINT PRIMARY KEY,
			name TEXT NOT NULL,
			gender TEXT NOT NULL DEFAULT 'unspecified',
			birth_date DATE NOT NULL,
			age INT NOT NULL,
			country TEXT NOT NULL,
			city TEXT NOT NULL,
			description TEXT NOT NULL,
			emoji_code TEXT NOT NULL DEFAULT '',
			photos TEXT[] NOT NULL DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)`,
		r.table,
	)

	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *ProfileRepository) ensureColumns(ctx context.Context) error {
	if r == nil || r.db == nil {
		return errors.New("postgres profile repository is not configured")
	}

	queries := []string{
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS photos TEXT[] NOT NULL DEFAULT '{}'", r.table),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS gender TEXT NOT NULL DEFAULT '%s'", r.table, domain.GenderUnspecified),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS birth_date DATE", r.table),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS country TEXT NOT NULL DEFAULT ''", r.table),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS city TEXT NOT NULL DEFAULT ''", r.table),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS emoji_code TEXT NOT NULL DEFAULT ''", r.table),
	}
	for _, query := range queries {
		if _, err := r.db.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProfileRepository) ensureSchema(ctx context.Context) error {
	if r == nil || r.db == nil {
		return errors.New("postgres profile repository is not configured")
	}

	if r.schema == "" || r.schema == "public" {
		return nil
	}

	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", quoteIdentifier(r.schema))
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func profileTable(schema string) string {
	schema = normalizeSchema(schema)
	return quoteIdentifier(schema) + "." + quoteIdentifier("profiles")
}

func normalizeSchema(schema string) string {
	schema = strings.TrimSpace(schema)
	if schema == "" {
		return "public"
	}

	return schema
}
