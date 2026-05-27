package app

import (
	"net/http"
	"testing"
	"time"

	"eshkere/internal/config"

	"github.com/alicebob/miniredis/v2"
)

func TestInitRedisAndDBHelpers(t *testing.T) {
	mr := miniredis.RunT(t)

	pool, err := initRedis(config.RedisConfig{
		Host:             mr.Host(),
		Port:             mustAtoi(t, mr.Port()),
		MaxIdle:          1,
		MaxActive:        2,
		IdleTimeout:      time.Minute,
		Wait:             true,
		ConnectTimeout:   time.Second,
		ReadTimeout:      time.Second,
		WriteTimeout:     time.Second,
		PingAfterIdleFor: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("initRedis success path: %v", err)
	}
	t.Cleanup(func() { _ = pool.Close() })

	conn := pool.Get()
	if _, err := conn.Do("PING"); err != nil {
		t.Fatalf("redis ping through pool: %v", err)
	}
	_ = conn.Close()

	if _, err := initRedis(config.RedisConfig{
		Host:           "127.0.0.1",
		Port:           1,
		MaxIdle:        1,
		MaxActive:      1,
		IdleTimeout:    time.Second,
		Wait:           true,
		ConnectTimeout: 100 * time.Millisecond,
		ReadTimeout:    100 * time.Millisecond,
		WriteTimeout:   100 * time.Millisecond,
	}); err == nil {
		t.Fatal("expected initRedis to fail on unreachable port")
	}

	if _, err := initDB(config.PostgresConfig{
		Host:     "127.0.0.1",
		Port:     1,
		Database: "db",
		Username: "user",
		Password: "pass",
	}); err == nil {
		t.Fatal("expected initDB to fail on unreachable postgres")
	}
}

func TestShutdownClosesServer(t *testing.T) {
	a := &App{cfg: &config.Config{GracefulTimeout: time.Second}}
	srv := &http.Server{Addr: "127.0.0.1:0", Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()
	time.Sleep(100 * time.Millisecond)

	if err := a.shutdown([]*http.Server{srv}); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			t.Fatalf("unexpected server error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("server did not shut down")
	}
}

func mustAtoi(t *testing.T, s string) int {
	t.Helper()
	var n int
	for _, ch := range s {
		n = n*10 + int(ch-'0')
	}
	return n
}
