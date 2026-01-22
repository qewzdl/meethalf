package telegram

import (
	"context"
	"errors"
	"strings"

	"meethalf-telegram-bot/internal/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Sender interface {
	Send(ctx context.Context, msg domain.OutgoingMessage) (int, error)
	Delete(ctx context.Context, chatID int64, messageID int) error
}

type BotSender struct {
	bot *tgbotapi.BotAPI
}

func NewSender(bot *tgbotapi.BotAPI) *BotSender {
	return &BotSender{bot: bot}
}

func (s *BotSender) Send(ctx context.Context, msg domain.OutgoingMessage) (int, error) {
	if s == nil || s.bot == nil {
		return 0, errors.New("telegram sender is not configured")
	}

	if err := ctx.Err(); err != nil {
		return 0, err
	}

	var callbackErr error
	if msg.CallbackQueryID != "" {
		callback := tgbotapi.NewCallback(msg.CallbackQueryID, msg.CallbackText)
		_, callbackErr = s.bot.Request(callback)
	}

	if msg.ChatID == 0 {
		return 0, callbackErr
	}

	if len(msg.PhotoIDs) > 0 {
		photoIDs := make([]string, 0, len(msg.PhotoIDs))
		for _, id := range msg.PhotoIDs {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			photoIDs = append(photoIDs, id)
		}
		if len(photoIDs) == 0 {
			return 0, callbackErr
		}

		if len(photoIDs) == 1 {
			photo := tgbotapi.NewPhoto(msg.ChatID, tgbotapi.FileID(photoIDs[0]))
			photo.Caption = msg.Text
			photo.ParseMode = msg.ParseMode
			if msg.InlineKeyboard != nil {
				keyboard := make([][]tgbotapi.InlineKeyboardButton, 0, len(msg.InlineKeyboard.Buttons))
				for _, row := range msg.InlineKeyboard.Buttons {
					buttons := make([]tgbotapi.InlineKeyboardButton, 0, len(row))
					for _, button := range row {
						if button.CallbackData == "" {
							continue
						}
						buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(button.Text, button.CallbackData))
					}
					if len(buttons) > 0 {
						keyboard = append(keyboard, buttons)
					}
				}
				if len(keyboard) > 0 {
					photo.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
				}
			}

			sent, err := s.bot.Send(photo)
			return sent.MessageID, errors.Join(callbackErr, err)
		}

		media := make([]interface{}, 0, len(photoIDs))
		captionApplied := false
		for _, id := range photoIDs {
			item := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(id))
			if !captionApplied && msg.Text != "" {
				item.Caption = msg.Text
				item.ParseMode = msg.ParseMode
				captionApplied = true
			}
			media = append(media, item)
		}
		if len(media) == 0 {
			return 0, callbackErr
		}

		group := tgbotapi.NewMediaGroup(msg.ChatID, media)
		messages, err := s.bot.SendMediaGroup(group)
		messageID := 0
		if len(messages) > 0 {
			messageID = messages[0].MessageID
		}
		return messageID, errors.Join(callbackErr, err)
	}

	if msg.Text == "" {
		return 0, callbackErr
	}

	message := tgbotapi.NewMessage(msg.ChatID, msg.Text)
	message.ParseMode = msg.ParseMode
	message.DisableWebPagePreview = msg.DisablePreview
	if msg.InlineKeyboard != nil {
		keyboard := make([][]tgbotapi.InlineKeyboardButton, 0, len(msg.InlineKeyboard.Buttons))
		for _, row := range msg.InlineKeyboard.Buttons {
			buttons := make([]tgbotapi.InlineKeyboardButton, 0, len(row))
			for _, button := range row {
				if button.CallbackData == "" {
					continue
				}
				buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(button.Text, button.CallbackData))
			}
			if len(buttons) > 0 {
				keyboard = append(keyboard, buttons)
			}
		}

		message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	}

	sent, err := s.bot.Send(message)
	return sent.MessageID, errors.Join(callbackErr, err)
}

func (s *BotSender) Delete(ctx context.Context, chatID int64, messageID int) error {
	if s == nil || s.bot == nil {
		return errors.New("telegram sender is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if chatID == 0 || messageID == 0 {
		return nil
	}

	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := s.bot.Request(deleteMsg)
	return err
}
