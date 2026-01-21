package telegram

import (
	"context"
	"log"
	"time"

	"meethalf-telegram-bot/internal/domain"
	"meethalf-telegram-bot/internal/usecase/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	usecase bot.Usecase
	sender  Sender
	logger  *log.Logger
}

func NewHandler(usecase bot.Usecase, sender Sender, logger *log.Logger) *Handler {
	return &Handler{
		usecase: usecase,
		sender:  sender,
		logger:  logger,
	}
}

func (h *Handler) Handle(ctx context.Context, update tgbotapi.Update) {
	msg, ok := h.toIncoming(update)
	if !ok {
		return
	}

	response, err := h.usecase.Handle(ctx, msg)
	if err != nil && h.logger != nil {
		h.logger.Printf("usecase error: %v", err)
	}

	if err := h.sender.Send(ctx, response); err != nil && h.logger != nil {
		h.logger.Printf("send response error: %v", err)
	}
}

func (h *Handler) toIncoming(update tgbotapi.Update) (domain.IncomingMessage, bool) {
	if update.Message == nil {
		return domain.IncomingMessage{}, false
	}

	text := update.Message.Text
	if text == "" {
		return domain.IncomingMessage{}, false
	}

	user := domain.User{}
	if update.Message.From != nil {
		user = domain.User{
			ID:           int64(update.Message.From.ID),
			Username:     update.Message.From.UserName,
			FirstName:    update.Message.From.FirstName,
			LastName:     update.Message.From.LastName,
			LanguageCode: update.Message.From.LanguageCode,
		}
	}

	command := ""
	arguments := ""
	if update.Message.IsCommand() {
		command = update.Message.Command()
		arguments = update.Message.CommandArguments()
	}

	receivedAt := time.Unix(int64(update.Message.Date), 0).UTC()
	if update.Message.Date == 0 {
		receivedAt = time.Now().UTC()
	}

	return domain.IncomingMessage{
		ChatID:     update.Message.Chat.ID,
		User:       user,
		Text:       text,
		Command:    command,
		Arguments:  arguments,
		ReceivedAt: receivedAt,
	}, true
}
