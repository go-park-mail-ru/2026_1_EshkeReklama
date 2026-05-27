package yookassa

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestPaymentMethodTitleAndAmountHelpers(t *testing.T) {
	if got := PaymentMethodTitle(PaymentMethod{}); got != "" {
		t.Fatalf("expected empty payment method title, got %q", got)
	}
	if got := PaymentMethodTitle(PaymentMethod{Card: Card{Last4: "1234", CardType: "mastercard"}}); got != "Mastercard •••• 1234" {
		t.Fatalf("unexpected normalized brand: %q", got)
	}
	if got := PaymentMethodTitle(PaymentMethod{Card: Card{Last4: "1234", CardType: "unknownBrand"}}); got != "Unknownbrand •••• 1234" {
		t.Fatalf("unexpected fallback brand: %q", got)
	}
	if got := PaymentMethodTitle(PaymentMethod{Card: Card{Last4: "5678"}}); got != "Bank card •••• 5678" {
		t.Fatalf("unexpected default brand: %q", got)
	}

	amount, err := AmountToRubles(Amount{Value: "1500.99"})
	if err != nil || amount != 1500 {
		t.Fatalf("AmountToRubles mismatch amount=%d err=%v", amount, err)
	}
	if _, err := AmountToRubles(Amount{Value: ""}); err == nil {
		t.Fatalf("expected empty amount error")
	}
	if formatRubles(42) != "42.00" {
		t.Fatalf("unexpected ruble format")
	}
	if normalizeCardBrand("visa") != "Visa" || normalizeCardBrand("JCB") != "JCB" {
		t.Fatalf("unexpected brand normalization")
	}
	if len(randomKey()) != 32 {
		t.Fatalf("expected 32-char random key")
	}
}

func TestDecodeWebhook(t *testing.T) {
	n, err := DecodeWebhook(strings.NewReader(`{"event":"payment.succeeded","object":{"id":"pay_1"}}`))
	if err != nil {
		t.Fatalf("DecodeWebhook: %v", err)
	}
	if n.Event != "payment.succeeded" || n.Object.ID != "pay_1" {
		t.Fatalf("unexpected notification: %+v", n)
	}
	if _, err := DecodeWebhook(strings.NewReader(`{`)); err == nil {
		t.Fatalf("expected decode error")
	}
}

func TestClientDoJSONAndPaymentFlows(t *testing.T) {
	var lastMethod, lastPath, lastAuth, lastIdempotence string
	var lastBody string
	client := NewClient(Config{ShopID: "shop", SecretKey: "secret", ReturnURL: "https://return"})
	client.baseURL = "https://mock.yookassa.test"
	client.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		lastMethod = r.Method
		lastPath = r.URL.Path
		lastAuth = r.Header.Get("Authorization")
		lastIdempotence = r.Header.Get("Idempotence-Key")
		if r.Body != nil {
			body, _ := io.ReadAll(r.Body)
			lastBody = string(body)
		} else {
			lastBody = ""
		}

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Request:    r,
			Body:       io.NopCloser(strings.NewReader("")),
		}

		switch r.URL.Path {
		case "/payments":
			resp.Header.Set("Content-Type", "application/json")
			resp.Body = io.NopCloser(strings.NewReader(`{"id":"pay_1","status":"pending","amount":{"value":"1500.00","currency":"RUB"},"confirmation":{"confirmation_url":"https://confirm"}}`))
		case "/payments/pay_1":
			resp.Header.Set("Content-Type", "application/json")
			resp.Body = io.NopCloser(strings.NewReader(`{"id":"pay_1","status":"succeeded","amount":{"value":"1500.00","currency":"RUB"}}`))
		default:
			resp.StatusCode = http.StatusBadRequest
			resp.Body = io.NopCloser(strings.NewReader("boom"))
		}
		return resp, nil
	})}

	payment, err := client.CreateRedirectPayment(context.Background(), 1500, "top up", map[string]string{"a": "b"})
	if err != nil {
		t.Fatalf("CreateRedirectPayment: %v", err)
	}
	if payment.ID != "pay_1" || lastMethod != http.MethodPost || lastPath != "/payments" {
		t.Fatalf("unexpected payment request/result: %+v method=%s path=%s", payment, lastMethod, lastPath)
	}
	if lastAuth == "" || lastIdempotence == "" || !strings.Contains(lastBody, `"return_url":"https://return"`) {
		t.Fatalf("expected auth, idempotence and return_url in request, got auth=%q key=%q body=%s", lastAuth, lastIdempotence, lastBody)
	}

	autopay, err := client.CreateAutopayPayment(context.Background(), 2300, "pm_1", "auto", nil)
	if err != nil {
		t.Fatalf("CreateAutopayPayment: %v", err)
	}
	if autopay.ID != "pay_1" || !strings.Contains(lastBody, `"payment_method_id":"pm_1"`) {
		t.Fatalf("unexpected autopay request body: %s", lastBody)
	}

	got, err := client.GetPayment(context.Background(), "pay_1")
	if err != nil {
		t.Fatalf("GetPayment: %v", err)
	}
	if got.Status != "succeeded" || lastMethod != http.MethodGet || lastPath != "/payments/pay_1" {
		t.Fatalf("unexpected payment lookup: %+v method=%s path=%s", got, lastMethod, lastPath)
	}
}

func TestClientDoJSONErrors(t *testing.T) {
	client := NewClient(Config{})
	if client.Enabled() {
		t.Fatalf("expected disabled client")
	}
	if _, err := client.GetPayment(context.Background(), "pay_1"); err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("expected not configured error, got %v", err)
	}

	client = NewClient(Config{ShopID: "shop", SecretKey: "secret"})
	client.baseURL = "https://mock.yookassa.test"
	client.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     make(http.Header),
			Request:    r,
			Body:       io.NopCloser(strings.NewReader("fail whale")),
		}, nil
	})}

	if _, err := client.GetPayment(context.Background(), "pay_1"); err == nil || !strings.Contains(err.Error(), "status=400") {
		t.Fatalf("expected status error, got %v", err)
	}
}
