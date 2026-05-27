package main

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestInitRedisAndPostgresHelpers(t *testing.T) {
	mr := miniredis.RunT(t)

	pool, err := initRedis(mr.Addr(), "")
	if err != nil {
		t.Fatalf("initRedis success path: %v", err)
	}
	t.Cleanup(func() { _ = pool.Close() })

	if _, err := initRedis("127.0.0.1:1", ""); err == nil {
		t.Fatal("expected initRedis to fail")
	}

	if _, err := initPostgres("postgres://user:pass@127.0.0.1:1/db?sslmode=disable&connect_timeout=1"); err == nil {
		t.Fatal("expected initPostgres to fail")
	}

	if got := envDuration("AUTH_TEST_DUR_MISSING", 3*time.Second); got != 3*time.Second {
		t.Fatalf("unexpected missing env duration: %v", got)
	}
}
