package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redis "github.com/gomodule/redigo/redis"
)

func TestRedisStore(t *testing.T) {
	mr := miniredis.RunT(t)
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", mr.Addr())
		},
	}

	store := NewRedisStore(pool)
	sess := Session{AdvertiserID: 9, ExpiresAt: time.Now().UTC()}

	if err := store.Save(context.Background(), "sid-1", sess, time.Hour); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := store.Get(context.Background(), "sid-1")
	if err != nil || got.AdvertiserID != 9 {
		t.Fatalf("get: %#v %v", got, err)
	}
	if err := store.Delete(context.Background(), "sid-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := store.Get(context.Background(), "sid-1"); !errors.Is(err, ErrStoreSessionNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if sessionKey("abc") != "session:abc" {
		t.Fatalf("unexpected session key")
	}
}
