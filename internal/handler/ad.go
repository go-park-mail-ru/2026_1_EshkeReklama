package handlers

import (
	"eshkere/internal/handler/dto"
	"eshkere/internal/middleware"
	"eshkere/pkg/httpx"
	"eshkere/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (a *API) RegisterAdsHandlers(r *mux.Router) {
	adsGroup := r.PathPrefix("/ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id}/ads").Subrouter()

	adsGroup.Use(middleware.Auth(a.sessionManager))
	adsGroup.HandleFunc("", a.CreateAd).Methods(http.MethodPost)
	adsGroup.HandleFunc("/{ad_id}", a.UpdateAd).Methods(http.MethodPut)
	adsGroup.HandleFunc("", a.ListAds).Methods(http.MethodGet)
	adsGroup.HandleFunc("/{ad_id}", a.DeleteAd).Methods(http.MethodDelete)
}

// CreateAd создаёт объявление в группе.
// @Summary      Создание объявления
// @Tags         ads
// @Accept       json
// @Produce      json
// @Param        ad_campaign_id  path      int                 true  "ID рекламной кампании"
// @Param        ad_group_id     path      int                 true  "ID группы объявлений"
// @Param        body            body      dto.CreateAdRequest true  "Параметры объявления"
// @Success      200             {object}  dto.CreateAdResponse
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id}/ads [post]
// @Security     CookieAuth
func (a *API) CreateAd(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqLogger := logger.GetLoggerFromCtx(ctx)

	var req dto.CreateAdRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		reqLogger.Warnw("invalid create ad payload", "error", err.Error())
		httpx.BadRequest(w, "invalid request")
		return
	}

	ad, err := req.ToModel()
	if err != nil {
		reqLogger.Warnw("invalid create ad model", "error", err.Error())
		httpx.BadRequest(w, "invalid request")
		return
	}

	groupID, err := strconv.Atoi(mux.Vars(r)["ad_group_id"])
	if err != nil {
		reqLogger.Warnw("invalid ad_group_id", "error", err.Error())
		httpx.BadRequest(w, "invalid ad_group_id")
		return
	}
	ad.AdGroupID = groupID

	createdAd, err := a.service.CreateAd(ctx, ad)
	if err != nil {
		reqLogger.Warnw("create ad rejected", "error", err.Error(), "ad_group_id", groupID)
		httpx.BadRequest(w, err.Error())
		return
	}

	httpx.JSON(w, http.StatusOK, dto.CreateAdResponse{
		ID: createdAd.ID,
	})
}

// UpdateAd обновляет объявление.
// @Summary      Обновление объявления
// @Tags         ads
// @Accept       json
// @Produce      json
// @Param        ad_campaign_id  path      int                    true  "ID рекламной кампании"
// @Param        ad_group_id     path      int                    true  "ID группы объявлений"
// @Param        ad_id           path      int                    true  "ID объявления"
// @Param        body            body      dto.UpdateAdRequest    true  "Поля для обновления"
// @Success      200             {object}  httpx.Success
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id}/ads/{ad_id} [put]
// @Security     CookieAuth
func (a *API) UpdateAd(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqLogger := logger.GetLoggerFromCtx(ctx)

	adID, err := strconv.Atoi(mux.Vars(r)["ad_id"])
	if err != nil {
		reqLogger.Warnw("invalid ad_id", "error", err.Error())
		httpx.BadRequest(w, "invalid ad_id")
		return
	}

	var req dto.UpdateAdRequest
	if err = httpx.DecodeJSON(r, &req); err != nil {
		reqLogger.Warnw("invalid update ad payload", "error", err.Error(), "ad_id", adID)
		httpx.BadRequest(w, "invalid request")
		return
	}

	//ad, err := req.ToModel(adID)
	//if err != nil {
	//	httpx.InternalError(w, err.Error())
	//	return
	//}

	err = a.service.UpdateAd(ctx, adID, req)
	if err != nil {
		reqLogger.Errorw("failed to update ad", "error", err.Error(), "ad_id", adID)
		httpx.InternalError(w, "internal error")
		return
	}

	httpx.JSON(w, http.StatusOK, nil)
}

// ListAds возвращает объявления группы.
// @Summary      Список объявлений группы
// @Tags         ads
// @Produce      json
// @Param        ad_campaign_id  path  int  true  "ID рекламной кампании"
// @Param        ad_group_id     path  int  true  "ID группы объявлений"
// @Success      200             {object}  dto.ListAdsResponse
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id}/ads [get]
// @Security     CookieAuth
func (a *API) ListAds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqLogger := logger.GetLoggerFromCtx(ctx)

	groupID, err := strconv.Atoi(mux.Vars(r)["ad_group_id"])
	if err != nil {
		reqLogger.Warnw("invalid ad_group_id", "error", err.Error())
		httpx.BadRequest(w, "invalid ad_group_id")
		return
	}

	ads, err := a.service.ListAds(ctx, groupID)
	if err != nil {
		reqLogger.Errorw("failed to list ads", "error", err.Error(), "ad_group_id", groupID)
		httpx.InternalError(w, "internal error")
		return
	}

	adsResponse := make([]*dto.AdResponse, 0, len(ads))
	for _, ad := range ads {
		adsResponse = append(adsResponse, dto.ToAdResponse(ad))
	}

	httpx.JSON(w, http.StatusOK, dto.ListAdsResponse{
		GroupID: groupID,
		Ads:     adsResponse,
	})
}

// DeleteAd удаляет объявление.
// @Summary      Удаление объявления
// @Tags         ads
// @Produce      json
// @Param        ad_campaign_id  path  int  true  "ID рекламной кампании"
// @Param        ad_group_id     path  int  true  "ID группы объявлений"
// @Param        ad_id           path  int  true  "ID объявления"
// @Success      200             {object}  httpx.Success
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id}/ads/{ad_id} [delete]
// @Security     CookieAuth
func (a *API) DeleteAd(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqLogger := logger.GetLoggerFromCtx(ctx)

	adID, err := strconv.Atoi(mux.Vars(r)["ad_id"])
	if err != nil {
		reqLogger.Warnw("invalid ad_id", "error", err.Error())
		httpx.BadRequest(w, "invalid ad_id")
		return
	}

	err = a.service.DeleteAd(ctx, adID)
	if err != nil {
		reqLogger.Errorw("failed to delete ad", "error", err.Error(), "ad_id", adID)
		httpx.InternalError(w, "internal error")
		return
	}

	httpx.JSON(w, http.StatusOK, nil)
}
