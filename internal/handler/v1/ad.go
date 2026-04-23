package v1

import (
	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/pkg/httpx"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (a *API) RegisterAdsHandlers(r *mux.Router) {
	adsGroup := r.PathPrefix("/ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id}/ads").Subrouter()

	adsGroup.Use(middleware.Auth(a.sessionManager))
	adsGroup.HandleFunc("", a.CreateAd).Methods(http.MethodPost)
	adsGroup.HandleFunc("", a.ListAds).Methods(http.MethodGet)
	adsGroup.HandleFunc("/{ad_id}", a.UpdateAd).Methods(http.MethodPut)
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

	req, err := newJSONRequest[dto.CreateAdRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	groupID, err := strconv.Atoi(mux.Vars(r)["ad_group_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing group id", err)
		return
	}

	createdAd, err := a.service.CreateAd(ctx, req.ToInput(groupID))
	if err != nil {
		handler.HandleError(w, r, "creating ad", err)
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

	adID, err := strconv.Atoi(mux.Vars(r)["ad_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing ad id", err)
		return
	}

	req, err := newJSONRequest[dto.UpdateAdRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	err = a.service.UpdateAd(ctx, req.ToInput(adID))
	if err != nil {
		handler.HandleError(w, r, "updating id", err)
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

	groupID, err := strconv.Atoi(mux.Vars(r)["ad_group_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing group id", err)
		return
	}

	ads, err := a.service.ListAds(ctx, groupID)
	if err != nil {
		handler.HandleError(w, r, "listing ads", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.ToListAdsResponse(groupID, ads))
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

	adID, err := strconv.Atoi(mux.Vars(r)["ad_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing ad id", err)
		return
	}

	err = a.service.DeleteAd(ctx, adID)
	if err != nil {
		handler.HandleError(w, r, "deleting ad", err)
		return
	}

	httpx.JSON(w, http.StatusOK, nil)
}
