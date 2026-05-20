package middleware

import (
	"context"
	"net/http"

	"eshkere/internal/models"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"
)

type AdminChecker interface {
	GetAdvertiserByID(ctx context.Context, id int) (*models.Advertiser, error)
}

func IsAdmin(checker AdminChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if checker == nil {
				httpx.InternalError(w, "internal error")
				return
			}

			advertiserID, err := ctxutils.AdvertiserIDFromContext(r.Context())
			if err != nil {
				httpx.Unauthorized(w, "unauthorized")
				return
			}

			advertiser, err := checker.GetAdvertiserByID(r.Context(), advertiserID)
			if err != nil {
				httpx.InternalError(w, "internal error")
				return
			}
			if advertiser == nil {
				httpx.Forbidden(w, "forbidden")
				return
			}
			if advertiser.Role != models.AdvertiserRoleAdmin {
				httpx.Forbidden(w, "forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
