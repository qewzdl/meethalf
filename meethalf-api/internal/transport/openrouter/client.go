package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"meethalf-api/internal/domain"
	"meethalf-api/internal/usecase/matching"
)

const defaultBaseURL = "https://openrouter.ai/api/v1"

type Client struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func New(apiKey, baseURL, model string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 12 * time.Second
	}

	trimmedBase := strings.TrimSpace(baseURL)
	if trimmedBase == "" {
		trimmedBase = defaultBaseURL
	}
	trimmedBase = strings.TrimRight(trimmedBase, "/")

	return &Client{
		apiKey:  strings.TrimSpace(apiKey),
		baseURL: trimmedBase,
		model:   strings.TrimSpace(model),
		client:  &http.Client{Timeout: timeout},
	}
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type aiResponse struct {
	Gender    string   `json:"gender"`
	MinAge    *int     `json:"min_age"`
	MaxAge    *int     `json:"max_age"`
	Country   string   `json:"country"`
	City      string   `json:"city"`
	EmojiCode string   `json:"emoji_code"`
	Keywords  []string `json:"keywords"`
}

func (c *Client) Analyze(ctx context.Context, input string) (matching.AIQuery, error) {
	if c == nil || c.client == nil {
		return matching.AIQuery{}, errors.New("openrouter client is not configured")
	}
	if c.apiKey == "" {
		return matching.AIQuery{}, errors.New("openrouter api key is missing")
	}
	if c.model == "" {
		return matching.AIQuery{}, errors.New("openrouter model is missing")
	}
	if err := ctx.Err(); err != nil {
		return matching.AIQuery{}, err
	}

	trimmedInput := strings.TrimSpace(input)
	if trimmedInput == "" {
		return matching.AIQuery{}, errors.New("input is empty")
	}

	payload := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: analysisPrompt},
			{Role: "user", Content: trimmedInput},
		},
		Temperature: 0.2,
		MaxTokens:   400,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return matching.AIQuery{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return matching.AIQuery{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return matching.AIQuery{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message := readErrorMessage(resp.Body)
		if message == "" {
			message = resp.Status
		}
		return matching.AIQuery{}, fmt.Errorf("openrouter request failed: %s", message)
	}

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return matching.AIQuery{}, err
	}
	if len(result.Choices) == 0 {
		return matching.AIQuery{}, errors.New("openrouter response has no choices")
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)
	parsed, err := parseAIResponse(content)
	if err != nil {
		return matching.AIQuery{}, err
	}

	return matching.AIQuery{
		Gender:    domain.Gender(strings.TrimSpace(parsed.Gender)),
		MinAge:    parsed.MinAge,
		MaxAge:    parsed.MaxAge,
		Country:   domain.Country(strings.TrimSpace(parsed.Country)),
		City:      strings.TrimSpace(parsed.City),
		EmojiCode: domain.ProfileEmojiCode(strings.TrimSpace(parsed.EmojiCode)),
		Keywords:  parsed.Keywords,
	}, nil
}

func parseAIResponse(content string) (aiResponse, error) {
	if strings.TrimSpace(content) == "" {
		return aiResponse{}, errors.New("openrouter response is empty")
	}

	var payload aiResponse
	if err := json.Unmarshal([]byte(content), &payload); err == nil {
		return payload, nil
	}

	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		snippet := content[start : end+1]
		if err := json.Unmarshal([]byte(snippet), &payload); err == nil {
			return payload, nil
		}
	}

	return aiResponse{}, errors.New("openrouter response is not valid json")
}

func readErrorMessage(r io.Reader) string {
	body, err := io.ReadAll(r)
	if err != nil || len(body) == 0 {
		return ""
	}
	return strings.TrimSpace(string(body))
}

const analysisPrompt = `You extract dating preferences from a user message and must return ONLY valid JSON.

Return this exact JSON shape:
{
  "gender": "male|female|other|unspecified",
  "min_age": integer or null,
  "max_age": integer or null,
  "country": "russia|kazakhstan|belarus|ukraine" or "",
  "city": string or "",
  "emoji_code": "LDR|STR|ANA|CRT|COM|EMP|MED|PRF|RSR|INN|EXE|ADV|CNT|RLS|MOT|SKP" or "",
  "keywords": ["keyword1", "keyword2"]
}

Rules:
- Do not invent details. If not specified, use empty string or null.
- Gender: use "unspecified" when the user says "any" or does not specify.
- Ages: integers only.
- Keywords: short, up to 8 items.
- Use English city names if possible. Supported cities:
  Moscow, Saint Petersburg, Novosibirsk, Krasnodar, Omsk, Rostov-on-Don, Perm, Krasnoyarsk, Yekaterinburg, Kazan,
  Nizhny Novgorod, Ufa, Chelyabinsk, Samara, Voronezh, Volgograd,
  Astana, Almaty, Semey, Pavlodar, Shymkent, Aktobe, Karaganda, Taraz, Ust-Kamenogorsk, Atyrau,
  Minsk, Gomel, Mogilev, Vitebsk, Grodno, Brest, Bobruisk, Baranovichi, Borisov,
  Kyiv, Kharkiv, Odesa, Dnipro, Donetsk, Lviv, Zaporizhzhia, Kryvyi Rih, Mykolaiv, Luhansk, Mariupol, Kherson,
  Vinnytsia, Poltava, Chernihiv, Cherkasy, Zhytomyr, Sumy, Khmelnytskyi, Rivne, Ternopil, Ivano-Frankivsk,
  Lutsk, Chernivtsi.
`
