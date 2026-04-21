package middleware

import (
	"context"
	"eshkere/pkg/ctxutils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eshkere/internal/session"
)

type memStore struct {
	m map[string]session.Session
}

func newMemStore() *memStore { return &memStore{m: make(map[string]session.Session)} }

func (s *memStore) Save(_ context.Context, sessionID string, sess session.Session, _ time.Duration) error {
	s.m[sessionID] = sess
	return nil
}
func (s *memStore) Get(_ context.Context, sessionID string) (session.Session, error) {
	v, ok := s.m[sessionID]
	if !ok {
		return session.Session{}, session.ErrStoreSessionNotFound
	}
	return v, nil
}
func (s *memStore) Delete(_ context.Context, sessionID string) error {
	delete(s.m, sessionID)
	return nil
}

func TestAuthMiddleware_UnauthorizedAndOK(t *testing.T) {
	sm := session.NewManager(
		newMemStore(),
		24*time.Hour,
		session.CookieConfig{Name: "sid", Path: "/", HTTPOnly: true, SameSite: http.SameSiteLaxMode},
	)

	protected := Auth(sm)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	createRR := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/", nil)
	if err := sm.Create(createRR, createReq, 5); err != nil {
		t.Fatalf("Create: %v", err)
	}
	c := createRR.Result().Cookies()[0]

	okRR := httptest.NewRecorder()
	okReq := httptest.NewRequest(http.MethodGet, "/", nil)
	okReq.AddCookie(c)
	protected.ServeHTTP(okRR, okReq)
	if okRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", okRR.Code)
	}
}
