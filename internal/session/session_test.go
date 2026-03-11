package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestManager_CreateGetDestroy(t *testing.T) {
	m := NewManager()

	// Create
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
	if cookies[0].Name != CookieName || cookies[0].Value == "" {
		t.Fatalf("unexpected cookie: %#v", cookies[0])
	}

	// Get
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

	// Destroy
	destroyReq := httptest.NewRequest(http.MethodPost, "/", nil)
	destroyReq.AddCookie(cookies[0])
	destroyRR := httptest.NewRecorder()
	m.Destroy(destroyRR, destroyReq)
	destroyResp := destroyRR.Result()
	defer destroyResp.Body.Close()
	dCookies := destroyResp.Cookies()
	if len(dCookies) != 1 || dCookies[0].MaxAge != -1 {
		t.Fatalf("expected cookie deletion, got %#v", dCookies)
	}
}

func TestManager_GetExpiredSessionDeletesAndReturnsNotFound(t *testing.T) {
	m := NewManager()
	m.sessions["id"] = Session{AdvertiserID: 1, ExpiresAt: time.Now().Add(-time.Minute)}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "id"})
	rr := httptest.NewRecorder()

	_, err := m.Get(rr, req)
	if err != ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}

	if _, ok := m.sessions["id"]; ok {
		t.Fatalf("expected expired session to be deleted")
	}
}

func TestManager_CreateDeletesOldSessionByCookie(t *testing.T) {
	m := NewManager()
	m.sessions["old"] = Session{AdvertiserID: 1, ExpiresAt: time.Now().Add(time.Hour)}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "old"})
	rr := httptest.NewRecorder()

	if err := m.Create(rr, req, 2); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, ok := m.sessions["old"]; ok {
		t.Fatalf("expected old session to be deleted")
	}
}
