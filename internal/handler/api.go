package handlers

import (
	"eshkere/internal/session"

	"github.com/gorilla/mux"
)

type Service interface {
}

type APIConfig struct {
	SessionManager *session.Manager
	Service        Service
}

type API struct {
	sessionManager *session.Manager
	service        Service
}

func NewAPI(config APIConfig) *API {
	return &API{
		sessionManager: config.SessionManager,
	}
}

func (a *API) RegisterRoutes(r *mux.Router) {
	a.RegisterAdvertiserHandlers(r)
	a.RegisterAdsHandlers(r)
}
