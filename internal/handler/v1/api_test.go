package v1

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	errs "eshkere/internal/errors"
	handlers "eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/internal/models"
	"eshkere/internal/service"
	serviceinput "eshkere/internal/service/input"

	"github.com/gorilla/mux"
)

const testCookieName = "session_id"

// stubAuthClient simulates the auth service for handler tests.
type stubAuthClient struct {
	sessions            map[string]int64 // sessionID → advertiserID
	credentials         map[int64]authTestCredentials
	registerFn          func(ctx context.Context, email, phone, password string) (int64, string, int64, error)
	loginFn             func(ctx context.Context, identifier, password string) (int64, string, int64, error)
	loginVKIDFn         func(ctx context.Context, code, deviceID, codeVerifier string) (int64, string, int64, error)
	validateFn          func(ctx context.Context, sessionID string) (int64, error)
	logoutFn            func(ctx context.Context, sessionID string) error
	getCredentialsFn    func(ctx context.Context, advertiserID int64) (string, string, error)
	updateCredentialsFn func(ctx context.Context, advertiserID int64, email, phone string) (string, string, error)
}

type authTestCredentials struct {
	email string
	phone string
}

func newStubAuthClient() *stubAuthClient {
	return &stubAuthClient{
		sessions:    make(map[string]int64),
		credentials: make(map[int64]authTestCredentials),
	}
}

func (c *stubAuthClient) addSession(sessionID string, advertiserID int64) {
	c.sessions[sessionID] = advertiserID
}

func (c *stubAuthClient) setCredentials(advertiserID int64, email, phone string) {
	c.credentials[advertiserID] = authTestCredentials{email: email, phone: phone}
}

func (c *stubAuthClient) Register(ctx context.Context, email, phone, password string) (int64, string, int64, error) {
	if c.registerFn != nil {
		return c.registerFn(ctx, email, phone, password)
	}
	return 0, "", 0, nil
}

func (c *stubAuthClient) Login(ctx context.Context, identifier, password string) (int64, string, int64, error) {
	if c.loginFn != nil {
		return c.loginFn(ctx, identifier, password)
	}
	return 0, "", 0, nil
}

func (c *stubAuthClient) LoginVKID(ctx context.Context, code, deviceID, codeVerifier string) (int64, string, int64, error) {
	if c.loginVKIDFn != nil {
		return c.loginVKIDFn(ctx, code, deviceID, codeVerifier)
	}
	return 0, "", 0, nil
}

func (c *stubAuthClient) ValidateSession(ctx context.Context, sessionID string) (int64, error) {
	if c.validateFn != nil {
		return c.validateFn(ctx, sessionID)
	}
	advID, ok := c.sessions[sessionID]
	if !ok {
		return 0, errs.ErrSessionNotFound
	}
	return advID, nil
}

func (c *stubAuthClient) Logout(ctx context.Context, sessionID string) error {
	if c.logoutFn != nil {
		return c.logoutFn(ctx, sessionID)
	}
	delete(c.sessions, sessionID)
	return nil
}

func (c *stubAuthClient) GetCredentials(ctx context.Context, advertiserID int64) (string, string, error) {
	if c.getCredentialsFn != nil {
		return c.getCredentialsFn(ctx, advertiserID)
	}
	cred := c.credentials[advertiserID]
	return cred.email, cred.phone, nil
}

func (c *stubAuthClient) UpdateCredentials(ctx context.Context, advertiserID int64, email, phone string) (string, string, error) {
	if c.updateCredentialsFn != nil {
		return c.updateCredentialsFn(ctx, advertiserID, email, phone)
	}
	c.setCredentials(advertiserID, email, phone)
	return email, phone, nil
}

// stubService implements the Service interface for handler tests.
type stubService struct {
	createAdvertiserProfileFn   func(ctx context.Context, id int64, name, email string) error
	getAdvertiserByIDFn         func(ctx context.Context, id int) (*models.Advertiser, error)
	updateAdvertiserProfileFn   func(ctx context.Context, in *serviceinput.UpdateAdvertiserProfile) (*models.Advertiser, error)
	updateAdvertiserAvatarFn    func(ctx context.Context, advertiserID int, avatar []byte, avatarExt, avatarContentType string) (*models.Advertiser, error)
	topUpAdvertiserBalanceFn    func(ctx context.Context, advertiserID int, amount int64) (int64, error)
	generateFeedLinkFn          func(ctx context.Context, campaignID int) (string, error)
	getAdsByFeedTokenFn         func(ctx context.Context, token string) ([]*models.Ad, error)
	requestAdFn                 func(ctx context.Context, embedToken, visitorID string) (*service.AdRequestResult, error)
	clickAdFn                   func(ctx context.Context, requestID string) (string, error)
	createAdCampaignFn          func(ctx context.Context, in *serviceinput.CreateAdCampaign) (*models.AdCampaign, error)
	updateAdCampaignFn          func(ctx context.Context, advertiserID int, in *serviceinput.UpdateAdCampaign) error
	listAdCampaignsFn           func(ctx context.Context, advertiserID int) ([]*models.AdCampaign, error)
	deleteAdCampaignFn          func(ctx context.Context, advertiserID, campaignID int) error
	createAdGroupFn             func(ctx context.Context, advertiserID int, in *serviceinput.CreateAdGroup) (*models.AdGroup, error)
	updateAdGroupFn             func(ctx context.Context, advertiserID int, in *serviceinput.UpdateAdGroup) error
	listAdGroupsFn              func(ctx context.Context, advertiserID, campaignID int) ([]*models.AdGroup, error)
	deleteAdGroupFn             func(ctx context.Context, advertiserID, groupID int) error
	createAdFn                  func(ctx context.Context, advertiserID int, in *serviceinput.CreateAd) (*models.Ad, error)
	getAdByIDFn                 func(ctx context.Context, adID int) (*models.Ad, error)
	listModerationAdsFn         func(ctx context.Context) ([]*models.Ad, error)
	updateAdFn                  func(ctx context.Context, advertiserID int, in *serviceinput.UpdateAd) error
	updateAdModerationStatusFn  func(ctx context.Context, in *serviceinput.UpdateAdStatus) error
	listAdsFn                   func(ctx context.Context, advertiserID, groupID int) ([]*models.Ad, error)
	deleteAdFn                  func(ctx context.Context, advertiserID, adID int) error
	createAppealFn              func(ctx context.Context, in *serviceinput.CreateAppeal) (*models.Appeal, error)
	listAppealsFn               func(ctx context.Context, advertiserID int) ([]*models.Appeal, error)
	getAppealByIDFn             func(ctx context.Context, appealID int) (*models.Appeal, error)
	createPartnerProfileFn      func(ctx context.Context, in *serviceinput.CreatePartnerProfile) error
	getPartnerByIDFn            func(ctx context.Context, id int) (*models.Partner, error)
	updatePartnerProfileFn      func(ctx context.Context, in *serviceinput.UpdatePartnerProfile) (*models.Partner, error)
	createPartnerSiteFn         func(ctx context.Context, in *serviceinput.CreatePartnerSite) (*models.PartnerSite, error)
	getPartnerSiteFn            func(ctx context.Context, siteID int) (*models.PartnerSite, error)
	listPartnerSitesFn          func(ctx context.Context, partnerID int) ([]*models.PartnerSite, error)
	updatePartnerSiteFn         func(ctx context.Context, in *serviceinput.UpdatePartnerSite) (*models.PartnerSite, error)
	deletePartnerSiteFn         func(ctx context.Context, partnerID, siteID int) error
	createPartnerBlockFn        func(ctx context.Context, partnerID int, in *serviceinput.CreatePartnerBlock) (*models.PartnerBlock, error)
	getPartnerBlockFn           func(ctx context.Context, partnerID, siteID, blockID int) (*models.PartnerBlock, []*models.PartnerBlockGeoRule, error)
	listPartnerBlocksFn         func(ctx context.Context, partnerID, siteID int) ([]*models.PartnerBlock, error)
	updatePartnerBlockMetaFn    func(ctx context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockMeta) (*models.PartnerBlock, error)
	updatePartnerBlockGeneralFn func(ctx context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockGeneral) (*models.PartnerBlock, error)
	updatePartnerBlockGeoFn     func(ctx context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockGeography) ([]*models.PartnerBlockGeoRule, error)
	updatePartnerBlockSelfAdFn  func(ctx context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockSelfAd) (*models.PartnerBlock, error)
	deletePartnerBlockFn        func(ctx context.Context, partnerID, siteID, blockID int) error
	getPartnerBlockEmbedCodeFn  func(ctx context.Context, partnerID, siteID, blockID int, baseURL string) (string, string, error)
}

func (s *stubService) CreateAdvertiserProfile(ctx context.Context, id int64, name, email string) error {
	if s.createAdvertiserProfileFn != nil {
		return s.createAdvertiserProfileFn(ctx, id, name, email)
	}
	return nil
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

func (s *stubService) GetAdByFeedToken(ctx context.Context, token string) (*models.Ad, error) {
	if s.getAdsByFeedTokenFn != nil {
		ads, err := s.getAdsByFeedTokenFn(ctx, token)
		if err != nil {
			return nil, err
		}
		if len(ads) == 0 {
			return &models.Ad{}, nil
		}
		return ads[0], nil
	}
	return nil, nil
}

func (s *stubService) GetAdsByFeedToken(ctx context.Context, token string) ([]*models.Ad, error) {
	if s.getAdsByFeedTokenFn != nil {
		return s.getAdsByFeedTokenFn(ctx, token)
	}
	return nil, nil
}

func (s *stubService) RequestAd(ctx context.Context, embedToken, visitorID string) (*service.AdRequestResult, error) {
	if s.requestAdFn != nil {
		return s.requestAdFn(ctx, embedToken, visitorID)
	}
	return nil, nil
}

func (s *stubService) ClickAd(ctx context.Context, requestID string) (string, error) {
	if s.clickAdFn != nil {
		return s.clickAdFn(ctx, requestID)
	}
	return "", nil
}

func (s *stubService) CreateAdCampaign(ctx context.Context, in *serviceinput.CreateAdCampaign) (*models.AdCampaign, error) {
	if s.createAdCampaignFn != nil {
		return s.createAdCampaignFn(ctx, in)
	}
	return nil, nil
}

func (s *stubService) UpdateAdCampaign(ctx context.Context, advertiserID int, in *serviceinput.UpdateAdCampaign) error {
	if s.updateAdCampaignFn != nil {
		return s.updateAdCampaignFn(ctx, advertiserID, in)
	}
	return nil
}

func (s *stubService) ListAdCampaigns(ctx context.Context, advertiserID int) ([]*models.AdCampaign, error) {
	if s.listAdCampaignsFn != nil {
		return s.listAdCampaignsFn(ctx, advertiserID)
	}
	return nil, nil
}

func (s *stubService) DeleteAdCampaign(ctx context.Context, advertiserID, campaignID int) error {
	if s.deleteAdCampaignFn != nil {
		return s.deleteAdCampaignFn(ctx, advertiserID, campaignID)
	}
	return nil
}

func (s *stubService) CreateAdGroup(ctx context.Context, advertiserID int, in *serviceinput.CreateAdGroup) (*models.AdGroup, error) {
	if s.createAdGroupFn != nil {
		return s.createAdGroupFn(ctx, advertiserID, in)
	}
	return nil, nil
}

func (s *stubService) UpdateAdGroup(ctx context.Context, advertiserID int, in *serviceinput.UpdateAdGroup) error {
	if s.updateAdGroupFn != nil {
		return s.updateAdGroupFn(ctx, advertiserID, in)
	}
	return nil
}

func (s *stubService) ListAdGroups(ctx context.Context, advertiserID, campaignID int) ([]*models.AdGroup, error) {
	if s.listAdGroupsFn != nil {
		return s.listAdGroupsFn(ctx, advertiserID, campaignID)
	}
	return nil, nil
}

func (s *stubService) DeleteAdGroup(ctx context.Context, advertiserID, groupID int) error {
	if s.deleteAdGroupFn != nil {
		return s.deleteAdGroupFn(ctx, advertiserID, groupID)
	}
	return nil
}

func (s *stubService) CreateAd(ctx context.Context, advertiserID int, in *serviceinput.CreateAd) (*models.Ad, error) {
	if s.createAdFn != nil {
		return s.createAdFn(ctx, advertiserID, in)
	}
	return nil, nil
}

func (s *stubService) GetAdByID(ctx context.Context, adID int) (*models.Ad, error) {
	if s.getAdByIDFn != nil {
		return s.getAdByIDFn(ctx, adID)
	}
	return nil, nil
}

func (s *stubService) ListModerationAds(ctx context.Context) ([]*models.Ad, error) {
	if s.listModerationAdsFn != nil {
		return s.listModerationAdsFn(ctx)
	}
	return nil, nil
}

func (s *stubService) UpdateAd(ctx context.Context, advertiserID int, in *serviceinput.UpdateAd) error {
	if s.updateAdFn != nil {
		return s.updateAdFn(ctx, advertiserID, in)
	}
	return nil
}

func (s *stubService) UpdateAdModerationStatus(ctx context.Context, in *serviceinput.UpdateAdStatus) error {
	if s.updateAdModerationStatusFn != nil {
		return s.updateAdModerationStatusFn(ctx, in)
	}
	return nil
}

func (s *stubService) ListAds(ctx context.Context, advertiserID, groupID int) ([]*models.Ad, error) {
	if s.listAdsFn != nil {
		return s.listAdsFn(ctx, advertiserID, groupID)
	}
	return nil, nil
}

func (s *stubService) DeleteAd(ctx context.Context, advertiserID, adID int) error {
	if s.deleteAdFn != nil {
		return s.deleteAdFn(ctx, advertiserID, adID)
	}
	return nil
}

func (s *stubService) CreateAppeal(ctx context.Context, in *serviceinput.CreateAppeal) (*models.Appeal, error) {
	if s.createAppealFn != nil {
		return s.createAppealFn(ctx, in)
	}
	return nil, nil
}

func (s *stubService) ListAppeals(ctx context.Context, advertiserID int) ([]*models.Appeal, error) {
	if s.listAppealsFn != nil {
		return s.listAppealsFn(ctx, advertiserID)
	}
	return nil, nil
}

func (s *stubService) GetAppealByID(ctx context.Context, appealID int) (*models.Appeal, error) {
	if s.getAppealByIDFn != nil {
		return s.getAppealByIDFn(ctx, appealID)
	}
	return nil, nil
}

func (s *stubService) CreatePartnerProfile(ctx context.Context, in *serviceinput.CreatePartnerProfile) error {
	if s.createPartnerProfileFn != nil {
		return s.createPartnerProfileFn(ctx, in)
	}
	return nil
}

func (s *stubService) GetPartnerByID(ctx context.Context, id int) (*models.Partner, error) {
	if s.getPartnerByIDFn != nil {
		return s.getPartnerByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *stubService) UpdatePartnerProfile(ctx context.Context, in *serviceinput.UpdatePartnerProfile) (*models.Partner, error) {
	if s.updatePartnerProfileFn != nil {
		return s.updatePartnerProfileFn(ctx, in)
	}
	return nil, nil
}

func (s *stubService) CreatePartnerSite(ctx context.Context, in *serviceinput.CreatePartnerSite) (*models.PartnerSite, error) {
	if s.createPartnerSiteFn != nil {
		return s.createPartnerSiteFn(ctx, in)
	}
	return nil, nil
}

func (s *stubService) GetPartnerSite(ctx context.Context, siteID int) (*models.PartnerSite, error) {
	if s.getPartnerSiteFn != nil {
		return s.getPartnerSiteFn(ctx, siteID)
	}
	return nil, nil
}

func (s *stubService) ListPartnerSites(ctx context.Context, partnerID int) ([]*models.PartnerSite, error) {
	if s.listPartnerSitesFn != nil {
		return s.listPartnerSitesFn(ctx, partnerID)
	}
	return nil, nil
}

func (s *stubService) UpdatePartnerSite(ctx context.Context, in *serviceinput.UpdatePartnerSite) (*models.PartnerSite, error) {
	if s.updatePartnerSiteFn != nil {
		return s.updatePartnerSiteFn(ctx, in)
	}
	return nil, nil
}

func (s *stubService) DeletePartnerSite(ctx context.Context, partnerID, siteID int) error {
	if s.deletePartnerSiteFn != nil {
		return s.deletePartnerSiteFn(ctx, partnerID, siteID)
	}
	return nil
}

func (s *stubService) CreatePartnerBlock(ctx context.Context, partnerID int, in *serviceinput.CreatePartnerBlock) (*models.PartnerBlock, error) {
	if s.createPartnerBlockFn != nil {
		return s.createPartnerBlockFn(ctx, partnerID, in)
	}
	return nil, nil
}

func (s *stubService) GetPartnerBlock(ctx context.Context, partnerID, siteID, blockID int) (*models.PartnerBlock, []*models.PartnerBlockGeoRule, error) {
	if s.getPartnerBlockFn != nil {
		return s.getPartnerBlockFn(ctx, partnerID, siteID, blockID)
	}
	return nil, nil, nil
}

func (s *stubService) ListPartnerBlocks(ctx context.Context, partnerID, siteID int) ([]*models.PartnerBlock, error) {
	if s.listPartnerBlocksFn != nil {
		return s.listPartnerBlocksFn(ctx, partnerID, siteID)
	}
	return nil, nil
}

func (s *stubService) UpdatePartnerBlockMeta(ctx context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockMeta) (*models.PartnerBlock, error) {
	if s.updatePartnerBlockMetaFn != nil {
		return s.updatePartnerBlockMetaFn(ctx, partnerID, siteID, in)
	}
	return nil, nil
}

func (s *stubService) UpdatePartnerBlockGeneralSettings(ctx context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockGeneral) (*models.PartnerBlock, error) {
	if s.updatePartnerBlockGeneralFn != nil {
		return s.updatePartnerBlockGeneralFn(ctx, partnerID, siteID, in)
	}
	return nil, nil
}

func (s *stubService) UpdatePartnerBlockGeographySettings(ctx context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockGeography) ([]*models.PartnerBlockGeoRule, error) {
	if s.updatePartnerBlockGeoFn != nil {
		return s.updatePartnerBlockGeoFn(ctx, partnerID, siteID, in)
	}
	return nil, nil
}

func (s *stubService) UpdatePartnerBlockSelfAdSettings(ctx context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockSelfAd) (*models.PartnerBlock, error) {
	if s.updatePartnerBlockSelfAdFn != nil {
		return s.updatePartnerBlockSelfAdFn(ctx, partnerID, siteID, in)
	}
	return nil, nil
}

func (s *stubService) DeletePartnerBlock(ctx context.Context, partnerID, siteID, blockID int) error {
	if s.deletePartnerBlockFn != nil {
		return s.deletePartnerBlockFn(ctx, partnerID, siteID, blockID)
	}
	return nil
}

func (s *stubService) GetPartnerBlockEmbedCode(ctx context.Context, partnerID, siteID, blockID int, baseURL string) (string, string, error) {
	if s.getPartnerBlockEmbedCodeFn != nil {
		return s.getPartnerBlockEmbedCodeFn(ctx, partnerID, siteID, blockID, baseURL)
	}
	return "", "", nil
}

func (s *stubService) ListPartnerCountries(ctx context.Context) []service.DictionaryItem { return nil }
func (s *stubService) ListPartnerRegistrationRegions(ctx context.Context, countryCode string) []service.DictionaryItem {
	return nil
}
func (s *stubService) ListPartnerCooperationForms(ctx context.Context) []service.DictionaryItem {
	return nil
}
func (s *stubService) ListPartnerPayoutCurrencies(ctx context.Context) []service.DictionaryItem {
	return nil
}
func (s *stubService) ListPartnerBlockTypes(ctx context.Context) []service.BlockTypeDictionaryItem {
	return nil
}
func (s *stubService) GetPartnerGeoTree(ctx context.Context) []*service.GeoTreeNode { return nil }

func newTestRouter(ac *stubAuthClient, svc Service) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	r.Use(middleware.CSRF(middleware.CSRFConfig{
		CookieName: "csrf_token",
		HeaderName: "X-CSRF-Token",
		SkipPaths:  []string{"/ad/request"},
	}))
	r.HandleFunc("/__ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)
	handlers.Register(r, NewAPI(APIConfig{
		AuthClient: ac,
		Service:    svc,
		CookieConfig: CookieConfig{
			Name:     testCookieName,
			Path:     "/",
			HTTPOnly: true,
		},
		VKIDConfig: VKIDConfig{
			ClientID:           54559987,
			RedirectURI:        "http://localhost:8000/advertisers/login/vk/callback",
			AuthDomain:         "id.vk.ru",
			Scope:              "email phone",
			DefaultRedirectURL: "http://localhost:8080/app",
			ErrorRedirectURL:   "http://localhost:8080/login",
		},
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

// createSessionCookie adds a session to the stub auth client and returns the cookie.
func createSessionCookie(t *testing.T, ac *stubAuthClient, advertiserID int) *http.Cookie {
	t.Helper()
	sessionID := "test-session-" + string(rune('0'+advertiserID))
	ac.addSession(sessionID, int64(advertiserID))
	return &http.Cookie{Name: testCookieName, Value: sessionID}
}

func TestRegister_OK(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)

	ac.registerFn = func(_ context.Context, email, phone, password string) (int64, string, int64, error) {
		if email != "a@a.test" || phone != "+70000000000" || password != "secret" {
			t.Fatalf("unexpected register args: email=%s phone=%s password=%s", email, phone, password)
		}
		ac.setCredentials(99, email, phone)
		return 99, "sess-abc", 9999999999, nil
	}

	body := `{"email":"a@a.test","phone":"+70000000000","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/advertisers/register", bytes.NewBufferString(body))
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
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)

	ac.loginFn = func(_ context.Context, identifier, password string) (int64, string, int64, error) {
		if identifier != "test@mail.com" {
			t.Fatalf("unexpected identifier: %s", identifier)
		}
		if password == "bad" {
			return 0, "", 0, errs.ErrInvalidCredentials
		}
		if password == "ok" {
			return 1, "sess-ok", 9999999999, nil
		}
		t.Fatalf("unexpected password: %s", password)
		return 0, "", 0, nil
	}
	ac.setCredentials(1, "test@mail.com", "9000000000")

	req := httptest.NewRequest(http.MethodPost, "/advertisers/login", bytes.NewBufferString(`{"identifier":"test@mail.com","password":"bad"}`))
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d body=%s", rr.Code, rr.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/advertisers/login", bytes.NewBufferString(`{"identifier":"test@mail.com","password":"ok"}`))
	req2.AddCookie(csrf)
	req2.Header.Set("X-CSRF-Token", csrf.Value)
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr2.Code, rr2.Body.String())
	}
}

func TestLoginVKID_CreatesProfileForFirstLogin(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)

	ac.loginVKIDFn = func(_ context.Context, code, deviceID, codeVerifier string) (int64, string, int64, error) {
		if code != "vk-code" || deviceID != "device-1" || codeVerifier != "verifier-1" {
			t.Fatalf("unexpected vk id payload: code=%s deviceID=%s codeVerifier=%s", code, deviceID, codeVerifier)
		}
		return 7, "vk-sess", 9999999999, nil
	}
	ac.setCredentials(7, "vk@example.com", "9000000001")

	svc.getAdvertiserByIDFn = func(_ context.Context, id int) (*models.Advertiser, error) {
		if id != 7 {
			t.Fatalf("unexpected advertiser id: %d", id)
		}
		return nil, errs.NotFoundError
	}
	svc.createAdvertiserProfileFn = func(_ context.Context, id int64, name, email string) error {
		if id != 7 || name != "" || email != "vk@example.com" {
			t.Fatalf("unexpected profile create payload: id=%d name=%q email=%q", id, name, email)
		}
		return nil
	}

	req := httptest.NewRequest(http.MethodPost, "/advertisers/login/vk", bytes.NewBufferString(`{"code":"vk-code","device_id":"device-1","code_verifier":"verifier-1"}`))
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

func TestBeginVKIDLogin_RedirectsToVK(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	req := httptest.NewRequest(http.MethodGet, "/advertisers/login/vk", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("expected 302 got %d body=%s", rr.Code, rr.Body.String())
	}

	location := rr.Result().Header.Get("Location")
	if !strings.HasPrefix(location, "https://id.vk.ru/authorize?") {
		t.Fatalf("unexpected redirect location: %s", location)
	}

	if len(rr.Result().Cookies()) == 0 {
		t.Fatalf("expected oauth cookies to be set")
	}
}

func TestLoginVKIDCallback_RedirectsBackAndSetsSession(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	ac.loginVKIDFn = func(_ context.Context, code, deviceID, codeVerifier string) (int64, string, int64, error) {
		if code != "vk-code" || deviceID != "device-1" || codeVerifier != "verifier-1" {
			t.Fatalf("unexpected vk callback payload: code=%s deviceID=%s codeVerifier=%s", code, deviceID, codeVerifier)
		}
		return 8, "vk-sess-callback", 9999999999, nil
	}
	ac.setCredentials(8, "vk2@example.com", "9000000002")

	svc.getAdvertiserByIDFn = func(_ context.Context, id int) (*models.Advertiser, error) {
		if id != 8 {
			t.Fatalf("unexpected advertiser id: %d", id)
		}
		return nil, errs.NotFoundError
	}
	svc.createAdvertiserProfileFn = func(_ context.Context, id int64, name, email string) error {
		if id != 8 || email != "vk2@example.com" {
			t.Fatalf("unexpected create advertiser profile payload: id=%d email=%s", id, email)
		}
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/advertisers/login/vk/callback?code=vk-code&device_id=device-1&state=state-1", nil)
	req.AddCookie(&http.Cookie{Name: vkidStateCookieName, Value: "state-1"})
	req.AddCookie(&http.Cookie{Name: vkidVerifierCookieName, Value: "verifier-1"})
	req.AddCookie(&http.Cookie{Name: vkidRedirectCookieName, Value: base64.RawURLEncoding.EncodeToString([]byte("http://localhost:8080/app"))})

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("expected 302 got %d body=%s", rr.Code, rr.Body.String())
	}
	if got := rr.Result().Header.Get("Location"); got != "http://localhost:8080/app" {
		t.Fatalf("unexpected redirect location: %s", got)
	}
	if rr.Result().Header.Get("Set-Cookie") == "" {
		t.Fatalf("expected session cookie")
	}
}

func TestMe_UnauthorizedAndOK(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)

	svc.getAdvertiserByIDFn = func(_ context.Context, id int) (*models.Advertiser, error) {
		if id != 1 {
			t.Fatalf("unexpected advertiser id: %d", id)
		}
		return &models.Advertiser{
			ID:      1,
			Name:    "Test",
			Balance: 100,
		}, nil
	}
	ac.setCredentials(1, "test@mail.com", "9000000000")

	req := httptest.NewRequest(http.MethodGet, "/advertisers/me", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d body=%s", rr.Code, rr.Body.String())
	}

	sess := createSessionCookie(t, ac, 1)

	req2 := httptest.NewRequest(http.MethodGet, "/advertisers/me", nil)
	req2.AddCookie(sess)
	req2.AddCookie(csrf)
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr2.Code, rr2.Body.String())
	}
}

func TestLogout_AlwaysOK(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)

	req := httptest.NewRequest(http.MethodPost, "/advertisers/logout", nil)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestBalance_GetAndTopUp(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

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

	getReq := httptest.NewRequest(http.MethodGet, "/advertisers/balance", nil)
	getReq.AddCookie(sess)
	getReq.AddCookie(csrf)
	getRR := httptest.NewRecorder()
	r.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", getRR.Code, getRR.Body.String())
	}

	topupReq := httptest.NewRequest(http.MethodPost, "/advertisers/balance/topup", bytes.NewBufferString(`{"amount":150}`))
	topupReq.AddCookie(sess)
	topupReq.AddCookie(csrf)
	topupReq.Header.Set("X-CSRF-Token", csrf.Value)
	topupRR := httptest.NewRecorder()
	r.ServeHTTP(topupRR, topupReq)
	if topupRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", topupRR.Code, topupRR.Body.String())
	}
}

func TestListAds_UnauthorizedAndEmptyList(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)

	svc.listAdsFn = func(_ context.Context, advertiserID, groupID int) ([]*models.Ad, error) {
		if advertiserID != 1 {
			t.Fatalf("unexpected advertiser id: %d", advertiserID)
		}
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

	sess := createSessionCookie(t, ac, 1)

	req2 := httptest.NewRequest(http.MethodGet, "/ad_campaigns/1/ad_groups/2/ads", nil)
	req2.AddCookie(sess)
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
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

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
