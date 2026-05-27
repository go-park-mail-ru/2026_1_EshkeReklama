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

func TestPartnerIDFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), PartnerIDKey, 17)
	if got, err := PartnerIDFromContext(ctx); err != nil || got != 17 {
		t.Fatalf("expected partner id 17, got %d err=%v", got, err)
	}

	badCtx := context.WithValue(context.Background(), PartnerIDKey, "17")
	if _, err := PartnerIDFromContext(badCtx); err == nil {
		t.Fatalf("expected error for wrong partner id type")
	}
}
