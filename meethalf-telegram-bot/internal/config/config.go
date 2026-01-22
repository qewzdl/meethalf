package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type BotConfig struct {
	Token          string
	Debug          bool
	AllowedUpdates []string
	PollingTimeout time.Duration
}

type WorkerConfig struct {
	PoolSize  int
	QueueSize int
}

type SessionConfig struct {
	Store string
	TTL   time.Duration
}

type RedisConfig struct {
	Enabled        bool
	Host           string
	Port           string
	Username       string
	Password       string
	DB             int
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	PoolSize       int
	MinIdleConns   int
}

type APIConfig struct {
	BaseURL string
	Timeout time.Duration
}

type Config struct {
	Env     string
	Bot     BotConfig
	Workers WorkerConfig
	Session SessionConfig
	Redis   RedisConfig
	API     APIConfig
}

func Load() Config {
	loadEnvFile(".env")

	redisConnectTimeout := getDuration("REDIS_CONNECT_TIMEOUT", "5s")
	redisReadTimeout := getDuration("REDIS_READ_TIMEOUT", "3s")
	redisWriteTimeout := getDuration("REDIS_WRITE_TIMEOUT", "3s")
	apiTimeout := getDuration("API_TIMEOUT", "5s")

	return Config{
		Env: getEnv("APP_ENV", "dev"),
		Bot: BotConfig{
			Token:          getEnv("BOT_TOKEN", ""),
			Debug:          getEnvBool("BOT_DEBUG", false),
			AllowedUpdates: getEnvCSV("BOT_ALLOWED_UPDATES", "message,callback_query"),
			PollingTimeout: getDuration("BOT_POLLING_TIMEOUT", "10s"),
		},
		Workers: WorkerConfig{
			PoolSize:  getEnvInt("BOT_WORKERS", 4),
			QueueSize: getEnvInt("BOT_QUEUE_SIZE", 100),
		},
		Session: SessionConfig{
			Store: getEnv("SESSION_STORE", "memory"),
			TTL:   getDuration("SESSION_TTL", "24h"),
		},
		Redis: RedisConfig{
			Enabled:        getEnvBool("REDIS_ENABLED", false),
			Host:           getEnv("REDIS_HOST", "localhost"),
			Port:           getEnv("REDIS_PORT", "6379"),
			Username:       getEnv("REDIS_USERNAME", ""),
			Password:       getEnv("REDIS_PASSWORD", ""),
			DB:             getEnvInt("REDIS_DB", 0),
			ConnectTimeout: redisConnectTimeout,
			ReadTimeout:    redisReadTimeout,
			WriteTimeout:   redisWriteTimeout,
			PoolSize:       getEnvInt("REDIS_POOL_SIZE", 0),
			MinIdleConns:   getEnvInt("REDIS_MIN_IDLE_CONNS", 0),
		},
		API: APIConfig{
			BaseURL: getEnv("API_BASE_URL", "http://localhost:8080"),
			Timeout: apiTimeout,
		},
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getDuration(key, fallback string) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		value = fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		parsed, _ = time.ParseDuration(fallback)
	}

	return parsed
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvCSV(key, fallback string) []string {
	value := os.Getenv(key)
	if value == "" {
		value = fallback
	}

	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		items = append(items, trimmed)
	}

	return items
}
