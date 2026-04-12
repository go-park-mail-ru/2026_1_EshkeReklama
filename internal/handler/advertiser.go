package handlers

import (
	"database/sql"
	"errors"
	"eshkere/internal/handler/dto"
	"eshkere/internal/middleware"
	"eshkere/internal/models"
	"eshkere/internal/service"
	"eshkere/pkg/httpx"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func (a *API) RegisterAdvertiserHandlers(r *mux.Router) {
	g := r.PathPrefix("/advertiser").Subrouter()

	g.HandleFunc("/register", a.Register).Methods(http.MethodPost)
	g.HandleFunc("/login", a.Login).Methods(http.MethodPost)
	g.HandleFunc("/logout", a.Logout).Methods(http.MethodPost)

	g.Handle("/me", middleware.Auth(a.sessionManager)(http.HandlerFunc(a.Me))).Methods(http.MethodGet)
}

// @Summary      Регистрация рекламодателя
// @Description  Создает новый аккаунт и открывает сессию
// @Tags         advertiser
// @Accept       json
// @Produce      json
// @Param        input body      dto.RegisterRequest  true  "Данные для регистрации"
// @Success      200   {object}  dto.RegisterResponse
// @Failure      400   {object}  httpx.Error "Invalid request или User already exists"
// @Failure      500   {object}  httpx.Error
// @Router       /advertiser/register [post]
func (a *API) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.RegisterRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	adv, err := a.service.RegisterAdvertiser(ctx, req.Name, req.Email, req.Phone, req.Password)
	if err != nil {
		a.handleRegisterError(w, err)
		return
	}

	if err := a.sessionManager.Create(w, r, adv.ID); err != nil {
		httpx.InternalError(w, "internal error")
		return
	}

	httpx.JSON(w, http.StatusOK, dto.RegisterResponse{
		ID:    adv.ID,
		Email: adv.Email,
		Phone: adv.Phone,
	})
}

func (a *API) handleRegisterError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrEmailTaken):
		httpx.BadRequest(w, "email already registered")
	case errors.Is(err, service.ErrPhoneTaken):
		httpx.BadRequest(w, "phone already registered")
	case errors.Is(err, service.ErrInvalidAdvertiserArg):
		httpx.BadRequest(w, err.Error())
	default:
		httpx.InternalError(w, "internal error")
	}
}

// @Summary      Вход рекламодателя
// @Description  Аутентифицирует рекламодателя по email или телефону и паролю
// @Tags         advertiser
// @Accept       json
// @Produce      json
// @Param        input body      dto.LoginRequest  true  "Данные для входа"
// @Success      200   {object}  dto.LoginResponse
// @Failure      400   {object}  httpx.Error "Invalid identifier или password"
// @Failure      401   {object}  httpx.Error "Неверные учётные данные"
// @Failure      500   {object}  httpx.Error
// @Router       /advertiser/login [post]
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.LoginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	adv, err := a.service.AuthenticateAdvertiser(ctx, req.Identifier, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			httpx.Unauthorized(w, "invalid credentials")
			return
		}
		httpx.InternalError(w, "internal error")
		return
	}

	if err := a.sessionManager.Create(w, r, adv.ID); err != nil {
		httpx.InternalError(w, "internal error")
		return
	}

	httpx.JSON(w, http.StatusOK, dto.LoginResponse{
		ID:    adv.ID,
		Email: adv.Email,
		Phone: adv.Phone,
	})
}

// @Summary      Профиль рекламодателя
// @Description  Возвращает данные текущего пользователя по сессии
// @Tags         advertiser
// @Produce      json
// @Success      200   {object}  dto.AdvertiserProfileResponse
// @Failure      401   {object}  httpx.Error
// @Failure      404   {object}  httpx.Error
// @Failure      500   {object}  httpx.Error
// @Router       /advertiser/me [get]
// @Security     CookieAuth
func (a *API) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := middleware.AdvertiserIDFromContext(ctx)
	if err != nil {
		httpx.Unauthorized(w, "unauthorized")
		return
	}

	adv, err := a.service.GetAdvertiserByID(ctx, advertiserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpx.NotFound(w, "advertiser not found")
			return
		}
		httpx.InternalError(w, "internal error")
		return
	}

	httpx.JSON(w, http.StatusOK, advertiserToProfile(adv))
}

func advertiserToProfile(adv *models.Advertiser) dto.AdvertiserProfileResponse {
	if adv == nil {
		return dto.AdvertiserProfileResponse{}
	}
	return dto.AdvertiserProfileResponse{
		ID:        adv.ID,
		Name:      adv.Name,
		Email:     adv.Email,
		Phone:     adv.Phone,
		Balance:   adv.Balance,
		CreatedAt: adv.CreatedAt.Format(time.RFC3339),
	}
}

// @Summary      Выход рекламодателя
// @Description  Завершает сессию текущего рекламодателя
// @Tags         advertiser
// @Produce      json
// @Success      200   {object}  map[string]string
// @Failure      500   {object}  httpx.Error
// @Router       /advertiser/logout [post]
func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	if err := a.sessionManager.Destroy(w, r); err != nil {
		httpx.InternalError(w, "internal error")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]string{
		"message": "logout ok",
	})
}
