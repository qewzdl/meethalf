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
		`INSERT INTO %s (user_id, username, name, gender, birth_date, age, country, city, description, emoji_code, photos, is_hidden, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		 ON CONFLICT (user_id) DO UPDATE SET
			username = EXCLUDED.username,
			name = EXCLUDED.name,
			gender = EXCLUDED.gender,
			birth_date = EXCLUDED.birth_date,
			age = EXCLUDED.age,
			country = EXCLUDED.country,
			city = EXCLUDED.city,
			description = EXCLUDED.description,
			emoji_code = EXCLUDED.emoji_code,
			photos = EXCLUDED.photos,
			is_hidden = EXCLUDED.is_hidden,
			updated_at = EXCLUDED.updated_at
		 RETURNING created_at, updated_at, is_banned, is_moderator`,
		r.table,
	)

	var createdAt time.Time
	var updatedAt time.Time
	var isBanned bool
	var isModerator bool
	if err := r.db.QueryRowContext(
		ctx,
		query,
		profile.UserID,
		profile.Username,
		profile.Name,
		profile.Gender,
		profile.BirthDate,
		profile.Age,
		profile.Country,
		profile.City,
		profile.Description,
		profile.EmojiCode,
		profile.Photos,
		profile.IsHidden,
		now,
		now,
	).Scan(&createdAt, &updatedAt, &isBanned, &isModerator); err != nil {
		return domain.Profile{}, err
	}

	profile.CreatedAt = createdAt
	profile.UpdatedAt = updatedAt
	profile.IsBanned = isBanned
	profile.IsModerator = isModerator
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
		`SELECT user_id, username, name, gender, birth_date, age, country, city, description, emoji_code, photos, is_hidden, is_banned, is_moderator, created_at, updated_at
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
		&stored.Username,
		&stored.Name,
		&stored.Gender,
		&birthDate,
		&age,
		&stored.Country,
		&stored.City,
		&stored.Description,
		&stored.EmojiCode,
		photosScanner,
		&stored.IsHidden,
		&stored.IsBanned,
		&stored.IsModerator,
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

func (r *ProfileRepository) UpdateVisibility(ctx context.Context, userID int64, isHidden bool) error {
	if r == nil || r.db == nil {
		return errors.New("postgres profile repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	query := fmt.Sprintf(
		`UPDATE %s
		 SET is_hidden = $1, updated_at = $2
		 WHERE user_id = $3`,
		r.table,
	)

	result, err := r.db.ExecContext(ctx, query, isHidden, time.Now().UTC(), userID)
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

func (r *ProfileRepository) UpdateBanStatus(ctx context.Context, userID int64, isBanned bool) error {
	if r == nil || r.db == nil {
		return errors.New("postgres profile repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	query := fmt.Sprintf(
		`UPDATE %s
		 SET is_banned = $1, updated_at = $2
		 WHERE user_id = $3`,
		r.table,
	)

	result, err := r.db.ExecContext(ctx, query, isBanned, time.Now().UTC(), userID)
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

func (r *ProfileRepository) UpdateModeratorStatus(ctx context.Context, userID int64, isModerator bool) error {
	if r == nil || r.db == nil {
		return errors.New("postgres profile repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	query := fmt.Sprintf(
		`UPDATE %s
		 SET is_moderator = $1, updated_at = $2
		 WHERE user_id = $3`,
		r.table,
	)

	result, err := r.db.ExecContext(ctx, query, isModerator, time.Now().UTC(), userID)
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

func (r *ProfileRepository) GetUserIDByUsername(ctx context.Context, username string) (int64, error) {
	if r == nil || r.db == nil {
		return 0, errors.New("postgres profile repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return 0, err
	}

	query := fmt.Sprintf(
		`SELECT user_id
		 FROM %s
		 WHERE LOWER(username) = LOWER($1)
		 ORDER BY updated_at DESC
		 LIMIT 1`,
		r.table,
	)

	var userID int64
	if err := r.db.QueryRowContext(ctx, query, username).Scan(&userID); err != nil {
		return 0, err
	}

	return userID, nil
}

func (r *ProfileRepository) GetUserSummary(ctx context.Context, userID int64) (domain.UserSummary, error) {
	if r == nil || r.db == nil {
		return domain.UserSummary{}, errors.New("postgres profile repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return domain.UserSummary{}, err
	}

	query := fmt.Sprintf(
		`SELECT user_id, username, name, is_hidden, is_banned, is_moderator, created_at, updated_at
		 FROM %s
		 WHERE user_id = $1`,
		r.table,
	)

	var user domain.UserSummary
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.Name,
		&user.IsHidden,
		&user.IsBanned,
		&user.IsModerator,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return domain.UserSummary{}, err
	}

	return user, nil
}

func (r *ProfileRepository) ListUsers(ctx context.Context, limit, offset int, onlyBanned, onlyModerators bool) ([]domain.UserSummary, int, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("postgres profile repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	conditions := make([]string, 0, 2)
	if onlyBanned {
		conditions = append(conditions, "is_banned = TRUE")
	}
	if onlyModerators {
		conditions = append(conditions, "is_moderator = TRUE")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s%s`, r.table, whereClause)
	total := 0
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []domain.UserSummary{}, 0, nil
	}

	query := fmt.Sprintf(
		`SELECT user_id, username, name, is_hidden, is_banned, is_moderator, created_at, updated_at
		 FROM %s%s
		 ORDER BY user_id ASC
		 LIMIT $1 OFFSET $2`,
		r.table,
		whereClause,
	)

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]domain.UserSummary, 0, limit)
	for rows.Next() {
		var user domain.UserSummary
		if err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.Name,
			&user.IsHidden,
			&user.IsBanned,
			&user.IsModerator,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *ProfileRepository) ListReportedUsers(ctx context.Context, limit, offset int) ([]domain.ReportedUserSummary, int, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("postgres profile repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	interactionsTable := matchInteractionsTable(r.schema)
	countQuery := fmt.Sprintf(
		`SELECT COUNT(DISTINCT i.target_id)
		 FROM %s i
		 JOIN %s p ON p.user_id = i.target_id
		 WHERE i.action = $1`,
		interactionsTable,
		r.table,
	)

	total := 0
	if err := r.db.QueryRowContext(ctx, countQuery, domain.MatchActionReport).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []domain.ReportedUserSummary{}, 0, nil
	}

	query := fmt.Sprintf(
		`SELECT p.user_id, p.username, p.name, p.is_hidden, p.is_banned, p.is_moderator, p.created_at, p.updated_at,
				COUNT(i.viewer_id) AS report_count
		 FROM %s i
		 JOIN %s p ON p.user_id = i.target_id
		 WHERE i.action = $1
		 GROUP BY p.user_id, p.username, p.name, p.is_hidden, p.is_banned, p.is_moderator, p.created_at, p.updated_at
		 ORDER BY report_count DESC, p.updated_at DESC, p.user_id ASC
		 LIMIT $2 OFFSET $3`,
		interactionsTable,
		r.table,
	)

	rows, err := r.db.QueryContext(ctx, query, domain.MatchActionReport, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]domain.ReportedUserSummary, 0, limit)
	for rows.Next() {
		var user domain.ReportedUserSummary
		if err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.Name,
			&user.IsHidden,
			&user.IsBanned,
			&user.IsModerator,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.ReportCount,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *ProfileRepository) ensureTable(ctx context.Context) error {
	query := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (
			user_id BIGINT PRIMARY KEY,
			username TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL,
			gender TEXT NOT NULL DEFAULT 'unspecified',
			birth_date DATE NOT NULL,
			age INT NOT NULL,
			country TEXT NOT NULL,
			city TEXT NOT NULL,
			description TEXT NOT NULL,
			emoji_code TEXT NOT NULL DEFAULT '',
			photos TEXT[] NOT NULL DEFAULT '{}',
			is_hidden BOOLEAN NOT NULL DEFAULT FALSE,
			is_banned BOOLEAN NOT NULL DEFAULT FALSE,
			is_moderator BOOLEAN NOT NULL DEFAULT FALSE,
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
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS username TEXT NOT NULL DEFAULT ''", r.table),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS photos TEXT[] NOT NULL DEFAULT '{}'", r.table),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS gender TEXT NOT NULL DEFAULT '%s'", r.table, domain.GenderUnspecified),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS birth_date DATE", r.table),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS country TEXT NOT NULL DEFAULT ''", r.table),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS city TEXT NOT NULL DEFAULT ''", r.table),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS emoji_code TEXT NOT NULL DEFAULT ''", r.table),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS is_hidden BOOLEAN NOT NULL DEFAULT FALSE", r.table),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS is_banned BOOLEAN NOT NULL DEFAULT FALSE", r.table),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS is_moderator BOOLEAN NOT NULL DEFAULT FALSE", r.table),
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
