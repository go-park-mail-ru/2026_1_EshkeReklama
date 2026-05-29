package v1

import (
	"net/http"

	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"

	"github.com/gorilla/mux"
)

func (a *API) RegisterSubscriptionHandlers(r *mux.Router) {
	sub := r.PathPrefix("/subscription").Subrouter()
	sub.Use(middleware.Auth(a.authClient, a.cookieConfig.Name))

	sub.HandleFunc("", a.GetTariffInfo).Methods(http.MethodGet)
	sub.HandleFunc("", a.PurchaseProSubscription).Methods(http.MethodPost)
}

// GetTariffInfo возвращает информацию о текущем тарифе и лимитах.
// @Summary      Информация о тарифе
// @Tags         subscription
// @Produce      json
// @Success      200  {object}  service.TariffInfo
// @Failure      401  {object}  httpx.Error
// @Router       /api/subscription [get]
// @Security     CookieAuth
func (a *API) GetTariffInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}

	info, err := a.service.GetTariffInfo(ctx, advertiserID)
	if err != nil {
		handler.HandleError(w, r, "get tariff info", err)
		return
	}

	httpx.JSON(w, http.StatusOK, info)
}

// PurchaseProSubscription списывает стоимость Pro с баланса и активирует подписку.
// @Summary      Оформить Pro-подписку с баланса
// @Tags         subscription
// @Produce      json
// @Success      200  {object}  service.TariffInfo
// @Failure      401  {object}  httpx.Error
// @Failure      402  {object}  httpx.Error  "Недостаточно средств на балансе"
// @Router       /api/subscription [post]
// @Security     CookieAuth
func (a *API) PurchaseProSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}

	info, err := a.service.PurchaseProSubscription(ctx, advertiserID)
	if err != nil {
		handler.HandleError(w, r, "purchase pro subscription", err)
		return
	}

	httpx.JSON(w, http.StatusOK, info)
}
