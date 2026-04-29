package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	redis "github.com/gomodule/redigo/redis"
)

var ErrStoreSessionNotFound = errors.New("session not found in store")

type RedisStore struct {
	pool *redis.Pool
}

func NewRedisStore(pool *redis.Pool) *RedisStore {
	return &RedisStore{pool: pool}
}

func (s *RedisStore) Save(_ context.Context, sessionID string, sess Session, ttl time.Duration) error {
	conn := s.pool.Get()
	defer conn.Close()

	payload, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	_, err = conn.Do("SET", sessionKey(sessionID), payload, "EX", int(ttl.Seconds()))
	if err != nil {
		return fmt.Errorf("save session: %w", err)
	}

	return nil
}

func (s *RedisStore) Get(_ context.Context, sessionID string) (Session, error) {
	conn := s.pool.Get()
	defer conn.Close()

	payload, err := redis.Bytes(conn.Do("GET", sessionKey(sessionID)))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return Session{}, ErrStoreSessionNotFound
		}
		return Session{}, fmt.Errorf("get session: %w", err)
	}

	var sess Session
	if err := json.Unmarshal(payload, &sess); err != nil {
		return Session{}, fmt.Errorf("unmarshal session: %w", err)
	}

	return sess, nil
}

func (s *RedisStore) Delete(_ context.Context, sessionID string) error {
	conn := s.pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", sessionKey(sessionID))
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}

func sessionKey(sessionID string) string {
	return "session:" + sessionID
}
