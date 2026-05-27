package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redigo "github.com/gomodule/redigo/redis"
)

func newMiniRedisPool(t *testing.T) *redigo.Pool {
	t.Helper()
	mr := miniredis.RunT(t)
	return &redigo.Pool{
		Dial: func() (redigo.Conn, error) {
			return redigo.Dial("tcp", mr.Addr())
		},
	}
}

func TestEmailVerificationStore(t *testing.T) {
	pool := newMiniRedisPool(t)
	store := NewEmailVerificationStore(pool)

	if err := store.Save(context.Background(), EmailVerificationRecord{}, time.Minute); err != nil {
		t.Fatalf("save empty record: %v", err)
	}

	record := EmailVerificationRecord{AdvertiserID: 7, Email: "USER@example.com ", Code: "123456"}
	if err := store.Save(context.Background(), record, time.Minute); err != nil {
		t.Fatalf("save record: %v", err)
	}

	got, err := store.Get(context.Background(), " user@example.com ")
	if err != nil {
		t.Fatalf("get record: %v", err)
	}
	if got.AdvertiserID != 7 || got.Email != "user@example.com" || got.Code != "123456" {
		t.Fatalf("unexpected record: %+v", got)
	}

	if emailVerificationKey(" User@example.com ") != "emailverify:user@example.com" {
		t.Fatalf("unexpected email verification key")
	}
	if normalizeVerificationEmail(" User@example.com ") != "user@example.com" {
		t.Fatalf("unexpected normalized email")
	}

	if err := store.Delete(context.Background(), "user@example.com"); err != nil {
		t.Fatalf("delete record: %v", err)
	}
	if _, err := store.Get(context.Background(), "user@example.com"); err == nil {
		t.Fatal("expected record to be deleted")
	}
}

func TestPasswordResetStoreAndNotificationDedupe(t *testing.T) {
	pool := newMiniRedisPool(t)
	resetStore := NewPasswordResetStore(pool)
	dedupe := NewNotificationDedupeStore(pool)

	if err := resetStore.Save(context.Background(), PasswordResetRecord{}, time.Minute); err != nil {
		t.Fatalf("save empty reset record: %v", err)
	}

	record := PasswordResetRecord{AdvertiserID: 11, Email: "reset@example.com", Code: "654321"}
	if err := resetStore.Save(context.Background(), record, time.Minute); err != nil {
		t.Fatalf("save reset record: %v", err)
	}

	got, err := resetStore.Get(context.Background(), 11)
	if err != nil {
		t.Fatalf("get reset record: %v", err)
	}
	if got.AdvertiserID != 11 || got.Email != "reset@example.com" || got.Code != "654321" {
		t.Fatalf("unexpected reset record: %+v", got)
	}
	if passwordResetKey(11) != "passwordreset:11" {
		t.Fatalf("unexpected password reset key")
	}

	first, err := dedupe.MarkOnce(context.Background(), "dedupe:key", time.Minute)
	if err != nil || !first {
		t.Fatalf("mark once first=%v err=%v", first, err)
	}
	second, err := dedupe.MarkOnce(context.Background(), "dedupe:key", time.Minute)
	if err != nil || second {
		t.Fatalf("mark once second=%v err=%v", second, err)
	}

	if err := resetStore.Delete(context.Background(), 11); err != nil {
		t.Fatalf("delete reset record: %v", err)
	}
	if _, err := resetStore.Get(context.Background(), 11); err == nil {
		t.Fatal("expected reset record to be deleted")
	}
}
