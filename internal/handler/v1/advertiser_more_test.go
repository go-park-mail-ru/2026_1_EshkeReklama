package v1

import (
	"bytes"
	"eshkere/internal/handler"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"eshkere/internal/models"

	"go.uber.org/mock/gomock"
)

func TestAdvertiser_UpdateProfile_WithoutAvatar(t *testing.T) {
	sm := newTestSessionManager()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := handlers.NewMockService(ctrl)
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)
	sess := handlers.createSessionCookie(t, sm, 1)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("name", "New Name")
	_ = w.WriteField("email", "NEW@MAIL.TEST")
	_ = w.WriteField("phone", "+7 900 123-45-67")
	_ = w.Close()

	svc.EXPECT().
		UpdateAdvertiserProfile(gomock.Any(), 1, "New Name", "NEW@MAIL.TEST", "+7 900 123-45-67").
		DoAndReturn(func(_ any, _ int, name, email, phone string) (*models.Advertiser, error) {
			return &models.Advertiser{ID: 1, Name: name, Email: email, Phone: phone}, nil
		})

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := handlers.NewMockService(ctrl)
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)
	sess := handlers.createSessionCookie(t, sm, 1)

	svc.EXPECT().GenerateFeedLink(gomock.Any(), 1).Return("tok", nil)

	req := httptest.NewRequest(http.MethodPost, "/advertiser/feed-link", nil)
	req.AddCookie(sess)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
}
