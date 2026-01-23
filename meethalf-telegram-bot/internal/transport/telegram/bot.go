package telegram

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotConfig struct {
	Token       string
	Debug       bool
	APIEndpoint string
	ProxyURL    string
}

func NewBot(cfg BotConfig) (*tgbotapi.BotAPI, error) {
	token := strings.TrimSpace(cfg.Token)
	if token == "" {
		return nil, errors.New("bot token is required")
	}

	apiEndpoint := normalizeAPIEndpoint(cfg.APIEndpoint)
	client, err := newHTTPClient(cfg.ProxyURL)
	if err != nil {
		return nil, err
	}

	bot, err := tgbotapi.NewBotAPIWithClient(token, apiEndpoint, client)
	if err != nil {
		return nil, fmt.Errorf("telegram api error: %s", redactToken(err.Error(), token))
	}

	bot.Debug = cfg.Debug
	return bot, nil
}

func normalizeAPIEndpoint(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return tgbotapi.APIEndpoint
	}

	trimmed = strings.TrimRight(trimmed, "/")
	if strings.Contains(trimmed, "%s") {
		return trimmed
	}
	if strings.HasSuffix(trimmed, "/bot") {
		return trimmed + "%s/%s"
	}

	return trimmed + "/bot%s/%s"
}

func newHTTPClient(proxyURL string) (*http.Client, error) {
	baseTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, errors.New("default http transport is unavailable")
	}
	transport := baseTransport.Clone()

	trimmed := strings.TrimSpace(proxyURL)
	if trimmed != "" {
		parsed, err := url.Parse(trimmed)
		if err != nil {
			return nil, fmt.Errorf("invalid bot proxy url: %w", err)
		}
		transport.Proxy = http.ProxyURL(parsed)
	} else {
		transport.Proxy = http.ProxyFromEnvironment
	}

	return &http.Client{Transport: transport}, nil
}

func redactToken(message, token string) string {
	if token == "" || message == "" {
		return message
	}

	return strings.ReplaceAll(message, token, "<redacted>")
}
