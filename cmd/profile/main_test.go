package main

import "testing"

func TestEnvOr(t *testing.T) {
	t.Setenv("PROFILE_TEST", "value")
	if got := envOr("PROFILE_TEST", "fallback"); got != "value" {
		t.Fatalf("unexpected envOr: %s", got)
	}
	if got := envOr("PROFILE_TEST_MISSING", "fallback"); got != "fallback" {
		t.Fatalf("unexpected fallback: %s", got)
	}
}
