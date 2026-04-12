package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eshkere/internal/handler/dto"
	"eshkere/internal/models"
	"eshkere/internal/service"
	"eshkere/internal/session"

	"github.com/gorilla/mux"
)

type stubService struct{}

func (stubService) RegisterAdvertiser(_ context.Context, _, email, phone, password string) (*models.Advertiser, error) {
	if email == "" || password == "" {
		return nil, service.ErrInvalidAdvertiserArg
	}
	return &models.Advertiser{ID: 99, Name: "u", Email: email, Phone: phone}, nil
}

func (stubService) AuthenticateAdvertiser(_ context.Context, _, password string) (*models.Advertiser, error) {
	if password == "bad" {
		return nil, service.ErrInvalidCredentials
	}
	return &models.Advertiser{ID: 1, Email: "test@mail.com", Phone: "9000000000"}, nil
}

func (stubService) GetAdvertiserByID(_ context.Context, id int) (*models.Advertiser, error) {
	if id != 1 {
		return nil, sql.ErrNoRows
	}
	return &models.Advertiser{
		ID:      1,
		Name:    "Test",
		Email:   "test@mail.com",
		Phone:   "9000000000",
		Balance: 100,
	}, nil
}

func (stubService) CreateAd(context.Context, *models.Ad) (*models.Ad, error) {
	return &models.Ad{}, nil
}

func (stubService) UpdateAd(context.Context, int, dto.UpdateAdRequest) error { return nil }

func (stubService) ListAds(context.Context, int) ([]*models.Ad, error) { return nil, nil }

func (stubService) DeleteAd(context.Context, int) error { return nil }

func (stubService) CreateAdCampaign(context.Context, *models.AdCampaign) (*models.AdCampaign, error) {
	return &models.AdCampaign{}, nil
}

func (stubService) UpdateAdCampaign(context.Context, int, dto.UpdateAdCampaignRequest) error {
	return nil
}

func (stubService) ListAdCampaigns(context.Context, int) ([]*models.AdCampaign, error) {
	return nil, nil
}

func (stubService) DeleteAdCampaign(context.Context, int) error { return nil }

func (stubService) CreateAdGroup(context.Context, *models.AdGroup) (*models.AdGroup, error) {
	return &models.AdGroup{}, nil
}

func (stubService) UpdateAdGroup(context.Context, int, dto.UpdateAdGroupRequest) error { return nil }

func (stubService) ListAdGroups(context.Context, int) ([]*models.AdGroup, error) { return nil, nil }

func (stubService) DeleteAdGroup(context.Context, int) error { return nil }

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

func newTestRouter(sm *session.Manager) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	Register(r, NewAPI(APIConfig{
		SessionManager: sm,
		Service:        stubService{},
	}))
	return r
}

func TestRegister_OK(t *testing.T) {
	sm := newTestSessionManager()
	r := newTestRouter(sm)

	body := `{"email":"a@a.test","phone":"+70000000000","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/advertiser/register", bytes.NewBufferString(body))
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
	r := newTestRouter(sm)

	req := httptest.NewRequest(http.MethodPost, "/advertiser/login", bytes.NewBufferString(`{"identifier":"test@mail.com","password":"bad"}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d body=%s", rr.Code, rr.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/advertiser/login", bytes.NewBufferString(`{"identifier":"test@mail.com","password":"ok"}`))
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr2.Code, rr2.Body.String())
	}
}

func TestMe_UnauthorizedAndOK(t *testing.T) {
	sm := newTestSessionManager()
	r := newTestRouter(sm)

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
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr2.Code, rr2.Body.String())
	}
}

func TestLogout_AlwaysOK(t *testing.T) {
	sm := newTestSessionManager()
	r := newTestRouter(sm)

	req := httptest.NewRequest(http.MethodPost, "/advertiser/logout", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestListAds_UnauthorizedAndEmptyList(t *testing.T) {
	sm := newTestSessionManager()
	r := newTestRouter(sm)

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
