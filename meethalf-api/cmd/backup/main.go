package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"meethalf-api/internal/config"
	"meethalf-api/internal/storage/postgres"
	"meethalf-api/internal/usecase/backup"
)

func main() {
	outPath := flag.String("out", "", "output backup file path")
	flag.Parse()

	cfg := config.Load()
	repo := postgres.NewBackupRepository(cfg.DB)
	uc := backup.New(repo)

	ctx := context.Background()
	snapshot, err := uc.Create(ctx)
	if err != nil {
		log.Fatalf("backup failed: %v", err)
	}

	filename := strings.TrimSpace(*outPath)
	if filename == "" {
		filename = snapshot.Filename
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("create backup file: %v", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, snapshot.Reader); err != nil {
		log.Fatalf("write backup: %v", err)
	}
	if err := snapshot.Reader.Close(); err != nil {
		log.Fatalf("finalize backup: %v", err)
	}

	fmt.Printf("Backup saved to %s\n", filename)
}
