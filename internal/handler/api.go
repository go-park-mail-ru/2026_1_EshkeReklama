package handlers

import (
	"eshkere/internal/session"

	"github.com/gorilla/mux"
)

type APIConfig struct {
	SessionManager *session.Manager
}

type API struct {
	sessionManager *session.Manager
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
