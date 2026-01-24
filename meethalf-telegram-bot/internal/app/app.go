package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	redisgo "github.com/redis/go-redis/v9"
	"meethalf-telegram-bot/internal/config"
	"meethalf-telegram-bot/internal/storage/memory"
	redisstorage "meethalf-telegram-bot/internal/storage/redis"
	"meethalf-telegram-bot/internal/transport/api"
	"meethalf-telegram-bot/internal/transport/telegram"
	botusecase "meethalf-telegram-bot/internal/usecase/bot"
)

type App struct {
	cfg     config.Config
	logger  *log.Logger
	poller  *telegram.Poller
	botName string
	redis   *redisgo.Client
}

func New(cfg config.Config, logger *log.Logger) (*App, error) {
	if strings.TrimSpace(cfg.Bot.Token) == "" {
		return nil, errors.New("BOT_TOKEN is required")
	}

	bot, err := telegram.NewBot(telegram.BotConfig{
		Token:       cfg.Bot.Token,
		Debug:       cfg.Bot.Debug,
		APIEndpoint: cfg.Bot.APIEndpoint,
		ProxyURL:    cfg.Bot.ProxyURL,
	})
	if err != nil {
		return nil, fmt.Errorf("init telegram bot: %w", err)
	}

	var redisClient *redisgo.Client
	store := strings.ToLower(strings.TrimSpace(cfg.Session.Store))
	var sessionRepo botusecase.SessionRepository
	var draftRepo botusecase.ProfileDraftRepository
	var deleteConfirmRepo botusecase.ProfileDeletionConfirmationRepository
	var adminActionRepo botusecase.AdminActionRepository

	switch store {
	case "", "memory", "inmemory":
		sessionRepo = memory.NewSessionRepository()
		draftRepo = memory.NewProfileDraftRepository()
		deleteConfirmRepo = memory.NewProfileDeletionConfirmationRepository()
		adminActionRepo = memory.NewAdminActionRepository(cfg.Session.TTL)
	case "redis":
		if !cfg.Redis.Enabled {
			return nil, errors.New("redis session store requested but redis is not enabled")
		}
		client, err := redisstorage.New(cfg.Redis)
		if err != nil {
			return nil, err
		}
		redisClient = client
		sessionRepo = redisstorage.NewSessionRepository(redisClient, cfg.Session.TTL)
		draftRepo = redisstorage.NewProfileDraftRepository(redisClient, cfg.Session.TTL)
		deleteConfirmRepo = redisstorage.NewProfileDeletionConfirmationRepository(redisClient, cfg.Session.TTL)
		adminActionRepo = redisstorage.NewAdminActionRepository(redisClient, cfg.Session.TTL)
	default:
		return nil, fmt.Errorf("unsupported session store: %s", store)
	}

	profileClient := api.NewProfileClient(cfg.API.BaseURL, cfg.API.Timeout)
	searchClient := api.NewSearchClient(cfg.API.BaseURL, cfg.API.Timeout)
	adminClient := api.NewAdminClient(cfg.API.BaseURL, cfg.API.Timeout)
	usecase := botusecase.New(sessionRepo, draftRepo, deleteConfirmRepo, adminActionRepo, profileClient, searchClient, adminClient, cfg.Bot.AdminUsernames)
	sender := telegram.NewSender(bot)
	handler := telegram.NewHandler(usecase, sender, logger)
	pool := telegram.NewWorkerPool(cfg.Workers.PoolSize, cfg.Workers.QueueSize, handler)
	poller := telegram.NewPoller(bot, pool, cfg.Bot.AllowedUpdates, cfg.Bot.PollingTimeout, logger)

	return &App{
		cfg:     cfg,
		logger:  logger,
		poller:  poller,
		botName: bot.Self.UserName,
		redis:   redisClient,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	if a == nil || a.poller == nil {
		return errors.New("app is not initialized")
	}

	if a.logger != nil {
		name := strings.TrimSpace(a.botName)
		if name == "" {
			name = "bot"
		}
		a.logger.Printf("telegram bot %s started", name)
	}

	err := a.poller.Run(ctx)
	a.closeRedis()
	return err
}

func (a *App) closeRedis() {
	if a == nil || a.redis == nil {
		return
	}

	if err := a.redis.Close(); err != nil && a.logger != nil {
		a.logger.Printf("failed to close redis: %v", err)
	}
}
