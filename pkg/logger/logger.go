package logger

import (
	"context"
	"eshkere/pkg"

	"go.uber.org/zap"
)

const (
	loggerKey pkg.ContextKey = "logger"
)

func CtxWithLogger(parent context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(parent, loggerKey, logger)
}

func GetLoggerFromCtx(ctx context.Context) *zap.SugaredLogger {
	v := ctx.Value(loggerKey)
	switch l := v.(type) {
	case *zap.SugaredLogger:
		return l
	}

	return zap.NewNop().Sugar()
}
