package v1

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"eshkere/internal/models"
)

func TestAdvertiser_UpdateProfile_WithoutAvatar(t *testing.T) {
	sm := newTestSessionManager()
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, sm, 1)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("name", "New Name")
	_ = w.WriteField("email", "NEW@MAIL.TEST")
	_ = w.WriteField("phone", "+7 900 123-45-67")
	_ = w.Close()

	svc.updateAdvertiserProfileFn = func(_ context.Context, advertiserID int, name, email, phone string) (*models.Advertiser, error) {
		if advertiserID != 1 {
			t.Fatalf("unexpected advertiser id: %d", advertiserID)
		}
		return &models.Advertiser{ID: 1, Name: name, Email: email, Phone: phone}, nil
	}

	req := httptest.NewRequest(http.MethodPut, "/advertiser/me", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.AddCookie(sess)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestAdvertiser_GenerateFeedLink_OK(t *testing.T) {
	sm := newTestSessionManager()
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, sm, 1)

	svc.generateFeedLinkFn = func(_ context.Context, campaignID int) (string, error) {
		if campaignID != 1 {
			t.Fatalf("unexpected campaign id: %d", campaignID)
		}
		return "tok", nil
	}

	req := httptest.NewRequest(http.MethodPost, "/ad_campaigns/1/feed", nil)
	req.AddCookie(sess)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d body=%s", rr.Code, rr.Body.String())
	}
}
