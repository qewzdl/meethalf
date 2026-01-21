package ratelimit

import (
	"errors"
	"sync"
	"time"
)

type Limiter interface {
	Allow(key string) bool
}

type Config struct {
	Requests int
	Window   time.Duration
	Burst    int
}

type InMemory struct {
	mu              sync.Mutex
	rate            float64
	burst           float64
	cleanupInterval time.Duration
	expiry          time.Duration
	nextCleanup     time.Time
	clients         map[string]*client
}

type client struct {
	tokens   float64
	lastSeen time.Time
}

func NewInMemory(cfg Config) (*InMemory, error) {
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

	cleanupInterval := cfg.Window
	if cleanupInterval < time.Second {
		cleanupInterval = time.Second
	}

	expiry := cfg.Window * 2
	if expiry < cleanupInterval {
		expiry = cleanupInterval * 2
	}

	now := time.Now()
	return &InMemory{
		rate:            ratePerSecond,
		burst:           float64(burst),
		cleanupInterval: cleanupInterval,
		expiry:          expiry,
		nextCleanup:     now.Add(cleanupInterval),
		clients:         make(map[string]*client),
	}, nil
}

func (l *InMemory) Allow(key string) bool {
	if l == nil {
		return true
	}

	if key == "" {
		key = "unknown"
	}

	now := time.Now()
	l.mu.Lock()
	if l.cleanupInterval > 0 && now.After(l.nextCleanup) {
		l.cleanup(now)
		l.nextCleanup = now.Add(l.cleanupInterval)
	}

	entry := l.clients[key]
	if entry == nil {
		entry = &client{
			tokens:   l.burst,
			lastSeen: now,
		}
		l.clients[key] = entry
	}

	allowed := entry.allow(now, l.rate, l.burst)
	l.mu.Unlock()
	return allowed
}

func (l *InMemory) cleanup(now time.Time) {
	if l.expiry <= 0 {
		return
	}

	expiry := now.Add(-l.expiry)
	for key, entry := range l.clients {
		if entry == nil || entry.lastSeen.Before(expiry) {
			delete(l.clients, key)
		}
	}
}

func (c *client) allow(now time.Time, rate, burst float64) bool {
	if c.lastSeen.IsZero() {
		c.lastSeen = now
	}

	elapsed := now.Sub(c.lastSeen).Seconds()
	if elapsed > 0 {
		c.tokens += elapsed * rate
		if c.tokens > burst {
			c.tokens = burst
		}
	}

	c.lastSeen = now

	if c.tokens >= 1 {
		c.tokens -= 1
		return true
	}

	return false
}

type Noop struct{}

func (Noop) Allow(string) bool {
	return true
}
