package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	goredis "github.com/gomodule/redigo/redis"
)

type PasswordResetRecord struct {
	AdvertiserID int64
	Email        string
	Code         string
}

type PasswordResetStore struct {
	pool *goredis.Pool
}

func NewPasswordResetStore(pool *goredis.Pool) *PasswordResetStore {
	return &PasswordResetStore{pool: pool}
}

func (s *PasswordResetStore) Save(ctx context.Context, record PasswordResetRecord, ttl time.Duration) error {
	if s == nil || record.AdvertiserID <= 0 || strings.TrimSpace(record.Email) == "" || strings.TrimSpace(record.Code) == "" || ttl <= 0 {
		return nil
	}

	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("get redis connection: %w", err)
	}
	defer conn.Close()

	key := passwordResetKey(record.AdvertiserID)
	if _, err := conn.Do("HMSET", key,
		"advertiser_id", record.AdvertiserID,
		"email", normalizeVerificationEmail(record.Email),
		"code", record.Code,
	); err != nil {
		return fmt.Errorf("save password reset: %w", err)
	}
	if _, err := conn.Do("EXPIRE", key, int(ttl.Seconds())); err != nil {
		return fmt.Errorf("expire password reset: %w", err)
	}
	return nil
}

func (s *PasswordResetStore) Get(ctx context.Context, advertiserID int64) (*PasswordResetRecord, error) {
	if s == nil || advertiserID <= 0 {
		return nil, goredis.ErrNil
	}

	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get redis connection: %w", err)
	}
	defer conn.Close()

	values, err := goredis.StringMap(conn.Do("HGETALL", passwordResetKey(advertiserID)))
	if err != nil {
		return nil, fmt.Errorf("get password reset: %w", err)
	}
	if len(values) == 0 {
		return nil, goredis.ErrNil
	}

	id, _ := strconv.ParseInt(values["advertiser_id"], 10, 64)
	return &PasswordResetRecord{
		AdvertiserID: id,
		Email:        values["email"],
		Code:         values["code"],
	}, nil
}

func (s *PasswordResetStore) Delete(ctx context.Context, advertiserID int64) error {
	if s == nil || advertiserID <= 0 {
		return nil
	}

	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("get redis connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.Do("DEL", passwordResetKey(advertiserID)); err != nil {
		return fmt.Errorf("delete password reset: %w", err)
	}
	return nil
}

func passwordResetKey(advertiserID int64) string {
	return "passwordreset:" + strconv.FormatInt(advertiserID, 10)
}
