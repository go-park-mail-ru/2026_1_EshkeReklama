package v1

import (
	"net/http"

	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"

	"github.com/gorilla/mux"
)

func (a *API) RegisterPartnerIncomeHandlers(r *mux.Router) {
	partnerIncome := r.PathPrefix("/partners/income").Subrouter()
	partnerIncome.Use(middleware.Auth(a.authClient, a.cookieConfig.Name))
	partnerIncome.HandleFunc("/stats", a.GetPartnerIncomeStats).Methods(http.MethodGet)
}

func (a *API) GetPartnerIncomeStats(w http.ResponseWriter, r *http.Request) {
	partnerID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "partner id from ctx", err)
		return
	}

	from, to, err := parseStatsPeriod(r)
	if err != nil {
		httpx.BadRequest(w, "invalid stats period")
		return
	}

	stats, err := a.service.GetPartnerIncomeStats(r.Context(), partnerID, from, to)
	if err != nil {
		handler.HandleError(w, r, "getting partner income stats", err)
		return
	}

	httpx.JSON(w, http.StatusOK, stats)
}
