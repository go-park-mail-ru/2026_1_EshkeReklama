package ctxutils

import (
	"context"
	"testing"
)

func TestRequestID_Roundtrip(t *testing.T) {
	id := NewRequestID()
	if id == "" {
		t.Fatalf("expected non-empty id")
	}

	ctx := CtxWithRequestID(context.Background(), id)
	if got := RequestIDFromCtx(ctx); got != id {
		t.Fatalf("expected %q got %q", id, got)
	}
}

func TestAdvertiserIDFromContext_NotFound(t *testing.T) {
	if _, err := AdvertiserIDFromContext(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}
