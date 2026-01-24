package telegram

import (
	"context"
	"log"
	"strings"
	"sync"
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

	callbackMeta, hasCallbackMeta := h.callbackMessageMeta(update)
	if hasCallbackMeta {
		if err := h.deleteCallbackMessage(ctx, callbackMeta); err != nil && h.logger != nil {
			h.logger.Printf("delete callback message error: %v", err)
		}
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
		messageID, err := h.sender.Send(ctx, response)
		if err != nil {
			if h.logger != nil {
				if response.Kind != "" {
					h.logger.Printf("notification send failed: kind=%s chat_id=%d err=%v", response.Kind, response.ChatID, err)
				} else {
					h.logger.Printf("send response error: %v", err)
				}
			}
			continue
		}
		if h.logger != nil && response.Kind != "" {
			h.logger.Printf("notification sent: kind=%s chat_id=%d message_id=%d", response.Kind, response.ChatID, messageID)
		}
		if response.CleanupFromMessageID > 0 && messageID > 0 {
			if loadingMessageID != 0 && response.CleanupFromMessageID <= loadingMessageID && loadingMessageID <= messageID {
				loadingMessageID = 0
			}
			h.deleteMessageRange(ctx, response.ChatID, response.CleanupFromMessageID, messageID)
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
		MessageID:  update.Message.MessageID,
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
	arguments := ""
	switch {
	case strings.HasPrefix(text, domain.CommandProfileVisibility+":"):
		command = domain.CommandProfileVisibility
		arguments = strings.TrimPrefix(text, domain.CommandProfileVisibility+":")
	case strings.HasPrefix(text, domain.CommandSearchGender+":"):
		command = domain.CommandSearchGender
		arguments = strings.TrimPrefix(text, domain.CommandSearchGender+":")
	case strings.HasPrefix(text, domain.CommandSearchAccuracy+":"):
		command = domain.CommandSearchAccuracy
		arguments = strings.TrimPrefix(text, domain.CommandSearchAccuracy+":")
	case strings.HasPrefix(text, domain.CommandMatchLike+":"):
		command = domain.CommandMatchLike
		arguments = strings.TrimPrefix(text, domain.CommandMatchLike+":")
	case strings.HasPrefix(text, domain.CommandMatchDislike+":"):
		command = domain.CommandMatchDislike
		arguments = strings.TrimPrefix(text, domain.CommandMatchDislike+":")
	case strings.HasPrefix(text, domain.CommandMatchReport+":"):
		command = domain.CommandMatchReport
		arguments = strings.TrimPrefix(text, domain.CommandMatchReport+":")
	case strings.HasPrefix(text, domain.CommandMatchViewProfile+":"):
		command = domain.CommandMatchViewProfile
		arguments = strings.TrimPrefix(text, domain.CommandMatchViewProfile+":")
	case strings.HasPrefix(text, domain.CommandAdminUsers+":"):
		command = domain.CommandAdminUsers
		arguments = strings.TrimPrefix(text, domain.CommandAdminUsers+":")
	case strings.HasPrefix(text, domain.CommandAdminBannedUsers+":"):
		command = domain.CommandAdminBannedUsers
		arguments = strings.TrimPrefix(text, domain.CommandAdminBannedUsers+":")
	case strings.HasPrefix(text, domain.CommandAdminModerators+":"):
		command = domain.CommandAdminModerators
		arguments = strings.TrimPrefix(text, domain.CommandAdminModerators+":")
	case strings.HasPrefix(text, domain.CommandAdminReports+":"):
		command = domain.CommandAdminReports
		arguments = strings.TrimPrefix(text, domain.CommandAdminReports+":")
	default:
		switch text {
		case domain.CommandProfile:
			command = domain.CommandProfile
		case domain.CommandProfileSetupBack:
			command = domain.CommandProfileSetupBack
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
		case domain.CommandCancel:
			command = domain.CommandCancel
		case domain.CommandSearchStart:
			command = domain.CommandSearchStart
		case domain.CommandSearchRefresh:
			command = domain.CommandSearchRefresh
		case domain.CommandSearchGender:
			command = domain.CommandSearchGender
		case domain.CommandMatchPrevious:
			command = domain.CommandMatchPrevious
		case domain.CommandAdminMenu:
			command = domain.CommandAdminMenu
		case domain.CommandAdminUsers:
			command = domain.CommandAdminUsers
		case domain.CommandAdminBannedUsers:
			command = domain.CommandAdminBannedUsers
		case domain.CommandAdminModerators:
			command = domain.CommandAdminModerators
		case domain.CommandAdminReports:
			command = domain.CommandAdminReports
		case domain.CommandAdminBan:
			command = domain.CommandAdminBan
		case domain.CommandAdminUnban:
			command = domain.CommandAdminUnban
		case domain.CommandAdminModerator:
			command = domain.CommandAdminModerator
		case domain.CommandAdminUnmoderator:
			command = domain.CommandAdminUnmoderator
		}
	}

	receivedAt := time.Now().UTC()
	if callback.Message.Date != 0 {
		receivedAt = time.Unix(int64(callback.Message.Date), 0).UTC()
	}

	return domain.IncomingMessage{
		ChatID:     callback.Message.Chat.ID,
		MessageID:  callback.Message.MessageID,
		User:       user,
		Text:       text,
		Command:    command,
		Arguments:  arguments,
		PhotoIDs:   nil,
		ReceivedAt: receivedAt,
	}, true
}

type callbackMessageMeta struct {
	chatID    int64
	messageID int
}

func (h *Handler) callbackMessageMeta(update tgbotapi.Update) (callbackMessageMeta, bool) {
	if update.CallbackQuery == nil || update.CallbackQuery.Message == nil {
		return callbackMessageMeta{}, false
	}

	chatID := update.CallbackQuery.Message.Chat.ID
	messageID := update.CallbackQuery.Message.MessageID
	if chatID == 0 || messageID == 0 {
		return callbackMessageMeta{}, false
	}

	return callbackMessageMeta{
		chatID:    chatID,
		messageID: messageID,
	}, true
}

func (h *Handler) deleteCallbackMessage(ctx context.Context, meta callbackMessageMeta) error {
	if meta.chatID == 0 || meta.messageID == 0 {
		return nil
	}
	return h.sender.Delete(ctx, meta.chatID, meta.messageID)
}

func (h *Handler) deleteMessageRange(ctx context.Context, chatID int64, fromID, toID int) {
	if h == nil || h.sender == nil {
		return
	}
	if chatID == 0 || fromID == 0 || toID == 0 || toID < fromID {
		return
	}

	const maxParallelDeletes = 8
	total := toID - fromID + 1
	if total <= 0 {
		return
	}

	workers := total
	if workers > maxParallelDeletes {
		workers = maxParallelDeletes
	}

	messageIDs := make(chan int, total)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for messageID := range messageIDs {
				if err := h.sender.Delete(ctx, chatID, messageID); err != nil && h.logger != nil {
					h.logger.Printf("delete message error: chat_id=%d message_id=%d err=%v", chatID, messageID, err)
				}
			}
		}()
	}

	for messageID := fromID; messageID <= toID; messageID++ {
		messageIDs <- messageID
	}
	close(messageIDs)
	wg.Wait()
}
