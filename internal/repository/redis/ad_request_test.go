package redis

import (
	"context"
	"testing"
	"time"

	"eshkere/internal/service"

	"github.com/alicebob/miniredis/v2"
	redigo "github.com/gomodule/redigo/redis"
)

func TestAdRequestStore(t *testing.T) {
	mr := miniredis.RunT(t)
	pool := &redigo.Pool{
		Dial: func() (redigo.Conn, error) {
			return redigo.Dial("tcp", mr.Addr())
		},
	}
	store := NewAdRequestStore(pool)

	if err := store.Save(context.Background(), service.AdRequestRecord{}, time.Hour); err != nil {
		t.Fatalf("expected nil for empty request id, got %v", err)
	}

	record := service.AdRequestRecord{
		RequestID: "req-1",
		VisitorID: "visitor-1",
		AdID:      3,
		TopicID:   7,
		TargetURL: "https://example.com",
	}
	if err := store.Save(context.Background(), record, time.Hour); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := store.Get(context.Background(), "req-1")
	if err != nil || got.RequestID != "req-1" || got.AdID != 3 || got.TopicID != 7 {
		t.Fatalf("get: %#v %v", got, err)
	}
	if _, err := store.Get(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty request id")
	}
	if adRequestKey("req-1") != "adreq:req-1" {
		t.Fatalf("unexpected redis key")
	}
}
