package middleware

import (
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/logger"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const requestIDHeader = "X-Request-ID"

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func RequestContext(baseLogger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDHeader)
			if requestID == "" {
				requestID = ctxutils.NewRequestID()
			}

			w.Header().Set(requestIDHeader, requestID)

			reqLogger := baseLogger.With(ctxutils.RequestIDKey, requestID)
			ctx := logger.CtxWithLogger(r.Context(), reqLogger)
			ctx = ctxutils.CtxWithRequestID(ctx, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AccessLog() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rec, r)

			requestLogger := logger.GetLoggerFromCtx(r.Context())
			requestLogger.Infow(
				"http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.statusCode,
				"duration_ms", time.Since(start).Milliseconds(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}
