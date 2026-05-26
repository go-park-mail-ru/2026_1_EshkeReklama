package middleware

import (
	"context"
	"errors"
	"net/http"

	errs "eshkere/internal/errors"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"
)

type SessionValidator interface {
	ValidateSession(ctx context.Context, sessionID string) (advertiserID int64, err error)
}

func Auth(validator SessionValidator, cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieName)
			if err != nil {
				httpx.Unauthorized(w, "unauthorized")
				return
			}

			advertiserID, err := validator.ValidateSession(r.Context(), cookie.Value)
			if err != nil {
				if errors.Is(err, errs.ErrSessionNotFound) {
					httpx.Unauthorized(w, "unauthorized")
					return
				}
				httpx.InternalError(w, "internal error")
				return
			}

			ctx := context.WithValue(r.Context(), ctxutils.AdvertiserIDKey, int(advertiserID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
