package v1

import (
	"database/sql"
	"net"
	"net/http"
	"strings"

	"eshkere/internal/handler"
	"eshkere/internal/handler/v1/dto"
	"eshkere/internal/models"
	"eshkere/internal/yookassa"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"

	"github.com/gorilla/mux"
)

var yookassaWebhookIPNets = mustParseCIDRs([]string{
	"185.71.76.0/27",
	"185.71.77.0/27",
	"77.75.153.0/25",
	"77.75.156.11/32",
	"77.75.156.35/32",
	"77.75.154.128/25",
	"2a02:5180::/32",
})

func (a *API) RegisterPaymentHandlers(r *mux.Router) {
	r.HandleFunc("/webhook/yookassa", a.YookassaWebhook).Methods(http.MethodPost)
}

func (a *API) CreateBalancePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser id from ctx", err)
		return
	}

	req, err := newJSONRequest[dto.CreatePaymentRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	result, err := a.service.CreateBalancePayment(ctx, advertiserID, req.Amount)
	if err != nil {
		handler.HandleError(w, r, "creating balance payment", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.CreatePaymentResponse{PaymentURL: result.PaymentURL})
}

func (a *API) GetAutopaySettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser id from ctx", err)
		return
	}

	settings, err := a.service.GetAdvertiserAutopaySettings(ctx, advertiserID)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser autopay settings", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.AutopaySettingsResponse{
		Enabled:   settings.Enabled,
		Threshold: settings.ThresholdAmount,
		Limit:     settings.TopUpAmount,
	})
}

func (a *API) UpdateAutopaySettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser id from ctx", err)
		return
	}

	req, err := newJSONRequest[dto.AutopaySettingsRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	err = a.service.UpdateAdvertiserAutopaySettings(ctx, &models.AdvertiserAutopaySettings{
		AdvertiserID:    advertiserID,
		Enabled:         req.Enabled,
		ThresholdAmount: req.Threshold,
		TopUpAmount:     req.Limit,
	})
	if err != nil {
		handler.HandleError(w, r, "updating advertiser autopay settings", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.AutopaySettingsResponse{
		Enabled:   req.Enabled,
		Threshold: req.Threshold,
		Limit:     req.Limit,
	})
}

func (a *API) GetNotificationSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser id from ctx", err)
		return
	}

	settings, err := a.service.GetAdvertiserNotificationSettings(ctx, advertiserID)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser notification settings", err)
		return
	}

	httpx.JSON(w, http.StatusOK, toNotificationSettingsResponse(settings))
}

func (a *API) UpdateNotificationSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser id from ctx", err)
		return
	}

	req, err := newJSONRequest[dto.NotificationSettingsRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	settings := &models.AdvertiserNotificationSettings{
		AdvertiserID:      advertiserID,
		EmailEnabled:      req.EmailEnabled,
		TelegramEnabled:   req.TelegramEnabled,
		WarningThreshold:  req.WarningThreshold,
		CriticalThreshold: req.CriticalThreshold,
	}
	if chatID := strings.TrimSpace(req.TelegramChatID); chatID != "" {
		settings.TelegramChatID = sql.NullString{String: chatID, Valid: true}
	}

	err = a.service.UpdateAdvertiserNotificationSettings(ctx, settings)
	if err != nil {
		handler.HandleError(w, r, "updating advertiser notification settings", err)
		return
	}

	httpx.JSON(w, http.StatusOK, toNotificationSettingsResponse(settings))
}

func toNotificationSettingsResponse(settings *models.AdvertiserNotificationSettings) dto.NotificationSettingsResponse {
	resp := dto.NotificationSettingsResponse{
		EmailEnabled:      settings.EmailEnabled,
		TelegramEnabled:   settings.TelegramEnabled,
		WarningThreshold:  settings.WarningThreshold,
		CriticalThreshold: settings.CriticalThreshold,
	}
	if settings.TelegramChatID.Valid {
		resp.TelegramChatID = settings.TelegramChatID.String
	}
	return resp
}

func (a *API) YookassaWebhook(w http.ResponseWriter, r *http.Request) {
	if !isTrustedYookassaIP(requestIP(r)) {
		httpx.Forbidden(w, "forbidden")
		return
	}

	notification, err := yookassa.DecodeWebhook(r.Body)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}
	if notification.Event != "payment.succeeded" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if _, err := a.service.CompletePaymentByWebhook(r.Context(), notification.Object.ID); err != nil {
		handler.HandleError(w, r, "processing yookassa webhook", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func requestIP(r *http.Request) net.IP {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		value := strings.TrimSpace(r.Header.Get(header))
		if value == "" {
			continue
		}
		ipText := strings.TrimSpace(strings.Split(value, ",")[0])
		if ip := net.ParseIP(ipText); ip != nil {
			return ip
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return net.ParseIP(host)
	}
	return net.ParseIP(r.RemoteAddr)
}

func isTrustedYookassaIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	for _, network := range yookassaWebhookIPNets {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func mustParseCIDRs(values []string) []*net.IPNet {
	out := make([]*net.IPNet, 0, len(values))
	for _, value := range values {
		_, network, err := net.ParseCIDR(value)
		if err == nil {
			out = append(out, network)
		}
	}
	return out
}
