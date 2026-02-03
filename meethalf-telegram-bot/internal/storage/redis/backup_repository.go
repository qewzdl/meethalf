package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"meethalf-telegram-bot/internal/domain"

	redisgo "github.com/redis/go-redis/v9"
)

const (
	defaultBackupPrefix  = "meethalf:"
	defaultBackupScanCnt = int64(500)
)

type BackupRepository struct {
	client    *redisgo.Client
	prefix    string
	scanCount int64
}

func NewBackupRepository(client *redisgo.Client, prefix string) *BackupRepository {
	trimmed := strings.TrimSpace(prefix)
	if trimmed == "" {
		trimmed = defaultBackupPrefix
	}

	return &BackupRepository{
		client:    client,
		prefix:    trimmed,
		scanCount: defaultBackupScanCnt,
	}
}

func (r *BackupRepository) Dump(ctx context.Context) ([]domain.BackupEntry, error) {
	if r == nil || r.client == nil {
		return nil, errors.New("redis backup repository is not configured")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	pattern := r.prefix + "*"
	iter := r.client.Scan(ctx, 0, pattern, r.scanCount).Iterator()
	entries := make([]domain.BackupEntry, 0, 256)

	for iter.Next(ctx) {
		key := iter.Val()
		entry, err := r.dumpKey(ctx, key)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func (r *BackupRepository) Restore(ctx context.Context, entries []domain.BackupEntry) error {
	if r == nil || r.client == nil {
		return errors.New("redis backup repository is not configured")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Key, r.prefix) {
			return fmt.Errorf("unexpected key outside backup prefix: %s", entry.Key)
		}
		if !isSupportedBackupType(entry.Type) {
			return fmt.Errorf("unsupported redis type: %s", entry.Type)
		}
	}

	if err := r.clearPrefix(ctx); err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	for _, entry := range entries {
		switch entry.Type {
		case "string":
			pipe.Set(ctx, entry.Key, entry.StringValue, 0)
		case "hash":
			if len(entry.HashValue) > 0 {
				pipe.HSet(ctx, entry.Key, entry.HashValue)
			}
		default:
			return fmt.Errorf("unsupported redis type: %s", entry.Type)
		}

		if entry.TTLMillis > 0 {
			pipe.PExpire(ctx, entry.Key, time.Duration(entry.TTLMillis)*time.Millisecond)
		}
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (r *BackupRepository) dumpKey(ctx context.Context, key string) (domain.BackupEntry, error) {
	keyType, err := r.client.Type(ctx, key).Result()
	if err != nil {
		return domain.BackupEntry{}, err
	}

	ttlMillis, err := r.keyTTL(ctx, key)
	if err != nil {
		return domain.BackupEntry{}, err
	}

	entry := domain.BackupEntry{
		Key:       key,
		Type:      keyType,
		TTLMillis: ttlMillis,
	}

	switch keyType {
	case "string":
		value, err := r.client.Get(ctx, key).Result()
		if err != nil {
			return domain.BackupEntry{}, err
		}
		entry.StringValue = value
	case "hash":
		values, err := r.client.HGetAll(ctx, key).Result()
		if err != nil {
			return domain.BackupEntry{}, err
		}
		entry.HashValue = values
	default:
		return domain.BackupEntry{}, fmt.Errorf("unsupported redis type: %s", keyType)
	}

	return entry, nil
}

func (r *BackupRepository) keyTTL(ctx context.Context, key string) (int64, error) {
	ttl, err := r.client.PTTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if ttl <= 0 {
		return 0, nil
	}

	return ttl.Milliseconds(), nil
}

func (r *BackupRepository) clearPrefix(ctx context.Context) error {
	pattern := r.prefix + "*"
	iter := r.client.Scan(ctx, 0, pattern, r.scanCount).Iterator()
	batch := make([]string, 0, int(r.scanCount))

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := r.client.Del(ctx, batch...).Err(); err != nil {
			return err
		}
		batch = batch[:0]
		return nil
	}

	for iter.Next(ctx) {
		batch = append(batch, iter.Val())
		if int64(len(batch)) >= r.scanCount {
			if err := flush(); err != nil {
				return err
			}
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}

	return flush()
}

func isSupportedBackupType(value string) bool {
	return value == "string" || value == "hash"
}
