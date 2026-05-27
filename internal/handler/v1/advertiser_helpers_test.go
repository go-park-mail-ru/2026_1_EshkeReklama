package v1

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	errs "eshkere/internal/errors"
	"eshkere/internal/handler/v1/dto"
	"eshkere/internal/models"
	svc "eshkere/internal/service"
	serviceinput "eshkere/internal/service/input"
)

func TestHasAdvertiserProfileChanges(t *testing.T) {
	if hasAdvertiserProfileChanges(&dto.UpdateAdvertiserProfileRequest{}) {
		t.Fatal("expected no profile changes for empty request")
	}

	name := "Igor"
	if !hasAdvertiserProfileChanges(&dto.UpdateAdvertiserProfileRequest{Name: &name}) {
		t.Fatal("expected profile changes when name is set")
	}
}

func TestResolveUpdatedContacts(t *testing.T) {
	ctx := context.Background()
	ac := newStubAuthClient()
	api := NewAPI(APIConfig{AuthClient: ac})

	ac.setCredentials(7, "old@mail.test", "+79990000000")

	email, phone, err := api.resolveUpdatedContacts(ctx, 7, &dto.UpdateAdvertiserProfileRequest{})
	if err != nil {
		t.Fatalf("resolve without changes: %v", err)
	}
	if email != "old@mail.test" || phone != "+79990000000" {
		t.Fatalf("unexpected original credentials: email=%q phone=%q", email, phone)
	}

	newEmail := "new@mail.test"
	newPhone := "+79991112233"
	email, phone, err = api.resolveUpdatedContacts(ctx, 7, &dto.UpdateAdvertiserProfileRequest{
		Email: &newEmail,
		Phone: &newPhone,
	})
	if err != nil {
		t.Fatalf("resolve with updates: %v", err)
	}
	if email != newEmail || phone != newPhone {
		t.Fatalf("unexpected updated credentials: email=%q phone=%q", email, phone)
	}

	ac.getCredentialsFn = func(context.Context, int64) (string, string, error) {
		return "", "", errors.New("auth unavailable")
	}
	if _, _, err = api.resolveUpdatedContacts(ctx, 7, &dto.UpdateAdvertiserProfileRequest{
		Email: &newEmail,
	}); err == nil {
		t.Fatal("expected credentials lookup error")
	}
}

func TestEnsureAdvertiserProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("existing profile", func(t *testing.T) {
		ac := newStubAuthClient()
		svcStub := &stubService{}
		api := NewAPI(APIConfig{AuthClient: ac, Service: svcStub})

		svcStub.getAdvertiserByIDFn = func(_ context.Context, id int) (*models.Advertiser, error) {
			if id != 11 {
				t.Fatalf("unexpected advertiser id: %d", id)
			}
			return &models.Advertiser{ID: 11}, nil
		}
		svcStub.createAdvertiserProfileFn = func(_ context.Context, _ int64, _, _ string) error {
			t.Fatal("create profile should not be called when profile already exists")
			return nil
		}

		if err := api.ensureAdvertiserProfile(ctx, 11, "Ivan", "Ivanov", "ivan@test.dev"); err != nil {
			t.Fatalf("ensure existing profile: %v", err)
		}
	})

	t.Run("create and update surname", func(t *testing.T) {
		ac := newStubAuthClient()
		svcStub := &stubService{}
		api := NewAPI(APIConfig{AuthClient: ac, Service: svcStub})

		createCalled := false
		updateCalled := false

		svcStub.getAdvertiserByIDFn = func(_ context.Context, id int) (*models.Advertiser, error) {
			if id != 12 {
				t.Fatalf("unexpected advertiser id: %d", id)
			}
			return nil, errs.NotFoundError
		}
		svcStub.createAdvertiserProfileFn = func(_ context.Context, id int64, name, email string) error {
			createCalled = true
			if id != 12 || name != "Petr" || email != "petr@test.dev" {
				t.Fatalf("unexpected create args: id=%d name=%q email=%q", id, name, email)
			}
			return nil
		}
		svcStub.updateAdvertiserProfileFn = func(_ context.Context, in *serviceinput.UpdateAdvertiserProfile) (*models.Advertiser, error) {
			updateCalled = true
			if in.AdvertiserID != 12 || in.Surname == nil || *in.Surname != "Petrov" {
				t.Fatalf("unexpected update input: %+v", in)
			}
			return &models.Advertiser{ID: 12}, nil
		}

		if err := api.ensureAdvertiserProfile(ctx, 12, "Petr", "Petrov", "petr@test.dev"); err != nil {
			t.Fatalf("ensure missing profile: %v", err)
		}
		if !createCalled {
			t.Fatal("expected create profile to be called")
		}
		if !updateCalled {
			t.Fatal("expected update profile to be called for surname")
		}
	})

	t.Run("unexpected lookup error", func(t *testing.T) {
		ac := newStubAuthClient()
		svcStub := &stubService{}
		api := NewAPI(APIConfig{AuthClient: ac, Service: svcStub})

		svcStub.getAdvertiserByIDFn = func(_ context.Context, _ int) (*models.Advertiser, error) {
			return nil, errors.New("db unavailable")
		}

		if err := api.ensureAdvertiserProfile(ctx, 13, "Sergey", "", "sergey@test.dev"); err == nil {
			t.Fatal("expected lookup error to be returned")
		}
	})
}

func TestDeliveryAlertResponse(t *testing.T) {
	tests := []struct {
		name    string
		alert   *svc.AdvertiserDeliveryAlert
		title   string
		message string
	}{
		{
			name:    "nil",
			alert:   nil,
			title:   "",
			message: "",
		},
		{
			name: "low balance",
			alert: &svc.AdvertiserDeliveryAlert{
				Level:             svc.DeliveryAlertLowBalance,
				ActiveCampaigns:   5,
				AffectedCampaigns: 1,
			},
			title:   "Баланс на исходе",
			message: "Показы идут, но средств мало. Пополните баланс заранее, чтобы не прерывать рекламу.",
		},
		{
			name: "at risk",
			alert: &svc.AdvertiserDeliveryAlert{
				Level:             svc.DeliveryAlertAtRisk,
				ActiveCampaigns:   4,
				AffectedCampaigns: 2,
			},
			title:   "Реклама под риском остановки",
			message: "Активные кампании еще работают, но при текущем балансе могут скоро остановиться.",
		},
		{
			name: "partially stopped",
			alert: &svc.AdvertiserDeliveryAlert{
				Level:             svc.DeliveryAlertPartiallyStopped,
				ActiveCampaigns:   3,
				AffectedCampaigns: 2,
			},
			title:   "Часть кампаний остановлена",
			message: "У части активных кампаний уже недостаточно средств для продолжения показов.",
		},
		{
			name: "fully stopped",
			alert: &svc.AdvertiserDeliveryAlert{
				Level:             svc.DeliveryAlertFullyStopped,
				ActiveCampaigns:   2,
				AffectedCampaigns: 2,
			},
			title:   "Реклама остановлена из-за нехватки средств",
			message: "У активных кампаний недостаточно средств. Пополните баланс, чтобы возобновить показы.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deliveryAlertResponse(tt.alert)
			if tt.alert == nil {
				if got != nil {
					t.Fatalf("expected nil response, got %+v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected non-nil response")
			}
			if got.Level != string(tt.alert.Level) {
				t.Fatalf("unexpected level: %q", got.Level)
			}
			if got.ActiveCampaigns != tt.alert.ActiveCampaigns || got.AffectedCampaigns != tt.alert.AffectedCampaigns {
				t.Fatalf("unexpected counters: %+v", got)
			}
			if got.Title != tt.title || got.Message != tt.message {
				t.Fatalf("unexpected text: title=%q message=%q", got.Title, got.Message)
			}
		})
	}
}

func TestSessionCookieHelpers(t *testing.T) {
	api := NewAPI(APIConfig{
		CookieConfig: CookieConfig{
			Name:     testCookieName,
			Path:     "/",
			HTTPOnly: true,
			Secure:   true,
		},
	})

	rr := httptest.NewRecorder()
	expiresAt := time.Now().Add(time.Hour)
	api.setSessionCookie(rr, "session-123", expiresAt)

	resp := rr.Result()
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie to be set")
	}
	if cookies[0].Name != testCookieName || cookies[0].Value != "session-123" {
		t.Fatalf("unexpected cookie: %+v", cookies[0])
	}
	if cookies[0].Path != "/" || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected cookie attrs: %+v", cookies[0])
	}

	clearRR := httptest.NewRecorder()
	api.clearSessionCookie(clearRR)
	clearCookies := clearRR.Result().Cookies()
	if len(clearCookies) == 0 {
		t.Fatal("expected clearing cookie to be set")
	}
	if clearCookies[0].Name != testCookieName || clearCookies[0].MaxAge != -1 {
		t.Fatalf("unexpected clearing cookie: %+v", clearCookies[0])
	}
}
