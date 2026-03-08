package handlers

import (
	"eshkere/internal/session"

	"github.com/gorilla/mux"
)

type Service interface {
	SaveLoginData()
	// методы из сервисного слоя
}

type APIConfig struct {
	Service        Service
	SessionManager *session.Manager
}

type API struct {
	service        Service
	sessionManager *session.Manager
}

func NewAPI(config APIConfig) *API {
	return &API{
		service:        config.Service,
		sessionManager: config.SessionManager,
	}
}

func (a *API) RegisterRoutes(r *mux.Router) {
	a.RegisterAdvertiserHandlers(r)
	// другие хендлеры
}
