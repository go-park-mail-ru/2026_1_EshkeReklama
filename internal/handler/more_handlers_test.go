package handler_test

import (
	"bytes"
	"context"
	errs "eshkere/internal/errors"
	handlers "eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"go.uber.org/mock/gomock"
)

const testCookieName = "session_id"

type stubAuthClient struct {
	sessions map[string]int64
}

func newStubAuthClient() *stubAuthClient {
	return &stubAuthClient{sessions: make(map[string]int64)}
}

func (c *stubAuthClient) addSession(id string, advID int64) { c.sessions[id] = advID }

func (c *stubAuthClient) Register(_ context.Context, _, _, _ string) (int64, string, int64, error) {
	return 0, "", 0, nil
}
func (c *stubAuthClient) Login(_ context.Context, _, _ string) (int64, string, int64, error) {
	return 0, "", 0, nil
}
func (c *stubAuthClient) LoginVKID(_ context.Context, _, _, _ string) (int64, string, int64, error) {
	return 0, "", 0, nil
}
func (c *stubAuthClient) ValidateSession(_ context.Context, sid string) (int64, error) {
	advID, ok := c.sessions[sid]
	if !ok {
		return 0, errs.ErrSessionNotFound
	}
	return advID, nil
}
func (c *stubAuthClient) Logout(_ context.Context, sid string) error {
	delete(c.sessions, sid)
	return nil
}
func (c *stubAuthClient) GetCredentials(_ context.Context, _ int64) (string, string, error) {
	return "", "", nil
}
func (c *stubAuthClient) UpdateCredentials(_ context.Context, _ int64, email, phone string) (string, string, error) {
	return email, phone, nil
}

func newTestRouter(ac *stubAuthClient, svc v1.Service) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	r.Use(middleware.CSRF(middleware.CSRFConfig{
		CookieName: "csrf_token",
		HeaderName: "X-CSRF-Token",
	}))
	r.HandleFunc("/__ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)
	handlers.Register(r, v1.NewAPI(v1.APIConfig{
		AuthClient: ac,
		Service:    svc,
		CookieConfig: v1.CookieConfig{
			Name:     testCookieName,
			Path:     "/",
			HTTPOnly: true,
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

func createSessionCookie(t *testing.T, ac *stubAuthClient, advertiserID int) *http.Cookie {
	t.Helper()
	sessionID := "test-session-" + string(rune('0'+advertiserID))
	ac.addSession(sessionID, int64(advertiserID))
	return &http.Cookie{Name: testCookieName, Value: sessionID}
}

func TestAdCampaign_CRUD(t *testing.T) {
	ac := newStubAuthClient()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := handlers.NewMockService(ctrl)
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	svc.EXPECT().
		CreateAdCampaign(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, in *serviceinput.CreateAdCampaign) (*models.AdCampaign, error) {
			if in.AdvertiserID != 1 || in.Name != "camp" {
				t.Fatalf("unexpected campaign input: %+v", in)
			}
			return &models.AdCampaign{ID: 42}, nil
		})

	createReq := httptest.NewRequest(http.MethodPost, "/ad_campaigns", bytes.NewBufferString(`{"name":"camp"}`))
	createReq.AddCookie(sess)
	createReq.AddCookie(csrf)
	createReq.Header.Set("X-CSRF-Token", csrf.Value)
	createRR := httptest.NewRecorder()
	r.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", createRR.Code, createRR.Body.String())
	}

	svc.EXPECT().
		ListAdCampaigns(gomock.Any(), 1).
		Return([]*models.AdCampaign{{ID: 1, AdvertiserID: 1, Status: models.AdStatusWorking, Name: "c"}}, nil)

	listReq := httptest.NewRequest(http.MethodGet, "/ad_campaigns", nil)
	listReq.AddCookie(sess)
	listReq.AddCookie(csrf)
	listRR := httptest.NewRecorder()
	r.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", listRR.Code, listRR.Body.String())
	}

	newName := "new"
	svc.EXPECT().
		UpdateAdCampaign(gomock.Any(), &serviceinput.UpdateAdCampaign{ID: 42, Name: &newName}).
		Return(nil)

	updReq := httptest.NewRequest(http.MethodPut, "/ad_campaigns/42", bytes.NewBufferString(`{"name":"new"}`))
	updReq.AddCookie(sess)
	updReq.AddCookie(csrf)
	updReq.Header.Set("X-CSRF-Token", csrf.Value)
	updRR := httptest.NewRecorder()
	r.ServeHTTP(updRR, updReq)
	if updRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", updRR.Code, updRR.Body.String())
	}

	svc.EXPECT().DeleteAdCampaign(gomock.Any(), 42).Return(nil)

	delReq := httptest.NewRequest(http.MethodDelete, "/ad_campaigns/42", nil)
	delReq.AddCookie(sess)
	delReq.AddCookie(csrf)
	delReq.Header.Set("X-CSRF-Token", csrf.Value)
	delRR := httptest.NewRecorder()
	r.ServeHTTP(delRR, delReq)
	if delRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", delRR.Code, delRR.Body.String())
	}
}

func TestAdGroup_And_Ads_CRUD(t *testing.T) {
	ac := newStubAuthClient()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := handlers.NewMockService(ctrl)
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	svc.EXPECT().
		CreateAdGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, in *serviceinput.CreateAdGroup) (*models.AdGroup, error) {
			if in.AdCampaignID != 10 || in.Name != "g" {
				t.Fatalf("unexpected group input: %+v", in)
			}
			return &models.AdGroup{ID: 5}, nil
		})

	createGroupReq := httptest.NewRequest(http.MethodPost, "/ad_campaigns/10/ad_groups", bytes.NewBufferString(`{"topic_id":1,"region_id":2,"name":"g","age_from":18,"age_to":25,"gender":"any"}`))
	createGroupReq.AddCookie(sess)
	createGroupReq.AddCookie(csrf)
	createGroupReq.Header.Set("X-CSRF-Token", csrf.Value)
	createGroupRR := httptest.NewRecorder()
	r.ServeHTTP(createGroupRR, createGroupReq)
	if createGroupRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", createGroupRR.Code, createGroupRR.Body.String())
	}

	svc.EXPECT().ListAdGroups(gomock.Any(), 10).Return([]*models.AdGroup{}, nil)
	listGroupReq := httptest.NewRequest(http.MethodGet, "/ad_campaigns/10/ad_groups", nil)
	listGroupReq.AddCookie(sess)
	listGroupReq.AddCookie(csrf)
	listGroupRR := httptest.NewRecorder()
	r.ServeHTTP(listGroupRR, listGroupReq)
	if listGroupRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", listGroupRR.Code, listGroupRR.Body.String())
	}

	svc.EXPECT().UpdateAdGroup(gomock.Any(), gomock.Any()).Return(nil)
	updGroupReq := httptest.NewRequest(http.MethodPut, "/ad_campaigns/10/ad_groups/5", bytes.NewBufferString(`{"name":"g2"}`))
	updGroupReq.AddCookie(sess)
	updGroupReq.AddCookie(csrf)
	updGroupReq.Header.Set("X-CSRF-Token", csrf.Value)
	updGroupRR := httptest.NewRecorder()
	r.ServeHTTP(updGroupRR, updGroupReq)
	if updGroupRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", updGroupRR.Code, updGroupRR.Body.String())
	}

	svc.EXPECT().DeleteAdGroup(gomock.Any(), 5).Return(nil)
	delGroupReq := httptest.NewRequest(http.MethodDelete, "/ad_campaigns/10/ad_groups/5", nil)
	delGroupReq.AddCookie(sess)
	delGroupReq.AddCookie(csrf)
	delGroupReq.Header.Set("X-CSRF-Token", csrf.Value)
	delGroupRR := httptest.NewRecorder()
	r.ServeHTTP(delGroupRR, delGroupReq)
	if delGroupRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", delGroupRR.Code, delGroupRR.Body.String())
	}

	svc.EXPECT().
		CreateAd(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, in *serviceinput.CreateAd) (*models.Ad, error) {
			if in.AdGroupID != 2 || in.Title != "t" {
				t.Fatalf("unexpected ad input: %+v", in)
			}
			return &models.Ad{ID: 9}, nil
		})

	var createAdBody bytes.Buffer
	createAdWriter := multipart.NewWriter(&createAdBody)
	_ = createAdWriter.WriteField("title", "t")
	_ = createAdWriter.WriteField("short_desc", "s")
	_ = createAdWriter.WriteField("target_url", "u")
	_ = createAdWriter.Close()
	createAdReq := httptest.NewRequest(http.MethodPost, "/ad_campaigns/1/ad_groups/2/ads", &createAdBody)
	createAdReq.Header.Set("Content-Type", createAdWriter.FormDataContentType())
	createAdReq.AddCookie(sess)
	createAdReq.AddCookie(csrf)
	createAdReq.Header.Set("X-CSRF-Token", csrf.Value)
	createAdRR := httptest.NewRecorder()
	r.ServeHTTP(createAdRR, createAdReq)
	if createAdRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", createAdRR.Code, createAdRR.Body.String())
	}

	svc.EXPECT().UpdateAd(gomock.Any(), gomock.Any()).Return(nil)
	var updAdBody bytes.Buffer
	updAdWriter := multipart.NewWriter(&updAdBody)
	_ = updAdWriter.WriteField("title", "t2")
	_ = updAdWriter.Close()
	updAdReq := httptest.NewRequest(http.MethodPut, "/ad_campaigns/1/ad_groups/2/ads/9", &updAdBody)
	updAdReq.Header.Set("Content-Type", updAdWriter.FormDataContentType())
	updAdReq.AddCookie(sess)
	updAdReq.AddCookie(csrf)
	updAdReq.Header.Set("X-CSRF-Token", csrf.Value)
	updAdRR := httptest.NewRecorder()
	r.ServeHTTP(updAdRR, updAdReq)
	if updAdRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", updAdRR.Code, updAdRR.Body.String())
	}

	svc.EXPECT().DeleteAd(gomock.Any(), 9).Return(nil)
	delAdReq := httptest.NewRequest(http.MethodDelete, "/ad_campaigns/1/ad_groups/2/ads/9", nil)
	delAdReq.AddCookie(sess)
	delAdReq.AddCookie(csrf)
	delAdReq.Header.Set("X-CSRF-Token", csrf.Value)
	delAdRR := httptest.NewRecorder()
	r.ServeHTTP(delAdRR, delAdReq)
	if delAdRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", delAdRR.Code, delAdRR.Body.String())
	}
}
