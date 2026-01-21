package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"meethalf-telegram-bot/internal/domain"
)

type Usecase interface {
	Handle(ctx context.Context, msg domain.IncomingMessage) (domain.OutgoingMessage, error)
}

type SessionRepository interface {
	Touch(ctx context.Context, session domain.Session) error
}

type service struct {
	sessions SessionRepository
	helpText string
}

const defaultHelpText = "Commands:\n/start - greet by name\n/help - show help\n/ping - health check"

func New(sessions SessionRepository) Usecase {
	return &service{
		sessions: sessions,
		helpText: defaultHelpText,
	}
}

func (s *service) Handle(ctx context.Context, msg domain.IncomingMessage) (domain.OutgoingMessage, error) {
	var touchErr error
	if s != nil && s.sessions != nil {
		touchErr = s.sessions.Touch(ctx, domain.Session{
			UserID:   msg.User.ID,
			ChatID:   msg.ChatID,
			LastSeen: s.now(msg.ReceivedAt),
		})
	}

	text := s.replyText(msg)
	return domain.OutgoingMessage{
		ChatID: msg.ChatID,
		Text:   text,
	}, touchErr
}

func (s *service) replyText(msg domain.IncomingMessage) string {
	switch msg.Command {
	case domain.CommandStart:
		return s.startGreeting(msg.User) + "\n" + s.helpText
	case domain.CommandHelp:
		return s.helpText
	case domain.CommandPing:
		return fmt.Sprintf("pong (%s)", time.Now().UTC().Format(time.RFC3339))
	case "":
		return s.helpText
	default:
		return "Unknown command.\n" + s.helpText
	}
}

func (s *service) startGreeting(user domain.User) string {
	fullName := s.userFullName(user)
	if fullName == "" {
		return "Welcome to Meethalf bot."
	}

	return fmt.Sprintf("Welcome to Meethalf bot, %s.", fullName)
}

func (s *service) userFullName(user domain.User) string {
	first := strings.TrimSpace(user.FirstName)
	last := strings.TrimSpace(user.LastName)

	if first == "" && last == "" {
		return ""
	}
	if last == "" {
		return first
	}
	if first == "" {
		return last
	}

	return first + " " + last
}

func (s *service) now(fallback time.Time) time.Time {
	if fallback.IsZero() {
		return time.Now().UTC()
	}

	return fallback.UTC()
}
