package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"meethalf-telegram-bot/internal/config"
	redisstorage "meethalf-telegram-bot/internal/storage/redis"
	"meethalf-telegram-bot/internal/usecase/backup"
)

func main() {
	inPath := flag.String("in", "", "input backup file path")
	flag.Parse()

	filename := strings.TrimSpace(*inPath)
	if filename == "" {
		log.Fatal("input backup file is required")
	}

	cfg := config.Load()
	store := strings.ToLower(strings.TrimSpace(cfg.Session.Store))
	if store != "redis" || !cfg.Redis.Enabled {
		log.Fatal("restore requires SESSION_STORE=redis with REDIS_ENABLED=true")
	}

	client, err := redisstorage.New(cfg.Redis)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer client.Close()

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("open backup file: %v", err)
	}
	defer file.Close()

	repo := redisstorage.NewBackupRepository(client, "")
	uc := backup.New(repo)

	if err := uc.Import(context.Background(), file); err != nil {
		log.Fatalf("restore failed: %v", err)
	}

	fmt.Printf("Backup restored from %s\n", filename)
}
