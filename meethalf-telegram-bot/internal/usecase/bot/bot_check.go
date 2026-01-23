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

func (s *service) botCheckOptions(answer int) []int {
	if answer <= 0 {
		return nil
	}

	minValue := answer - botCheckOptionsSpread
	if minValue < botCheckOptionsMinValue {
		minValue = botCheckOptionsMinValue
	}
	maxValue := answer + botCheckOptionsSpread
	if maxValue < minValue {
		maxValue = minValue
	}

	pool := make([]int, 0, maxValue-minValue)
	for value := minValue; value <= maxValue; value++ {
		if value == answer {
			continue
		}
		pool = append(pool, value)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(pool), func(i, j int) {
		pool[i], pool[j] = pool[j], pool[i]
	})

	options := make([]int, 0, botCheckOptionsCount)
	options = append(options, answer)
	for i := 0; i < len(pool) && len(options) < botCheckOptionsCount; i++ {
		options = append(options, pool[i])
	}

	for candidate := maxValue + 1; len(options) < botCheckOptionsCount; candidate++ {
		options = append(options, candidate)
	}

	rng.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	return options
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
