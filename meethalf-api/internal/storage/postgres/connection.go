package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"meethalf-api/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func New(cfg config.DBConfig) (*sql.DB, error) {
	return NewWithDatabase(cfg, cfg.Name)
}

func NewWithDatabase(cfg config.DBConfig, dbName string) (*sql.DB, error) {
	cfg.Name = dbName
	dsn := buildDSN(cfg)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	ctx := context.Background()
	if cfg.ConnectTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.ConnectTimeout)
		defer cancel()
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

func buildDSN(cfg config.DBConfig) string {
	host := cfg.Host
	if host == "" {
		host = "localhost"
	}

	port := cfg.Port
	if port == "" {
		port = "5432"
	}

	user := cfg.User
	if user == "" {
		user = "postgres"
	}

	dbName := cfg.Name
	if dbName == "" {
		dbName = "postgres"
	}

	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	query := url.Values{}
	query.Set("sslmode", sslMode)
	if cfg.ConnectTimeout > 0 {
		query.Set("connect_timeout", strconv.Itoa(int(cfg.ConnectTimeout.Seconds())))
	}

	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, cfg.Password),
		Host:     net.JoinHostPort(host, port),
		Path:     dbName,
		RawQuery: query.Encode(),
	}

	return u.String()
}
