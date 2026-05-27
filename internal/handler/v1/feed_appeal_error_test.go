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
)

func TestCreateFeed_ErrorPaths(t *testing.T) {
	t.Run("invalid campaign id", func(t *testing.T) {
		ac := newStubAuthClient()
		svc := &stubService{}
		r := newTestRouter(ac, svc)

		csrf := getCSRF(t, r)
		sess := createSessionCookie(t, ac, 1)

		req := httptest.NewRequest(http.MethodPost, "/ad_campaigns/nope/feed", nil)
		req.AddCookie(sess)
		req.AddCookie(csrf)
		req.Header.Set("X-CSRF-Token", csrf.Value)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("service error", func(t *testing.T) {
		ac := newStubAuthClient()
		svc := &stubService{}
		r := newTestRouter(ac, svc)

		csrf := getCSRF(t, r)
		sess := createSessionCookie(t, ac, 1)

		svc.generateFeedLinkFn = func(_ context.Context, campaignID int) (string, error) {
			if campaignID != 5 {
				t.Fatalf("unexpected campaign id: %d", campaignID)
			}
			return "", errors.New("feed generation failed")
		}

		req := httptest.NewRequest(http.MethodPost, "/ad_campaigns/5/feed", nil)
		req.AddCookie(sess)
		req.AddCookie(csrf)
		req.Header.Set("X-CSRF-Token", csrf.Value)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d body=%s", rr.Code, rr.Body.String())
		}
	})
}

func TestAppeal_ErrorPaths(t *testing.T) {
	t.Run("create invalid request", func(t *testing.T) {
		ac := newStubAuthClient()
		svc := &stubService{}
		r := newTestRouter(ac, svc)

		csrf := getCSRF(t, r)

		req := httptest.NewRequest(http.MethodPost, "/appeals", bytes.NewBufferString(`{"title":"oops"}`))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(csrf)
		req.Header.Set("X-CSRF-Token", csrf.Value)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("list appeals service error", func(t *testing.T) {
		ac := newStubAuthClient()
		svc := &stubService{}
		r := newTestRouter(ac, svc)

		csrf := getCSRF(t, r)
		sess := createSessionCookie(t, ac, 1)

		svc.listAppealsFn = func(_ context.Context, advertiserID int) ([]*models.Appeal, error) {
			if advertiserID != 1 {
				t.Fatalf("unexpected advertiser id: %d", advertiserID)
			}
			return nil, errors.New("list failed")
		}

		req := httptest.NewRequest(http.MethodGet, "/appeals", nil)
		req.AddCookie(sess)
		req.AddCookie(csrf)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("get appeal invalid id", func(t *testing.T) {
		ac := newStubAuthClient()
		svc := &stubService{}
		r := newTestRouter(ac, svc)

		csrf := getCSRF(t, r)
		sess := createSessionCookie(t, ac, 1)

		req := httptest.NewRequest(http.MethodGet, "/appeals/not-an-id", nil)
		req.AddCookie(sess)
		req.AddCookie(csrf)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("get appeal not found", func(t *testing.T) {
		ac := newStubAuthClient()
		svc := &stubService{}
		r := newTestRouter(ac, svc)

		csrf := getCSRF(t, r)
		sess := createSessionCookie(t, ac, 1)

		svc.getAppealByIDFn = func(_ context.Context, appealID int) (*models.Appeal, error) {
			if appealID != 42 {
				t.Fatalf("unexpected appeal id: %d", appealID)
			}
			return nil, errs.NotFoundError
		}

		req := httptest.NewRequest(http.MethodGet, "/appeals/42", nil)
		req.AddCookie(sess)
		req.AddCookie(csrf)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d body=%s", rr.Code, rr.Body.String())
		}
	})
}
