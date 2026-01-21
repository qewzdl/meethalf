package main

import (
	"context"
	"os"

	"meethalf-api/internal/config"
	"meethalf-api/internal/logger"
	"meethalf-api/internal/storage/postgres"
	"meethalf-api/internal/usecase/database"
)

func main() {
	cfg := config.Load()
	log := logger.New()

	adminCfg := cfg.DB
	adminCfg.User = cfg.DB.AdminUser
	adminCfg.Password = cfg.DB.AdminPassword

	adminDB, err := postgres.NewWithDatabase(adminCfg, cfg.DB.AdminName)
	if err != nil {
		log.Printf("failed to connect admin db: %v", err)
		os.Exit(1)
	}
	defer func() {
		if err := adminDB.Close(); err != nil {
			log.Printf("failed to close db: %v", err)
		}
	}()

	repo := postgres.NewDatabaseRepository(adminDB)
	uc := database.New(repo, database.Settings{
		Name:     cfg.DB.Name,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Schema:   cfg.DB.Schema,
	})

	ctx := context.Background()
	if cfg.DB.ConnectTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.DB.ConnectTimeout)
		defer cancel()
	}

	if err := uc.Ensure(ctx); err != nil {
		log.Printf("failed to ensure database: %v", err)
		os.Exit(1)
	}

	log.Printf("database %s is ready", cfg.DB.Name)
}
