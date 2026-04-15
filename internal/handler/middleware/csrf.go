package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"eshkere/pkg/httpx"
	"net/http"
	"net/url"
	"strings"
)

type CSRFConfig struct {
	CookieName string
	HeaderName string
	Secure     bool
}

func CSRF(cfg CSRFConfig) func(http.Handler) http.Handler {
	cookieName := cfg.CookieName
	if cookieName == "" {
		cookieName = "csrf_token"
	}
	headerName := cfg.HeaderName
	if headerName == "" {
		headerName = "X-CSRF-Token"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, hasToken := readCSRFCookie(r, cookieName)
			if !hasToken {
				token = generateCSRFToken()
				http.SetCookie(w, &http.Cookie{
					Name:     cookieName,
					Value:    token,
					Path:     "/",
					Secure:   cfg.Secure,
					HttpOnly: false, // must be readable by JS to send as header
					SameSite: http.SameSiteLaxMode,
				})
			}

			if !isUnsafeMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			// Basic origin check when header is present (browser requests).
			// This complements double-submit cookie validation.
			if origin := r.Header.Get("Origin"); origin != "" {
				if !sameOrigin(origin, r.Host) {
					httpx.ErrorJSON(w, http.StatusForbidden, "csrf blocked")
					return
				}
			}

			reqToken := r.Header.Get(headerName)
			if reqToken == "" || token == "" || subtleConstantTimeStringEq(reqToken, token) == false {
				httpx.ErrorJSON(w, http.StatusForbidden, "csrf token required")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func readCSRFCookie(r *http.Request, cookieName string) (string, bool) {
	c, err := r.Cookie(cookieName)
	if err != nil || c == nil || c.Value == "" {
		return "", false
	}
	return c.Value, true
}

func generateCSRFToken() string {
	var b [32]byte
	_, _ = rand.Read(b[:])
	return base64.RawURLEncoding.EncodeToString(b[:])
}

func isUnsafeMethod(m string) bool {
	switch m {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func sameOrigin(originHeader string, host string) bool {
	u, err := url.Parse(originHeader)
	if err != nil {
		return false
	}
	originHost := u.Host
	if originHost == "" {
		return false
	}
	// Normalize host header (may include port)
	return strings.EqualFold(originHost, host)
}

// minimal constant-time compare for short tokens
func subtleConstantTimeStringEq(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}
