package telegram

import (
	"context"
	"log"
	"strings"
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

	callbackHandled := false
	loadingChatID := int64(0)
	loadingMessageID := 0
	if provider, ok := h.usecase.(bot.LoadingMessageProvider); ok {
		loading, shouldSend, err := provider.LoadingMessage(ctx, msg)
		if err != nil && h.logger != nil {
			h.logger.Printf("loading message error: %v", err)
		}
		if shouldSend {
			if update.CallbackQuery != nil {
				loading.CallbackQueryID = update.CallbackQuery.ID
				callbackHandled = true
			}
			messageID, err := h.sender.Send(ctx, loading)
			if err != nil && h.logger != nil {
				h.logger.Printf("send loading message error: %v", err)
			}
			if messageID != 0 {
				loadingChatID = loading.ChatID
				loadingMessageID = messageID
			}
		}
	}

	responses, err := h.usecase.Handle(ctx, msg)
	if err != nil && h.logger != nil {
		h.logger.Printf("usecase error: %v", err)
	}

	if update.CallbackQuery != nil && len(responses) > 0 && !callbackHandled {
		responses[0].CallbackQueryID = update.CallbackQuery.ID
	}

	for _, response := range responses {
		if _, err := h.sender.Send(ctx, response); err != nil && h.logger != nil {
			h.logger.Printf("send response error: %v", err)
		}
	}

	if loadingMessageID != 0 {
		if err := h.sender.Delete(ctx, loadingChatID, loadingMessageID); err != nil && h.logger != nil {
			h.logger.Printf("delete loading message error: %v", err)
		}
	}
}

func (h *Handler) toIncoming(update tgbotapi.Update) (domain.IncomingMessage, bool) {
	if update.Message == nil {
		return h.fromCallback(update.CallbackQuery)
	}

	text := update.Message.Text
	if text == "" && update.Message.Caption != "" {
		text = update.Message.Caption
	}

	photoIDs := []string{}
	if len(update.Message.Photo) > 0 {
		photo := update.Message.Photo[len(update.Message.Photo)-1]
		if photo.FileID != "" {
			photoIDs = append(photoIDs, photo.FileID)
		}
	}

	if text == "" && len(photoIDs) == 0 {
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
		PhotoIDs:   photoIDs,
		ReceivedAt: receivedAt,
	}, true
}

func (h *Handler) fromCallback(callback *tgbotapi.CallbackQuery) (domain.IncomingMessage, bool) {
	if callback == nil || callback.Message == nil {
		return domain.IncomingMessage{}, false
	}

	user := domain.User{}
	if callback.From != nil {
		user = domain.User{
			ID:           int64(callback.From.ID),
			Username:     callback.From.UserName,
			FirstName:    callback.From.FirstName,
			LastName:     callback.From.LastName,
			LanguageCode: callback.From.LanguageCode,
		}
	}

	text := strings.TrimSpace(callback.Data)
	command := ""
	switch text {
	case domain.CommandProfile:
		command = domain.CommandProfile
	case domain.CommandProfileView:
		command = domain.CommandProfileView
	case domain.CommandProfilePreview:
		command = domain.CommandProfilePreview
	case domain.CommandProfileEdit:
		command = domain.CommandProfileEdit
	case domain.CommandProfileEditName:
		command = domain.CommandProfileEditName
	case domain.CommandProfileEditGender:
		command = domain.CommandProfileEditGender
	case domain.CommandProfileEditBirthDate:
		command = domain.CommandProfileEditBirthDate
	case domain.CommandProfileEditCountry:
		command = domain.CommandProfileEditCountry
	case domain.CommandProfileEditCity:
		command = domain.CommandProfileEditCity
	case domain.CommandProfileEditDesc:
		command = domain.CommandProfileEditDesc
	case domain.CommandProfileEditEmoji:
		command = domain.CommandProfileEditEmoji
	case domain.CommandProfileEditPhotos:
		command = domain.CommandProfileEditPhotos
	case domain.CommandProfileSettings:
		command = domain.CommandProfileSettings
	case domain.CommandProfileDelete:
		command = domain.CommandProfileDelete
	case domain.CommandProfileDeleteConfirm:
		command = domain.CommandProfileDeleteConfirm
	case domain.CommandProfileDeleteCancel:
		command = domain.CommandProfileDeleteCancel
	}

	receivedAt := time.Now().UTC()
	if callback.Message.Date != 0 {
		receivedAt = time.Unix(int64(callback.Message.Date), 0).UTC()
	}

	return domain.IncomingMessage{
		ChatID:     callback.Message.Chat.ID,
		User:       user,
		Text:       text,
		Command:    command,
		Arguments:  "",
		PhotoIDs:   nil,
		ReceivedAt: receivedAt,
	}, true
}
