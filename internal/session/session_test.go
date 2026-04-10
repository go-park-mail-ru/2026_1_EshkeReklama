package session

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testCookieName = "session_id"

type memoryStore struct {
	sessions map[string]Session
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		sessions: make(map[string]Session),
	}
}

func (s *memoryStore) Save(_ context.Context, sessionID string, sess Session, _ time.Duration) error {
	s.sessions[sessionID] = sess
	return nil
}

func (s *memoryStore) Get(_ context.Context, sessionID string) (Session, error) {
	sess, ok := s.sessions[sessionID]
	if !ok {
		return Session{}, ErrStoreSessionNotFound
	}

	return sess, nil
}

func (s *memoryStore) Delete(_ context.Context, sessionID string) error {
	delete(s.sessions, sessionID)
	return nil
}

func newTestManager(store Store) *Manager {
	return NewManager(store, 24*time.Hour, CookieConfig{
		Name:     testCookieName,
		Path:     "/",
		HTTPOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func TestManager_CreateGetDestroy(t *testing.T) {
	store := newMemoryStore()
	m := newTestManager(store)

	createReq := httptest.NewRequest(http.MethodPost, "/", nil)
	createRR := httptest.NewRecorder()
	if err := m.Create(createRR, createReq, 7); err != nil {
		t.Fatalf("Create: %v", err)
	}
	createResp := createRR.Result()
	defer createResp.Body.Close()

	cookies := createResp.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Name != testCookieName || cookies[0].Value == "" {
		t.Fatalf("unexpected cookie: %#v", cookies[0])
	}

	getReq := httptest.NewRequest(http.MethodGet, "/", nil)
	getReq.AddCookie(cookies[0])
	getRR := httptest.NewRecorder()

	sess, err := m.Get(getRR, getReq)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if sess.AdvertiserID != 7 {
		t.Fatalf("expected advertiserID 7, got %d", sess.AdvertiserID)
	}
	getResp := getRR.Result()
	defer getResp.Body.Close()
	if len(getResp.Cookies()) != 1 {
		t.Fatalf("expected refreshed cookie to be set")
	}

	destroyReq := httptest.NewRequest(http.MethodPost, "/", nil)
	destroyReq.AddCookie(cookies[0])
	destroyRR := httptest.NewRecorder()
	if err := m.Destroy(destroyRR, destroyReq); err != nil {
		t.Fatalf("Destroy: %v", err)
	}
	destroyResp := destroyRR.Result()
	defer destroyResp.Body.Close()

	dCookies := destroyResp.Cookies()
	if len(dCookies) != 1 || dCookies[0].MaxAge != -1 {
		t.Fatalf("expected cookie deletion, got %#v", dCookies)
	}
}

func TestManager_GetMissingSessionReturnsNotFound(t *testing.T) {
	m := newTestManager(newMemoryStore())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: testCookieName, Value: "id"})
	rr := httptest.NewRecorder()

	_, err := m.Get(rr, req)
	if err != ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestManager_CreateDeletesOldSessionByCookie(t *testing.T) {
	store := newMemoryStore()
	store.sessions["old"] = Session{AdvertiserID: 1, ExpiresAt: time.Now().Add(time.Hour)}

	m := newTestManager(store)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: testCookieName, Value: "old"})
	rr := httptest.NewRecorder()

	if err := m.Create(rr, req, 2); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, ok := store.sessions["old"]; ok {
		t.Fatalf("expected old session to be deleted")
	}
}
