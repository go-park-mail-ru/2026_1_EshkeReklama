package handlers

import (
	"database/sql"
	"errors"
	"eshkere/internal/handler/dto"
	"eshkere/internal/middleware"
	"eshkere/pkg/httpx"
	"eshkere/pkg/logger"
	"net/http"

	"github.com/gorilla/mux"
)

func (a *API) RegisterAdvertiserHandlers(r *mux.Router) {
	groups := r.PathPrefix("/advertiser").Subrouter()

	groups.HandleFunc("/register", a.Register).Methods(http.MethodPost)
	groups.HandleFunc("/login", a.Login).Methods(http.MethodPost)
	groups.HandleFunc("/logout", a.Logout).Methods(http.MethodPost)
	groups.Handle("/balance", middleware.Auth(a.sessionManager)(http.HandlerFunc(a.GetBalance))).Methods(http.MethodGet)
	groups.Handle("/balance/topup", middleware.Auth(a.sessionManager)(http.HandlerFunc(a.TopUpBalance))).Methods(http.MethodPost)

	groups.Handle("/me", middleware.Auth(a.sessionManager)(http.HandlerFunc(a.Me))).Methods(http.MethodGet)
	groups.Handle("/me", middleware.Auth(a.sessionManager)(http.HandlerFunc(a.UpdateProfile))).Methods(http.MethodPut)
	groups.Handle("/me/avatar", middleware.Auth(a.sessionManager)(http.HandlerFunc(a.UpdateAvatar))).Methods(http.MethodPut)

	groups.Handle("/feed-link", middleware.Auth(a.sessionManager)(http.HandlerFunc(a.GenerateFeedLink))).Methods(http.MethodPost)
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
	reqLogger := logger.GetLoggerFromCtx(ctx)

	req, err := newJSONRequest[dto.RegisterRequest](r)
	if err != nil {
		reqLogger.Warnw("invalid register payload", "error", err.Error())
		httpx.BadRequest(w, "invalid request")
		return
	}

	adv, err := a.service.RegisterAdvertiser(ctx, req.Name, req.Email, req.Phone, req.Password)
	if err != nil {
		statusCode, clientMessage, isExpected := convertDomainError(err)
		if isExpected {
			reqLogger.Warnw("register failed", "error", err.Error())
		} else {
			reqLogger.Errorw("register failed", "error", err.Error())
		}
		httpx.ErrorJSON(w, statusCode, clientMessage)
		return
	}

	if err = a.sessionManager.Create(w, r, adv.ID); err != nil {
		reqLogger.Errorw("failed to create session", "error", err.Error(), "advertiser_id", adv.ID)
		httpx.InternalError(w, "internal error")
		return
	}

	httpx.JSON(w, http.StatusOK, dto.RegisterResponse{
		ID:    adv.ID,
		Email: adv.Email,
		Phone: adv.Phone,
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
// @Router       /advertiser/login [post]
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqLogger := logger.GetLoggerFromCtx(ctx)

	req, err := newJSONRequest[dto.LoginRequest](r)
	if err != nil {
		reqLogger.Warnw("invalid login payload", "error", err.Error())
		httpx.BadRequest(w, "invalid request")
		return
	}

	adv, err := a.service.AuthenticateAdvertiser(ctx, req.Identifier, req.Password)
	if err != nil {
		statusCode, clientMessage, isExpected := convertDomainError(err)
		if isExpected {
			reqLogger.Warnw("login failed", "error", err.Error(), "identifier", req.Identifier)
		} else {
			reqLogger.Errorw("login failed", "error", err.Error(), "identifier", req.Identifier)
		}
		httpx.ErrorJSON(w, statusCode, clientMessage)
		return
	}

	if err = a.sessionManager.Create(w, r, adv.ID); err != nil {
		reqLogger.Errorw("failed to create session", "error", err.Error(), "advertiser_id", adv.ID)
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
	reqLogger := logger.GetLoggerFromCtx(ctx)

	advertiserID, err := middleware.AdvertiserIDFromContext(ctx)
	if err != nil {
		reqLogger.Warnw("unauthorized profile request", "error", err.Error())
		httpx.Unauthorized(w, "unauthorized")
		return
	}

	adv, err := a.service.GetAdvertiserByID(ctx, advertiserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			reqLogger.Warnw("advertiser not found", "error", err.Error(), "advertiser_id", advertiserID)
			httpx.NotFound(w, "advertiser not found")
			return
		}
		reqLogger.Errorw("failed to load advertiser profile", "error", err.Error(), "advertiser_id", advertiserID)
		httpx.InternalError(w, "internal error")
		return
	}

	httpx.JSON(w, http.StatusOK, dto.AdvertiserToProfile(adv))
}

// @Summary      Обновление профиля рекламодателя
// @Description  Обновляет данные текущего рекламодателя по сессии
// @Tags         advertiser
// @Accept       multipart/form-data
// @Produce      json
// @Param        name   formData  string  false  "Имя рекламодателя"
// @Param        email  formData  string  false  "Email рекламодателя"
// @Param        phone  formData  string  false  "Телефон рекламодателя"
// @Success      200    {object}  dto.AdvertiserProfileResponse
// @Failure      400    {object}  httpx.Error
// @Failure      401    {object}  httpx.Error
// @Failure      500    {object}  httpx.Error
// @Router       /advertiser/me [put]
// @Security     CookieAuth
func (a *API) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqLogger := logger.GetLoggerFromCtx(ctx)

	advertiserID, err := middleware.AdvertiserIDFromContext(ctx)
	if err != nil {
		reqLogger.Warnw("unauthorized update profile request", "error", err.Error())
		httpx.Unauthorized(w, "unauthorized")
		return
	}

	if err = r.ParseMultipartForm(5 << 20); err != nil {
		reqLogger.Warnw("invalid multipart payload", "error", err.Error(), "advertiser_id", advertiserID)
		httpx.BadRequest(w, "invalid multipart payload")
		return
	}

	adv, err := a.service.UpdateAdvertiserProfile(
		ctx,
		advertiserID,
		r.FormValue("name"),
		r.FormValue("email"),
		r.FormValue("phone"),
	)
	if err != nil {
		statusCode, clientMessage, isExpected := convertDomainError(err)
		if isExpected {
			reqLogger.Warnw("update profile rejected", "error", err.Error(), "advertiser_id", advertiserID)
		} else {
			reqLogger.Errorw("update profile failed", "error", err.Error(), "advertiser_id", advertiserID)
		}
		httpx.ErrorJSON(w, statusCode, clientMessage)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.AdvertiserToProfile(adv))
}

// @Summary      Генерация уникальной feed-ссылки
// @Description  Создает или обновляет уникальную публичную ссылку на feed объявлений текущего рекламодателя
// @Tags         advertiser
// @Produce      json
// @Success      200  {object}  dto.FeedLinkResponse
// @Failure      401  {object}  httpx.Error
// @Failure      500  {object}  httpx.Error
// @Router       /advertiser/feed-link [post]
// @Security     CookieAuth
func (a *API) GenerateFeedLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqLogger := logger.GetLoggerFromCtx(ctx)

	advertiserID, err := middleware.AdvertiserIDFromContext(ctx)
	if err != nil {
		reqLogger.Warnw("unauthorized generate feed link request", "error", err.Error())
		httpx.Unauthorized(w, "unauthorized")
		return
	}

	token, err := a.service.GenerateFeedLink(ctx, advertiserID)
	if err != nil {
		reqLogger.Errorw("failed to generate feed link", "error", err.Error(), "advertiser_id", advertiserID)
		httpx.InternalError(w, "internal error")
		return
	}

	httpx.JSON(w, http.StatusOK, dto.FeedLinkResponse{
		URL: "/feed/" + token,
	})
}

// @Summary      Выход рекламодателя
// @Description  Завершает сессию текущего рекламодателя
// @Tags         advertiser
// @Produce      json
// @Success      200   {object}  map[string]string
// @Failure      500   {object}  httpx.Error
// @Router       /advertiser/logout [post]
func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	reqLogger := logger.GetLoggerFromCtx(r.Context())

	if err := a.sessionManager.Destroy(w, r); err != nil {
		reqLogger.Errorw("failed to destroy session", "error", err.Error())
		httpx.InternalError(w, "internal error")
		return
	}

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
// @Router       /advertiser/balance [get]
// @Security     CookieAuth
func (a *API) GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqLogger := logger.GetLoggerFromCtx(ctx)

	advertiserID, err := middleware.AdvertiserIDFromContext(ctx)
	if err != nil {
		reqLogger.Warnw("unauthorized get balance request", "error", err.Error())
		httpx.Unauthorized(w, "unauthorized")
		return
	}

	adv, err := a.service.GetAdvertiserByID(ctx, advertiserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			reqLogger.Warnw("advertiser not found for balance", "error", err.Error(), "advertiser_id", advertiserID)
			httpx.NotFound(w, "advertiser not found")
			return
		}

		reqLogger.Errorw("failed to get balance", "error", err.Error(), "advertiser_id", advertiserID)
		httpx.InternalError(w, "internal error")
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
// @Router       /advertiser/balance/topup [post]
// @Security     CookieAuth
func (a *API) TopUpBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqLogger := logger.GetLoggerFromCtx(ctx)

	advertiserID, err := middleware.AdvertiserIDFromContext(ctx)
	if err != nil {
		reqLogger.Warnw("unauthorized top up request", "error", err.Error())
		httpx.Unauthorized(w, "unauthorized")
		return
	}

	req, err := newJSONRequest[dto.TopUpBalanceRequest](r)
	if err != nil {
		reqLogger.Warnw("invalid top up payload", "error", err.Error(), "advertiser_id", advertiserID)
		httpx.BadRequest(w, "invalid request")
		return
	}

	balance, err := a.service.TopUpAdvertiserBalance(ctx, advertiserID, req.Amount)
	if err != nil {
		statusCode, clientMessage, isExpected := convertDomainError(err)
		if isExpected {
			reqLogger.Warnw("balance top up rejected", "error", err.Error(), "advertiser_id", advertiserID, "amount", req.Amount)
		} else {
			reqLogger.Errorw("balance top up failed", "error", err.Error(), "advertiser_id", advertiserID, "amount", req.Amount)
		}
		httpx.ErrorJSON(w, statusCode, clientMessage)
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
// @Router       /advertiser/me/avatar [put]
// @Security     CookieAuth
func (a *API) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqLogger := logger.GetLoggerFromCtx(ctx)

	advertiserID, err := middleware.AdvertiserIDFromContext(ctx)
	if err != nil {
		reqLogger.Warnw("unauthorized update avatar request", "error", err.Error())
		httpx.Unauthorized(w, "unauthorized")
		return
	}

	if err = r.ParseMultipartForm(maxAvatarSize); err != nil {
		reqLogger.Warnw("invalid multipart payload", "error", err.Error(), "advertiser_id", advertiserID)
		httpx.BadRequest(w, "invalid multipart payload")
		return
	}

	file, fileHeader, err := r.FormFile("avatar")
	if err != nil {
		reqLogger.Warnw("avatar file is required", "error", err.Error(), "advertiser_id", advertiserID)
		httpx.BadRequest(w, "avatar file is required")
		return
	}
	defer file.Close()

	uploaded, err := ParseAndValidateImage(fileHeader)
	if err != nil {
		reqLogger.Warnw("invalid avatar file", "error", err.Error(), "advertiser_id", advertiserID)
		httpx.BadRequest(w, "invalid avatar file")
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
		statusCode, clientMessage, isExpected := convertDomainError(err)
		if isExpected {
			reqLogger.Warnw("update avatar rejected", "error", err.Error(), "advertiser_id", advertiserID)
		} else {
			reqLogger.Errorw("update avatar failed", "error", err.Error(), "advertiser_id", advertiserID)
		}
		httpx.ErrorJSON(w, statusCode, clientMessage)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.AdvertiserToProfile(adv))
}
