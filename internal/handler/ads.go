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
	adsGroup.HandleFunc("/", a.ListAds).Methods(http.MethodGet)
}

func (a *API) ListAds(w http.ResponseWriter, r *http.Request) {
	advertiserID, err := middleware.AdvertiserIDFromContext(r.Context())
	if err != nil {
		httpx.Unauthorized(w, "unauthorized")
		return
	}

	// примерчики
	ads := []dto.AdResponse{
		{
			ID:          1,
			Title:       "iPhone 14",
			Description: "Телефон в хорошем состоянии",
			Price:       70000,
		},
		{
			ID:          2,
			Title:       "MacBook Air M1",
			Description: "Ноутбук для работы и учебы",
			Price:       85000,
		},
		{
			ID:          3,
			Title:       "PlayStation 5",
			Description: "Приставка, почти не использовалась",
			Price:       50000,
		},
	}

	resp := dto.ListAdsResponse{
		AdvertiserID: advertiserID,
		Ads:          ads,
	}

	httpx.JSON(w, http.StatusOK, resp)
}
