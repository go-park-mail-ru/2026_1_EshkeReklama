package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

const AdvertiserGroupURI = "/advertiser"

const (
	RegisterURI = "/register"
	LoginURI    = "/login"
	LogoutURI   = "/logout"
)

func (a *API) RegisterAdvertiserHandlers(r *mux.Router) {
	advertiserGroup := r.PathPrefix(AdvertiserGroupURI).Subrouter()

	advertiserGroup.HandleFunc(RegisterURI, a.Register).Methods(http.MethodPost)
	advertiserGroup.HandleFunc(LoginURI, a.Login).Methods(http.MethodPost)
	advertiserGroup.HandleFunc(LogoutURI, a.Logout).Methods(http.MethodPost)
}

func (a *API) Register(w http.ResponseWriter, r *http.Request) {
	return
}

func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	return
}

func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	return
}
