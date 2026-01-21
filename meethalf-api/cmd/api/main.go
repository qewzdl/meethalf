package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"meethalf-api/internal/app"
	"meethalf-api/internal/config"
	"meethalf-api/internal/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New()

	application, err := app.New(cfg, log)
	if err != nil {
		log.Printf("failed to init app: %v", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := application.Run(ctx); err != nil {
		log.Printf("server stopped with error: %v", err)
		os.Exit(1)
	}
}
