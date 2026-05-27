package v1

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	"eshkere/internal/service"
)

func TestBalanceAutomationHandlers(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{
		createBalancePaymentFn: func(_ context.Context, advertiserID int, amount int64) (*service.BalancePaymentResult, error) {
			if advertiserID != 1 || amount != 1500 {
				t.Fatalf("unexpected payment input: advertiser=%d amount=%d", advertiserID, amount)
			}
			return &service.BalancePaymentResult{PaymentURL: "https://pay.example/redirect"}, nil
		},
		getAutopaySettingsFn: func(_ context.Context, advertiserID int) (*models.AdvertiserAutopaySettings, error) {
			if advertiserID != 1 {
				t.Fatalf("unexpected advertiser for autopay: %d", advertiserID)
			}
			return &models.AdvertiserAutopaySettings{AdvertiserID: 1, Enabled: true, ThresholdAmount: 700, TopUpAmount: 2300}, nil
		},
		updateAutopaySettingsFn: func(_ context.Context, settings *models.AdvertiserAutopaySettings) error {
			if !settings.Enabled || settings.ThresholdAmount != 900 || settings.TopUpAmount != 3000 {
				t.Fatalf("unexpected autopay settings: %+v", settings)
			}
			return nil
		},
		getNotificationSettingsFn: func(_ context.Context, advertiserID int) (*models.AdvertiserNotificationSettings, error) {
			if advertiserID != 1 {
				t.Fatalf("unexpected advertiser for notifications: %d", advertiserID)
			}
			return &models.AdvertiserNotificationSettings{AdvertiserID: 1, EmailEnabled: true, WarningThreshold: 400, CriticalThreshold: 100}, nil
		},
		updateNotificationSettingsFn: func(_ context.Context, settings *models.AdvertiserNotificationSettings) error {
			if !settings.EmailEnabled || settings.WarningThreshold != 500 || settings.CriticalThreshold != 200 {
				t.Fatalf("unexpected notification settings: %+v", settings)
			}
			return nil
		},
		completePaymentByWebhookFn: func(_ context.Context, paymentID string) (*service.WebhookResult, error) {
			if paymentID != "pay_123" {
				t.Fatalf("unexpected payment id: %s", paymentID)
			}
			return &service.WebhookResult{}, nil
		},
	}
	r := newTestRouter(ac, svc)
	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	tests := []struct {
		method string
		path   string
		body   string
		want   string
	}{
		{method: http.MethodPost, path: "/advertisers/balance/payment/create", body: `{"amount":1500}`, want: `"payment_url":"https://pay.example/redirect"`},
		{method: http.MethodGet, path: "/advertisers/balance/autopay", want: `"threshold":700`},
		{method: http.MethodPost, path: "/advertisers/balance/autopay", body: `{"enabled":true,"threshold":900,"limit":3000}`, want: `"limit":3000`},
		{method: http.MethodGet, path: "/advertisers/notification-settings", want: `"warning_threshold":400`},
		{method: http.MethodPut, path: "/advertisers/notification-settings", body: `{"email_enabled":true,"warning_threshold":500,"critical_threshold":200}`, want: `"critical_threshold":200`},
	}

	for _, tc := range tests {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.AddCookie(sess)
		req.AddCookie(csrf)
		if tc.method != http.MethodGet {
			req.Header.Set("X-CSRF-Token", csrf.Value)
		}
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s %s: expected 200 got %d body=%s", tc.method, tc.path, rr.Code, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), tc.want) {
			t.Fatalf("%s %s: body %s does not contain %s", tc.method, tc.path, rr.Body.String(), tc.want)
		}
	}
}

func TestBalanceAutomationHandlers_Errors(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{
		completePaymentByWebhookFn: func(_ context.Context, _ string) (*service.WebhookResult, error) {
			return nil, errs.NotImplementedError
		},
	}
	api := NewAPI(APIConfig{AuthClient: ac, Service: svc})

	forbiddenWebhook := httptest.NewRequest(http.MethodPost, "/api/webhook/yookassa", bytes.NewBufferString(`{"event":"payment.succeeded","object":{"id":"pay_123"}}`))
	forbiddenWebhook.RemoteAddr = "10.0.0.2:1234"
	forbiddenRR := httptest.NewRecorder()
	api.YookassaWebhook(forbiddenRR, forbiddenWebhook)
	if forbiddenRR.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got %d body=%s", forbiddenRR.Code, forbiddenRR.Body.String())
	}

	badWebhook := httptest.NewRequest(http.MethodPost, "/api/webhook/yookassa", bytes.NewBufferString(`{`))
	badWebhook.RemoteAddr = "185.71.76.1:1234"
	badWebhookRR := httptest.NewRecorder()
	api.YookassaWebhook(badWebhookRR, badWebhook)
	if badWebhookRR.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d body=%s", badWebhookRR.Code, badWebhookRR.Body.String())
	}

	ignoredWebhook := httptest.NewRequest(http.MethodPost, "/api/webhook/yookassa", bytes.NewBufferString(`{"event":"payment.waiting_for_capture","object":{"id":"pay_123"}}`))
	ignoredWebhook.RemoteAddr = "185.71.76.2:1234"
	ignoredRR := httptest.NewRecorder()
	api.YookassaWebhook(ignoredRR, ignoredWebhook)
	if ignoredRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", ignoredRR.Code, ignoredRR.Body.String())
	}

	failedWebhook := httptest.NewRequest(http.MethodPost, "/api/webhook/yookassa", bytes.NewBufferString(`{"event":"payment.succeeded","object":{"id":"pay_123"}}`))
	failedWebhook.RemoteAddr = "185.71.76.3:1234"
	failedRR := httptest.NewRecorder()
	api.YookassaWebhook(failedRR, failedWebhook)
	if failedRR.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501 got %d body=%s", failedRR.Code, failedRR.Body.String())
	}
}

func TestWebhookRequestIPAndTrustHelpers(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/webhook/yookassa", nil)
	req.Header.Set("X-Forwarded-For", "185.71.76.5, 10.0.0.1")
	if got := requestIP(req).String(); got != "185.71.76.5" {
		t.Fatalf("unexpected forwarded ip: %s", got)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/webhook/yookassa", nil)
	req2.Header.Set("X-Real-IP", "77.75.156.11")
	if got := requestIP(req2).String(); got != "77.75.156.11" {
		t.Fatalf("unexpected real ip: %s", got)
	}

	req3 := httptest.NewRequest(http.MethodPost, "/api/webhook/yookassa", nil)
	req3.RemoteAddr = "2a02:5180::1:443"
	if requestIP(req3) == nil {
		t.Fatalf("expected remote addr ip")
	}

	if !isTrustedYookassaIP(requestIP(req)) {
		t.Fatalf("expected trusted yookassa ip")
	}
	if isTrustedYookassaIP(nil) {
		t.Fatalf("nil ip must not be trusted")
	}
	if len(mustParseCIDRs([]string{"bad", "185.71.76.0/27"})) != 1 {
		t.Fatalf("expected invalid cidr to be skipped")
	}
}
