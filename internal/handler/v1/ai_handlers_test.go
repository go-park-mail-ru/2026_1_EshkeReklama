package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	errs "eshkere/internal/errors"
	"eshkere/internal/service"
)

func TestGenerateAdText_OK(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{
		generateAdTextFn: func(_ context.Context, advertiserID int, in service.GenerateAdTextInput) (*service.GeneratedAdText, error) {
			if advertiserID != 7 {
				t.Fatalf("unexpected advertiser id: %d", advertiserID)
			}
			if in.ProductName != "CRM для фитнес-клуба" || in.HeadlineMaxLen != 60 || in.BodyMaxLen != 150 {
				t.Fatalf("unexpected input: %+v", in)
			}
			return &service.GeneratedAdText{
				Headline: "Автоматизируйте фитнес-клуб",
				Body:     "CRM поможет вести клиентов, записи и оплату в одном окне.",
			}, nil
		},
	}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	session := createSessionCookie(t, ac, 7)

	req := httptest.NewRequest(http.MethodPost, "/api/ai/ad-text", bytes.NewBufferString(`{
		"product_name":"CRM для фитнес-клуба",
		"product_description":"Сервис для расписания, абонементов и уведомлений",
		"tone":"professional",
		"headline_max_len":60,
		"body_max_len":150
	}`))
	req.AddCookie(csrf)
	req.AddCookie(session)
	req.Header.Set("X-CSRF-Token", csrf.Value)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Data service.GeneratedAdText `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.Headline == "" || resp.Data.Body == "" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestGenerateAdVariants_ProRequired(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{
		generateAdVariantsFn: func(_ context.Context, _ int, _ service.GenerateAdVariantsInput) (*service.GeneratedAdVariants, error) {
			return nil, errs.ErrProRequired
		},
	}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	session := createSessionCookie(t, ac, 4)

	req := httptest.NewRequest(http.MethodPost, "/api/ai/ad-variants", bytes.NewBufferString(`{
		"product_name":"Онлайн-школа",
		"product_description":"Курсы английского для взрослых",
		"tone":"friendly",
		"count":3,
		"headline_max_len":60,
		"body_max_len":150
	}`))
	req.AddCookie(csrf)
	req.AddCookie(session)
	req.Header.Set("X-CSRF-Token", csrf.Value)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusPaymentRequired {
		t.Fatalf("expected 402 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestGenerateAdImage_OK(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{
		generateAdImageFn: func(_ context.Context, advertiserID int, in service.GenerateAdImageInput) (*service.GeneratedAdImage, error) {
			if advertiserID != 9 || in.Prompt == "" {
				t.Fatalf("unexpected input: advertiserID=%d input=%+v", advertiserID, in)
			}
			return &service.GeneratedAdImage{
				Images: []service.GeneratedAdImageVariant{
					{ImageURL: "https://cdn.example.com/generated/ad-image-1.png"},
					{ImageURL: "https://cdn.example.com/generated/ad-image-2.png"},
					{ImageURL: "https://cdn.example.com/generated/ad-image-3.png"},
				},
			}, nil
		},
	}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	session := createSessionCookie(t, ac, 9)

	req := httptest.NewRequest(http.MethodPost, "/api/ai/ad-image", bytes.NewBufferString(`{
		"prompt":"Светлый баннер с ноутбуком и графиком роста",
		"style":"clean",
		"format":"feed",
		"generation_key":"draft-1"
	}`))
	req.AddCookie(csrf)
	req.AddCookie(session)
	req.Header.Set("X-CSRF-Token", csrf.Value)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
}
