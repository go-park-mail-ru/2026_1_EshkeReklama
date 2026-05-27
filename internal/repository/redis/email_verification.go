package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	goredis "github.com/gomodule/redigo/redis"
)

type EmailVerificationRecord struct {
	AdvertiserID int64
	Email        string
	Code         string
}

type EmailVerificationStore struct {
	pool *goredis.Pool
}

func NewEmailVerificationStore(pool *goredis.Pool) *EmailVerificationStore {
	return &EmailVerificationStore{pool: pool}
}

func (s *EmailVerificationStore) Save(ctx context.Context, record EmailVerificationRecord, ttl time.Duration) error {
	if s == nil || record.AdvertiserID <= 0 || strings.TrimSpace(record.Email) == "" || strings.TrimSpace(record.Code) == "" || ttl <= 0 {
		return nil
	}

	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("get redis connection: %w", err)
	}
	defer conn.Close()

	key := emailVerificationKey(record.Email)
	if _, err := conn.Do("HMSET", key,
		"advertiser_id", record.AdvertiserID,
		"email", normalizeVerificationEmail(record.Email),
		"code", record.Code,
	); err != nil {
		return fmt.Errorf("save email verification: %w", err)
	}
	if _, err := conn.Do("EXPIRE", key, int(ttl.Seconds())); err != nil {
		return fmt.Errorf("expire email verification: %w", err)
	}
	return nil
}

func (s *EmailVerificationStore) Get(ctx context.Context, email string) (*EmailVerificationRecord, error) {
	if s == nil {
		return nil, goredis.ErrNil
	}

	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get redis connection: %w", err)
	}
	defer conn.Close()

	values, err := goredis.StringMap(conn.Do("HGETALL", emailVerificationKey(email)))
	if err != nil {
		return nil, fmt.Errorf("get email verification: %w", err)
	}
	if len(values) == 0 {
		return nil, goredis.ErrNil
	}

	advertiserID, _ := strconv.ParseInt(values["advertiser_id"], 10, 64)
	return &EmailVerificationRecord{
		AdvertiserID: advertiserID,
		Email:        values["email"],
		Code:         values["code"],
	}, nil
}

func (s *EmailVerificationStore) Delete(ctx context.Context, email string) error {
	if s == nil {
		return nil
	}

	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("get redis connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.Do("DEL", emailVerificationKey(email)); err != nil {
		return fmt.Errorf("delete email verification: %w", err)
	}
	return nil
}

func emailVerificationKey(email string) string {
	return "emailverify:" + normalizeVerificationEmail(email)
}

func normalizeVerificationEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
