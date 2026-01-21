package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

type DatabaseRepository struct {
	db *sql.DB
}

func NewDatabaseRepository(db *sql.DB) *DatabaseRepository {
	return &DatabaseRepository{db: db}
}

func (r *DatabaseRepository) DatabaseExists(ctx context.Context, name string) (bool, error) {
	if r == nil || r.db == nil {
		return false, errors.New("postgres db is not initialized")
	}

	if name == "" {
		return false, errors.New("database name is empty")
	}

	const query = "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	var exists bool
	if err := r.db.QueryRowContext(ctx, query, name).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *DatabaseRepository) RoleExists(ctx context.Context, name string) (bool, error) {
	if r == nil || r.db == nil {
		return false, errors.New("postgres db is not initialized")
	}

	if name == "" {
		return false, errors.New("role name is empty")
	}

	const query = "SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname = $1)"
	var exists bool
	if err := r.db.QueryRowContext(ctx, query, name).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *DatabaseRepository) CreateRole(ctx context.Context, name string) error {
	if r == nil || r.db == nil {
		return errors.New("postgres db is not initialized")
	}

	if name == "" {
		return errors.New("role name is empty")
	}

	query := fmt.Sprintf(
		"CREATE ROLE %s WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT",
		quoteIdentifier(name),
	)
	_, err := r.db.ExecContext(ctx, query)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "42710" {
		return nil
	}

	return err
}

func (r *DatabaseRepository) UpdateRolePassword(ctx context.Context, name, password string) error {
	if r == nil || r.db == nil {
		return errors.New("postgres db is not initialized")
	}

	if name == "" {
		return errors.New("role name is empty")
	}

	if password == "" {
		return errors.New("role password is empty")
	}

	query := fmt.Sprintf(
		"ALTER ROLE %s WITH PASSWORD %s",
		quoteIdentifier(name),
		quoteLiteral(password),
	)
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *DatabaseRepository) CreateDatabase(ctx context.Context, name, owner string) error {
	if r == nil || r.db == nil {
		return errors.New("postgres db is not initialized")
	}

	if name == "" {
		return errors.New("database name is empty")
	}

	if owner == "" {
		return errors.New("database owner is empty")
	}

	query := fmt.Sprintf("CREATE DATABASE %s OWNER %s", quoteIdentifier(name), quoteIdentifier(owner))
	_, err := r.db.ExecContext(ctx, query)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "42P04" {
		return nil
	}

	return err
}

func (r *DatabaseRepository) SetDatabaseOwner(ctx context.Context, name, owner string) error {
	if r == nil || r.db == nil {
		return errors.New("postgres db is not initialized")
	}

	if name == "" {
		return errors.New("database name is empty")
	}

	if owner == "" {
		return errors.New("database owner is empty")
	}

	query := fmt.Sprintf("ALTER DATABASE %s OWNER TO %s", quoteIdentifier(name), quoteIdentifier(owner))
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *DatabaseRepository) RevokePublicDatabasePrivileges(ctx context.Context, name string) error {
	if r == nil || r.db == nil {
		return errors.New("postgres db is not initialized")
	}

	if name == "" {
		return errors.New("database name is empty")
	}

	query := fmt.Sprintf("REVOKE ALL ON DATABASE %s FROM PUBLIC", quoteIdentifier(name))
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *DatabaseRepository) GrantDatabasePrivileges(ctx context.Context, name, role string) error {
	if r == nil || r.db == nil {
		return errors.New("postgres db is not initialized")
	}

	if name == "" {
		return errors.New("database name is empty")
	}

	if role == "" {
		return errors.New("database role is empty")
	}

	query := fmt.Sprintf(
		"GRANT CONNECT, TEMPORARY ON DATABASE %s TO %s",
		quoteIdentifier(name),
		quoteIdentifier(role),
	)
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *DatabaseRepository) SchemaExists(ctx context.Context, name string) (bool, error) {
	if r == nil || r.db == nil {
		return false, errors.New("postgres db is not initialized")
	}

	if name == "" {
		return false, errors.New("schema name is empty")
	}

	const query = "SELECT EXISTS(SELECT 1 FROM pg_namespace WHERE nspname = $1)"
	var exists bool
	if err := r.db.QueryRowContext(ctx, query, name).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *DatabaseRepository) CreateSchema(ctx context.Context, name, owner string) error {
	if r == nil || r.db == nil {
		return errors.New("postgres db is not initialized")
	}

	if name == "" {
		return errors.New("schema name is empty")
	}

	if owner == "" {
		return errors.New("schema owner is empty")
	}

	query := fmt.Sprintf("CREATE SCHEMA %s AUTHORIZATION %s", quoteIdentifier(name), quoteIdentifier(owner))
	_, err := r.db.ExecContext(ctx, query)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "42P06" {
		return nil
	}

	return err
}

func (r *DatabaseRepository) SetSchemaOwner(ctx context.Context, name, owner string) error {
	if r == nil || r.db == nil {
		return errors.New("postgres db is not initialized")
	}

	if name == "" {
		return errors.New("schema name is empty")
	}

	if owner == "" {
		return errors.New("schema owner is empty")
	}

	query := fmt.Sprintf("ALTER SCHEMA %s OWNER TO %s", quoteIdentifier(name), quoteIdentifier(owner))
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *DatabaseRepository) RevokePublicSchemaPrivileges(ctx context.Context, name string) error {
	if r == nil || r.db == nil {
		return errors.New("postgres db is not initialized")
	}

	if name == "" {
		return errors.New("schema name is empty")
	}

	query := fmt.Sprintf("REVOKE ALL ON SCHEMA %s FROM PUBLIC", quoteIdentifier(name))
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *DatabaseRepository) GrantSchemaPrivileges(ctx context.Context, name, role string) error {
	if r == nil || r.db == nil {
		return errors.New("postgres db is not initialized")
	}

	if name == "" {
		return errors.New("schema name is empty")
	}

	if role == "" {
		return errors.New("schema role is empty")
	}

	query := fmt.Sprintf(
		"GRANT USAGE, CREATE ON SCHEMA %s TO %s",
		quoteIdentifier(name),
		quoteIdentifier(role),
	)
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *DatabaseRepository) SetRoleSearchPath(ctx context.Context, role, schema string) error {
	if r == nil || r.db == nil {
		return errors.New("postgres db is not initialized")
	}

	if role == "" {
		return errors.New("role name is empty")
	}

	if schema == "" {
		return errors.New("schema name is empty")
	}

	query := fmt.Sprintf(
		"ALTER ROLE %s SET search_path = %s",
		quoteIdentifier(role),
		quoteIdentifier(schema),
	)
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *DatabaseRepository) RevokeDefaultPublicPrivileges(ctx context.Context, role, schema string) error {
	if r == nil || r.db == nil {
		return errors.New("postgres db is not initialized")
	}

	if role == "" {
		return errors.New("role name is empty")
	}

	if schema == "" {
		return errors.New("schema name is empty")
	}

	roleIdent := quoteIdentifier(role)
	schemaIdent := quoteIdentifier(schema)
	statements := []string{
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES FOR ROLE %s IN SCHEMA %s REVOKE ALL ON TABLES FROM PUBLIC", roleIdent, schemaIdent),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES FOR ROLE %s IN SCHEMA %s REVOKE ALL ON SEQUENCES FROM PUBLIC", roleIdent, schemaIdent),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES FOR ROLE %s IN SCHEMA %s REVOKE ALL ON FUNCTIONS FROM PUBLIC", roleIdent, schemaIdent),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES FOR ROLE %s IN SCHEMA %s REVOKE ALL ON TYPES FROM PUBLIC", roleIdent, schemaIdent),
	}

	for _, stmt := range statements {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}

	return nil
}

func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func quoteLiteral(value string) string {
	return `'` + strings.ReplaceAll(value, `'`, `''`) + `'`
}
