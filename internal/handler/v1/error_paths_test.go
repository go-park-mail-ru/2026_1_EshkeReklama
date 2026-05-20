package v1

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	errs "eshkere/internal/errors"

	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

func TestHandlers_BadRequests(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	// invalid ad_group_id (route var not int)
	req := httptest.NewRequest(http.MethodPost, "/ad_campaigns/1/ad_groups/zzz/ads", nil)
	req.AddCookie(sess)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d body=%s", rr.Code, rr.Body.String())
	}

	// invalid ad_campaign_id for ad group
	req2 := httptest.NewRequest(http.MethodPost, "/ad_campaigns/nope/ad_groups", bytes.NewBufferString(`{"topic_id":1,"region_id":2,"name":"g","age_from":18,"age_to":25,"gender":"any"}`))
	req2.AddCookie(sess)
	req2.AddCookie(csrf)
	req2.Header.Set("X-CSRF-Token", csrf.Value)
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d body=%s", rr2.Code, rr2.Body.String())
	}

	// invalid json for ad_campaigns create
	req3 := httptest.NewRequest(http.MethodPost, "/ad_campaigns", bytes.NewBufferString(`{"name":`))
	req3.AddCookie(sess)
	req3.AddCookie(csrf)
	req3.Header.Set("X-CSRF-Token", csrf.Value)
	rr3 := httptest.NewRecorder()
	r.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d body=%s", rr3.Code, rr3.Body.String())
	}

	// service error in create campaign -> 400
	svc.createAdCampaignFn = func(_ context.Context, _ *serviceinput.CreateAdCampaign) (*models.AdCampaign, error) {
		return nil, errs.BadRequestError
	}
	req4 := httptest.NewRequest(http.MethodPost, "/ad_campaigns", bytes.NewBufferString(`{"name":"camp"}`))
	req4.AddCookie(sess)
	req4.AddCookie(csrf)
	req4.Header.Set("X-CSRF-Token", csrf.Value)
	rr4 := httptest.NewRecorder()
	r.ServeHTTP(rr4, req4)
	if rr4.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d body=%s", rr4.Code, rr4.Body.String())
	}
}

func TestFeed_NotFound(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	svc.getAdsByFeedTokenFn = func(_ context.Context, token string) ([]*models.Ad, error) {
		if token != "missing" {
			t.Fatalf("unexpected token: %s", token)
		}
		return nil, errs.NotFoundError
	}

	req := httptest.NewRequest(http.MethodGet, "/feed/missing", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestFeed_InternalError(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	svc.getAdsByFeedTokenFn = func(_ context.Context, token string) ([]*models.Ad, error) {
		if token != "tok" {
			t.Fatalf("unexpected token: %s", token)
		}
		return nil, errors.New("db down")
	}

	req := httptest.NewRequest(http.MethodGet, "/feed/tok", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d body=%s", rr.Code, rr.Body.String())
	}
}
