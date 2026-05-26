package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	errs "eshkere/internal/errors"
	"eshkere/pkg/ctxutils"
)

type stubValidator struct {
	sessions map[string]int64
}

func (v *stubValidator) ValidateSession(_ context.Context, sid string) (int64, error) {
	advID, ok := v.sessions[sid]
	if !ok {
		return 0, errs.ErrSessionNotFound
	}
	return advID, nil
}

func TestAuthMiddleware_UnauthorizedAndOK(t *testing.T) {
	validator := &stubValidator{sessions: map[string]int64{"valid-sid": 5}}

	protected := Auth(validator, "sid")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := ctxutils.AdvertiserIDFromContext(r.Context())
		if err != nil || id != 5 {
			t.Fatalf("expected advertiser id 5, got id=%d err=%v", id, err)
		}
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	protected.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", rr.Code)
	}

	okRR := httptest.NewRecorder()
	okReq := httptest.NewRequest(http.MethodGet, "/", nil)
	okReq.AddCookie(&http.Cookie{Name: "sid", Value: "valid-sid"})
	protected.ServeHTTP(okRR, okReq)
	if okRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", okRR.Code)
	}
}
