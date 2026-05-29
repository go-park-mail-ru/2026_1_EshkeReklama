package v1

import (
	"net/http"
	"strconv"

	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/pkg/httpx"

	"github.com/gorilla/mux"
)

func (a *API) RegisterAdminHandlers(r *mux.Router) {
	admins := r.PathPrefix("/admin").Subrouter()

	admins.Use(middleware.Auth(a.authClient, a.cookieConfig.Name))
	admins.Use(middleware.IsAdmin(a.service))

	admins.HandleFunc("/ads/{ad_id}", a.GetAdminAdByID).Methods(http.MethodGet)
	admins.HandleFunc("/ads/{ad_id}/status", a.UpdateAdStatus).Methods(http.MethodPatch)
	admins.HandleFunc("/ads", a.ListAdminAds).Methods(http.MethodGet)

	a.RegisterAdminSupportHandlers(r)
}

// GetAdminAdByID возвращает объявление для просмотра администратором.
// @Summary      Просмотр объявления администратором
// @Description  Возвращает объявление по ID для модерации
// @Tags         admin
// @Produce      json
// @Param        ad_id  path      int  true  "ID объявления"
// @Success      200    {object}  dto.AdResponse
// @Failure      400    {object}  httpx.Error
// @Failure      401    {object}  httpx.Error
// @Failure      403    {object}  httpx.Error
// @Failure      404    {object}  httpx.Error
// @Failure      500    {object}  httpx.Error
// @Router       /api/admin/ads/{ad_id} [get]
// @Security     CookieAuth
func (a *API) GetAdminAdByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	adID, err := strconv.Atoi(mux.Vars(r)["ad_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing ad id", err)
		return
	}

	ad, err := a.service.GetAdByID(ctx, adID)
	if err != nil {
		handler.HandleError(w, r, "getting ad by id", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.ToAdResponse(ad))
}

// UpdateAdStatus меняет статус объявления по решению администратора.
// @Summary      Решение модерации объявления
// @Description  Одобряет или отклоняет объявление; при одобрении итоговый статус зависит от доступных средств
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        ad_id  path      int                        true  "ID объявления"
// @Param        body   body      dto.UpdateAdStatusRequest  true  "Решение модерации"
// @Success      200    {object}  httpx.Success
// @Failure      400    {object}  httpx.Error
// @Failure      401    {object}  httpx.Error
// @Failure      403    {object}  httpx.Error
// @Failure      404    {object}  httpx.Error
// @Failure      500    {object}  httpx.Error
// @Router       /api/admin/ads/{ad_id}/status [patch]
// @Security     CookieAuth
func (a *API) UpdateAdStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	adID, err := strconv.Atoi(mux.Vars(r)["ad_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing ad id", err)
		return
	}

	req, err := newJSONRequest[dto.UpdateAdStatusRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	if err = a.service.UpdateAdModerationStatus(ctx, req.ToInput(adID)); err != nil {
		handler.HandleError(w, r, "updating ad status", err)
		return
	}

	httpx.JSON(w, http.StatusOK, nil)
}

// ListAdminAds возвращает очередь объявлений, ожидающих модерации.
// @Summary      Очередь объявлений на модерации
// @Description  Возвращает все объявления со статусом moderation
// @Tags         admin
// @Produce      json
// @Success      200  {object}  dto.ListAdminAdsResponse
// @Failure      401  {object}  httpx.Error
// @Failure      403  {object}  httpx.Error
// @Failure      500  {object}  httpx.Error
// @Router       /api/admin/ads [get]
// @Security     CookieAuth
func (a *API) ListAdminAds(w http.ResponseWriter, r *http.Request) {
	ads, err := a.service.ListModerationAds(r.Context())
	if err != nil {
		handler.HandleError(w, r, "listing admin ads", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.ToListAdminAdsResponse(ads))
}
