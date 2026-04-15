package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eshkere/internal/handler/dto"
	"eshkere/internal/models"
	"eshkere/internal/service"
	"eshkere/internal/session"

	"github.com/gorilla/mux"
	"go.uber.org/mock/gomock"
)

const testCookieName = "session_id"

type memoryStore struct {
	sessions map[string]session.Session
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		sessions: make(map[string]session.Session),
	}
}

func (s *memoryStore) Save(_ context.Context, sessionID string, sess session.Session, _ time.Duration) error {
	s.sessions[sessionID] = sess
	return nil
}

func (s *memoryStore) Get(_ context.Context, sessionID string) (session.Session, error) {
	sess, ok := s.sessions[sessionID]
	if !ok {
		return session.Session{}, session.ErrStoreSessionNotFound
	}

	return sess, nil
}

func (s *memoryStore) Delete(_ context.Context, sessionID string) error {
	delete(s.sessions, sessionID)
	return nil
}

func newTestSessionManager() *session.Manager {
	return session.NewManager(
		newMemoryStore(),
		24*time.Hour,
		session.CookieConfig{
			Name:     testCookieName,
			Path:     "/",
			HTTPOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
	)
}

func newTestRouter(sm *session.Manager, svc Service) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	r.Use(middleware.CSRF(middleware.CSRFConfig{
		CookieName: "csrf_token",
		HeaderName: "X-CSRF-Token",
	}))
	r.HandleFunc("/__ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)
	handlers.Register(r, NewAPI(APIConfig{
		SessionManager: sm,
		Service:        svc,
	}))
	return r
}

func getCSRF(t *testing.T, r *mux.Router) *http.Cookie {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/__ping", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	for _, c := range rr.Result().Cookies() {
		if c.Name == "csrf_token" && c.Value != "" {
			return c
		}
	}
	t.Fatalf("csrf_token cookie not set")
	return nil
}

func TestRegister_OK(t *testing.T) {
	sm := newTestSessionManager()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc := handlers.NewMockService(ctrl)
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)

	svc.EXPECT().
		RegisterAdvertiser(gomock.Any(), gomock.Any(), "a@a.test", "+70000000000", "secret").
		Return(&models.Advertiser{ID: 99, Email: "a@a.test", Phone: "+70000000000"}, nil)

	body := `{"email":"a@a.test","phone":"+70000000000","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/advertiser/register", bytes.NewBufferString(body))
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
	if rr.Result().Header.Get("Set-Cookie") == "" {
		t.Fatalf("expected session cookie")
	}
}

func TestLogin_UnauthorizedAndOK(t *testing.T) {
	sm := newTestSessionManager()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc := handlers.NewMockService(ctrl)
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)

	svc.EXPECT().
		AuthenticateAdvertiser(gomock.Any(), "test@mail.com", "bad").
		Return(nil, service.ErrInvalidCredentials)
	svc.EXPECT().
		AuthenticateAdvertiser(gomock.Any(), "test@mail.com", "ok").
		Return(&models.Advertiser{ID: 1, Email: "test@mail.com", Phone: "9000000000"}, nil)

	req := httptest.NewRequest(http.MethodPost, "/advertiser/login", bytes.NewBufferString(`{"identifier":"test@mail.com","password":"bad"}`))
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d body=%s", rr.Code, rr.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/advertiser/login", bytes.NewBufferString(`{"identifier":"test@mail.com","password":"ok"}`))
	req2.AddCookie(csrf)
	req2.Header.Set("X-CSRF-Token", csrf.Value)
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr2.Code, rr2.Body.String())
	}
}

func TestMe_UnauthorizedAndOK(t *testing.T) {
	sm := newTestSessionManager()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc := handlers.NewMockService(ctrl)
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)

	svc.EXPECT().
		GetAdvertiserByID(gomock.Any(), 1).
		Return(&models.Advertiser{
			ID:      1,
			Name:    "Test",
			Email:   "test@mail.com",
			Phone:   "9000000000",
			Balance: 100,
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/advertiser/me", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d body=%s", rr.Code, rr.Body.String())
	}

	createReq := httptest.NewRequest(http.MethodPost, "/", nil)
	createRR := httptest.NewRecorder()
	if err := sm.Create(createRR, createReq, 1); err != nil {
		t.Fatalf("Create: %v", err)
	}
	cookies := createRR.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected cookie")
	}

	req2 := httptest.NewRequest(http.MethodGet, "/advertiser/me", nil)
	req2.AddCookie(cookies[0])
	req2.AddCookie(csrf)
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr2.Code, rr2.Body.String())
	}
}

func TestLogout_AlwaysOK(t *testing.T) {
	sm := newTestSessionManager()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc := handlers.NewMockService(ctrl)
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)

	req := httptest.NewRequest(http.MethodPost, "/advertiser/logout", nil)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestBalance_GetAndTopUp(t *testing.T) {
	sm := newTestSessionManager()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc := handlers.NewMockService(ctrl)
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)

	svc.EXPECT().
		GetAdvertiserByID(gomock.Any(), 1).
		Return(&models.Advertiser{ID: 1, Balance: 100}, nil)
	svc.EXPECT().
		TopUpAdvertiserBalance(gomock.Any(), 1, int64(150)).
		Return(int64(250), nil)

	createReq := httptest.NewRequest(http.MethodPost, "/", nil)
	createRR := httptest.NewRecorder()
	if err := sm.Create(createRR, createReq, 1); err != nil {
		t.Fatalf("Create: %v", err)
	}

	cookies := createRR.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected cookie")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/advertiser/balance", nil)
	getReq.AddCookie(cookies[0])
	getReq.AddCookie(csrf)
	getRR := httptest.NewRecorder()
	r.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", getRR.Code, getRR.Body.String())
	}

	topupReq := httptest.NewRequest(http.MethodPost, "/advertiser/balance/topup", bytes.NewBufferString(`{"amount":150}`))
	topupReq.AddCookie(cookies[0])
	topupReq.AddCookie(csrf)
	topupReq.Header.Set("X-CSRF-Token", csrf.Value)
	topupRR := httptest.NewRecorder()
	r.ServeHTTP(topupRR, topupReq)
	if topupRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", topupRR.Code, topupRR.Body.String())
	}
}

func TestListAds_UnauthorizedAndEmptyList(t *testing.T) {
	sm := newTestSessionManager()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc := handlers.NewMockService(ctrl)
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)

	svc.EXPECT().
		ListAds(gomock.Any(), 2).
		Return([]*models.Ad{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/ad_campaigns/1/ad_groups/2/ads", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d body=%s", rr.Code, rr.Body.String())
	}

	createReq := httptest.NewRequest(http.MethodPost, "/", nil)
	createRR := httptest.NewRecorder()
	if err := sm.Create(createRR, createReq, 1); err != nil {
		t.Fatalf("Create: %v", err)
	}

	cookies := createRR.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected cookie")
	}

	req2 := httptest.NewRequest(http.MethodGet, "/ad_campaigns/1/ad_groups/2/ads", nil)
	req2.AddCookie(cookies[0])
	req2.AddCookie(csrf)
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr2.Code, rr2.Body.String())
	}

	var envelope struct {
		Data dto.ListAdsResponse `json:"data"`
	}
	if err := json.Unmarshal(rr2.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if envelope.Data.GroupID != 2 {
		t.Fatalf("expected group_id 2 got %d", envelope.Data.GroupID)
	}
	if len(envelope.Data.Ads) != 0 {
		t.Fatalf("expected empty ads list, got %d", len(envelope.Data.Ads))
	}
}

func TestFeed_EmptyList(t *testing.T) {
	sm := newTestSessionManager()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc := handlers.NewMockService(ctrl)
	r := newTestRouter(sm, svc)

	svc.EXPECT().
		GetAdsByFeedToken(gomock.Any(), "feed-token").
		Return([]*models.Ad{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/feed/feed-token", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
}
