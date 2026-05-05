package redis

import (
	"context"
	"eshkere/internal/service"
	"fmt"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

type AdRequestStore struct {
	pool *redis.Pool
}

func NewAdRequestStore(pool *redis.Pool) *AdRequestStore {
	return &AdRequestStore{pool: pool}
}

func (s *AdRequestStore) Save(ctx context.Context, record service.AdRequestRecord, ttl time.Duration) error {
	if record.RequestID == "" || ttl <= 0 {
		return nil
	}

	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("get redis connection: %w", err)
	}
	defer conn.Close()

	key := adRequestKey(record.RequestID)
	if _, err := conn.Do("HMSET", key,
		"visitor_id", record.VisitorID,
		"ad_id", record.AdID,
		"topic_id", record.TopicID,
		"target_url", record.TargetURL,
	); err != nil {
		return fmt.Errorf("save ad request: %w", err)
	}
	if _, err := conn.Do("EXPIRE", key, int(ttl.Seconds())); err != nil {
		return fmt.Errorf("expire ad request: %w", err)
	}
	return nil
}

func (s *AdRequestStore) Get(ctx context.Context, requestID string) (*service.AdRequestRecord, error) {
	if requestID == "" {
		return nil, redis.ErrNil
	}

	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get redis connection: %w", err)
	}
	defer conn.Close()

	values, err := redis.StringMap(conn.Do("HGETALL", adRequestKey(requestID)))
	if err != nil {
		return nil, fmt.Errorf("get ad request: %w", err)
	}
	if len(values) == 0 {
		return nil, redis.ErrNil
	}

	adID, _ := strconv.Atoi(values["ad_id"])
	topicID, _ := strconv.Atoi(values["topic_id"])
	return &service.AdRequestRecord{
		RequestID: requestID,
		VisitorID: values["visitor_id"],
		AdID:      adID,
		TopicID:   topicID,
		TargetURL: values["target_url"],
	}, nil
}

func adRequestKey(requestID string) string {
	return "adreq:" + requestID
}
