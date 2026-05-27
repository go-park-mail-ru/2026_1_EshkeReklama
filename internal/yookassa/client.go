package yookassa

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ShopID     string
	SecretKey  string
	ReturnURL  string
	WebhookURL string
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	shopID     string
	secretKey  string
	returnURL  string
	webhookURL string
}

type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type Card struct {
	Last4    string `json:"last4"`
	CardType string `json:"card_type"`
}

type PaymentMethod struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Card Card   `json:"card"`
}

type Confirmation struct {
	Type            string `json:"type"`
	ConfirmationURL string `json:"confirmation_url"`
	ReturnURL       string `json:"return_url,omitempty"`
}

type Payment struct {
	ID            string            `json:"id"`
	Status        string            `json:"status"`
	Paid          bool              `json:"paid"`
	Amount        Amount            `json:"amount"`
	Confirmation  Confirmation      `json:"confirmation"`
	PaymentMethod PaymentMethod     `json:"payment_method"`
	Metadata      map[string]string `json:"metadata"`
}

type createPaymentRequest struct {
	Amount            Amount            `json:"amount"`
	Capture           bool              `json:"capture"`
	Confirmation      *Confirmation     `json:"confirmation,omitempty"`
	Description       string            `json:"description,omitempty"`
	SavePaymentMethod bool              `json:"save_payment_method,omitempty"`
	PaymentMethodID   string            `json:"payment_method_id,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
}

type webhookNotification struct {
	Event  string  `json:"event"`
	Object Payment `json:"object"`
}

func NewClient(cfg Config) *Client {
	return &Client{
		baseURL: "https://api.yookassa.ru/v3",
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		shopID:     cfg.ShopID,
		secretKey:  cfg.SecretKey,
		returnURL:  cfg.ReturnURL,
		webhookURL: cfg.WebhookURL,
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.shopID != "" && c.secretKey != ""
}

func (c *Client) CreateRedirectPayment(ctx context.Context, amountRub int64, description string, metadata map[string]string) (*Payment, error) {
	req := createPaymentRequest{
		Amount: Amount{
			Value:    formatRubles(amountRub),
			Currency: "RUB",
		},
		Capture: true,
		Confirmation: &Confirmation{
			Type:      "redirect",
			ReturnURL: c.returnURL,
		},
		Description: description,
		//SavePaymentMethod: true,
		Metadata: metadata,
	}

	var payment Payment
	if err := c.doJSON(ctx, http.MethodPost, "/payments", req, &payment); err != nil {
		return nil, err
	}

	return &payment, nil
}

func (c *Client) CreateAutopayPayment(ctx context.Context, amountRub int64, paymentMethodID string, description string, metadata map[string]string) (*Payment, error) {
	req := createPaymentRequest{
		Amount: Amount{
			Value:    formatRubles(amountRub),
			Currency: "RUB",
		},
		Capture:         true,
		PaymentMethodID: paymentMethodID,
		Description:     description,
		Metadata:        metadata,
	}

	var payment Payment
	if err := c.doJSON(ctx, http.MethodPost, "/payments", req, &payment); err != nil {
		return nil, err
	}

	return &payment, nil
}

func (c *Client) GetPayment(ctx context.Context, paymentID string) (*Payment, error) {
	var payment Payment
	if err := c.doJSON(ctx, http.MethodGet, "/payments/"+paymentID, nil, &payment); err != nil {
		return nil, err
	}
	return &payment, nil
}

func DecodeWebhook(body io.Reader) (*webhookNotification, error) {
	var notification webhookNotification
	if err := json.NewDecoder(body).Decode(&notification); err != nil {
		return nil, fmt.Errorf("decode yookassa webhook: %w", err)
	}
	return &notification, nil
}

func PaymentMethodTitle(paymentMethod PaymentMethod) string {
	last4 := strings.TrimSpace(paymentMethod.Card.Last4)
	if last4 == "" {
		return ""
	}

	brand := strings.TrimSpace(paymentMethod.Card.CardType)
	if brand == "" {
		brand = "Bank card"
	} else {
		brand = normalizeCardBrand(brand)
	}

	return fmt.Sprintf("%s •••• %s", brand, last4)
}

func AmountToRubles(amount Amount) (int64, error) {
	value := strings.TrimSpace(amount.Value)
	if value == "" {
		return 0, fmt.Errorf("empty amount value")
	}

	parts := strings.SplitN(value, ".", 2)
	rubles, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse rubles: %w", err)
	}
	return rubles, nil
}

func (c *Client) doJSON(ctx context.Context, method string, path string, payload any, out any) error {
	if !c.Enabled() {
		return fmt.Errorf("yookassa is not configured")
	}

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal yookassa payload: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("create yookassa request: %w", err)
	}
	req.SetBasicAuth(c.shopID, c.secretKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotence-Key", randomKey())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do yookassa request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyData, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("yookassa %s %s failed: status=%d body=%s", method, path, resp.StatusCode, strings.TrimSpace(string(bodyData)))
	}

	if out == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode yookassa response: %w", err)
	}
	return nil
}

func formatRubles(amountRub int64) string {
	return fmt.Sprintf("%d.00", amountRub)
}

func randomKey() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func normalizeCardBrand(brand string) string {
	switch strings.ToLower(brand) {
	case "mastercard":
		return "Mastercard"
	case "visa":
		return "Visa"
	case "mir":
		return "Mir"
	case "unionpay":
		return "UnionPay"
	case "jcb":
		return "JCB"
	default:
		return strings.ToUpper(brand[:1]) + strings.ToLower(brand[1:])
	}
}
