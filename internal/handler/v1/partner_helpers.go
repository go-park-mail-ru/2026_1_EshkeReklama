package v1

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (a *API) setSessionCookieWithConfig(w http.ResponseWriter, cfg CookieConfig, sessionID string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.Name,
		Value:    sessionID,
		Path:     cfg.Path,
		Expires:  expiresAt,
		HttpOnly: cfg.HTTPOnly,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *API) clearSessionCookieWithConfig(w http.ResponseWriter, cfg CookieConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.Name,
		Value:    "",
		Path:     cfg.Path,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: cfg.HTTPOnly,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func requestBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if forwardedProto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwardedProto != "" {
		scheme = forwardedProto
	}
	host := r.Host
	if forwardedHost := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); forwardedHost != "" {
		host = forwardedHost
	}
	return fmt.Sprintf("%s://%s", scheme, host)
}
