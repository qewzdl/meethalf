package bot

import (
	"strings"

	"meethalf-telegram-bot/internal/domain"
)

const emojiKeyboardColumns = 4

type profileEmojiOption struct {
	Code  domain.ProfileEmojiCode
	Emoji string
}

var profileEmojiOptions = []profileEmojiOption{
	{Code: domain.ProfileEmojiLeader, Emoji: "👑"},
	{Code: domain.ProfileEmojiStrategist, Emoji: "🧠"},
	{Code: domain.ProfileEmojiAnalyst, Emoji: "🧩"},
	{Code: domain.ProfileEmojiCreator, Emoji: "🎨"},
	{Code: domain.ProfileEmojiCommunicator, Emoji: "🤝"},
	{Code: domain.ProfileEmojiEmpath, Emoji: "❤️"},
	{Code: domain.ProfileEmojiMediator, Emoji: "🕊"},
	{Code: domain.ProfileEmojiPerfectionist, Emoji: "🧼"},
	{Code: domain.ProfileEmojiResearcher, Emoji: "🧭"},
	{Code: domain.ProfileEmojiInnovator, Emoji: "💡"},
	{Code: domain.ProfileEmojiExecutor, Emoji: "🛠"},
	{Code: domain.ProfileEmojiAdventurer, Emoji: "🔥"},
	{Code: domain.ProfileEmojiContemplator, Emoji: "☕️"},
	{Code: domain.ProfileEmojiRealist, Emoji: "🧱"},
	{Code: domain.ProfileEmojiMotivator, Emoji: "🎯"},
	{Code: domain.ProfileEmojiSkeptic, Emoji: "🛡"},
}

var profileEmojiByCode = map[domain.ProfileEmojiCode]string{
	domain.ProfileEmojiLeader:        "👑",
	domain.ProfileEmojiStrategist:    "🧠",
	domain.ProfileEmojiAnalyst:       "🧩",
	domain.ProfileEmojiCreator:       "🎨",
	domain.ProfileEmojiCommunicator:  "🤝",
	domain.ProfileEmojiEmpath:        "❤️",
	domain.ProfileEmojiMediator:      "🕊",
	domain.ProfileEmojiPerfectionist: "🧼",
	domain.ProfileEmojiResearcher:    "🧭",
	domain.ProfileEmojiInnovator:     "💡",
	domain.ProfileEmojiExecutor:      "🛠",
	domain.ProfileEmojiAdventurer:    "🔥",
	domain.ProfileEmojiContemplator:  "☕️",
	domain.ProfileEmojiRealist:       "🧱",
	domain.ProfileEmojiMotivator:     "🎯",
	domain.ProfileEmojiSkeptic:       "🛡",
}

var profileEmojiCodeByEmoji = map[string]domain.ProfileEmojiCode{
	"👑":  domain.ProfileEmojiLeader,
	"🧠":  domain.ProfileEmojiStrategist,
	"🧩":  domain.ProfileEmojiAnalyst,
	"🎨":  domain.ProfileEmojiCreator,
	"🤝":  domain.ProfileEmojiCommunicator,
	"❤️": domain.ProfileEmojiEmpath,
	"🕊":  domain.ProfileEmojiMediator,
	"🧼":  domain.ProfileEmojiPerfectionist,
	"🧭":  domain.ProfileEmojiResearcher,
	"💡":  domain.ProfileEmojiInnovator,
	"🛠":  domain.ProfileEmojiExecutor,
	"🔥":  domain.ProfileEmojiAdventurer,
	"☕️": domain.ProfileEmojiContemplator,
	"🧱":  domain.ProfileEmojiRealist,
	"🎯":  domain.ProfileEmojiMotivator,
	"🛡":  domain.ProfileEmojiSkeptic,
}

func (s *service) emojiInlineKeyboard() *domain.InlineKeyboard {
	if len(profileEmojiOptions) == 0 {
		return nil
	}

	rows := make([][]domain.InlineButton, 0, (len(profileEmojiOptions)+emojiKeyboardColumns-1)/emojiKeyboardColumns)
	for i, option := range profileEmojiOptions {
		if i%emojiKeyboardColumns == 0 {
			rows = append(rows, []domain.InlineButton{})
		}
		rowIndex := len(rows) - 1
		rows[rowIndex] = append(rows[rowIndex], domain.InlineButton{
			Text:         option.Emoji,
			CallbackData: string(option.Code),
		})
	}

	return &domain.InlineKeyboard{Buttons: rows}
}

func (s *service) normalizeEmojiCode(value string) (domain.ProfileEmojiCode, bool) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", false
	}

	upper := strings.ToUpper(normalized)
	code := domain.ProfileEmojiCode(upper)
	if _, ok := profileEmojiByCode[code]; ok {
		return code, true
	}

	if code, ok := profileEmojiCodeByEmoji[normalized]; ok {
		return code, true
	}

	return "", false
}

func (s *service) emojiLabel(code domain.ProfileEmojiCode) string {
	if emoji, ok := profileEmojiByCode[code]; ok {
		return emoji
	}

	return "Not set"
}
