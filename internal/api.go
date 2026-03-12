package internal

import (
	"eshkere/internal/repository"
	"eshkere/internal/session"

	"github.com/gorilla/mux"
)

type APIConfig struct {
	SessionManager *session.Manager
	Repo           *repository.Repository
}
type API struct {
	sessionManager *session.Manager
	repo           *repository.Repository
}

func NewAPI(config APIConfig) *API {
	return &API{
		sessionManager: config.SessionManager,
		repo:           config.Repo,
	}
}

func (a *API) RegisterRoutes(r *mux.Router) {
	a.RegisterAdvertiserHandlers(r)
	a.RegisterAdsHandlers(r)
	// другие хендлеры
}
