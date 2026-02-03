package domain

import "time"

type BackupEntry struct {
	Key         string            `json:"key"`
	Type        string            `json:"type"`
	TTLMillis   int64             `json:"ttl_ms,omitempty"`
	StringValue string            `json:"string_value,omitempty"`
	HashValue   map[string]string `json:"hash_value,omitempty"`
}

type BackupSnapshot struct {
	Version     int           `json:"version"`
	GeneratedAt time.Time     `json:"generated_at"`
	Entries     []BackupEntry `json:"entries"`
}
