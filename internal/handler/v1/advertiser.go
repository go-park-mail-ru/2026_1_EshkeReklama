package v1

import (
	"context"
	"errors"
	"net/http"
	"time"

	errs "eshkere/internal/errors"
	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"

	"github.com/gorilla/mux"
)

func (a *API) RegisterAdvertiserHandlers(r *mux.Router) {
	advertisers := r.PathPrefix("/advertisers").Subrouter()

	advertisers.HandleFunc("/register", a.Register).Methods(http.MethodPost)
	advertisers.HandleFunc("/login", a.Login).Methods(http.MethodPost)
	advertisers.HandleFunc("/login/vk", a.LoginVKID).Methods(http.MethodPost)
	advertisers.HandleFunc("/logout", a.Logout).Methods(http.MethodPost)
	advertisers.Handle("/balance", middleware.Auth(a.authClient, a.cookieConfig.Name)(http.HandlerFunc(a.GetBalance))).Methods(http.MethodGet)
	advertisers.Handle("/balance/topup", middleware.Auth(a.authClient, a.cookieConfig.Name)(http.HandlerFunc(a.TopUpBalance))).Methods(http.MethodPost)

	advertisers.Handle("/me", middleware.Auth(a.authClient, a.cookieConfig.Name)(http.HandlerFunc(a.Me))).Methods(http.MethodGet)
	advertisers.Handle("/me", middleware.Auth(a.authClient, a.cookieConfig.Name)(http.HandlerFunc(a.UpdateProfile))).Methods(http.MethodPut)
	advertisers.Handle("/me/avatar", middleware.Auth(a.authClient, a.cookieConfig.Name)(http.HandlerFunc(a.UpdateAvatar))).Methods(http.MethodPut)
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
// @Router       /advertisers/register [post]
func (a *API) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := newJSONRequest[dto.RegisterRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	advID, sessionID, expiresAt, err := a.authClient.Register(ctx, req.Email, req.Phone, req.Password)
	if err != nil {
		handler.HandleError(w, r, "register advertiser", err)
		return
	}

	if err = a.service.CreateAdvertiserProfile(ctx, advID, req.Name, req.Email); err != nil {
		// компенсирующая операция: откатываем credentials в auth-сервисе
		_ = a.authClient.Logout(ctx, sessionID)
		handler.HandleError(w, r, "create advertiser profile", err)
		return
	}

	a.setSessionCookie(w, sessionID, time.Unix(expiresAt, 0))

	httpx.JSON(w, http.StatusOK, dto.RegisterResponse{
		ID:    int(advID),
		Email: req.Email,
		Phone: req.Phone,
	})
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
// @Router       /advertisers/login [post]
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := newJSONRequest[dto.LoginRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	advID, sessionID, expiresAt, err := a.authClient.Login(ctx, req.Identifier, req.Password)
	if err != nil {
		handler.HandleError(w, r, "auth advertiser", err)
		return
	}

	email, phone, err := a.authClient.GetCredentials(ctx, advID)
	if err != nil {
		handler.HandleError(w, r, "get advertiser credentials", err)
		return
	}

	a.setSessionCookie(w, sessionID, time.Unix(expiresAt, 0))

	httpx.JSON(w, http.StatusOK, dto.LoginResponse{
		ID:    int(advID),
		Email: email,
		Phone: phone,
	})
}

// @Summary      Вход рекламодателя через VK ID
// @Description  Аутентифицирует рекламодателя по frontend-driven VK ID SDK payload и открывает сессию
// @Tags         advertiser
// @Accept       json
// @Produce      json
// @Param        input body      dto.VKIDLoginRequest  true  "access_token и user_id от VK ID SDK"
// @Success      200   {object}  dto.LoginResponse
// @Failure      400   {object}  httpx.Error
// @Failure      401   {object}  httpx.Error
// @Failure      409   {object}  httpx.Error
// @Failure      500   {object}  httpx.Error
// @Router       /advertisers/login/vk [post]
func (a *API) LoginVKID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := newJSONRequest[dto.VKIDLoginRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	advID, sessionID, expiresAt, firstName, lastName, err := a.authClient.LoginVKID(ctx, req.AccessToken, req.UserID)
	if err != nil {
		handler.HandleError(w, r, "auth advertiser via vk id", err)
		return
	}

	email, phone, err := a.authClient.GetCredentials(ctx, advID)
	if err != nil {
		handler.HandleError(w, r, "get advertiser credentials", err)
		return
	}

	if err := a.ensureAdvertiserProfile(ctx, advID, firstName, lastName, email); err != nil {
		_ = a.authClient.Logout(ctx, sessionID)
		handler.HandleError(w, r, "ensure advertiser profile", err)
		return
	}

	a.setSessionCookie(w, sessionID, time.Unix(expiresAt, 0))

	httpx.JSON(w, http.StatusOK, dto.LoginResponse{
		ID:    int(advID),
		Email: email,
		Phone: phone,
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
// @Router       /advertisers/me [get]
// @Security     CookieAuth
func (a *API) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser id from ctx", err)
		return
	}

	adv, err := a.service.GetAdvertiserByID(ctx, advertiserID)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser by id", err)
		return
	}

	email, phone, err := a.authClient.GetCredentials(ctx, int64(advertiserID))
	if err != nil {
		handler.HandleError(w, r, "getting advertiser credentials", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.AdvertiserWithContactsToProfile(adv, email, phone))
}

// @Summary      Обновление профиля рекламодателя
// @Description  Обновляет данные текущего рекламодателя по сессии
// @Tags         advertiser
// @Accept       json
// @Produce      json
// @Param        body   body      dto.UpdateAdvertiserProfileRequest  true  "Поля для обновления профиля"
// @Success      200    {object}  dto.AdvertiserProfileResponse
// @Failure      400    {object}  httpx.Error
// @Failure      401    {object}  httpx.Error
// @Failure      500    {object}  httpx.Error
// @Router       /advertisers/me [put]
// @Security     CookieAuth
func (a *API) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser id from ctx", err)
		return
	}

	req, err := newJSONRequest[dto.UpdateAdvertiserProfileRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	email, phone, err := a.resolveUpdatedContacts(ctx, int64(advertiserID), req)
	if err != nil {
		handler.HandleError(w, r, "updating advertiser credentials", err)
		return
	}

	var adv *models.Advertiser
	if hasAdvertiserProfileChanges(req) {
		adv, err = a.service.UpdateAdvertiserProfile(ctx, req.ToInput(advertiserID))
		if err != nil {
			handler.HandleError(w, r, "updating profile", err)
			return
		}
	} else {
		adv, err = a.service.GetAdvertiserByID(ctx, advertiserID)
		if err != nil {
			handler.HandleError(w, r, "getting advertiser by id", err)
			return
		}
	}

	httpx.JSON(w, http.StatusOK, dto.AdvertiserWithContactsToProfile(adv, email, phone))
}

func hasAdvertiserProfileChanges(req *dto.UpdateAdvertiserProfileRequest) bool {
	return req.Name != nil || req.Surname != nil || req.Company != nil || req.City != nil || req.Tariff != nil
}

func (a *API) resolveUpdatedContacts(
	ctx context.Context,
	advertiserID int64,
	req *dto.UpdateAdvertiserProfileRequest,
) (string, string, error) {
	if req.Email == nil && req.Phone == nil {
		return a.authClient.GetCredentials(ctx, advertiserID)
	}

	email, phone, err := a.authClient.GetCredentials(ctx, advertiserID)
	if err != nil {
		return "", "", err
	}

	if req.Email != nil {
		email = *req.Email
	}
	if req.Phone != nil {
		phone = *req.Phone
	}

	return a.authClient.UpdateCredentials(ctx, advertiserID, email, phone)
}

func (a *API) ensureAdvertiserProfile(ctx context.Context, advertiserID int64, preferredName, preferredSurname, email string) error {
	if _, err := a.service.GetAdvertiserByID(ctx, int(advertiserID)); err == nil {
		return nil
	} else if !errors.Is(err, errs.NotFoundError) {
		return err
	}

	if err := a.service.CreateAdvertiserProfile(ctx, advertiserID, preferredName, email); err != nil {
		return err
	}
	if preferredSurname == "" {
		return nil
	}
	_, err := a.service.UpdateAdvertiserProfile(ctx, &serviceinput.UpdateAdvertiserProfile{
		AdvertiserID: int(advertiserID),
		Surname:      &preferredSurname,
	})
	return err
}

// @Summary      Выход рекламодателя
// @Description  Завершает сессию текущего рекламодателя
// @Tags         advertiser
// @Produce      json
// @Success      200   {object}  map[string]string
// @Failure      500   {object}  httpx.Error
// @Router       /advertisers/logout [post]
func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(a.cookieConfig.Name)
	if err != nil {
		httpx.JSON(w, http.StatusOK, map[string]string{"message": "logout ok"})
		return
	}

	if err := a.authClient.Logout(r.Context(), cookie.Value); err != nil {
		handler.HandleError(w, r, "destroying session", err)
		return
	}

	a.clearSessionCookie(w)

	httpx.JSON(w, http.StatusOK, map[string]string{
		"message": "logout ok",
	})
}

// @Summary      Баланс рекламодателя
// @Description  Возвращает текущий баланс рекламодателя по сессии
// @Tags         advertiser
// @Produce      json
// @Success      200  {object}  dto.BalanceResponse
// @Failure      401  {object}  httpx.Error
// @Failure      404  {object}  httpx.Error
// @Failure      500  {object}  httpx.Error
// @Router       /advertisers/balance [get]
// @Security     CookieAuth
func (a *API) GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser id from ctx", err)
		return
	}

	adv, err := a.service.GetAdvertiserByID(ctx, advertiserID)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser by id", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.BalanceResponse{
		Balance: adv.Balance,
	})
}

// @Summary      Пополнение баланса
// @Description  Пополняет баланс текущего рекламодателя
// @Tags         advertiser
// @Accept       json
// @Produce      json
// @Param        body  body      dto.TopUpBalanceRequest  true  "Сумма пополнения"
// @Success      200   {object}  dto.BalanceResponse
// @Failure      400   {object}  httpx.Error
// @Failure      401   {object}  httpx.Error
// @Failure      500   {object}  httpx.Error
// @Router       /advertisers/balance/topup [post]
// @Security     CookieAuth
func (a *API) TopUpBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser id from ctx", err)
		return
	}

	req, err := newJSONRequest[dto.TopUpBalanceRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	balance, err := a.service.TopUpAdvertiserBalance(ctx, advertiserID, req.Amount)
	if err != nil {
		handler.HandleError(w, r, "topping up advertiser balance", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.BalanceResponse{
		Balance: balance,
	})
}

// @Summary      Обновление аватара рекламодателя
// @Description  Загружает JPEG, PNG или WEBP аватар текущего рекламодателя
// @Tags         advertiser
// @Accept       multipart/form-data
// @Produce      json
// @Param        avatar  formData  file  true  "Файл аватара"
// @Success      200     {object}  dto.AdvertiserProfileResponse
// @Failure      400     {object}  httpx.Error
// @Failure      401     {object}  httpx.Error
// @Failure      500     {object}  httpx.Error
// @Router       /advertisers/me/avatar [put]
// @Security     CookieAuth
func (a *API) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser id from ctx", err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarSize)
	if err = r.ParseMultipartForm(maxAvatarSize); err != nil {
		handler.HandleError(w, r, "parsing multipart form", err)
		return
	}

	file, fileHeader, err := r.FormFile("avatar")
	if err != nil {
		handler.HandleError(w, r, "uploading avatar", err)
		return
	}
	defer file.Close()

	uploaded, err := ParseAndValidateImage(fileHeader)
	if err != nil {
		handler.HandleError(w, r, "parsing and validating image", err)
		return
	}

	adv, err := a.service.UpdateAdvertiserAvatar(
		ctx,
		advertiserID,
		uploaded.Data,
		uploaded.Ext,
		uploaded.ContentType,
	)
	if err != nil {
		handler.HandleError(w, r, "updating advertiser avatar", err)
		return
	}

	email, phone, err := a.authClient.GetCredentials(ctx, int64(advertiserID))
	if err != nil {
		handler.HandleError(w, r, "getting advertiser credentials", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.AdvertiserWithContactsToProfile(adv, email, phone))
}

func (a *API) setSessionCookie(w http.ResponseWriter, sessionID string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.cookieConfig.Name,
		Value:    sessionID,
		Path:     a.cookieConfig.Path,
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: a.cookieConfig.HTTPOnly,
		Secure:   a.cookieConfig.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *API) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.cookieConfig.Name,
		Value:    "",
		Path:     a.cookieConfig.Path,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: a.cookieConfig.HTTPOnly,
		Secure:   a.cookieConfig.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}
