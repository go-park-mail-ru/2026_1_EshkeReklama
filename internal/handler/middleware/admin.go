package middleware

import (
	"eshkere/pkg/httpx"
	"net/http"
)

const adminTokenHeader = "X-Admin-Token"

func AdminToken(expected string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expected == "" {
				httpx.InternalError(w, "admin token is not configured")
				return
			}

			if r.Header.Get(adminTokenHeader) != expected {
				httpx.Unauthorized(w, "unauthorized")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
