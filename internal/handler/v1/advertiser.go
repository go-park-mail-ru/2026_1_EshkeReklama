package v1

import (
	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"
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

	req, err := newJSONRequest[dto.RegisterRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	adv, err := a.service.RegisterAdvertiser(ctx, req.Name, req.Email, req.Phone, req.Password)
	if err != nil {
		handler.HandleError(w, r, "register advertiser", err)
		return
	}

	if err = a.sessionManager.Create(w, r, adv.ID); err != nil {
		handler.HandleError(w, r, "creating session", err)
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

	req, err := newJSONRequest[dto.LoginRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	adv, err := a.service.AuthenticateAdvertiser(ctx, req.Identifier, req.Password)
	if err != nil {
		handler.HandleError(w, r, "auth advertiser", err)
		return
	}

	if err = a.sessionManager.Create(w, r, adv.ID); err != nil {
		handler.HandleError(w, r, "creating session", err)
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

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser id from ctx", err)
		return
	}

	adv, err := a.service.GetAdvertiserByID(ctx, advertiserID)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser by id", err)
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

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser id from ctx", err)
		return
	}

	if err = r.ParseMultipartForm(5 << 20); err != nil {
		handler.HandleError(w, r, "parsing multipart form", err)
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
		handler.HandleError(w, r, "updating profile", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.AdvertiserToProfile(adv))
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
		handler.HandleError(w, r, "destroying session", err)
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
// @Router       /advertiser/balance/topup [post]
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
// @Router       /advertiser/me/avatar [put]
// @Security     CookieAuth
func (a *API) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser id from ctx", err)
		return
	}

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

	httpx.JSON(w, http.StatusOK, dto.AdvertiserToProfile(adv))
}
