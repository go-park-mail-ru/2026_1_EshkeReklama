package v1

import (
	"net/http"
	"strconv"

	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"

	"github.com/gorilla/mux"
)

const rollbackCampaignOnErrorQueryParam = "rollback_campaign_on_error"

func (a *API) RegisterAdGroupHandlers(r *mux.Router) {
	groups := r.PathPrefix("/ad_campaigns/{ad_campaign_id}/ad_groups").Subrouter()

	groups.Use(middleware.Auth(a.authClient, a.cookieConfig.Name))
	groups.HandleFunc("", a.CreateAdGroup).Methods(http.MethodPost)
	groups.HandleFunc("", a.ListAdGroups).Methods(http.MethodGet)
	groups.HandleFunc("/{ad_group_id}", a.UpdateAdGroup).Methods(http.MethodPut)
	groups.HandleFunc("/{ad_group_id}", a.DeleteAdGroup).Methods(http.MethodDelete)
}

// CreateAdGroup создаёт группу объявлений в кампании.
// @Summary      Создание группы объявлений
// @Tags         ad_groups
// @Accept       json
// @Produce      json
// @Param        ad_campaign_id  path      int                      true  "ID рекламной кампании"
// @Param        body            body      dto.CreateAdGroupRequest true  "Параметры группы"
// @Success      200             {object}  dto.CreateAdGroupResponse
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      404             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id}/ad_groups [post]
// @Security     CookieAuth
func (a *API) CreateAdGroup(w http.ResponseWriter, r *http.Request) {
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

	req, err := newJSONRequest[dto.CreateAdGroupRequest](r)
	if err != nil {
		a.rollbackCampaignOnGroupCreateError(w, r, advertiserID, campaignID)
		httpx.BadRequest(w, "invalid request")
		return
	}

	created, err := a.service.CreateAdGroup(ctx, advertiserID, req.ToInput(campaignID))
	if err != nil {
		if rollbackErr := a.rollbackCampaignOnGroupCreateError(w, r, advertiserID, campaignID); rollbackErr != nil {
			return
		}
		handler.HandleError(w, r, "creating group", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.CreateAdGroupResponse{
		ID: created.ID,
	})
}

func (a *API) rollbackCampaignOnGroupCreateError(w http.ResponseWriter, r *http.Request, advertiserID, campaignID int) error {
	if r.URL.Query().Get(rollbackCampaignOnErrorQueryParam) != "true" {
		return nil
	}

	if err := a.service.DeleteAdCampaign(r.Context(), advertiserID, campaignID); err != nil {
		handler.HandleError(w, r, "rolling back campaign after group create failure", err)
		return err
	}

	return nil
}

// UpdateAdGroup обновляет группу объявлений.
// @Summary      Обновление группы объявлений
// @Tags         ad_groups
// @Accept       json
// @Produce      json
// @Param        ad_campaign_id  path      int                       true  "ID рекламной кампании"
// @Param        ad_group_id     path      int                       true  "ID группы"
// @Param        body            body      dto.UpdateAdGroupRequest  true  "Поля для обновления"
// @Success      200             {object}  httpx.Success
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      404             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id} [put]
// @Security     CookieAuth
func (a *API) UpdateAdGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}

	groupID, err := strconv.Atoi(mux.Vars(r)["ad_group_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing group id", err)
		return
	}

	req, err := newJSONRequest[dto.UpdateAdGroupRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	if err = a.service.UpdateAdGroup(ctx, advertiserID, req.ToInput(groupID)); err != nil {
		handler.HandleError(w, r, "updating group", err)
		return
	}

	httpx.JSON(w, http.StatusOK, nil)
}

// ListAdGroups возвращает группы объявлений кампании.
// @Summary      Список групп объявлений
// @Tags         ad_groups
// @Produce      json
// @Param        ad_campaign_id  path  int  true  "ID рекламной кампании"
// @Success      200             {object}  dto.ListAdGroupsResponse
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      404             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id}/ad_groups [get]
// @Security     CookieAuth
func (a *API) ListAdGroups(w http.ResponseWriter, r *http.Request) {
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

	groups, err := a.service.ListAdGroups(ctx, advertiserID, campaignID)
	if err != nil {
		handler.HandleError(w, r, "listing groups", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.ToListAdGroupsResponse(campaignID, groups))
}

// DeleteAdGroup удаляет группу объявлений.
// @Summary      Удаление группы объявлений
// @Tags         ad_groups
// @Produce      json
// @Param        ad_campaign_id  path  int  true  "ID рекламной кампании"
// @Param        ad_group_id     path  int  true  "ID группы"
// @Success      200             {object}  httpx.Success
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      404             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id} [delete]
// @Security     CookieAuth
func (a *API) DeleteAdGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}

	groupID, err := strconv.Atoi(mux.Vars(r)["ad_group_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing group id", err)
		return
	}

	if err = a.service.DeleteAdGroup(ctx, advertiserID, groupID); err != nil {
		handler.HandleError(w, r, "deleting group", err)
		return
	}

	httpx.JSON(w, http.StatusOK, nil)
}
