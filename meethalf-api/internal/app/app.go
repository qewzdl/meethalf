package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"meethalf-api/internal/config"
	"meethalf-api/internal/ratelimit"
	"meethalf-api/internal/storage/postgres"
	redisstorage "meethalf-api/internal/storage/redis"
	"meethalf-api/internal/transport/httpserver"
	"meethalf-api/internal/transport/openrouter"
	"meethalf-api/internal/usecase/admin"
	"meethalf-api/internal/usecase/health"
	"meethalf-api/internal/usecase/matching"
	"meethalf-api/internal/usecase/profile"

	redisgo "github.com/redis/go-redis/v9"
)

type App struct {
	cfg    config.Config
	logger *log.Logger
	server *http.Server
	db     *sql.DB
	redis  *redisgo.Client
}

func New(cfg config.Config, logger *log.Logger) (*App, error) {
	db, err := postgres.New(cfg.DB)
	if err != nil {
		return nil, err
	}

	var redisClient *redisgo.Client
	if cfg.Redis.Enabled {
		client, err := redisstorage.New(cfg.Redis)
		if err != nil {
			_ = db.Close()
			return nil, err
		}
		redisClient = client
	}

	healthRepos := []health.Repository{postgres.NewHealthRepository(db, cfg.Health.DBTimeout)}
	if redisClient != nil {
		healthRepos = append(healthRepos, redisstorage.NewHealthRepository(redisClient, cfg.Health.RedisTimeout))
	}
	healthRepo := health.NewCompositeRepository(healthRepos...)
	healthUC := health.New(healthRepo)

	profileRepo, err := postgres.NewProfileRepository(db, cfg.DB.Schema)
	if err != nil {
		if redisClient != nil {
			_ = redisClient.Close()
		}
		_ = db.Close()
		return nil, err
	}
	profileUC := profile.New(profileRepo)
	adminUC := admin.New(profileRepo)
	matchingRepo, err := postgres.NewMatchingRepository(db, cfg.DB.Schema)
	if err != nil {
		if redisClient != nil {
			_ = redisClient.Close()
		}
		_ = db.Close()
		return nil, err
	}
	var aiAnalyzer matching.AIAnalyzer
	if cfg.OpenRouter.APIKey != "" {
		aiAnalyzer = openrouter.New(
			cfg.OpenRouter.APIKey,
			cfg.OpenRouter.BaseURL,
			cfg.OpenRouter.Model,
			cfg.OpenRouter.Timeout,
		)
	}
	matchingUC := matching.New(matchingRepo, aiAnalyzer)
	handlers := httpserver.NewHandlers(healthUC, profileUC, matchingUC, adminUC)

	var limiter ratelimit.Limiter
	if cfg.RateLimit.Enabled {
		rateCfg := ratelimit.Config{
			Requests: cfg.RateLimit.Requests,
			Window:   cfg.RateLimit.Window,
			Burst:    cfg.RateLimit.Burst,
		}

		store := strings.ToLower(strings.TrimSpace(cfg.RateLimit.Store))
		switch store {
		case "", "memory", "inmemory":
			rateLimiter, err := ratelimit.NewInMemory(rateCfg)
			if err != nil {
				if redisClient != nil {
					_ = redisClient.Close()
				}
				_ = db.Close()
				return nil, err
			}
			limiter = rateLimiter
		case "redis":
			if redisClient == nil {
				_ = db.Close()
				return nil, errors.New("redis rate limit store requested but redis is not enabled")
			}
			rateLimiter, err := ratelimit.NewRedis(redisClient, rateCfg)
			if err != nil {
				_ = redisClient.Close()
				_ = db.Close()
				return nil, err
			}
			limiter = rateLimiter
		default:
			if redisClient != nil {
				_ = redisClient.Close()
			}
			_ = db.Close()
			return nil, fmt.Errorf("unsupported rate limit store: %s", store)
		}
	}

	router := httpserver.NewRouter(handlers, limiter)

	server := &http.Server{
		Addr:         cfg.HTTP.Address(),
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	return &App{
		cfg:    cfg,
		logger: logger,
		server: server,
		db:     db,
		redis:  redisClient,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		if a.logger != nil {
			a.logger.Printf("http server listening on %s", a.server.Addr)
		}

		errCh <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.HTTP.ShutdownTimeout)
		defer cancel()
		err := a.server.Shutdown(shutdownCtx)
		a.closeRedis()
		a.closeDB()
		return err
	case err := <-errCh:
		a.closeRedis()
		a.closeDB()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func (a *App) closeDB() {
	if a == nil || a.db == nil {
		return
	}

	if err := a.db.Close(); err != nil && a.logger != nil {
		a.logger.Printf("failed to close db: %v", err)
	}
}

func (a *App) closeRedis() {
	if a == nil || a.redis == nil {
		return
	}

	if err := a.redis.Close(); err != nil && a.logger != nil {
		a.logger.Printf("failed to close redis: %v", err)
	}
}
