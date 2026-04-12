package handlers

import (
	"eshkere/internal/handler/dto"
	"eshkere/internal/middleware"
	"eshkere/pkg/httpx"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (a *API) RegisterAdCampaignHandlers(r *mux.Router) {
	campaigns := r.PathPrefix("/ad_campaigns").Subrouter()

	campaigns.Use(middleware.Auth(a.sessionManager))
	campaigns.HandleFunc("", a.CreateAdCampaign).Methods(http.MethodPost)
	campaigns.HandleFunc("", a.ListAdCampaigns).Methods(http.MethodGet)
	campaigns.HandleFunc("/{ad_campaign_id}", a.UpdateAdCampaign).Methods(http.MethodPut)
	campaigns.HandleFunc("/{ad_campaign_id}", a.DeleteAdCampaign).Methods(http.MethodDelete)
}

// CreateAdCampaign создаёт рекламную кампанию для текущего рекламодателя (ID из сессии).
// @Summary      Создание рекламной кампании
// @Description  Создаёт кампанию; рекламодатель определяется по сессии
// @Tags         ad_campaigns
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateAdCampaignRequest  true  "Параметры кампании"
// @Success      200   {object}  dto.CreateAdCampaignResponse
// @Failure      400   {object}  httpx.Error
// @Failure      401   {object}  httpx.Error
// @Failure      500   {object}  httpx.Error
// @Router       /ad_campaigns [post]
// @Security     CookieAuth
func (a *API) CreateAdCampaign(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := middleware.AdvertiserIDFromContext(ctx)
	if err != nil {
		httpx.Unauthorized(w, "unauthorized")
		return
	}

	var req dto.CreateAdCampaignRequest
	if err = httpx.DecodeJSON(r, &req); err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	campaign := req.ToModel(advertiserID)
	created, err := a.service.CreateAdCampaign(ctx, campaign)
	if err != nil {
		httpx.BadRequest(w, err.Error())
		return
	}

	httpx.JSON(w, http.StatusOK, dto.CreateAdCampaignResponse{
		ID: created.ID,
	})
}

// UpdateAdCampaign обновляет рекламную кампанию.
// @Summary      Обновление рекламной кампании
// @Description  Частичное обновление полей кампании
// @Tags         ad_campaigns
// @Accept       json
// @Produce      json
// @Param        ad_campaign_id  path      int                         true  "ID кампании"
// @Param        body            body      dto.UpdateAdCampaignRequest true  "Поля для обновления"
// @Success      200             {object}  httpx.Success
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id} [put]
// @Security     CookieAuth
func (a *API) UpdateAdCampaign(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	campaignID, err := strconv.Atoi(mux.Vars(r)["ad_campaign_id"])
	if err != nil {
		httpx.BadRequest(w, "invalid ad_campaign_id")
		return
	}

	var req dto.UpdateAdCampaignRequest
	if err = httpx.DecodeJSON(r, &req); err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	if err = a.service.UpdateAdCampaign(ctx, campaignID, req); err != nil {
		httpx.InternalError(w, err.Error())
		return
	}

	httpx.JSON(w, http.StatusOK, nil)
}

// ListAdCampaigns возвращает список кампаний текущего рекламодателя.
// @Summary      Список рекламных кампаний
// @Description  Кампании рекламодателя из сессии
// @Tags         ad_campaigns
// @Produce      json
// @Success      200   {object}  dto.ListAdCampaignsResponse
// @Failure      401   {object}  httpx.Error
// @Failure      500   {object}  httpx.Error
// @Router       /ad_campaigns [get]
// @Security     CookieAuth
func (a *API) ListAdCampaigns(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := middleware.AdvertiserIDFromContext(ctx)
	if err != nil {
		httpx.Unauthorized(w, "unauthorized")
		return
	}

	campaigns, err := a.service.ListAdCampaigns(ctx, advertiserID)
	if err != nil {
		httpx.InternalError(w, err.Error())
		return
	}

	out := make([]*dto.AdCampaignResponse, 0, len(campaigns))
	for _, c := range campaigns {
		out = append(out, dto.ToAdCampaignResponse(c))
	}

	httpx.JSON(w, http.StatusOK, dto.ListAdCampaignsResponse{
		AdvertiserID: advertiserID,
		Campaigns:    out,
	})
}

// DeleteAdCampaign удаляет рекламную кампанию.
// @Summary      Удаление рекламной кампании
// @Tags         ad_campaigns
// @Produce      json
// @Param        ad_campaign_id  path  int  true  "ID кампании"
// @Success      200             {object}  httpx.Success
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id} [delete]
// @Security     CookieAuth
func (a *API) DeleteAdCampaign(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	campaignID, err := strconv.Atoi(mux.Vars(r)["ad_campaign_id"])
	if err != nil {
		httpx.BadRequest(w, "invalid ad_campaign_id")
		return
	}

	if err = a.service.DeleteAdCampaign(ctx, campaignID); err != nil {
		httpx.InternalError(w, err.Error())
		return
	}

	httpx.JSON(w, http.StatusOK, nil)
}
