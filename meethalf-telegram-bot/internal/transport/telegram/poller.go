package telegram

import (
	"context"
	"errors"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Poller struct {
	bot            *tgbotapi.BotAPI
	pool           *WorkerPool
	allowedUpdates []string
	timeout        time.Duration
	logger         *log.Logger
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

	updates := p.bot.GetUpdatesChan(updateConfig)
	if p.pool != nil {
		p.pool.Start(ctx)
	}

	for {
		select {
		case <-ctx.Done():
			p.bot.StopReceivingUpdates()
			if p.pool != nil {
				p.pool.Stop()
			}
			return nil
		case update, ok := <-updates:
			if !ok {
				if p.pool != nil {
					p.pool.Stop()
				}
				return nil
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
