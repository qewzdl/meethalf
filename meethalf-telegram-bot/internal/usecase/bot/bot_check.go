package bot

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) ensureBotCheck(draft *domain.ProfileDraft, fallback time.Time) {
	if draft == nil {
		return
	}

	if strings.TrimSpace(draft.BotCheckQuestion) != "" && draft.BotCheckAnswer > 0 {
		return
	}

	s.resetBotCheck(draft, fallback)
}

func (s *service) resetBotCheck(draft *domain.ProfileDraft, fallback time.Time) {
	if draft == nil {
		return
	}

	question, answer := s.newBotCheckChallenge()
	draft.BotCheckQuestion = question
	draft.BotCheckAnswer = answer
	draft.BotCheckAttempts = 0
	draft.UpdatedAt = s.now(fallback)
}

func (s *service) newBotCheckChallenge() (string, int) {
	min := botCheckMinOperand
	max := botCheckMaxOperand
	if max < min {
		max = min
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	left := rng.Intn(max-min+1) + min
	right := rng.Intn(max-min+1) + min
	return fmt.Sprintf("%d + %d = ?", left, right), left + right
}

func (s *service) botCheckMatches(draft domain.ProfileDraft, value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || draft.BotCheckAnswer <= 0 {
		return false
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return false
	}

	return parsed == draft.BotCheckAnswer
}
