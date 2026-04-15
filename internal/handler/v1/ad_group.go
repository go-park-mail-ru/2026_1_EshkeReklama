package v1

import (
	handlers "eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/pkg/httpx"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (a *API) RegisterAdGroupHandlers(r *mux.Router) {
	groups := r.PathPrefix("/ad_campaigns/{ad_campaign_id}/ad_groups").Subrouter()

	groups.Use(middleware.Auth(a.sessionManager))
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
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id}/ad_groups [post]
// @Security     CookieAuth
func (a *API) CreateAdGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	campaignID, err := strconv.Atoi(mux.Vars(r)["ad_campaign_id"])
	if err != nil {
		handlers.HandleError(w, r, "parsing campaign id", err)
		return
	}

	req, err := newJSONRequest[dto.CreateAdGroupRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	group := req.ToModel(campaignID)
	created, err := a.service.CreateAdGroup(ctx, group)
	if err != nil {
		handlers.HandleError(w, r, "creating group", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.CreateAdGroupResponse{
		ID: created.ID,
	})
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
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id} [put]
// @Security     CookieAuth
func (a *API) UpdateAdGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	groupID, err := strconv.Atoi(mux.Vars(r)["ad_group_id"])
	if err != nil {
		handlers.HandleError(w, r, "parsing group id", err)
		return
	}

	req, err := newJSONRequest[dto.UpdateAdGroupRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	if err = a.service.UpdateAdGroup(ctx, groupID, req); err != nil {
		handlers.HandleError(w, r, "updating group", err)
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
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id}/ad_groups [get]
// @Security     CookieAuth
func (a *API) ListAdGroups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	campaignID, err := strconv.Atoi(mux.Vars(r)["ad_campaign_id"])
	if err != nil {
		handlers.HandleError(w, r, "parsing campaign id", err)
		return
	}

	groups, err := a.service.ListAdGroups(ctx, campaignID)
	if err != nil {
		handlers.HandleError(w, r, "listing groups", err)
		return
	}

	out := make([]*dto.AdGroupResponse, 0, len(groups))
	for _, g := range groups {
		out = append(out, dto.ToAdGroupResponse(g))
	}

	httpx.JSON(w, http.StatusOK, dto.ListAdGroupsResponse{
		AdCampaignID: campaignID,
		Groups:       out,
	})
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
// @Failure      500             {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id} [delete]
// @Security     CookieAuth
func (a *API) DeleteAdGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	groupID, err := strconv.Atoi(mux.Vars(r)["ad_group_id"])
	if err != nil {
		handlers.HandleError(w, r, "parsing group id", err)
		return
	}

	if err = a.service.DeleteAdGroup(ctx, groupID); err != nil {
		handlers.HandleError(w, r, "deleting group", err)
		return
	}

	httpx.JSON(w, http.StatusOK, nil)
}
