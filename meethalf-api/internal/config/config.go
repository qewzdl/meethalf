package config

import (
	"os"
	"strconv"
	"time"
)

type HTTPConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func (h HTTPConfig) Address() string {
	if h.Host == "" {
		return ":" + h.Port
	}

	return h.Host + ":" + h.Port
}

type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	Schema          string
	AdminName       string
	AdminUser       string
	AdminPassword   string
	SSLMode         string
	ConnectTimeout  time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
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

type HealthConfig struct {
	DBTimeout    time.Duration
	RedisTimeout time.Duration
}

type RateLimitConfig struct {
	Enabled  bool
	Store    string
	Requests int
	Window   time.Duration
	Burst    int
}

type OpenRouterConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

type MatchingConfig struct {
	AlgorithmVersion string
}

type Config struct {
	Env        string
	HTTP       HTTPConfig
	DB         DBConfig
	Redis      RedisConfig
	Health     HealthConfig
	RateLimit  RateLimitConfig
	OpenRouter OpenRouterConfig
	Matching   MatchingConfig
}

func Load() Config {
	dbConnectTimeout := getDuration("DB_CONNECT_TIMEOUT", "5s")
	dbConnMaxLifetime := getDuration("DB_CONN_MAX_LIFETIME", "30m")
	dbConnMaxIdleTime := getDuration("DB_CONN_MAX_IDLE_TIME", "5m")
	redisConnectTimeout := getDuration("REDIS_CONNECT_TIMEOUT", "5s")
	redisReadTimeout := getDuration("REDIS_READ_TIMEOUT", "3s")
	redisWriteTimeout := getDuration("REDIS_WRITE_TIMEOUT", "3s")
	openRouterTimeout := getDuration("OPENROUTER_TIMEOUT", "12s")

	return Config{
		Env: getEnv("APP_ENV", "dev"),
		HTTP: HTTPConfig{
			Host:            getEnv("HTTP_HOST", "0.0.0.0"),
			Port:            getEnv("HTTP_PORT", "8080"),
			ReadTimeout:     getDuration("HTTP_READ_TIMEOUT", "5s"),
			WriteTimeout:    getDuration("HTTP_WRITE_TIMEOUT", "10s"),
			IdleTimeout:     getDuration("HTTP_IDLE_TIMEOUT", "30s"),
			ShutdownTimeout: getDuration("HTTP_SHUTDOWN_TIMEOUT", "10s"),
		},
		DB: DBConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "meethalf_app"),
			Password:        getEnv("DB_PASSWORD", "meethalf_app"),
			Name:            getEnv("DB_NAME", "meethalf"),
			Schema:          getEnv("DB_SCHEMA", "meethalf"),
			AdminName:       getEnv("DB_ADMIN_NAME", "postgres"),
			AdminUser:       getEnv("DB_ADMIN_USER", "postgres"),
			AdminPassword:   getEnv("DB_ADMIN_PASSWORD", "postgres"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			ConnectTimeout:  dbConnectTimeout,
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: dbConnMaxLifetime,
			ConnMaxIdleTime: dbConnMaxIdleTime,
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
		Health: HealthConfig{
			DBTimeout:    getDuration("HEALTH_DB_TIMEOUT", dbConnectTimeout.String()),
			RedisTimeout: getDuration("HEALTH_REDIS_TIMEOUT", redisReadTimeout.String()),
		},
		RateLimit: RateLimitConfig{
			Enabled:  getEnvBool("RATE_LIMIT_ENABLED", true),
			Store:    getEnv("RATE_LIMIT_STORE", "memory"),
			Requests: getEnvInt("RATE_LIMIT_REQUESTS", 60),
			Window:   getDuration("RATE_LIMIT_WINDOW", "1m"),
			Burst:    getEnvInt("RATE_LIMIT_BURST", 60),
		},
		OpenRouter: OpenRouterConfig{
			APIKey:  getEnv("OPENROUTER_API_KEY", ""),
			BaseURL: getEnv("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
			Model:   getEnv("OPENROUTER_MODEL", "openai/gpt-4o-mini"),
			Timeout: openRouterTimeout,
		},
		Matching: MatchingConfig{
			AlgorithmVersion: getEnv("MATCHING_ALGORITHM_VERSION", "matching_v1"),
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
