package internal

import (
	"eshkere/pkg/httpx"
	"net/http"

	"github.com/gorilla/mux"
)

const AdsGroupURI = "/ads"

func (a *API) RegisterAdsHandlers(r *mux.Router) {
	adsGroup := r.PathPrefix(AdsGroupURI).Subrouter()

	adsGroup.Use(Auth(a.sessionManager))
	adsGroup.HandleFunc("", a.ListAds).Methods(http.MethodGet)
}

// @Summary      Список рекламных кампаний
// @Description  Возвращает список всех кампаний текущего рекламодателя
// @Tags         ads
// @Produce      json
// @Success      200   {object}  ListAdsResponse
// @Failure      401   {object}  httpx.Error "Unauthorized"
// @Router       /ads [get]
// @Security     CookieAuth
func (a *API) ListAds(w http.ResponseWriter, r *http.Request) {
	advertiserID, err := AdvertiserIDFromContext(r.Context())
	if err != nil {
		httpx.Unauthorized(w, "unauthorized")
		return
	}

	// Достаем кампании именно для этого рекламодателя
	unparsedAds := a.repo.Ads.ListByID(advertiserID)

	parsedAds := make([]AdResponse, len(unparsedAds))
	for i, ad := range unparsedAds {
		parsedAds[i] = AdResponse{
			ID:           ad.ID,
			Title:        ad.Title,
			TargetAction: ad.TargetAction,
			Price:        ad.Price,
		}
	}

	resp := ListAdsResponse{
		AdvertiserID: advertiserID,
		Ads:          parsedAds,
	}

	httpx.JSON(w, http.StatusOK, resp)
}
