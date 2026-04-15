package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestCSRF_SetsCookieOnGET(t *testing.T) {
	r := mux.NewRouter()
	r.Use(CSRF(CSRFConfig{CookieName: "csrf_token", HeaderName: "X-CSRF-Token"}))
	r.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}

	cookies := rr.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "csrf_token" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected csrf_token cookie to be set")
	}
}

func TestCSRF_BlocksUnsafeWithoutHeader(t *testing.T) {
	r := mux.NewRouter()
	r.Use(CSRF(CSRFConfig{CookieName: "csrf_token", HeaderName: "X-CSRF-Token"}))
	r.HandleFunc("/mut", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodPost, http.MethodGet)

	// first request to obtain cookie
	getReq := httptest.NewRequest(http.MethodGet, "/mut", nil)
	getRR := httptest.NewRecorder()
	r.ServeHTTP(getRR, getReq)

	var csrfCookie *http.Cookie
	for _, c := range getRR.Result().Cookies() {
		if c.Name == "csrf_token" {
			csrfCookie = c
			break
		}
	}
	if csrfCookie == nil || csrfCookie.Value == "" {
		t.Fatalf("expected csrf cookie")
	}

	postReq := httptest.NewRequest(http.MethodPost, "/mut", nil)
	postReq.AddCookie(csrfCookie)
	postRR := httptest.NewRecorder()
	r.ServeHTTP(postRR, postReq)

	if postRR.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got %d body=%s", postRR.Code, postRR.Body.String())
	}
}

func TestCSRF_AllowsUnsafeWithMatchingHeaderAndCookie(t *testing.T) {
	r := mux.NewRouter()
	r.Use(CSRF(CSRFConfig{CookieName: "csrf_token", HeaderName: "X-CSRF-Token"}))
	r.HandleFunc("/mut", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodPost, http.MethodGet)

	// get cookie
	getReq := httptest.NewRequest(http.MethodGet, "/mut", nil)
	getRR := httptest.NewRecorder()
	r.ServeHTTP(getRR, getReq)

	var csrfCookie *http.Cookie
	for _, c := range getRR.Result().Cookies() {
		if c.Name == "csrf_token" {
			csrfCookie = c
			break
		}
	}
	if csrfCookie == nil || csrfCookie.Value == "" {
		t.Fatalf("expected csrf cookie")
	}

	postReq := httptest.NewRequest(http.MethodPost, "/mut", nil)
	postReq.AddCookie(csrfCookie)
	postReq.Header.Set("X-CSRF-Token", csrfCookie.Value)
	postRR := httptest.NewRecorder()
	r.ServeHTTP(postRR, postReq)

	if postRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", postRR.Code, postRR.Body.String())
	}
}
