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
	reqLogger := logger.GetLoggerFromCtx(ctx)

	campaignID, err := strconv.Atoi(mux.Vars(r)["ad_campaign_id"])
	if err != nil {
		reqLogger.Warnw("invalid ad_campaign_id", "error", err.Error())
		httpx.BadRequest(w, "invalid ad_campaign_id")
		return
	}

	req, err := newJSONRequest[dto.CreateAdGroupRequest](r)
	if err != nil {
		reqLogger.Warnw("invalid create ad group payload", "error", err.Error(), "ad_campaign_id", campaignID)
		httpx.BadRequest(w, "invalid request")
		return
	}

	group := req.ToModel(campaignID)
	created, err := a.service.CreateAdGroup(ctx, group)
	if err != nil {
		reqLogger.Warnw("create ad group rejected", "error", err.Error(), "ad_campaign_id", campaignID)
		httpx.BadRequest(w, err.Error())
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
	reqLogger := logger.GetLoggerFromCtx(ctx)

	groupID, err := strconv.Atoi(mux.Vars(r)["ad_group_id"])
	if err != nil {
		reqLogger.Warnw("invalid ad_group_id", "error", err.Error())
		httpx.BadRequest(w, "invalid ad_group_id")
		return
	}

	req, err := newJSONRequest[dto.UpdateAdGroupRequest](r)
	if err != nil {
		reqLogger.Warnw("invalid update ad group payload", "error", err.Error(), "ad_group_id", groupID)
		httpx.BadRequest(w, "invalid request")
		return
	}

	if err = a.service.UpdateAdGroup(ctx, groupID, *req); err != nil {
		reqLogger.Errorw("failed to update ad group", "error", err.Error(), "ad_group_id", groupID)
		httpx.InternalError(w, "internal error")
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
	reqLogger := logger.GetLoggerFromCtx(ctx)

	campaignID, err := strconv.Atoi(mux.Vars(r)["ad_campaign_id"])
	if err != nil {
		reqLogger.Warnw("invalid ad_campaign_id", "error", err.Error())
		httpx.BadRequest(w, "invalid ad_campaign_id")
		return
	}

	groups, err := a.service.ListAdGroups(ctx, campaignID)
	if err != nil {
		reqLogger.Errorw("failed to list ad groups", "error", err.Error(), "ad_campaign_id", campaignID)
		httpx.InternalError(w, "internal error")
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
	reqLogger := logger.GetLoggerFromCtx(ctx)

	groupID, err := strconv.Atoi(mux.Vars(r)["ad_group_id"])
	if err != nil {
		reqLogger.Warnw("invalid ad_group_id", "error", err.Error())
		httpx.BadRequest(w, "invalid ad_group_id")
		return
	}

	if err = a.service.DeleteAdGroup(ctx, groupID); err != nil {
		reqLogger.Errorw("failed to delete ad group", "error", err.Error(), "ad_group_id", groupID)
		httpx.InternalError(w, "internal error")
		return
	}

	httpx.JSON(w, http.StatusOK, nil)
}
