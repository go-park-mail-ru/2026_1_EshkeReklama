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

func TestAdCampaignStatus_InvalidStatus(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	req := httptest.NewRequest(http.MethodPatch, "/ad_campaigns/5/status", bytes.NewBufferString(`{"status":"working"}`))
	req.AddCookie(sess)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestAdCampaignStatus_ServiceError(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	svc.turnOffAdCampaignFn = func(_ context.Context, advertiserID, campaignID int) error {
		if advertiserID != 1 || campaignID != 5 {
			t.Fatalf("unexpected ids: advertiser=%d campaign=%d", advertiserID, campaignID)
		}
		return errs.NotFoundError
	}

	req := httptest.NewRequest(http.MethodPatch, "/ad_campaigns/5/status", bytes.NewBufferString(`{"status":"turned_off"}`))
	req.AddCookie(sess)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestAdGroupCreate_ServiceErrorWithoutRollback(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	svc.createAdGroupFn = func(_ context.Context, advertiserID int, in *serviceinput.CreateAdGroup) (*models.AdGroup, error) {
		if advertiserID != 1 {
			t.Fatalf("unexpected advertiser id: %d", advertiserID)
		}
		if in.AdCampaignID != 10 || in.Topic != "Авто" || in.Region != "Москва" {
			t.Fatalf("unexpected input: %+v", in)
		}
		return nil, errs.BadRequestError
	}
	svc.deleteAdCampaignFn = func(_ context.Context, advertiserID, campaignID int) error {
		t.Fatalf("delete campaign should not be called without rollback flag, got advertiser=%d campaign=%d", advertiserID, campaignID)
		return nil
	}

	req := httptest.NewRequest(http.MethodPost, "/ad_campaigns/10/ad_groups", bytes.NewBufferString(`{"topic":"Авто","region":"Москва","name":"g","age_from":18,"age_to":25,"gender":"any"}`))
	req.AddCookie(sess)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestAdGroupCreate_RollbackDeleteFailure(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	createCalled := false
	deleteCalled := false

	svc.createAdGroupFn = func(_ context.Context, advertiserID int, in *serviceinput.CreateAdGroup) (*models.AdGroup, error) {
		createCalled = true
		if advertiserID != 1 || in.AdCampaignID != 10 {
			t.Fatalf("unexpected input: advertiser=%d campaign=%d", advertiserID, in.AdCampaignID)
		}
		return nil, errs.BadRequestError
	}
	svc.deleteAdCampaignFn = func(_ context.Context, advertiserID, campaignID int) error {
		deleteCalled = true
		if advertiserID != 1 || campaignID != 10 {
			t.Fatalf("unexpected ids: advertiser=%d campaign=%d", advertiserID, campaignID)
		}
		return errors.New("rollback failed")
	}

	req := httptest.NewRequest(http.MethodPost, "/ad_campaigns/10/ad_groups?rollback_campaign_on_error=true", bytes.NewBufferString(`{"topic":"Авто","region":"Москва","name":"g","age_from":18,"age_to":25,"gender":"any"}`))
	req.AddCookie(sess)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d body=%s", rr.Code, rr.Body.String())
	}
	if !createCalled {
		t.Fatal("expected createAdGroup to be called")
	}
	if !deleteCalled {
		t.Fatal("expected deleteAdCampaign to be called")
	}
}

func TestAdCampaignDelete_ServiceError(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	svc.deleteAdCampaignFn = func(_ context.Context, advertiserID, campaignID int) error {
		if advertiserID != 1 || campaignID != 99 {
			t.Fatalf("unexpected ids: advertiser=%d campaign=%d", advertiserID, campaignID)
		}
		return errs.NotFoundError
	}

	req := httptest.NewRequest(http.MethodDelete, "/ad_campaigns/99", nil)
	req.AddCookie(sess)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d body=%s", rr.Code, rr.Body.String())
	}
}
