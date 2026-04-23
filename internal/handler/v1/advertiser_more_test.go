package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eshkere/internal/handler/v1/dto"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

func TestAdvertiser_UpdateProfile_WithoutAvatar(t *testing.T) {
	sm := newTestSessionManager()
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, sm, 1)

	body, err := json.Marshal(dto.UpdateAdvertiserProfileRequest{
		Name:  "New Name",
		Email: "NEW@MAIL.TEST",
		Phone: "+7 900 123-45-67",
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	svc.updateAdvertiserProfileFn = func(_ context.Context, in *serviceinput.UpdateAdvertiserProfile) (*models.Advertiser, error) {
		if in.AdvertiserID != 1 {
			t.Fatalf("unexpected advertiser id: %d", in.AdvertiserID)
		}
		return &models.Advertiser{ID: 1, Name: in.Name, Email: in.Email, Phone: in.Phone}, nil
	}

	req := httptest.NewRequest(http.MethodPut, "/advertiser/me", bytes.NewReader(body))
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
