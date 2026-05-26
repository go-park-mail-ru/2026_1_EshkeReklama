package v1

import (
	"net/http"
	"strconv"
	"time"

	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"

	"github.com/gorilla/mux"
)

func (a *API) RegisterAdvertiserStatsHandlers(r *mux.Router) {
	stats := r.PathPrefix("/ad_campaigns").Subrouter()

	stats.Use(middleware.Auth(a.authClient, a.cookieConfig.Name))
	stats.HandleFunc("/{ad_campaign_id}/stats", a.GetCampaignStats).Methods(http.MethodGet)
	stats.HandleFunc("/{ad_campaign_id}/ad_groups/{ad_group_id}/stats", a.GetGroupStats).Methods(http.MethodGet)
	stats.HandleFunc("/{ad_campaign_id}/ad_groups/{ad_group_id}/ads/{ad_id}/stats", a.GetAdStats).Methods(http.MethodGet)
}

// GetCampaignStats возвращает статистику рекламной кампании.
// @Summary      Статистика кампании
// @Description  Возвращает показы, клики, CTR, расходы, CPC, вознаграждение партнерам, выручку платформы, динамику по дням и разбивки по группам/площадкам
// @Tags         stats
// @Produce      json
// @Param        ad_campaign_id  path      int     true   "ID кампании"
// @Param        from            query     string  false  "Дата начала периода в формате YYYY-MM-DD"
// @Param        to              query     string  false  "Дата конца периода в формате YYYY-MM-DD"
// @Success      200             {object}  dto.CampaignStatsResponse
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      404             {object}  httpx.Error
// @Failure      501             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id}/stats [get]
// @Security     CookieAuth
func (a *API) GetCampaignStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}
	campaignID, err := strconv.Atoi(mux.Vars(r)["ad_campaign_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing campaign id", err)
		return
	}
	from, to, err := parseStatsPeriod(r)
	if err != nil {
		httpx.BadRequest(w, "invalid stats period")
		return
	}

	stats, err := a.service.GetCampaignStats(ctx, advertiserID, campaignID, from, to)
	if err != nil {
		handler.HandleError(w, r, "getting campaign stats", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.ToCampaignStatsResponse(stats))
}

// GetGroupStats возвращает статистику группы объявлений.
// @Summary      Статистика группы объявлений
// @Description  Возвращает показатели группы, динамику по дням и разбивки по объявлениям/площадкам
// @Tags         stats
// @Produce      json
// @Param        ad_campaign_id  path      int     true   "ID кампании"
// @Param        ad_group_id     path      int     true   "ID группы объявлений"
// @Param        from            query     string  false  "Дата начала периода в формате YYYY-MM-DD"
// @Param        to              query     string  false  "Дата конца периода в формате YYYY-MM-DD"
// @Success      200             {object}  dto.GroupStatsResponse
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      404             {object}  httpx.Error
// @Failure      501             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id}/stats [get]
// @Security     CookieAuth
func (a *API) GetGroupStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}
	vars := mux.Vars(r)
	campaignID, err := strconv.Atoi(vars["ad_campaign_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing campaign id", err)
		return
	}
	groupID, err := strconv.Atoi(vars["ad_group_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing group id", err)
		return
	}
	from, to, err := parseStatsPeriod(r)
	if err != nil {
		httpx.BadRequest(w, "invalid stats period")
		return
	}

	stats, err := a.service.GetGroupStats(ctx, advertiserID, campaignID, groupID, from, to)
	if err != nil {
		handler.HandleError(w, r, "getting group stats", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.ToGroupStatsResponse(stats))
}

// GetAdStats возвращает статистику объявления.
// @Summary      Статистика объявления
// @Description  Возвращает показатели объявления, динамику по дням и разбивку по площадкам
// @Tags         stats
// @Produce      json
// @Param        ad_campaign_id  path      int     true   "ID кампании"
// @Param        ad_group_id     path      int     true   "ID группы объявлений"
// @Param        ad_id           path      int     true   "ID объявления"
// @Param        from            query     string  false  "Дата начала периода в формате YYYY-MM-DD"
// @Param        to              query     string  false  "Дата конца периода в формате YYYY-MM-DD"
// @Success      200             {object}  dto.AdStatsResponse
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      404             {object}  httpx.Error
// @Failure      501             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id}/ads/{ad_id}/stats [get]
// @Security     CookieAuth
func (a *API) GetAdStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}
	vars := mux.Vars(r)
	campaignID, err := strconv.Atoi(vars["ad_campaign_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing campaign id", err)
		return
	}
	groupID, err := strconv.Atoi(vars["ad_group_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing group id", err)
		return
	}
	adID, err := strconv.Atoi(vars["ad_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing ad id", err)
		return
	}
	from, to, err := parseStatsPeriod(r)
	if err != nil {
		httpx.BadRequest(w, "invalid stats period")
		return
	}

	stats, err := a.service.GetAdStats(ctx, advertiserID, campaignID, groupID, adID, from, to)
	if err != nil {
		handler.HandleError(w, r, "getting ad stats", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.ToAdStatsResponse(stats))
}

func parseStatsPeriod(r *http.Request) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	to := dateOnly(now)
	from := to.AddDate(0, 0, -29)

	query := r.URL.Query()
	if rawTo := query.Get("to"); rawTo != "" {
		parsed, err := time.Parse("2006-01-02", rawTo)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		to = parsed
	}
	if rawFrom := query.Get("from"); rawFrom != "" {
		parsed, err := time.Parse("2006-01-02", rawFrom)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		from = parsed
	}
	if from.After(to) {
		return time.Time{}, time.Time{}, strconv.ErrSyntax
	}
	return dateOnly(from), dateOnly(to), nil
}

func dateOnly(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
