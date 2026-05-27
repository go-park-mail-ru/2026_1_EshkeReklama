package main

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func TestInitRedis(t *testing.T) {
	mr := miniredis.RunT(t)

	pool, err := initRedis(mr.Addr(), "")
	if err != nil {
		t.Fatalf("initRedis success path: %v", err)
	}
	t.Cleanup(func() { _ = pool.Close() })

	if _, err := initRedis("127.0.0.1:1", ""); err == nil {
		t.Fatal("expected initRedis to fail")
	}
}
