package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"eshkere/pkg"

	"go.uber.org/zap"
)

func TestSameOrigin(t *testing.T) {
	if !sameOrigin("http://example.com:80", "example.com:80") {
		t.Fatalf("expected same origin")
	}
	if sameOrigin("http://evil.com", "example.com") {
		t.Fatalf("expected different origin")
	}
}

func TestAccessLog_WrapsStatusCode(t *testing.T) {
	h := AccessLog()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusTeapot {
		t.Fatalf("expected 418 got %d", rr.Code)
	}
}

func TestRequestContext_SetsRequestID(t *testing.T) {
	h := RequestContext(zap.NewNop().Sugar())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if pkg.RequestIDFromCtx(r.Context()) == "" {
			t.Fatalf("expected request id in context")
		}
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
}
