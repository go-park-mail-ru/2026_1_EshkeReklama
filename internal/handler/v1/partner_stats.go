package v1

import (
	"net/http"

	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"

	"github.com/gorilla/mux"
)

func (a *API) RegisterPartnerStatsHandlers(r *mux.Router) {
	partnerIncome := r.PathPrefix("/partners/income").Subrouter()
	partnerIncome.Use(middleware.PartnerAuth(a.authClient, a.cookieConfig.Name))
	partnerIncome.HandleFunc("/stats", a.GetPartnerIncomeStats).Methods(http.MethodGet)
}

// GetPartnerIncomeStats возвращает статистику дохода партнера.
// @Summary      Статистика дохода партнера
// @Description  Возвращает показы, начисленное вознаграждение, eCPM и детализацию по дням, сайтам и рекламным блокам
// @Tags         partner_income
// @Produce      json
// @Param        from  query     string  false  "Дата начала периода в формате YYYY-MM-DD"
// @Param        to    query     string  false  "Дата конца периода в формате YYYY-MM-DD"
// @Success      200   {object}  dto.PartnerIncomeStatsResponse
// @Failure      400   {object}  httpx.Error
// @Failure      401   {object}  httpx.Error
// @Failure      500   {object}  httpx.Error
// @Router       /partners/income/stats [get]
// @Security     CookieAuth
func (a *API) GetPartnerIncomeStats(w http.ResponseWriter, r *http.Request) {
	partnerID, err := ctxutils.PartnerIDFromContext(r.Context())
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

	httpx.JSON(w, http.StatusOK, dto.ToPartnerIncomeStatsResponse(stats))
}
