package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/gomodule/redigo/redis"
)

type AIImageGenerationStore struct {
	pool *goredis.Pool
}

func NewAIImageGenerationStore(pool *goredis.Pool) *AIImageGenerationStore {
	return &AIImageGenerationStore{pool: pool}
}

func (s *AIImageGenerationStore) Reserve(ctx context.Context, key string, ttl time.Duration) (int, bool, error) {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("get redis conn: %w", err)
	}
	defer conn.Close()

	attempt, err := goredis.Int(conn.Do("INCR", aiImageGenerationKey(key)))
	if err != nil {
		return 0, false, fmt.Errorf("incr ai image generation counter: %w", err)
	}
	if attempt == 1 && ttl > 0 {
		if _, err := conn.Do("EXPIRE", aiImageGenerationKey(key), int(ttl.Seconds())); err != nil {
			return 0, false, fmt.Errorf("expire ai image generation counter: %w", err)
		}
	}

	return attempt, attempt <= 4, nil
}

func aiImageGenerationKey(key string) string {
	return "ai:imagegen:" + key
}
