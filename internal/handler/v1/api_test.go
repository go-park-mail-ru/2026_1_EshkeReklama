package v1

import (
	"bytes"
	"context"
	"encoding/json"
	errs "eshkere/internal/errors"
	handlers "eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eshkere/internal/handler/v1/dto"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"eshkere/internal/session"

	"github.com/gorilla/mux"
)

const testCookieName = "session_id"

type memoryStore struct {
	sessions map[string]session.Session
}

type stubService struct {
	registerAdvertiserFn      func(ctx context.Context, name, email, phone, password string) (*models.Advertiser, error)
	authenticateAdvertiserFn  func(ctx context.Context, identifier, password string) (*models.Advertiser, error)
	getAdvertiserByIDFn       func(ctx context.Context, id int) (*models.Advertiser, error)
	updateAdvertiserProfileFn func(ctx context.Context, in *serviceinput.UpdateAdvertiserProfile) (*models.Advertiser, error)
	updateAdvertiserAvatarFn  func(ctx context.Context, advertiserID int, avatar []byte, avatarExt, avatarContentType string) (*models.Advertiser, error)
	topUpAdvertiserBalanceFn  func(ctx context.Context, advertiserID int, amount int64) (int64, error)
	generateFeedLinkFn        func(ctx context.Context, campaignID int) (string, error)
	getAdsByFeedTokenFn       func(ctx context.Context, token string) ([]*models.Ad, error)
	createAdCampaignFn        func(ctx context.Context, in *serviceinput.CreateAdCampaign) (*models.AdCampaign, error)
	updateAdCampaignFn        func(ctx context.Context, in *serviceinput.UpdateAdCampaign) error
	listAdCampaignsFn         func(ctx context.Context, advertiserID int) ([]*models.AdCampaign, error)
	deleteAdCampaignFn        func(ctx context.Context, campaignID int) error
	createAdGroupFn           func(ctx context.Context, in *serviceinput.CreateAdGroup) (*models.AdGroup, error)
	updateAdGroupFn           func(ctx context.Context, in *serviceinput.UpdateAdGroup) error
	listAdGroupsFn            func(ctx context.Context, campaignID int) ([]*models.AdGroup, error)
	deleteAdGroupFn           func(ctx context.Context, groupID int) error
	createAdFn                func(ctx context.Context, in *serviceinput.CreateAd) (*models.Ad, error)
	updateAdFn                func(ctx context.Context, in *serviceinput.UpdateAd) error
	listAdsFn                 func(ctx context.Context, groupID int) ([]*models.Ad, error)
	deleteAdFn                func(ctx context.Context, adID int) error
}

func (s *stubService) RegisterAdvertiser(ctx context.Context, name, email, phone, password string) (*models.Advertiser, error) {
	if s.registerAdvertiserFn != nil {
		return s.registerAdvertiserFn(ctx, name, email, phone, password)
	}
	return nil, nil
}

func (s *stubService) AuthenticateAdvertiser(ctx context.Context, identifier, password string) (*models.Advertiser, error) {
	if s.authenticateAdvertiserFn != nil {
		return s.authenticateAdvertiserFn(ctx, identifier, password)
	}
	return nil, nil
}

func (s *stubService) GetAdvertiserByID(ctx context.Context, id int) (*models.Advertiser, error) {
	if s.getAdvertiserByIDFn != nil {
		return s.getAdvertiserByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *stubService) UpdateAdvertiserProfile(ctx context.Context, in *serviceinput.UpdateAdvertiserProfile) (*models.Advertiser, error) {
	if s.updateAdvertiserProfileFn != nil {
		return s.updateAdvertiserProfileFn(ctx, in)
	}
	return nil, nil
}

func (s *stubService) UpdateAdvertiserAvatar(ctx context.Context, advertiserID int, avatar []byte, avatarExt, avatarContentType string) (*models.Advertiser, error) {
	if s.updateAdvertiserAvatarFn != nil {
		return s.updateAdvertiserAvatarFn(ctx, advertiserID, avatar, avatarExt, avatarContentType)
	}
	return nil, nil
}

func (s *stubService) TopUpAdvertiserBalance(ctx context.Context, advertiserID int, amount int64) (int64, error) {
	if s.topUpAdvertiserBalanceFn != nil {
		return s.topUpAdvertiserBalanceFn(ctx, advertiserID, amount)
	}
	return 0, nil
}

func (s *stubService) GenerateFeedLink(ctx context.Context, campaignID int) (string, error) {
	if s.generateFeedLinkFn != nil {
		return s.generateFeedLinkFn(ctx, campaignID)
	}
	return "", nil
}

func (s *stubService) GetAdsByFeedToken(ctx context.Context, token string) ([]*models.Ad, error) {
	if s.getAdsByFeedTokenFn != nil {
		return s.getAdsByFeedTokenFn(ctx, token)
	}
	return nil, nil
}

func (s *stubService) CreateAdCampaign(ctx context.Context, in *serviceinput.CreateAdCampaign) (*models.AdCampaign, error) {
	if s.createAdCampaignFn != nil {
		return s.createAdCampaignFn(ctx, in)
	}
	return nil, nil
}

func (s *stubService) UpdateAdCampaign(ctx context.Context, in *serviceinput.UpdateAdCampaign) error {
	if s.updateAdCampaignFn != nil {
		return s.updateAdCampaignFn(ctx, in)
	}
	return nil
}

func (s *stubService) ListAdCampaigns(ctx context.Context, advertiserID int) ([]*models.AdCampaign, error) {
	if s.listAdCampaignsFn != nil {
		return s.listAdCampaignsFn(ctx, advertiserID)
	}
	return nil, nil
}

func (s *stubService) DeleteAdCampaign(ctx context.Context, campaignID int) error {
	if s.deleteAdCampaignFn != nil {
		return s.deleteAdCampaignFn(ctx, campaignID)
	}
	return nil
}

func (s *stubService) CreateAdGroup(ctx context.Context, in *serviceinput.CreateAdGroup) (*models.AdGroup, error) {
	if s.createAdGroupFn != nil {
		return s.createAdGroupFn(ctx, in)
	}
	return nil, nil
}

func (s *stubService) UpdateAdGroup(ctx context.Context, in *serviceinput.UpdateAdGroup) error {
	if s.updateAdGroupFn != nil {
		return s.updateAdGroupFn(ctx, in)
	}
	return nil
}

func (s *stubService) ListAdGroups(ctx context.Context, campaignID int) ([]*models.AdGroup, error) {
	if s.listAdGroupsFn != nil {
		return s.listAdGroupsFn(ctx, campaignID)
	}
	return nil, nil
}

func (s *stubService) DeleteAdGroup(ctx context.Context, groupID int) error {
	if s.deleteAdGroupFn != nil {
		return s.deleteAdGroupFn(ctx, groupID)
	}
	return nil
}

func (s *stubService) CreateAd(ctx context.Context, in *serviceinput.CreateAd) (*models.Ad, error) {
	if s.createAdFn != nil {
		return s.createAdFn(ctx, in)
	}
	return nil, nil
}

func (s *stubService) UpdateAd(ctx context.Context, in *serviceinput.UpdateAd) error {
	if s.updateAdFn != nil {
		return s.updateAdFn(ctx, in)
	}
	return nil
}

func (s *stubService) ListAds(ctx context.Context, groupID int) ([]*models.Ad, error) {
	if s.listAdsFn != nil {
		return s.listAdsFn(ctx, groupID)
	}
	return nil, nil
}

func (s *stubService) DeleteAd(ctx context.Context, adID int) error {
	if s.deleteAdFn != nil {
		return s.deleteAdFn(ctx, adID)
	}
	return nil
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

func createSessionCookie(t *testing.T, sm *session.Manager, advertiserID int) *http.Cookie {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rr := httptest.NewRecorder()
	if err := sm.Create(rr, req, advertiserID); err != nil {
		t.Fatalf("Create session: %v", err)
	}
	cookies := rr.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie")
	}
	return cookies[0]
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
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)

	svc.registerAdvertiserFn = func(_ context.Context, _ string, email, phone, password string) (*models.Advertiser, error) {
		if email != "a@a.test" || phone != "+70000000000" || password != "secret" {
			t.Fatalf("unexpected register args: email=%s phone=%s password=%s", email, phone, password)
		}
		return &models.Advertiser{ID: 99, Email: email, Phone: phone}, nil
	}

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
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)

	svc.authenticateAdvertiserFn = func(_ context.Context, identifier, password string) (*models.Advertiser, error) {
		if identifier != "test@mail.com" {
			t.Fatalf("unexpected identifier: %s", identifier)
		}
		if password == "bad" {
			return nil, errs.ErrInvalidCredentials
		}
		if password == "ok" {
			return &models.Advertiser{ID: 1, Email: "test@mail.com", Phone: "9000000000"}, nil
		}
		t.Fatalf("unexpected password: %s", password)
		return nil, nil
	}

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
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)

	svc.getAdvertiserByIDFn = func(_ context.Context, id int) (*models.Advertiser, error) {
		if id != 1 {
			t.Fatalf("unexpected advertiser id: %d", id)
		}
		return &models.Advertiser{
			ID:      1,
			Name:    "Test",
			Email:   "test@mail.com",
			Phone:   "9000000000",
			Balance: 100,
		}, nil
	}

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
	svc := &stubService{}
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
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)

	svc.getAdvertiserByIDFn = func(_ context.Context, id int) (*models.Advertiser, error) {
		if id != 1 {
			t.Fatalf("unexpected advertiser id: %d", id)
		}
		return &models.Advertiser{ID: 1, Balance: 100}, nil
	}
	svc.topUpAdvertiserBalanceFn = func(_ context.Context, advertiserID int, amount int64) (int64, error) {
		if advertiserID != 1 || amount != 150 {
			t.Fatalf("unexpected topup args: advertiserID=%d amount=%d", advertiserID, amount)
		}
		return 250, nil
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
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)

	svc.listAdsFn = func(_ context.Context, groupID int) ([]*models.Ad, error) {
		if groupID != 2 {
			t.Fatalf("unexpected group id: %d", groupID)
		}
		return []*models.Ad{}, nil
	}

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
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	svc.getAdsByFeedTokenFn = func(_ context.Context, token string) ([]*models.Ad, error) {
		if token != "feed-token" {
			t.Fatalf("unexpected token: %s", token)
		}
		return []*models.Ad{}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/feed/feed-token", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
}
