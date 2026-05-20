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

func (a *API) RegisterPartnerSiteHandlers(r *mux.Router) {
	partnerSites := r.PathPrefix("/partners/sites").Subrouter()
	partnerSites.Use(middleware.Auth(a.authClient, a.cookieConfig.Name))
	partnerSites.HandleFunc("", a.ListPartnerSites).Methods(http.MethodGet)
	partnerSites.HandleFunc("", a.CreatePartnerSite).Methods(http.MethodPost)
	partnerSites.HandleFunc("/{site_id}", a.GetPartnerSite).Methods(http.MethodGet)
	partnerSites.HandleFunc("/{site_id}", a.UpdatePartnerSite).Methods(http.MethodPut)
	partnerSites.HandleFunc("/{site_id}", a.DeletePartnerSite).Methods(http.MethodDelete)
}

// @Summary      Список сайтов партнера
// @Description  Возвращает все сайты текущего партнера
// @Tags         partner_sites
// @Produce      json
// @Success      200  {object}  dto.ListPartnerSitesResponse
// @Failure      401  {object}  httpx.Error
// @Failure      500  {object}  httpx.Error
// @Router       /partners/sites [get]
// @Security     CookieAuth
func (a *API) ListPartnerSites(w http.ResponseWriter, r *http.Request) {
	partnerID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "advertiser id from ctx", err)
		return
	}
	sites, err := a.service.ListPartnerSites(r.Context(), partnerID)
	if err != nil {
		handler.HandleError(w, r, "list partner sites", err)
		return
	}
	httpx.JSON(w, http.StatusOK, dto.ListPartnerSitesResponse{
		PartnerID: partnerID,
		Sites:     dto.ToPartnerSiteResponses(sites),
	})
}

// @Summary      Создание сайта партнера
// @Description  Добавляет новый домен в раздел "Реклама на сайтах"
// @Tags         partner_sites
// @Accept       json
// @Produce      json
// @Param        input  body      dto.CreatePartnerSiteRequest  true  "Домен и название сайта"
// @Success      201    {object}  map[string]int
// @Failure      400    {object}  httpx.Error
// @Failure      401    {object}  httpx.Error
// @Failure      409    {object}  httpx.Error
// @Failure      500    {object}  httpx.Error
// @Router       /partners/sites [post]
// @Security     CookieAuth
func (a *API) CreatePartnerSite(w http.ResponseWriter, r *http.Request) {
	partnerID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "advertiser id from ctx", err)
		return
	}
	req, err := newJSONRequest[dto.CreatePartnerSiteRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}
	site, err := a.service.CreatePartnerSite(r.Context(), req.ToInput(partnerID))
	if err != nil {
		handler.HandleError(w, r, "create partner site", err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]int{"id": site.ID})
}

// @Summary      Сайт партнера
// @Description  Возвращает один сайт текущего партнера
// @Tags         partner_sites
// @Produce      json
// @Param        site_id  path      int  true  "ID сайта"
// @Success      200      {object}  dto.PartnerSiteResponse
// @Failure      400      {object}  httpx.Error
// @Failure      401      {object}  httpx.Error
// @Failure      404      {object}  httpx.Error
// @Failure      500      {object}  httpx.Error
// @Router       /partners/sites/{site_id} [get]
// @Security     CookieAuth
func (a *API) GetPartnerSite(w http.ResponseWriter, r *http.Request) {
	partnerID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "advertiser id from ctx", err)
		return
	}
	siteID, err := strconv.Atoi(mux.Vars(r)["site_id"])
	if err != nil {
		httpx.BadRequest(w, "invalid site id")
		return
	}
	site, err := a.service.GetPartnerSite(r.Context(), siteID)
	if err != nil {
		handler.HandleError(w, r, "get partner site", err)
		return
	}
	if site.PartnerID != partnerID {
		httpx.NotFound(w, "not found")
		return
	}
	httpx.JSON(w, http.StatusOK, dto.ToPartnerSiteResponse(site))
}

// @Summary      Обновление сайта партнера
// @Description  Обновляет домен и название сайта
// @Tags         partner_sites
// @Accept       json
// @Produce      json
// @Param        site_id  path      int                          true  "ID сайта"
// @Param        input    body      dto.UpdatePartnerSiteRequest true  "Поля сайта"
// @Success      200      {object}  dto.PartnerSiteResponse
// @Failure      400      {object}  httpx.Error
// @Failure      401      {object}  httpx.Error
// @Failure      404      {object}  httpx.Error
// @Failure      409      {object}  httpx.Error
// @Failure      500      {object}  httpx.Error
// @Router       /partners/sites/{site_id} [put]
// @Security     CookieAuth
func (a *API) UpdatePartnerSite(w http.ResponseWriter, r *http.Request) {
	partnerID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "advertiser id from ctx", err)
		return
	}
	siteID, err := strconv.Atoi(mux.Vars(r)["site_id"])
	if err != nil {
		httpx.BadRequest(w, "invalid site id")
		return
	}
	req, err := newJSONRequest[dto.UpdatePartnerSiteRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}
	site, err := a.service.UpdatePartnerSite(r.Context(), req.ToInput(partnerID, siteID))
	if err != nil {
		handler.HandleError(w, r, "update partner site", err)
		return
	}
	httpx.JSON(w, http.StatusOK, dto.ToPartnerSiteResponse(site))
}

// @Summary      Удаление сайта партнера
// @Description  Удаляет сайт текущего партнера
// @Tags         partner_sites
// @Produce      json
// @Param        site_id  path      int  true  "ID сайта"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  httpx.Error
// @Failure      401      {object}  httpx.Error
// @Failure      404      {object}  httpx.Error
// @Failure      500      {object}  httpx.Error
// @Router       /partners/sites/{site_id} [delete]
// @Security     CookieAuth
func (a *API) DeletePartnerSite(w http.ResponseWriter, r *http.Request) {
	partnerID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "advertiser id from ctx", err)
		return
	}
	siteID, err := strconv.Atoi(mux.Vars(r)["site_id"])
	if err != nil {
		httpx.BadRequest(w, "invalid site id")
		return
	}
	if err := a.service.DeletePartnerSite(r.Context(), partnerID, siteID); err != nil {
		handler.HandleError(w, r, "delete partner site", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
