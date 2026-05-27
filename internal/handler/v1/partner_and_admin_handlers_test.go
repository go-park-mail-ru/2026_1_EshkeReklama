package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

func TestPartnerHandlersAndPartnerSites(t *testing.T) {
	ac := newStubAuthClient()
	ac.registerFn = func(_ context.Context, email, phone, password string) (int64, string, int64, error) {
		return 12, "partner-session", time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC).Unix(), nil
	}
	ac.loginFn = func(_ context.Context, identifier, password string) (int64, string, int64, error) {
		return 12, "partner-login", time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC).Unix(), nil
	}
	ac.getCredentialsFn = func(_ context.Context, partnerID int64) (string, string, bool, error) {
		return "partner@test.dev", "+79990000000", true, nil
	}
	ac.updateCredentialsFn = func(_ context.Context, partnerID int64, email, phone string) (string, string, error) {
		return email, phone, nil
	}

	svc := &stubService{
		createPartnerProfileFn: func(_ context.Context, in *serviceinput.CreatePartnerProfile) error {
			if in.ID != 12 || in.CountryCode != "RU" {
				t.Fatalf("unexpected partner register input: %+v", in)
			}
			return nil
		},
		getPartnerByIDFn: func(_ context.Context, id int) (*models.Partner, error) {
			return &models.Partner{
				ID:                     id,
				LastName:               "Ivanov",
				FirstName:              "Ivan",
				MiddleName:             "Ivanovich",
				BirthDate:              time.Date(1990, 2, 3, 0, 0, 0, 0, time.UTC),
				CountryCode:            "RU",
				RegistrationRegionCode: "MSK",
				CooperationForm:        models.CooperationFormSelfEmployed,
				PayoutCurrency:         models.PayoutCurrencyRUB,
				Balance:                100,
				CreatedAt:              time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			}, nil
		},
		updatePartnerProfileFn: func(_ context.Context, in *serviceinput.UpdatePartnerProfile) (*models.Partner, error) {
			if in.PartnerID != 12 {
				t.Fatalf("unexpected partner update input: %+v", in)
			}
			return &models.Partner{
				ID:                     12,
				LastName:               "Petrov",
				FirstName:              "Ivan",
				MiddleName:             "Ivanovich",
				BirthDate:              time.Date(1990, 2, 3, 0, 0, 0, 0, time.UTC),
				CountryCode:            "RU",
				RegistrationRegionCode: "MSK",
				CooperationForm:        models.CooperationFormSelfEmployed,
				PayoutCurrency:         models.PayoutCurrencyRUB,
				Balance:                100,
				CreatedAt:              time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			}, nil
		},
		listPartnerSitesFn: func(_ context.Context, partnerID int) ([]*models.PartnerSite, error) {
			return []*models.PartnerSite{{ID: 1, PartnerID: partnerID, Domain: "site.test", SiteName: "Main", Status: models.PartnerSiteStatusActive, CreatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)}}, nil
		},
		createPartnerSiteFn: func(_ context.Context, in *serviceinput.CreatePartnerSite) (*models.PartnerSite, error) {
			return &models.PartnerSite{ID: 2, PartnerID: in.PartnerID, Domain: in.Domain, SiteName: in.SiteName, Status: models.PartnerSiteStatusDraft, CreatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)}, nil
		},
		getPartnerSiteFn: func(_ context.Context, siteID int) (*models.PartnerSite, error) {
			return &models.PartnerSite{ID: siteID, PartnerID: 12, Domain: "site.test", SiteName: "Main", Status: models.PartnerSiteStatusActive, CreatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)}, nil
		},
		updatePartnerSiteFn: func(_ context.Context, in *serviceinput.UpdatePartnerSite) (*models.PartnerSite, error) {
			return &models.PartnerSite{ID: in.ID, PartnerID: in.PartnerID, Domain: "updated.test", SiteName: "Updated", Status: models.PartnerSiteStatusBlocked, CreatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)}, nil
		},
		deletePartnerSiteFn: func(_ context.Context, partnerID, siteID int) error {
			if partnerID != 12 || siteID != 2 {
				t.Fatalf("unexpected delete site args: partner=%d site=%d", partnerID, siteID)
			}
			return nil
		},
	}
	r := newTestRouter(ac, svc)
	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 12)

	tests := []struct {
		method string
		path   string
		body   string
		want   string
		status int
	}{
		{method: http.MethodPost, path: "/partners/register", body: `{"last_name":"Ivanov","first_name":"Ivan","middle_name":"I","birth_date":"1990-02-03","email":"partner@test.dev","phone":"+79990000000","country_code":"RU","registration_region_code":"MSK","cooperation_form":"self_employed","payout_currency":"RUB","password":"secret1"}`, want: `"id":12`, status: http.StatusOK},
		{method: http.MethodPost, path: "/partners/login", body: `{"identifier":"partner@test.dev","password":"secret1"}`, want: `"email":"partner@test.dev"`, status: http.StatusOK},
		{method: http.MethodGet, path: "/partners/me", want: `"last_name":"Ivanov"`, status: http.StatusOK},
		{method: http.MethodPut, path: "/partners/me", body: `{"last_name":"Petrov","email":"new@test.dev"}`, want: `"last_name":"Petrov"`, status: http.StatusOK},
		{method: http.MethodGet, path: "/partners/sites", want: `"site_name":"Main"`, status: http.StatusOK},
		{method: http.MethodPost, path: "/partners/sites", body: `{"domain":"new.test","site_name":"New Site"}`, want: `"id":2`, status: http.StatusCreated},
		{method: http.MethodGet, path: "/partners/sites/2", want: `"domain":"site.test"`, status: http.StatusOK},
		{method: http.MethodPut, path: "/partners/sites/2", body: `{"domain":"updated.test","site_name":"Updated","status":"blocked"}`, want: `"status":"blocked"`, status: http.StatusOK},
		{method: http.MethodDelete, path: "/partners/sites/2", want: `"message":"deleted"`, status: http.StatusOK},
		{method: http.MethodPost, path: "/partners/logout", want: `"message":"logout ok"`, status: http.StatusOK},
	}

	for _, tc := range tests {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if strings.HasPrefix(tc.path, "/partners/me") || strings.HasPrefix(tc.path, "/partners/sites") || tc.path == "/partners/logout" {
			req.AddCookie(sess)
		}
		req.AddCookie(csrf)
		if tc.method != http.MethodGet {
			req.Header.Set("X-CSRF-Token", csrf.Value)
		}
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != tc.status {
			t.Fatalf("%s %s: expected %d got %d body=%s", tc.method, tc.path, tc.status, rr.Code, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), tc.want) {
			t.Fatalf("%s %s: body %s does not contain %s", tc.method, tc.path, rr.Body.String(), tc.want)
		}
	}
}

func TestAdminHandlers(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{
		getAdvertiserByIDFn: func(_ context.Context, id int) (*models.Advertiser, error) {
			return &models.Advertiser{ID: id, Role: models.AdvertiserRoleAdmin}, nil
		},
		getAdByIDFn: func(_ context.Context, adID int) (*models.Ad, error) {
			return &models.Ad{ID: adID, Title: "Ad title", Status: models.AdStatusModeration}, nil
		},
		updateAdModerationStatusFn: func(_ context.Context, in *serviceinput.UpdateAdStatus) error {
			if in.AdID != 9 {
				t.Fatalf("unexpected moderation update input: %+v", in)
			}
			return nil
		},
		listModerationAdsFn: func(_ context.Context) ([]*models.Ad, error) {
			return []*models.Ad{{ID: 9, Title: "Ad title", Status: models.AdStatusModeration}}, nil
		},
	}
	r := newTestRouter(ac, svc)
	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	tests := []struct {
		method string
		path   string
		body   string
		want   string
	}{
		{method: http.MethodGet, path: "/admin/ads/9", want: `"title":"Ad title"`},
		{method: http.MethodPatch, path: "/admin/ads/9/status", body: `{"status":"approve"}`, want: `{}`},
		{method: http.MethodGet, path: "/admin/ads", want: `"ads":[{"id":9`},
	}

	for _, tc := range tests {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.AddCookie(sess)
		req.AddCookie(csrf)
		if tc.method != http.MethodGet {
			req.Header.Set("X-CSRF-Token", csrf.Value)
		}
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s %s: expected 200 got %d body=%s", tc.method, tc.path, rr.Code, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), tc.want) {
			t.Fatalf("%s %s: body %s does not contain %s", tc.method, tc.path, rr.Body.String(), tc.want)
		}
	}
}
