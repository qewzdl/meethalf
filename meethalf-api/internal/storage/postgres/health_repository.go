package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type HealthRepository struct {
	db      *sql.DB
	timeout time.Duration
}

func NewHealthRepository(db *sql.DB, timeout time.Duration) *HealthRepository {
	return &HealthRepository{
		db:      db,
		timeout: timeout,
	}
}

func (r *HealthRepository) Name() string {
	return "postgres"
}

func (r *HealthRepository) Timeout() time.Duration {
	if r == nil {
		return 0
	}

	return r.timeout
}

func (r *HealthRepository) Ping(ctx context.Context) error {
	if r == nil || r.db == nil {
		return errors.New("postgres db is not initialized")
	}

	return r.db.PingContext(ctx)
}
