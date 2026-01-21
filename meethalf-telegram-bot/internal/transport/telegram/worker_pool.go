package telegram

import (
	"context"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type WorkerPool struct {
	size    int
	queue   chan tgbotapi.Update
	handler *Handler
	stop    sync.Once
}

func NewWorkerPool(size, queueSize int, handler *Handler) *WorkerPool {
	if size < 1 {
		size = 1
	}
	if queueSize < 1 {
		queueSize = 1
	}

	return &WorkerPool{
		size:    size,
		queue:   make(chan tgbotapi.Update, queueSize),
		handler: handler,
	}
}

func (p *WorkerPool) Start(ctx context.Context) {
	if p == nil || p.handler == nil {
		return
	}

	for i := 0; i < p.size; i++ {
		go p.worker(ctx)
	}
}

func (p *WorkerPool) Submit(update tgbotapi.Update) bool {
	if p == nil {
		return false
	}

	select {
	case p.queue <- update:
		return true
	default:
		return false
	}
}

func (p *WorkerPool) Stop() {
	if p == nil {
		return
	}

	p.stop.Do(func() {
		close(p.queue)
	})
}

func (p *WorkerPool) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-p.queue:
			if !ok {
				return
			}
			p.handler.Handle(ctx, update)
		}
	}
}
