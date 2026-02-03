package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"meethalf-api/internal/config"
	"meethalf-api/internal/storage/postgres"
	"meethalf-api/internal/usecase/backup"
)

func main() {
	inPath := flag.String("in", "", "input backup file path")
	flag.Parse()

	filename := strings.TrimSpace(*inPath)
	if filename == "" {
		log.Fatal("input backup file is required")
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("open backup file: %v", err)
	}
	defer file.Close()

	cfg := config.Load()
	repo := postgres.NewBackupRepository(cfg.DB)
	uc := backup.New(repo)

	ctx := context.Background()
	if err := uc.Restore(ctx, file); err != nil {
		log.Fatalf("restore failed: %v", err)
	}

	fmt.Printf("Backup restored from %s\n", filename)
}
