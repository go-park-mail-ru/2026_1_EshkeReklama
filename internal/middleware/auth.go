package middleware

import (
	"context"
	"errors"
	"eshkere/internal/session"
	"eshkere/pkg/httpx"
	"net/http"
)

func Auth(sm *session.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, err := sm.Get(w, r)
			if err != nil {
				if errors.Is(err, session.ErrSessionNotFound) {
					httpx.Unauthorized(w, "unauthorized")
					return
				}

				httpx.InternalError(w, "internal error")
				return
			}

			ctx := context.WithValue(r.Context(), AdvertiserIDKey, sess.AdvertiserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
