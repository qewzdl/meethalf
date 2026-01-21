package telegram

import (
	"context"
	"errors"

	"meethalf-telegram-bot/internal/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Sender interface {
	Send(ctx context.Context, msg domain.OutgoingMessage) error
}

type BotSender struct {
	bot *tgbotapi.BotAPI
}

func NewSender(bot *tgbotapi.BotAPI) *BotSender {
	return &BotSender{bot: bot}
}

func (s *BotSender) Send(ctx context.Context, msg domain.OutgoingMessage) error {
	if s == nil || s.bot == nil {
		return errors.New("telegram sender is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if msg.ChatID == 0 || msg.Text == "" {
		return nil
	}

	message := tgbotapi.NewMessage(msg.ChatID, msg.Text)
	message.ParseMode = msg.ParseMode
	message.DisableWebPagePreview = msg.DisablePreview

	_, err := s.bot.Send(message)
	return err
}
