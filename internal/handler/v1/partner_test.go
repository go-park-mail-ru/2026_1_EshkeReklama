package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"eshkere/internal/handler/v1/dto"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPartner_Register_Login_Me(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)

	ac.registerFn = func(_ context.Context, email, phone, password string) (int64, string, int64, error) {
		if email != "partner@example.com" || password != "secret123" {
			t.Fatalf("unexpected register args")
		}
		ac.setCredentials(44, "partner@example.com", "9001234567")
		return 44, "partner-sess", 9999999999, nil
	}
	svc.createPartnerProfileFn = func(_ context.Context, in *serviceinput.CreatePartnerProfile) error {
		if in.ID != 44 || in.LastName != "Иванов" || in.FirstName != "Иван" {
			t.Fatalf("unexpected partner profile input: %+v", in)
		}
		return nil
	}

	registerBody := `{"last_name":"Иванов","first_name":"Иван","middle_name":"Иванович","birth_date":"1997-09-12","email":"partner@example.com","phone":"+7 900 123-45-67","country_code":"RU","registration_region_code":"RU-MOW","cooperation_form":"self_employed","payout_currency":"RUB","password":"secret123"}`
	registerReq := httptest.NewRequest(http.MethodPost, "/partners/register", bytes.NewBufferString(registerBody))
	registerReq.AddCookie(csrf)
	registerReq.Header.Set("X-CSRF-Token", csrf.Value)
	registerRR := httptest.NewRecorder()
	r.ServeHTTP(registerRR, registerReq)
	if registerRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", registerRR.Code, registerRR.Body.String())
	}

	ac.addSession("partner-sess", 44)
	sess := &http.Cookie{Name: testCookieName, Value: "partner-sess"}
	svc.getPartnerByIDFn = func(_ context.Context, id int) (*models.Partner, error) {
		return &models.Partner{
			ID:                     id,
			LastName:               "Иванов",
			FirstName:              "Иван",
			MiddleName:             "Иванович",
			BirthDate:              time.Date(1997, 9, 12, 0, 0, 0, 0, time.UTC),
			CountryCode:            "RU",
			RegistrationRegionCode: "RU-MOW",
			CooperationForm:        models.CooperationFormSelfEmployed,
			PayoutCurrency:         models.PayoutCurrencyRUB,
			CreatedAt:              time.Unix(0, 0),
		}, nil
	}

	meReq := httptest.NewRequest(http.MethodGet, "/partners/me", nil)
	meReq.AddCookie(sess)
	meReq.AddCookie(csrf)
	meRR := httptest.NewRecorder()
	r.ServeHTTP(meRR, meReq)
	if meRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", meRR.Code, meRR.Body.String())
	}

	var meEnvelope struct {
		Data dto.PartnerProfileResponse `json:"data"`
	}
	if err := json.Unmarshal(meRR.Body.Bytes(), &meEnvelope); err != nil {
		t.Fatalf("unmarshal me: %v", err)
	}
	if meEnvelope.Data.Email != "partner@example.com" || meEnvelope.Data.LastName != "Иванов" {
		t.Fatalf("unexpected me response: %+v", meEnvelope.Data)
	}
}

func TestPartner_CreateSite_And_Block(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	ac.addSession("partner-site-sess", 44)
	sess := &http.Cookie{Name: testCookieName, Value: "partner-site-sess"}

	svc.createPartnerSiteFn = func(_ context.Context, in *serviceinput.CreatePartnerSite) (*models.PartnerSite, error) {
		if in.PartnerID != 44 || in.Domain != "example.com" {
			t.Fatalf("unexpected create site input: %+v", in)
		}
		return &models.PartnerSite{ID: 101, PartnerID: 44, Domain: "example.com", SiteName: "Example"}, nil
	}
	svc.createPartnerBlockFn = func(_ context.Context, partnerID int, in *serviceinput.CreatePartnerBlock) (*models.PartnerBlock, error) {
		if partnerID != 44 || in.PartnerSiteID != 101 || in.BlockType != models.PartnerBlockTypeBanner {
			t.Fatalf("unexpected create block input: partnerID=%d in=%+v", partnerID, in)
		}
		return &models.PartnerBlock{ID: 9001, PartnerSiteID: 101, Name: in.Name, BlockType: in.BlockType, Status: models.PartnerBlockStatusDraft}, nil
	}

	siteReq := httptest.NewRequest(http.MethodPost, "/partners/sites", bytes.NewBufferString(`{"domain":"example.com","site_name":"Example"}`))
	siteReq.AddCookie(sess)
	siteReq.AddCookie(csrf)
	siteReq.Header.Set("X-CSRF-Token", csrf.Value)
	siteRR := httptest.NewRecorder()
	r.ServeHTTP(siteRR, siteReq)
	if siteRR.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d body=%s", siteRR.Code, siteRR.Body.String())
	}

	blockReq := httptest.NewRequest(http.MethodPost, "/partners/sites/101/blocks", bytes.NewBufferString(`{"block_type":"banner","name":"Баннер (03.05.2026)"}`))
	blockReq.AddCookie(sess)
	blockReq.AddCookie(csrf)
	blockReq.Header.Set("X-CSRF-Token", csrf.Value)
	blockRR := httptest.NewRecorder()
	r.ServeHTTP(blockRR, blockReq)
	if blockRR.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d body=%s", blockRR.Code, blockRR.Body.String())
	}
}
