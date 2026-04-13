package pkg

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
)

type ContextKey string

const (
	AdvertiserIDKey ContextKey = "advertiser_id"
	requestIDKey    ContextKey = "request_id"
)

var ErrAdvertiserIDNotFound = errors.New("advertiser id not found in logger")

func AdvertiserIDFromContext(ctx context.Context) (int, error) {
	value := ctx.Value(AdvertiserIDKey)
	if value == nil {
		return 0, ErrAdvertiserIDNotFound
	}

	id, ok := value.(int)
	if !ok {
		return 0, ErrAdvertiserIDNotFound
	}

	return id, nil
}

func CtxWithRequestID(parent context.Context, requestID string) context.Context {
	return context.WithValue(parent, requestIDKey, requestID)
}

func RequestIDFromCtx(ctx context.Context) string {
	v := ctx.Value(requestIDKey)
	id, ok := v.(string)
	if !ok {
		return ""
	}

	return id
}

func NewRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}

	return hex.EncodeToString(b)
}
