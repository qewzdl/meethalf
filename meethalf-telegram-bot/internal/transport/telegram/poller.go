package telegram

import (
	"context"
	"errors"
	"log"
	"net"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Poller struct {
	bot            *tgbotapi.BotAPI
	pool           *WorkerPool
	allowedUpdates []string
	timeout        time.Duration
	logger         *log.Logger
	warnedDNS      bool
}

func NewPoller(bot *tgbotapi.BotAPI, pool *WorkerPool, allowedUpdates []string, timeout time.Duration, logger *log.Logger) *Poller {
	return &Poller{
		bot:            bot,
		pool:           pool,
		allowedUpdates: allowedUpdates,
		timeout:        timeout,
		logger:         logger,
	}
}

func (p *Poller) Run(ctx context.Context) error {
	if p == nil || p.bot == nil {
		return errors.New("telegram bot is not configured")
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.AllowedUpdates = p.allowedUpdates
	if p.timeout > 0 {
		updateConfig.Timeout = int(p.timeout.Seconds())
	}

	if p.pool != nil {
		p.pool.Start(ctx)
	}

	retryDelay := 3 * time.Second

	for {
		select {
		case <-ctx.Done():
			if p.pool != nil {
				p.pool.Stop()
			}
			return nil
		default:
		}

		updates, err := p.bot.GetUpdates(updateConfig)
		if err != nil {
			p.logUpdatesError(err, retryDelay)
			if !sleep(ctx, retryDelay) {
				if p.pool != nil {
					p.pool.Stop()
				}
				return nil
			}
			continue
		}

		for _, update := range updates {
			if update.UpdateID >= updateConfig.Offset {
				updateConfig.Offset = update.UpdateID + 1
			}

			if p.pool == nil {
				continue
			}
			if !p.pool.Submit(update) && p.logger != nil {
				p.logger.Printf("update dropped: queue is full")
			}
		}
	}
}

func (p *Poller) logUpdatesError(err error, retryDelay time.Duration) {
	if err == nil || p.logger == nil {
		return
	}

	message := err.Error()
	if p.bot != nil {
		message = redactToken(message, p.bot.Token)
	}

	if isDNSError(err) && !p.warnedDNS {
		p.logger.Printf("failed to get updates: %s (check DNS or set BOT_API_ENDPOINT/BOT_PROXY_URL); retrying in %s", message, retryDelay)
		p.warnedDNS = true
		return
	}

	p.logger.Printf("failed to get updates: %s; retrying in %s", message, retryDelay)
}

func isDNSError(err error) bool {
	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr)
}

func sleep(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return true
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
