package handlers

import (
	"eshkere/internal/handler/dto"
	"eshkere/pkg/httpx"
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
	var req dto.RegisterRequest

	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	if err := req.Validate(); err != nil {
		httpx.BadRequest(w, err.Error())
		return
	}

	// временно считаем, что пользователь создан
	advertiserID := 1

	if err := a.sessionManager.Create(w, advertiserID); err != nil {
		httpx.InternalError(w)
		return
	}

	resp := dto.RegisterResponse{
		ID:    advertiserID,
		Email: req.Email,
		Phone: req.Phone,
	}

	httpx.JSON(w, http.StatusOK, resp)
}

func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	if err := req.Validate(); err != nil {
		httpx.BadRequest(w, err.Error())
		return
	}

	// Временно считаем, что пользователь найден и пароль верный
	advertiserID := 1

	if err := a.sessionManager.Create(w, advertiserID); err != nil {
		httpx.InternalError(w)
		return
	}

	resp := dto.LoginResponse{
		ID:    advertiserID,
		Email: "test@example.com",
		Phone: "+79991234567",
	}

	httpx.JSON(w, http.StatusOK, resp)
}

func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	a.sessionManager.Destroy(w, r)

	httpx.JSON(w, http.StatusOK, map[string]string{
		"message": "logout ok",
	})
}
