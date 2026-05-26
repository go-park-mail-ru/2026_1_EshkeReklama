package logger

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestCtxWithLogger_And_GetLoggerFromCtx(t *testing.T) {
	l := zap.NewNop().Sugar()
	ctx := CtxWithLogger(context.Background(), l)
	if GetLoggerFromCtx(ctx) == nil {
		t.Fatalf("expected logger")
	}
	if GetLoggerFromCtx(context.Background()) == nil {
		t.Fatalf("expected default logger")
	}
}
