package main

import (
	"os"
	"testing"
	"time"
)

func TestEnvHelpersAndVKIDInit(t *testing.T) {
	t.Setenv("AUTH_TEST_STR", "value")
	if got := envOr("AUTH_TEST_STR", "fallback"); got != "value" {
		t.Fatalf("unexpected envOr: %s", got)
	}
	if got := envOr("AUTH_TEST_MISSING", "fallback"); got != "fallback" {
		t.Fatalf("unexpected fallback: %s", got)
	}

	t.Setenv("AUTH_TEST_DUR", "2m")
	if got := envDuration("AUTH_TEST_DUR", time.Second); got != 2*time.Minute {
		t.Fatalf("unexpected duration: %v", got)
	}
	t.Setenv("AUTH_TEST_DUR", "bad")
	if got := envDuration("AUTH_TEST_DUR", time.Second); got != time.Second {
		t.Fatalf("unexpected default duration: %v", got)
	}

	t.Setenv("AUTH_TEST_INT", "15")
	if got := envInt64("AUTH_TEST_INT", 1); got != 15 {
		t.Fatalf("unexpected int64: %d", got)
	}
	t.Setenv("AUTH_TEST_INT", "bad")
	if got := envInt64("AUTH_TEST_INT", 1); got != 1 {
		t.Fatalf("unexpected default int64: %d", got)
	}

	if client := initVKIDClient(0, "", 0); client != nil {
		t.Fatal("expected nil client for missing config")
	}
	if client := initVKIDClient(1, "id.vk.ru", time.Second); client == nil {
		t.Fatal("expected vkid client")
	}
	if client := initVKIDClient(1, "", 0); client == nil {
		t.Fatal("expected default-domain vkid client")
	}

	_ = os.Getenv("AUTH_TEST_STR")
}
