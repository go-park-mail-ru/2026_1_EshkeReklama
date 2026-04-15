package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_SetsHeadersForAllowedOrigin(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	h := CORS([]string{"http://a.test"})(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://a.test")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if !nextCalled {
		t.Fatalf("expected next handler to be called")
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "http://a.test" {
		t.Fatalf("expected allow-origin header to be set")
	}
	if rr.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("expected allow-credentials header to be set")
	}
}

func TestCORS_OptionsShortCircuit(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	h := CORS([]string{"http://a.test"})(next)

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://a.test")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if nextCalled {
		t.Fatalf("did not expect next handler to be called on OPTIONS")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatalf("expected allow-methods to be set")
	}
}
