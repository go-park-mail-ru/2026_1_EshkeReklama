package logger

import (
	"context"
	"eshkere/pkg/ctxutils"

	"go.uber.org/zap"
)

const (
	LoggerKey ctxutils.ContextKey = "logger"
)

func CtxWithLogger(parent context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(parent, LoggerKey, logger)
}

func GetLoggerFromCtx(ctx context.Context) *zap.SugaredLogger {
	v := ctx.Value(LoggerKey)
	switch l := v.(type) {
	case *zap.SugaredLogger:
		return l
	}

	return zap.Must(zap.NewDevelopment()).Sugar()
}
