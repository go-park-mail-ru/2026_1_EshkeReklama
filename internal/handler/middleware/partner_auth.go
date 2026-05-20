package middleware

import (
	"context"
	"errors"
	"net/http"

	errs "eshkere/internal/errors"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"
)

func PartnerAuth(validator SessionValidator, cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieName)
			if err != nil {
				httpx.Unauthorized(w, "unauthorized")
				return
			}

			partnerID, err := validator.ValidateSession(r.Context(), cookie.Value)
			if err != nil {
				if errors.Is(err, errs.ErrSessionNotFound) {
					httpx.Unauthorized(w, "unauthorized")
					return
				}
				httpx.InternalError(w, "internal error")
				return
			}

			ctx := context.WithValue(r.Context(), ctxutils.PartnerIDKey, int(partnerID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
