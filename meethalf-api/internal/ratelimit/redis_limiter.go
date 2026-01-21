package ratelimit

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisKeyPrefix = "rate_limit:"

var redisAllowScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local burst = tonumber(ARGV[3])
local expiry = tonumber(ARGV[4])

local data = redis.call("HMGET", key, "tokens", "ts")
local tokens = tonumber(data[1])
local ts = tonumber(data[2])

if tokens == nil then
  tokens = burst
  ts = now
else
  if ts == nil then
    ts = now
  end
  local elapsed = (now - ts) / 1000000000
  if elapsed > 0 then
    tokens = tokens + (elapsed * rate)
    if tokens > burst then
      tokens = burst
    end
  end
end

local allowed = 0
if tokens >= 1 then
  tokens = tokens - 1
  allowed = 1
end

redis.call("HSET", key, "tokens", tokens, "ts", now)
if expiry ~= nil and expiry > 0 then
  redis.call("PEXPIRE", key, expiry)
end

return allowed
`)

type Redis struct {
	client *redis.Client
	rate   float64
	burst  float64
	expiry time.Duration
	prefix string
}

func NewRedis(client *redis.Client, cfg Config) (*Redis, error) {
	if client == nil {
		return nil, errors.New("redis client is nil")
	}

	if cfg.Requests <= 0 {
		return nil, errors.New("rate limit requests must be positive")
	}

	if cfg.Window <= 0 {
		return nil, errors.New("rate limit window must be positive")
	}

	ratePerSecond := float64(cfg.Requests) / cfg.Window.Seconds()
	if ratePerSecond <= 0 {
		return nil, errors.New("rate limit rate must be positive")
	}

	burst := cfg.Burst
	if burst <= 0 {
		burst = cfg.Requests
	}

	expiry := cfg.Window * 2
	if expiry < time.Second {
		expiry = time.Second
	}

	return &Redis{
		client: client,
		rate:   ratePerSecond,
		burst:  float64(burst),
		expiry: expiry,
		prefix: redisKeyPrefix,
	}, nil
}

func (l *Redis) Allow(key string) bool {
	if l == nil || l.client == nil {
		return true
	}

	if key == "" {
		key = "unknown"
	}

	result, err := redisAllowScript.Run(
		context.Background(),
		l.client,
		[]string{l.prefix + key},
		time.Now().UnixNano(),
		l.rate,
		l.burst,
		l.expiry.Milliseconds(),
	).Int()
	if err != nil {
		return true
	}

	return result == 1
}
