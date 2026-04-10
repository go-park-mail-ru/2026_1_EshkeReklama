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

// @Summary      Регистрация рекламодателя
// @Description  Создает новый аккаунт и открывает сессию
// @Tags         advertiser
// @Accept       json
// @Produce      json
// @Param        input body      dto.RegisterRequest  true  "Данные для регистрации"
// @Success      200   {object}  dto.RegisterResponse
// @Failure      400   {object}  httpx.Error "Invalid request или User already exists"
// @Router       /advertiser/register [post]
func (a *API) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	httpx.NotImplemented(w, "register is not wired to storage yet")
}

// @Summary      Вход рекламодателя
// @Description  Аутентифицирует рекламодателя по email или телефону и паролю
// @Tags         advertiser
// @Accept       json
// @Produce      json
// @Param        input body      dto.LoginRequest  true  "Данные для входа"
// @Success      200   {object}  dto.LoginResponse
// @Failure      400   {object}  httpx.Error "Invalid identifier или password"
// @Router       /advertiser/login [post]
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	httpx.NotImplemented(w, "login is not wired to storage yet")
}

// @Summary      Выход рекламодателя
// @Description  Завершает сессию текущего рекламодателя
// @Tags         advertiser
// @Produce      json
// @Success      200   {object}  map[string]string
// @Router       /advertiser/logout [post]
// @Security     CookieAuth
func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	if err := a.sessionManager.Destroy(w, r); err != nil {
		httpx.InternalError(w)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]string{
		"message": "logout ok",
	})
}
