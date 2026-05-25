package v1

import (
	"net/http"
	"strconv"
	"time"

	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"

	"github.com/gorilla/mux"
)

func (a *API) RegisterStatsHandlers(r *mux.Router) {
	stats := r.PathPrefix("/ad_campaigns").Subrouter()

	stats.Use(middleware.Auth(a.authClient, a.cookieConfig.Name))
	stats.HandleFunc("/{ad_campaign_id}/stats", a.GetCampaignStats).Methods(http.MethodGet)
	stats.HandleFunc("/{ad_campaign_id}/ad_groups/{ad_group_id}/stats", a.GetGroupStats).Methods(http.MethodGet)
	stats.HandleFunc("/{ad_campaign_id}/ad_groups/{ad_group_id}/ads/{ad_id}/stats", a.GetAdStats).Methods(http.MethodGet)
}

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

	httpx.JSON(w, http.StatusOK, stats)
}

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

	httpx.JSON(w, http.StatusOK, stats)
}

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

	httpx.JSON(w, http.StatusOK, stats)
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
