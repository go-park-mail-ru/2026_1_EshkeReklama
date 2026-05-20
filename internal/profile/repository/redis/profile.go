package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	profileservice "eshkere/internal/profile/service"

	redis "github.com/gomodule/redigo/redis"
)

const profileTTL = 90 * 24 * time.Hour

type Repository struct {
	pool *redis.Pool
}

func New(pool *redis.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetTopics(ctx context.Context, visitorID string, limit int) ([]profileservice.TopicScore, error) {
	if visitorID == "" || limit <= 0 {
		return nil, nil
	}

	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get redis connection: %w", err)
	}
	defer conn.Close()

	values, err := redis.Values(conn.Do("ZREVRANGE", topicsKey(visitorID), 0, limit-1, "WITHSCORES"))
	if err != nil {
		return nil, fmt.Errorf("get profile topics: %w", err)
	}

	topics := make([]profileservice.TopicScore, 0, len(values)/2)
	for len(values) > 0 {
		var rawTopic string
		var score float64
		values, err = redis.Scan(values, &rawTopic, &score)
		if err != nil {
			return nil, fmt.Errorf("scan profile topic: %w", err)
		}

		topicID, err := strconv.Atoi(rawTopic)
		if err != nil {
			continue
		}
		topics = append(topics, profileservice.TopicScore{TopicID: topicID, Score: score})
	}

	return topics, nil
}

func (r *Repository) TrackTopic(ctx context.Context, visitorID string, topicID int, scoreDelta float64, now time.Time) error {
	if visitorID == "" || topicID <= 0 || scoreDelta == 0 {
		return nil
	}

	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("get redis connection: %w", err)
	}
	defer conn.Close()

	profileKey := profileKey(visitorID)
	topicKey := topicsKey(visitorID)
	nowUnix := now.Unix()
	ttlSeconds := int(profileTTL.Seconds())

	if err := conn.Send("MULTI"); err != nil {
		return fmt.Errorf("start redis tx: %w", err)
	}
	if err := conn.Send("HSETNX", profileKey, "created_at", nowUnix); err != nil {
		return fmt.Errorf("queue profile created_at: %w", err)
	}
	if err := conn.Send("HSET", profileKey, "last_seen_at", nowUnix); err != nil {
		return fmt.Errorf("queue profile last_seen_at: %w", err)
	}
	if err := conn.Send("ZINCRBY", topicKey, scoreDelta, topicID); err != nil {
		return fmt.Errorf("queue profile topic increment: %w", err)
	}
	if err := conn.Send("EXPIRE", profileKey, ttlSeconds); err != nil {
		return fmt.Errorf("queue profile ttl: %w", err)
	}
	if err := conn.Send("EXPIRE", topicKey, ttlSeconds); err != nil {
		return fmt.Errorf("queue profile topics ttl: %w", err)
	}
	if _, err := conn.Do("EXEC"); err != nil {
		return fmt.Errorf("track profile topic: %w", err)
	}

	return nil
}

func profileKey(visitorID string) string {
	return "profile:" + visitorID
}

func topicsKey(visitorID string) string {
	return profileKey(visitorID) + ":topics"
}
