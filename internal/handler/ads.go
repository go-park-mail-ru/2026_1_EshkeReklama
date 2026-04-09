package handlers

import (
	"eshkere/internal/handler/dto"
	"eshkere/internal/middleware"
	"eshkere/pkg/httpx"
	"net/http"

	"github.com/gorilla/mux"
)

const AdsGroupURI = "/ads"

func (a *API) RegisterAdsHandlers(r *mux.Router) {
	adsGroup := r.PathPrefix(AdsGroupURI).Subrouter()

	adsGroup.Use(middleware.Auth(a.sessionManager))
	adsGroup.HandleFunc("", a.ListAds).Methods(http.MethodGet)
}

// @Summary      Список рекламных кампаний
// @Description  Возвращает список всех кампаний текущего рекламодателя
// @Tags         ads
// @Produce      json
// @Success      200   {object}  dto.ListAdsResponse
// @Failure      401   {object}  httpx.Error "Unauthorized"
// @Router       /ads [get]
// @Security     CookieAuth
func (a *API) ListAds(w http.ResponseWriter, r *http.Request) {
	advertiserID, err := middleware.AdvertiserIDFromContext(r.Context())
	if err != nil {
		httpx.Unauthorized(w, "unauthorized")
		return
	}

	httpx.JSON(w, http.StatusOK, dto.ListAdsResponse{
		AdvertiserID: advertiserID,
		Ads:          []dto.AdResponse{},
	})
}
