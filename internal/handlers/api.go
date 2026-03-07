package handlers

import (
	"github.com/gorilla/mux"
)

type Service interface {
	SaveLoginData()
	// методы из сервисного слоя
}

type APIConfig struct {
	Service Service
	// Auth httpx.HandleFunc
}

type API struct {
	service Service
	// auth    httpx.HandlerFunc
}

func NewAPI(config APIConfig) *API {
	return &API{
		service: config.Service,
		// auth
	}
}

func (a *API) RegisterRoutes(r *mux.Router) {
	a.RegisterAdvertiserHandlers(r)
	// другие хендлеры
}
