package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"eshkere/internal/handler/v1/dto"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

func TestAdvertiser_UpdateProfile_WithoutAvatar(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	body, err := json.Marshal(dto.UpdateAdvertiserProfileRequest{
		Name:  &[]string{"New Name"}[0],
		Email: &[]string{"NEW@MAIL.TEST"}[0],
		Phone: &[]string{"+7 900 123-45-67"}[0],
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	svc.updateAdvertiserProfileFn = func(_ context.Context, in *serviceinput.UpdateAdvertiserProfile) (*models.Advertiser, error) {
		if in.AdvertiserID != 1 {
			t.Fatalf("unexpected advertiser id: %d", in.AdvertiserID)
		}
		name := ""
		if in.Name != nil {
			name = *in.Name
		}
		return &models.Advertiser{ID: 1, Name: name}, nil
	}
	ac.setCredentials(1, "NEW@MAIL.TEST", "+7 900 123-45-67")

	req := httptest.NewRequest(http.MethodPut, APIPrefix+"/advertisers/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sess)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestAdvertiser_UpdateProfile_IgnoresTariffField(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	svc.updateAdvertiserProfileFn = func(_ context.Context, in *serviceinput.UpdateAdvertiserProfile) (*models.Advertiser, error) {
		if in.AdvertiserID != 1 {
			t.Fatalf("unexpected advertiser id: %d", in.AdvertiserID)
		}
		return &models.Advertiser{ID: 1, Name: "Same", Tariff: models.TariffTypeBasic}, nil
	}

	req := httptest.NewRequest(http.MethodPut, APIPrefix+"/advertisers/me", bytes.NewReader([]byte(`{"tariff":"pro"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sess)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
	if body := rr.Body.String(); strings.Contains(body, `"tariff":"pro"`) {
		t.Fatalf("tariff must not be upgraded via profile, body=%s", body)
	}
}

func TestAdvertiser_GenerateFeedLink_OK(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	svc.generateFeedLinkFn = func(_ context.Context, campaignID int) (string, error) {
		if campaignID != 1 {
			t.Fatalf("unexpected campaign id: %d", campaignID)
		}
		return "tok", nil
	}

	req := httptest.NewRequest(http.MethodPost, APIPrefix+"/ad_campaigns/1/feed", nil)
	req.AddCookie(sess)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d body=%s", rr.Code, rr.Body.String())
	}
}
