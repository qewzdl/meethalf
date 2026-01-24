package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"meethalf-api/internal/domain"
	"meethalf-api/internal/usecase/matching"

	"github.com/jackc/pgx/v5/pgtype"
)

type MatchingRepository struct {
	db                *sql.DB
	schema            string
	sessionsTable     string
	historyTable      string
	interactionsTable string
	typeMap           *pgtype.Map
}

func NewMatchingRepository(db *sql.DB, schema string) (*MatchingRepository, error) {
	if db == nil {
		return nil, errors.New("postgres db is not initialized")
	}

	normalizedSchema := normalizeSchema(schema)
	repo := &MatchingRepository{
		db:                db,
		schema:            normalizedSchema,
		sessionsTable:     matchSessionsTable(normalizedSchema),
		historyTable:      matchHistoryTable(normalizedSchema),
		interactionsTable: matchInteractionsTable(normalizedSchema),
		typeMap:           pgtype.NewMap(),
	}

	if err := repo.ensureSchema(context.Background()); err != nil {
		return nil, err
	}
	if err := repo.ensureTables(context.Background()); err != nil {
		return nil, err
	}
	if err := repo.ensureColumns(context.Background()); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *MatchingRepository) GetProfile(ctx context.Context, userID int64) (domain.Profile, error) {
	if r == nil || r.db == nil {
		return domain.Profile{}, errors.New("postgres matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return domain.Profile{}, err
	}

	query := fmt.Sprintf(
		`SELECT user_id, name, gender, birth_date, age, country, city, description, emoji_code, photos, is_hidden, is_banned, created_at, updated_at
		 FROM %s
		 WHERE user_id = $1`,
		profileTable(r.schema),
	)

	row := r.db.QueryRowContext(ctx, query, userID)
	return r.scanProfile(row.Scan)
}

func (r *MatchingRepository) GetSession(ctx context.Context, viewerID int64) (domain.MatchSession, bool, error) {
	if r == nil || r.db == nil {
		return domain.MatchSession{}, false, errors.New("postgres matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return domain.MatchSession{}, false, err
	}

	query := fmt.Sprintf(
		`SELECT viewer_id, gender_filter, accuracy, session_version, current_index, created_at, updated_at
		 FROM %s
		 WHERE viewer_id = $1`,
		r.sessionsTable,
	)

	var session domain.MatchSession
	err := r.db.QueryRowContext(ctx, query, viewerID).Scan(
		&session.ViewerID,
		&session.GenderFilter,
		&session.Accuracy,
		&session.SessionVersion,
		&session.CurrentIndex,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.MatchSession{}, false, nil
		}
		return domain.MatchSession{}, false, err
	}

	return session, true, nil
}

func (r *MatchingRepository) SaveSession(ctx context.Context, session domain.MatchSession) error {
	if r == nil || r.db == nil {
		return errors.New("postgres matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (viewer_id, gender_filter, accuracy, session_version, current_index, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (viewer_id) DO UPDATE SET
			gender_filter = EXCLUDED.gender_filter,
			accuracy = EXCLUDED.accuracy,
			session_version = EXCLUDED.session_version,
			current_index = EXCLUDED.current_index,
			created_at = EXCLUDED.created_at,
			updated_at = EXCLUDED.updated_at`,
		r.sessionsTable,
	)

	_, err := r.db.ExecContext(
		ctx,
		query,
		session.ViewerID,
		session.GenderFilter,
		session.Accuracy,
		session.SessionVersion,
		session.CurrentIndex,
		session.CreatedAt,
		session.UpdatedAt,
	)
	return err
}

func (r *MatchingRepository) ResetHistory(ctx context.Context, viewerID int64) error {
	if r == nil || r.db == nil {
		return errors.New("postgres matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE viewer_id = $1`, r.historyTable)
	_, err := r.db.ExecContext(ctx, query, viewerID)
	return err
}

func (r *MatchingRepository) GetHistoryCandidate(ctx context.Context, viewerID int64, sessionVersion int64, position int) (domain.Profile, bool, error) {
	if r == nil || r.db == nil {
		return domain.Profile{}, false, errors.New("postgres matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return domain.Profile{}, false, err
	}

	query := fmt.Sprintf(
		`SELECT p.user_id, p.name, p.gender, p.birth_date, p.age, p.country, p.city, p.description, p.emoji_code, p.photos, p.is_hidden, p.is_banned, p.created_at, p.updated_at
		 FROM %s h
		 JOIN %s p ON p.user_id = h.candidate_id
		 WHERE h.viewer_id = $1 AND h.session_version = $2 AND h.position = $3
		   AND p.is_banned = FALSE`,
		r.historyTable,
		profileTable(r.schema),
	)

	row := r.db.QueryRowContext(ctx, query, viewerID, sessionVersion, position)
	candidate, err := r.scanProfile(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Profile{}, false, nil
		}
		return domain.Profile{}, false, err
	}

	return candidate, true, nil
}

func (r *MatchingRepository) SaveHistoryCandidate(ctx context.Context, viewerID int64, sessionVersion int64, position int, candidateID int64) error {
	if r == nil || r.db == nil {
		return errors.New("postgres matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (viewer_id, session_version, position, candidate_id, created_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (viewer_id, session_version, position) DO NOTHING`,
		r.historyTable,
	)

	_, err := r.db.ExecContext(ctx, query, viewerID, sessionVersion, position, candidateID, time.Now().UTC())
	return err
}

func (r *MatchingRepository) FindCandidate(ctx context.Context, params matching.CandidateParams) (domain.Profile, bool, error) {
	if r == nil || r.db == nil {
		return domain.Profile{}, false, errors.New("postgres matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return domain.Profile{}, false, err
	}

	genderFilter := strings.TrimSpace(string(params.GenderFilter))
	if genderFilter == string(domain.GenderUnspecified) {
		genderFilter = ""
	}

	scoreExpr := `(CASE WHEN p.country = $5 THEN 1 ELSE 0 END) + ` +
		`(CASE WHEN p.city = $6 THEN 1 ELSE 0 END) + ` +
		`(CASE WHEN ABS(p.age - $7::int) <= $9::int THEN 1 ELSE 0 END) + ` +
		`(CASE WHEN p.emoji_code = $8 THEN 1 ELSE 0 END)`

	query := fmt.Sprintf(
		`SELECT user_id, name, gender, birth_date, age, country, city, description, emoji_code, photos, is_hidden, is_banned, created_at, updated_at
		 FROM %s p
		 WHERE p.user_id <> $1
		   AND p.is_hidden = FALSE
		   AND p.is_banned = FALSE
		   AND ($2 = '' OR p.gender = $2)
		   AND NOT EXISTS (
				SELECT 1 FROM %s i
				WHERE i.viewer_id = $1 AND i.target_id = p.user_id
		   )
		   AND NOT EXISTS (
				SELECT 1 FROM %s h
				WHERE h.viewer_id = $1 AND h.session_version = $3 AND h.candidate_id = p.user_id
		   )
		   AND %s >= $4
		 ORDER BY %s DESC, p.updated_at DESC, p.user_id DESC
		 LIMIT 1`,
		profileTable(r.schema),
		r.interactionsTable,
		r.historyTable,
		scoreExpr,
		scoreExpr,
	)

	row := r.db.QueryRowContext(
		ctx,
		query,
		params.ViewerID,
		genderFilter,
		params.SessionVersion,
		params.Accuracy,
		params.ViewerCountry,
		params.ViewerCity,
		params.ViewerAge,
		params.ViewerEmoji,
		params.AgeWindow,
	)

	candidate, err := r.scanProfile(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Profile{}, false, nil
		}
		return domain.Profile{}, false, err
	}

	return candidate, true, nil
}

func (r *MatchingRepository) RecordAction(ctx context.Context, viewerID, targetID int64, action domain.MatchAction) error {
	if r == nil || r.db == nil {
		return errors.New("postgres matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	now := time.Now().UTC()
	var notifiedAt interface{}
	if action != domain.MatchActionLike {
		notifiedAt = now
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (viewer_id, target_id, action, created_at, notified_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (viewer_id, target_id) DO UPDATE SET
			action = EXCLUDED.action,
			created_at = EXCLUDED.created_at,
			notified_at = EXCLUDED.notified_at`,
		r.interactionsTable,
	)

	_, err := r.db.ExecContext(ctx, query, viewerID, targetID, action, now, notifiedAt)
	return err
}

func (r *MatchingRepository) HasAction(ctx context.Context, viewerID, targetID int64, action domain.MatchAction) (bool, error) {
	if r == nil || r.db == nil {
		return false, errors.New("postgres matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	query := fmt.Sprintf(
		`SELECT EXISTS (
			SELECT 1 FROM %s
			WHERE viewer_id = $1 AND target_id = $2 AND action = $3
		)`,
		r.interactionsTable,
	)

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, viewerID, targetID, action).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *MatchingRepository) ListPendingLikes(ctx context.Context, userID int64) ([]domain.Profile, []int64, error) {
	if r == nil || r.db == nil {
		return nil, nil, errors.New("postgres matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	query := fmt.Sprintf(
		`SELECT p.user_id, p.name, p.gender, p.birth_date, p.age, p.country, p.city, p.description, p.emoji_code, p.photos, p.is_hidden, p.is_banned, p.created_at, p.updated_at
		 FROM %s i
		 JOIN %s p ON p.user_id = i.viewer_id
		 WHERE i.target_id = $1 AND i.action = $2 AND i.notified_at IS NULL
		   AND p.is_banned = FALSE
		 ORDER BY i.created_at ASC`,
		r.interactionsTable,
		profileTable(r.schema),
	)

	rows, err := r.db.QueryContext(ctx, query, userID, domain.MatchActionLike)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	likes := make([]domain.Profile, 0)
	likerIDs := make([]int64, 0)
	for rows.Next() {
		profile, err := r.scanProfile(rows.Scan)
		if err != nil {
			return nil, nil, err
		}
		likes = append(likes, profile)
		likerIDs = append(likerIDs, profile.UserID)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return likes, likerIDs, nil
}

func (r *MatchingRepository) MarkLikesNotified(ctx context.Context, userID int64, likerIDs []int64) error {
	if r == nil || r.db == nil {
		return errors.New("postgres matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(likerIDs) == 0 {
		return nil
	}

	placeholders := make([]string, 0, len(likerIDs))
	args := make([]any, 0, len(likerIDs)+3)
	args = append(args, time.Now().UTC(), userID, domain.MatchActionLike)
	for i, id := range likerIDs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+4))
		args = append(args, id)
	}

	query := fmt.Sprintf(
		`UPDATE %s
		 SET notified_at = $1
		 WHERE target_id = $2 AND action = $3 AND notified_at IS NULL
		   AND viewer_id IN (%s)`,
		r.interactionsTable,
		strings.Join(placeholders, ", "),
	)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *MatchingRepository) scanProfile(scan func(dest ...any) error) (domain.Profile, error) {
	typeMap := r.typeMap
	if typeMap == nil {
		typeMap = pgtype.NewMap()
	}
	var photos []string
	photosScanner := typeMap.SQLScanner(&photos)

	var stored domain.Profile
	var birthDate sql.NullTime
	var age sql.NullInt32
	if err := scan(
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
		&stored.IsHidden,
		&stored.IsBanned,
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

func (r *MatchingRepository) ensureTables(ctx context.Context) error {
	queries := []string{
		fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s (
				viewer_id BIGINT PRIMARY KEY,
				gender_filter TEXT NOT NULL DEFAULT 'unspecified',
				accuracy INT NOT NULL DEFAULT 0,
				session_version BIGINT NOT NULL DEFAULT 1,
				current_index INT NOT NULL DEFAULT 0,
				created_at TIMESTAMPTZ NOT NULL,
				updated_at TIMESTAMPTZ NOT NULL
			)`,
			r.sessionsTable,
		),
		fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s (
				viewer_id BIGINT NOT NULL,
				session_version BIGINT NOT NULL,
				position INT NOT NULL,
				candidate_id BIGINT NOT NULL,
				created_at TIMESTAMPTZ NOT NULL,
				PRIMARY KEY (viewer_id, session_version, position)
			)`,
			r.historyTable,
		),
		fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s (
				viewer_id BIGINT NOT NULL,
				target_id BIGINT NOT NULL,
				action TEXT NOT NULL,
				created_at TIMESTAMPTZ NOT NULL,
				notified_at TIMESTAMPTZ NULL,
				PRIMARY KEY (viewer_id, target_id)
			)`,
			r.interactionsTable,
		),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (target_id, action, notified_at)", quoteIdentifier("match_interactions_target_idx"), r.interactionsTable),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (viewer_id, session_version, candidate_id)", quoteIdentifier("match_history_candidate_idx"), r.historyTable),
	}

	for _, query := range queries {
		if _, err := r.db.ExecContext(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

func (r *MatchingRepository) ensureColumns(ctx context.Context) error {
	if r == nil || r.db == nil {
		return errors.New("postgres matching repository is not configured")
	}

	queries := []string{
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS gender_filter TEXT NOT NULL DEFAULT '%s'", r.sessionsTable, domain.GenderUnspecified),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS accuracy INT NOT NULL DEFAULT 0", r.sessionsTable),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS session_version BIGINT NOT NULL DEFAULT 1", r.sessionsTable),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS current_index INT NOT NULL DEFAULT 0", r.sessionsTable),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ", r.sessionsTable),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ", r.sessionsTable),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS session_version BIGINT NOT NULL DEFAULT 1", r.historyTable),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS position INT NOT NULL DEFAULT 0", r.historyTable),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS candidate_id BIGINT NOT NULL DEFAULT 0", r.historyTable),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ", r.historyTable),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS action TEXT NOT NULL DEFAULT ''", r.interactionsTable),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ", r.interactionsTable),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS notified_at TIMESTAMPTZ", r.interactionsTable),
	}

	for _, query := range queries {
		if _, err := r.db.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func (r *MatchingRepository) ensureSchema(ctx context.Context) error {
	if r == nil || r.db == nil {
		return errors.New("postgres matching repository is not configured")
	}
	if r.schema == "" || r.schema == "public" {
		return nil
	}

	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", quoteIdentifier(r.schema))
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func matchSessionsTable(schema string) string {
	schema = normalizeSchema(schema)
	return quoteIdentifier(schema) + "." + quoteIdentifier("match_sessions")
}

func matchHistoryTable(schema string) string {
	schema = normalizeSchema(schema)
	return quoteIdentifier(schema) + "." + quoteIdentifier("match_history")
}

func matchInteractionsTable(schema string) string {
	schema = normalizeSchema(schema)
	return quoteIdentifier(schema) + "." + quoteIdentifier("match_interactions")
}
