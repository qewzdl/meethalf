package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"meethalf-telegram-bot/internal/config"
	redisstorage "meethalf-telegram-bot/internal/storage/redis"
	"meethalf-telegram-bot/internal/usecase/backup"
)

func main() {
	outPath := flag.String("out", "", "output backup file path")
	flag.Parse()

	cfg := config.Load()
	store := strings.ToLower(strings.TrimSpace(cfg.Session.Store))
	if store != "redis" || !cfg.Redis.Enabled {
		log.Fatal("backup requires SESSION_STORE=redis with REDIS_ENABLED=true")
	}

	client, err := redisstorage.New(cfg.Redis)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer client.Close()

	repo := redisstorage.NewBackupRepository(client, "")
	uc := backup.New(repo)

	filename := strings.TrimSpace(*outPath)
	if filename == "" {
		filename = fmt.Sprintf("meethalf-telegram-bot-backup-%s.json", time.Now().UTC().Format("20060102-150405"))
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("create backup file: %v", err)
	}
	defer file.Close()

	meta, err := uc.Export(context.Background(), file)
	if err != nil {
		log.Fatalf("backup failed: %v", err)
	}

	fmt.Printf("Backup saved to %s (entries: %d)\n", filename, meta.Entries)
}
