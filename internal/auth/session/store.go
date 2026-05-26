package session

import (
	"context"
	"time"
)

type Store interface {
	Save(ctx context.Context, sessionID string, sess Session, ttl time.Duration) error
	Get(ctx context.Context, sessionID string) (Session, error)
	Delete(ctx context.Context, sessionID string) error
}
