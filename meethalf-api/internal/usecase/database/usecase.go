package database

import (
	"context"
	"errors"
	"fmt"
)

type Usecase interface {
	Ensure(ctx context.Context) error
}

type Settings struct {
	Name     string
	User     string
	Password string
	Schema   string
}

type Repository interface {
	DatabaseExists(ctx context.Context, name string) (bool, error)
	CreateDatabase(ctx context.Context, name, owner string) error
	SetDatabaseOwner(ctx context.Context, name, owner string) error
	GrantDatabasePrivileges(ctx context.Context, name, role string) error
	RevokePublicDatabasePrivileges(ctx context.Context, name string) error
	RoleExists(ctx context.Context, name string) (bool, error)
	CreateRole(ctx context.Context, name string) error
	UpdateRolePassword(ctx context.Context, name, password string) error
	SchemaExists(ctx context.Context, name string) (bool, error)
	CreateSchema(ctx context.Context, name, owner string) error
	SetSchemaOwner(ctx context.Context, name, owner string) error
	RevokePublicSchemaPrivileges(ctx context.Context, name string) error
	GrantSchemaPrivileges(ctx context.Context, name, role string) error
	SetRoleSearchPath(ctx context.Context, role, schema string) error
	RevokeDefaultPublicPrivileges(ctx context.Context, role, schema string) error
}

type service struct {
	repo     Repository
	settings Settings
}

func New(repo Repository, settings Settings) Usecase {
	return &service{
		repo:     repo,
		settings: settings,
	}
}

func (s *service) Ensure(ctx context.Context) error {
	if s.repo == nil {
		return errors.New("database repository is nil")
	}

	if s.settings.Name == "" {
		return errors.New("database name is empty")
	}

	if s.settings.User == "" {
		return errors.New("database user is empty")
	}

	if s.settings.Password == "" {
		return errors.New("database password is empty")
	}

	schema := s.settings.Schema
	if schema == "" {
		schema = "public"
	}

	roleExists, err := s.repo.RoleExists(ctx, s.settings.User)
	if err != nil {
		return fmt.Errorf("check database role exists: %w", err)
	}

	if !roleExists {
		if err := s.repo.CreateRole(ctx, s.settings.User); err != nil {
			return fmt.Errorf("create database role: %w", err)
		}
	}

	if err := s.repo.UpdateRolePassword(ctx, s.settings.User, s.settings.Password); err != nil {
		return fmt.Errorf("update database role password: %w", err)
	}

	exists, err := s.repo.DatabaseExists(ctx, s.settings.Name)
	if err != nil {
		return fmt.Errorf("check database exists: %w", err)
	}

	if !exists {
		if err := s.repo.CreateDatabase(ctx, s.settings.Name, s.settings.User); err != nil {
			return fmt.Errorf("create database: %w", err)
		}
	}

	if err := s.repo.SetDatabaseOwner(ctx, s.settings.Name, s.settings.User); err != nil {
		return fmt.Errorf("set database owner: %w", err)
	}

	if err := s.repo.RevokePublicDatabasePrivileges(ctx, s.settings.Name); err != nil {
		return fmt.Errorf("revoke public database privileges: %w", err)
	}

	if err := s.repo.GrantDatabasePrivileges(ctx, s.settings.Name, s.settings.User); err != nil {
		return fmt.Errorf("grant database privileges: %w", err)
	}

	if schema != "public" {
		schemaExists, err := s.repo.SchemaExists(ctx, schema)
		if err != nil {
			return fmt.Errorf("check schema exists: %w", err)
		}

		if !schemaExists {
			if err := s.repo.CreateSchema(ctx, schema, s.settings.User); err != nil {
				return fmt.Errorf("create schema: %w", err)
			}
		}

		if err := s.repo.SetSchemaOwner(ctx, schema, s.settings.User); err != nil {
			return fmt.Errorf("set schema owner: %w", err)
		}
	}

	if err := s.repo.RevokePublicSchemaPrivileges(ctx, schema); err != nil {
		return fmt.Errorf("revoke public schema privileges: %w", err)
	}

	if err := s.repo.GrantSchemaPrivileges(ctx, schema, s.settings.User); err != nil {
		return fmt.Errorf("grant schema privileges: %w", err)
	}

	if err := s.repo.SetRoleSearchPath(ctx, s.settings.User, schema); err != nil {
		return fmt.Errorf("set role search_path: %w", err)
	}

	if err := s.repo.RevokeDefaultPublicPrivileges(ctx, s.settings.User, schema); err != nil {
		return fmt.Errorf("revoke default public privileges: %w", err)
	}

	return nil
}
