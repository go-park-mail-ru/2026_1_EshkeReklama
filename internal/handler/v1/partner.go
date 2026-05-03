package v1

import (
	"context"
	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func (a *API) RegisterPartnerHandlers(r *mux.Router) {
	groups := r.PathPrefix("/partners").Subrouter()
	groups.HandleFunc("/register", a.RegisterPartner).Methods(http.MethodPost)
	groups.HandleFunc("/login", a.LoginPartner).Methods(http.MethodPost)
	groups.HandleFunc("/logout", a.LogoutPartner).Methods(http.MethodPost)
	groups.Handle("/me", middleware.PartnerAuth(a.authClient, a.partnerCookieConfig.Name)(http.HandlerFunc(a.MePartner))).Methods(http.MethodGet)
	groups.Handle("/me", middleware.PartnerAuth(a.authClient, a.partnerCookieConfig.Name)(http.HandlerFunc(a.UpdatePartner))).Methods(http.MethodPut)
}

// @Summary      Регистрация партнера
// @Description  Создает профиль партнера и открывает сессию
// @Tags         partner
// @Accept       json
// @Produce      json
// @Param        input  body      dto.PartnerRegisterRequest  true  "Данные партнера"
// @Success      200    {object}  dto.PartnerAuthResponse
// @Failure      400    {object}  httpx.Error
// @Failure      409    {object}  httpx.Error
// @Failure      500    {object}  httpx.Error
// @Router       /partners/register [post]
func (a *API) RegisterPartner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, err := newJSONRequest[dto.PartnerRegisterRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	partnerID, sessionID, expiresAt, err := a.authClient.Register(ctx, req.Email, req.Phone, req.Password)
	if err != nil {
		handler.HandleError(w, r, "register partner", err)
		return
	}

	if err := a.service.CreatePartnerProfile(ctx, req.ToInput(partnerID)); err != nil {
		_ = a.authClient.Logout(ctx, sessionID)
		handler.HandleError(w, r, "create partner profile", err)
		return
	}

	a.setSessionCookieWithConfig(w, a.partnerCookieConfig, sessionID, time.Unix(expiresAt, 0))
	httpx.JSON(w, http.StatusOK, dto.PartnerAuthResponse{ID: int(partnerID), Email: req.Email, Phone: req.Phone})
}

// @Summary      Вход партнера
// @Description  Логин по email или телефону и паролю
// @Tags         partner
// @Accept       json
// @Produce      json
// @Param        input  body      dto.PartnerLoginRequest  true  "Учетные данные"
// @Success      200    {object}  dto.PartnerAuthResponse
// @Failure      400    {object}  httpx.Error
// @Failure      401    {object}  httpx.Error
// @Failure      500    {object}  httpx.Error
// @Router       /partners/login [post]
func (a *API) LoginPartner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, err := newJSONRequest[dto.PartnerLoginRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	partnerID, sessionID, expiresAt, err := a.authClient.Login(ctx, req.Identifier, req.Password)
	if err != nil {
		handler.HandleError(w, r, "login partner", err)
		return
	}
	email, phone, err := a.authClient.GetCredentials(ctx, partnerID)
	if err != nil {
		handler.HandleError(w, r, "get partner credentials", err)
		return
	}

	a.setSessionCookieWithConfig(w, a.partnerCookieConfig, sessionID, time.Unix(expiresAt, 0))
	httpx.JSON(w, http.StatusOK, dto.PartnerAuthResponse{ID: int(partnerID), Email: email, Phone: phone})
}

// @Summary      Профиль партнера
// @Description  Возвращает профиль текущего партнера по сессии
// @Tags         partner
// @Produce      json
// @Success      200  {object}  dto.PartnerProfileResponse
// @Failure      401  {object}  httpx.Error
// @Failure      404  {object}  httpx.Error
// @Failure      500  {object}  httpx.Error
// @Router       /partners/me [get]
// @Security     CookieAuth
func (a *API) MePartner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	partnerID, err := ctxutils.PartnerIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "partner id from ctx", err)
		return
	}
	partner, err := a.service.GetPartnerByID(ctx, partnerID)
	if err != nil {
		handler.HandleError(w, r, "get partner by id", err)
		return
	}
	email, phone, err := a.authClient.GetCredentials(ctx, int64(partnerID))
	if err != nil {
		handler.HandleError(w, r, "get partner credentials", err)
		return
	}
	httpx.JSON(w, http.StatusOK, dto.PartnerToProfile(partner, email, phone))
}

// @Summary      Обновление профиля партнера
// @Description  Обновляет данные текущего партнера
// @Tags         partner
// @Accept       json
// @Produce      json
// @Param        input  body      dto.UpdatePartnerProfileRequest  true  "Поля профиля"
// @Success      200    {object}  dto.PartnerProfileResponse
// @Failure      400    {object}  httpx.Error
// @Failure      401    {object}  httpx.Error
// @Failure      409    {object}  httpx.Error
// @Failure      500    {object}  httpx.Error
// @Router       /partners/me [put]
// @Security     CookieAuth
func (a *API) UpdatePartner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	partnerID, err := ctxutils.PartnerIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "partner id from ctx", err)
		return
	}
	req, err := newJSONRequest[dto.UpdatePartnerProfileRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}
	email, phone, err := a.resolveUpdatedPartnerContacts(ctx, int64(partnerID), req)
	if err != nil {
		handler.HandleError(w, r, "update partner credentials", err)
		return
	}

	partner, err := a.service.UpdatePartnerProfile(ctx, req.ToInput(partnerID))
	if err != nil {
		handler.HandleError(w, r, "update partner profile", err)
		return
	}
	httpx.JSON(w, http.StatusOK, dto.PartnerToProfile(partner, email, phone))
}

func (a *API) resolveUpdatedPartnerContacts(ctx context.Context, partnerID int64, req *dto.UpdatePartnerProfileRequest) (string, string, error) {
	if req.Email == nil && req.Phone == nil {
		return a.authClient.GetCredentials(ctx, partnerID)
	}
	email, phone, err := a.authClient.GetCredentials(ctx, partnerID)
	if err != nil {
		return "", "", err
	}
	if req.Email != nil {
		email = *req.Email
	}
	if req.Phone != nil {
		phone = *req.Phone
	}
	return a.authClient.UpdateCredentials(ctx, partnerID, email, phone)
}

// @Summary      Выход партнера
// @Description  Завершает текущую партнерскую сессию
// @Tags         partner
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  httpx.Error
// @Router       /partners/logout [post]
func (a *API) LogoutPartner(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(a.partnerCookieConfig.Name)
	if err == nil {
		if logoutErr := a.authClient.Logout(r.Context(), cookie.Value); logoutErr != nil {
			handler.HandleError(w, r, "logout partner", logoutErr)
			return
		}
	}
	a.clearSessionCookieWithConfig(w, a.partnerCookieConfig)
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "logout ok"})
}

var _ = context.Background
