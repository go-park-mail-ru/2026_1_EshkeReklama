package postgres

import (
	"context"
	"eshkere/pkg/logger"
	"time"
)

func logDBQuery(ctx context.Context, operation string, startedAt time.Time, err error) {
	fields := []any{
		"operation", operation,
		"duration_ms", time.Since(startedAt).Milliseconds(),
	}

	requestLogger := logger.GetLoggerFromCtx(ctx)
	if err != nil {
		requestLogger.Errorw("db query failed", append(fields, "error", err.Error())...)
		return
	}

	requestLogger.Infow("db query finished", fields...)
}
