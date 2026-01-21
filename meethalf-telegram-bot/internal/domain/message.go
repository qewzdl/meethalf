package domain

import "time"

const (
	CommandStart = "start"
	CommandHelp  = "help"
	CommandPing  = "ping"
)

type IncomingMessage struct {
	ChatID     int64
	User       User
	Text       string
	Command    string
	Arguments  string
	ReceivedAt time.Time
}

type OutgoingMessage struct {
	ChatID         int64
	Text           string
	ParseMode      string
	DisablePreview bool
}
