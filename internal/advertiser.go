package internal

import (
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
// @Param        input body      RegisterRequest  true  "Данные для регистрации"
// @Success      200   {object}  RegisterResponse
// @Failure      400   {object}  httpx.Error "Invalid request или User already exists"
// @Router       /advertiser/register [post]
func (a *API) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	if _, exists := a.repo.Users.GetByEmail(req.Email); exists {
		httpx.BadRequest(w, "user already exists")
		return
	}

	newUser := a.repo.Users.Create(req.Email, req.Phone, req.Password)

	if err := a.sessionManager.Create(w, r, newUser.ID); err != nil {
		httpx.InternalError(w)
		return
	}

	httpx.JSON(w, http.StatusOK, RegisterResponse{
		ID:    newUser.ID,
		Email: newUser.Email,
		Phone: newUser.Phone,
	})
}

// @Summary      Вход рекламодателя
// @Description  Аутентифицирует рекламодателя по email или телефону и паролю
// @Tags         advertiser
// @Accept       json
// @Produce      json
// @Param        input body      LoginRequest  true  "Данные для входа"
// @Success      200   {object}  LoginResponse
// @Failure      400   {object}  httpx.Error "Invalid identifier или password"
// @Router       /advertiser/login [post]
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	// Ищем пользователя по любому из признаков
	user, exists := a.repo.Users.GetByEmail(req.Identifier)
	if !exists {
		user, exists = a.repo.Users.GetByPhone(req.Identifier)
	}

	if !exists || user.Password != req.Password {
		httpx.BadRequest(w, "invalid identifier or password")
		return
	}

	if err := a.sessionManager.Create(w, r, user.ID); err != nil {
		httpx.InternalError(w)
		return
	}

	httpx.JSON(w, http.StatusOK, LoginResponse{
		ID:    user.ID,
		Email: user.Email,
		Phone: user.Phone,
	})
}

// @Summary      Выход рекламодателя
// @Description  Завершает сессию текущего рекламодателя
// @Tags         advertiser
// @Produce      json
// @Success      200   {object}  map[string]string
// @Router       /advertiser/logout [post]
// @Security     CookieAuth
func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	a.sessionManager.Destroy(w, r)

	httpx.JSON(w, http.StatusOK, map[string]string{
		"message": "logout ok",
	})
}
