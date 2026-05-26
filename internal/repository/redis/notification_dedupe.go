package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/gomodule/redigo/redis"
)

type NotificationDedupeStore struct {
	pool *goredis.Pool
}

func NewNotificationDedupeStore(pool *goredis.Pool) *NotificationDedupeStore {
	return &NotificationDedupeStore{pool: pool}
}

func (s *NotificationDedupeStore) MarkOnce(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return false, fmt.Errorf("get redis conn: %w", err)
	}
	defer conn.Close()

	reply, err := goredis.String(conn.Do("SET", key, "1", "EX", int(ttl.Seconds()), "NX"))
	if err != nil && err != goredis.ErrNil {
		return false, fmt.Errorf("set dedupe key: %w", err)
	}

	return reply == "OK", nil
}
