package v1

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	authsvc "eshkere/internal/auth/service"
	errs "eshkere/internal/errors"
	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/internal/models"
	redisrepo "eshkere/internal/repository/redis"
	svc "eshkere/internal/service"
	serviceinput "eshkere/internal/service/input"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"

	goredis "github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
)

func (a *API) RegisterAdvertiserHandlers(r *mux.Router) {
	advertisers := r.PathPrefix("/advertisers").Subrouter()

	advertisers.HandleFunc("/register", a.Register).Methods(http.MethodPost)
	advertisers.HandleFunc("/register/verify", a.VerifyRegistration).Methods(http.MethodPost)
	advertisers.HandleFunc("/password/reset", a.RequestPasswordReset).Methods(http.MethodPost)
	advertisers.HandleFunc("/password/reset/confirm", a.ConfirmPasswordReset).Methods(http.MethodPost)
	advertisers.HandleFunc("/login", a.Login).Methods(http.MethodPost)
	advertisers.HandleFunc("/login/vk", a.LoginVKID).Methods(http.MethodPost)
	advertisers.HandleFunc("/logout", a.Logout).Methods(http.MethodPost)
	advertisers.Handle("/balance", middleware.Auth(a.authClient, a.cookieConfig.Name)(http.HandlerFunc(a.GetBalance))).Methods(http.MethodGet)
	advertisers.Handle("/balance/topup", middleware.Auth(a.authClient, a.cookieConfig.Name)(http.HandlerFunc(a.TopUpBalance))).Methods(http.MethodPost)
	advertisers.Handle("/balance/payment/create", middleware.Auth(a.authClient, a.cookieConfig.Name)(http.HandlerFunc(a.CreateBalancePayment))).Methods(http.MethodPost)
	advertisers.Handle("/balance/autopay", middleware.Auth(a.authClient, a.cookieConfig.Name)(http.HandlerFunc(a.GetAutopaySettings))).Methods(http.MethodGet)
	advertisers.Handle("/balance/autopay", middleware.Auth(a.authClient, a.cookieConfig.Name)(http.HandlerFunc(a.UpdateAutopaySettings))).Methods(http.MethodPost)
	advertisers.Handle("/notification-settings", middleware.Auth(a.authClient, a.cookieConfig.Name)(http.HandlerFunc(a.GetNotificationSettings))).Methods(http.MethodGet)
	advertisers.Handle("/notification-settings", middleware.Auth(a.authClient, a.cookieConfig.Name)(http.HandlerFunc(a.UpdateNotificationSettings))).Methods(http.MethodPut)

	advertisers.Handle("/me", middleware.Auth(a.authClient, a.cookieConfig.Name)(http.HandlerFunc(a.Me))).Methods(http.MethodGet)
	advertisers.Handle("/me", middleware.Auth(a.authClient, a.cookieConfig.Name)(http.HandlerFunc(a.UpdateProfile))).Methods(http.MethodPut)
	advertisers.Handle("/me/password", middleware.Auth(a.authClient, a.cookieConfig.Name)(http.HandlerFunc(a.ChangePassword))).Methods(http.MethodPut)
	advertisers.Handle("/me/avatar", middleware.Auth(a.authClient, a.cookieConfig.Name)(http.HandlerFunc(a.UpdateAvatar))).Methods(http.MethodPut)
}

// @Summary      Регистрация рекламодателя
// @Description  Создает новый аккаунт и отправляет код подтверждения на email
// @Tags         advertiser
// @Accept       json
// @Produce      json
// @Param        input body      dto.RegisterRequest  true  "Данные для регистрации"
// @Success      200   {object}  dto.RegisterResponse
// @Failure      400   {object}  httpx.Error "Invalid request или User already exists"
// @Failure      500   {object}  httpx.Error
// @Router       /api/advertisers/register [post]
func (a *API) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := newJSONRequest[dto.RegisterRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	email := normalizeVerificationEmail(req.Email)

	advID, sessionID, _, err := a.authClient.Register(ctx, email, req.Phone, req.Password)
	if err != nil {
		if errors.Is(err, errs.ErrEmailTaken) {
			if resendErr := a.resendRegistrationCode(ctx, email); resendErr == nil {
				httpx.JSON(w, http.StatusAccepted, dto.RegisterResponse{
					Email:                email,
					Phone:                req.Phone,
					VerificationRequired: true,
					Message:              "Код подтверждения отправлен повторно",
				})
				return
			}
		}
		handler.HandleError(w, r, "register advertiser", err)
		return
	}

	if err = a.service.CreateAdvertiserProfile(ctx, advID, req.Name, email); err != nil {
		// компенсирующая операция: откатываем credentials в auth-сервисе
		_ = a.authClient.Logout(ctx, sessionID)
		handler.HandleError(w, r, "create advertiser profile", err)
		return
	}

	_ = a.authClient.Logout(ctx, sessionID)

	code, err := generateVerificationCode()
	if err != nil {
		handler.HandleError(w, r, "generate verification code", err)
		return
	}
	if err := a.saveAndSendVerificationCode(ctx, advID, email, code); err != nil {
		handler.HandleError(w, r, "send verification email", err)
		return
	}

	httpx.JSON(w, http.StatusAccepted, dto.RegisterResponse{
		ID:                   int(advID),
		Email:                email,
		Phone:                req.Phone,
		VerificationRequired: true,
		Message:              "Код подтверждения отправлен на почту",
	})
}

func (a *API) VerifyRegistration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := newJSONRequest[dto.VerifyRegistrationRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	email := normalizeVerificationEmail(req.Email)
	record, err := a.verificationStore.Get(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			handler.HandleError(w, r, "verify registration", errs.ErrInvalidVerifyCode)
			return
		}
		if err == goredis.ErrNil {
			handler.HandleError(w, r, "verify registration", errs.ErrInvalidVerifyCode)
			return
		}
		if errors.Is(err, goredis.ErrNil) {
			handler.HandleError(w, r, "verify registration", errs.ErrInvalidVerifyCode)
			return
		}
		handler.HandleError(w, r, "verify registration", err)
		return
	}
	if record.Code != strings.TrimSpace(req.Code) {
		handler.HandleError(w, r, "verify registration", errs.ErrInvalidVerifyCode)
		return
	}
	if err := a.verificationStore.Delete(ctx, email); err != nil {
		handler.HandleError(w, r, "verify registration", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.VerifyRegistrationResponse{
		Email:    email,
		Verified: true,
		Message:  "Почта успешно подтверждена",
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
// @Router       /api/advertisers/login [post]
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

	email, phone, _, err := a.authClient.GetCredentials(ctx, advID)
	if err != nil {
		handler.HandleError(w, r, "get advertiser credentials", err)
		return
	}
	if err := a.ensureEmailVerified(ctx, email, sessionID); err != nil {
		handler.HandleError(w, r, "verify advertiser email", err)
		return
	}

	a.setSessionCookie(w, sessionID, time.Unix(expiresAt, 0))

	httpx.JSON(w, http.StatusOK, dto.LoginResponse{
		ID:    int(advID),
		Email: email,
		Phone: phone,
	})
}

func (a *API) ensureEmailVerified(ctx context.Context, email, sessionID string) error {
	if a.verificationStore == nil {
		return nil
	}

	_, err := a.verificationStore.Get(ctx, email)
	if err == nil {
		_ = a.authClient.Logout(ctx, sessionID)
		return errs.ErrEmailNotVerified
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err == goredis.ErrNil || errors.Is(err, goredis.ErrNil) {
		return nil
	}
	return err
}

func (a *API) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := newJSONRequest[dto.PasswordResetRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	if a.credentialsManager == nil || a.passwordResetStore == nil || a.verificationEmailSender == nil {
		handler.HandleError(w, r, "request password reset", fmt.Errorf("password reset is not configured"))
		return
	}

	cred, err := a.credentialsManager.FindByIdentifier(ctx, req.Identifier)
	if err != nil {
		switch {
		case errors.Is(err, authsvc.ErrInvalidArg):
			handler.HandleError(w, r, "request password reset", errs.ErrInvalidAdvertiserArg)
			return
		case errors.Is(err, sql.ErrNoRows):
			httpx.JSON(w, http.StatusOK, dto.PasswordResetResponse{Message: "Если аккаунт существует, код для восстановления отправлен на почту"})
			return
		default:
			handler.HandleError(w, r, "request password reset", err)
			return
		}
	}

	if strings.TrimSpace(cred.Email) == "" {
		httpx.JSON(w, http.StatusOK, dto.PasswordResetResponse{Message: "Если аккаунт существует, код для восстановления отправлен на почту"})
		return
	}

	code, err := generateVerificationCode()
	if err != nil {
		handler.HandleError(w, r, "request password reset", err)
		return
	}

	if err := a.passwordResetStore.Save(ctx, redisrepo.PasswordResetRecord{
		AdvertiserID: cred.ID,
		Email:        cred.Email,
		Code:         code,
	}, a.passwordResetTTL); err != nil {
		handler.HandleError(w, r, "request password reset", err)
		return
	}
	if err := a.verificationEmailSender.SendPasswordResetCode(ctx, cred.Email, code); err != nil {
		handler.HandleError(w, r, "request password reset", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.PasswordResetResponse{Message: "Если аккаунт существует, код для восстановления отправлен на почту"})
}

func (a *API) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := newJSONRequest[dto.ConfirmPasswordResetRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	if a.credentialsManager == nil || a.passwordResetStore == nil {
		handler.HandleError(w, r, "confirm password reset", fmt.Errorf("password reset is not configured"))
		return
	}

	cred, err := a.credentialsManager.FindByIdentifier(ctx, req.Identifier)
	if err != nil {
		switch {
		case errors.Is(err, authsvc.ErrInvalidArg):
			handler.HandleError(w, r, "confirm password reset", errs.ErrInvalidAdvertiserArg)
			return
		case errors.Is(err, sql.ErrNoRows):
			handler.HandleError(w, r, "confirm password reset", errs.ErrInvalidVerifyCode)
			return
		default:
			handler.HandleError(w, r, "confirm password reset", err)
			return
		}
	}

	record, err := a.passwordResetStore.Get(ctx, cred.ID)
	if err != nil {
		if err == goredis.ErrNil || errors.Is(err, goredis.ErrNil) {
			handler.HandleError(w, r, "confirm password reset", errs.ErrInvalidVerifyCode)
			return
		}
		handler.HandleError(w, r, "confirm password reset", err)
		return
	}
	if record.Code != strings.TrimSpace(req.Code) {
		handler.HandleError(w, r, "confirm password reset", errs.ErrInvalidVerifyCode)
		return
	}

	if err := a.credentialsManager.SetPassword(ctx, cred.ID, req.NewPassword); err != nil {
		if errors.Is(err, authsvc.ErrInvalidArg) {
			handler.HandleError(w, r, "confirm password reset", errs.ErrInvalidAdvertiserArg)
			return
		}
		handler.HandleError(w, r, "confirm password reset", err)
		return
	}
	if err := a.passwordResetStore.Delete(ctx, cred.ID); err != nil {
		handler.HandleError(w, r, "confirm password reset", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.PasswordResetResponse{Message: "Пароль успешно обновлен"})
}

func (a *API) resendRegistrationCode(ctx context.Context, email string) error {
	record, err := a.verificationStore.Get(ctx, email)
	if err != nil {
		return err
	}

	code, err := generateVerificationCode()
	if err != nil {
		return err
	}
	return a.saveAndSendVerificationCode(ctx, record.AdvertiserID, email, code)
}

func (a *API) saveAndSendVerificationCode(ctx context.Context, advertiserID int64, email, code string) error {
	if a.verificationStore == nil || a.verificationEmailSender == nil {
		return fmt.Errorf("email verification is not configured")
	}
	if err := a.verificationStore.Save(ctx, redisrepo.EmailVerificationRecord{
		AdvertiserID: advertiserID,
		Email:        email,
		Code:         code,
	}, a.registrationVerifyTTL); err != nil {
		return err
	}
	return a.verificationEmailSender.SendEmailVerificationCode(ctx, email, code)
}

func generateVerificationCode() (string, error) {
	var code strings.Builder
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("generate verification digit: %w", err)
		}
		code.WriteByte(byte('0' + n.Int64()))
	}
	return code.String(), nil
}

func normalizeVerificationEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
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
// @Router       /api/advertisers/login/vk [post]
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

	email, phone, _, err := a.authClient.GetCredentials(ctx, advID)
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
// @Router       /api/advertisers/me [get]
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

	email, phone, canChangePassword, err := a.authClient.GetCredentials(ctx, int64(advertiserID))
	if err != nil {
		handler.HandleError(w, r, "getting advertiser credentials", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.AdvertiserWithContactsToProfile(adv, email, phone, canChangePassword))
}

// @Summary      Смена пароля рекламодателя
// @Description  Обновляет пароль текущего рекламодателя по сессии
// @Tags         advertiser
// @Accept       json
// @Produce      json
// @Param        body   body      dto.ChangePasswordRequest  true  "Текущий и новый пароль"
// @Success      204
// @Failure      400    {object}  httpx.Error
// @Failure      401    {object}  httpx.Error
// @Failure      422    {object}  httpx.Error "Смена пароля недоступна для VK ID аккаунтов без локального пароля"
// @Router       /advertisers/me/password [put]
// @Security     CookieAuth
func (a *API) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "getting advertiser id from ctx", err)
		return
	}

	req, err := newJSONRequest[dto.ChangePasswordRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	if err := a.authClient.ChangePassword(ctx, int64(advertiserID), req.CurrentPassword, req.NewPassword); err != nil {
		handler.HandleError(w, r, "changing advertiser password", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
// @Router       /api/advertisers/me [put]
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

	_, _, canChangePassword, err := a.authClient.GetCredentials(ctx, int64(advertiserID))
	if err != nil {
		handler.HandleError(w, r, "getting advertiser credentials", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.AdvertiserWithContactsToProfile(adv, email, phone, canChangePassword))
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
		email, phone, _, err := a.authClient.GetCredentials(ctx, advertiserID)
		return email, phone, err
	}

	email, phone, _, err := a.authClient.GetCredentials(ctx, advertiserID)
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
// @Router       /api/advertisers/logout [post]
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
// @Router       /api/advertisers/balance [get]
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

	campaigns, err := a.service.ListAdCampaigns(ctx, advertiserID)
	if err != nil {
		handler.HandleError(w, r, "listing advertiser campaigns", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.BalanceResponse{
		Balance:                 adv.Balance,
		SavedPaymentMethodID:    adv.SavedPaymentMethodID.String,
		SavedPaymentMethodTitle: adv.SavedPaymentMethodTitle.String,
		DeliveryAlert:           deliveryAlertResponse(svc.BuildAdvertiserDeliveryAlert(adv.Balance, campaigns)),
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
// @Router       /api/advertisers/balance/topup [post]
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

	campaigns, err := a.service.ListAdCampaigns(ctx, advertiserID)
	if err != nil {
		handler.HandleError(w, r, "listing advertiser campaigns", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.BalanceResponse{
		Balance:       balance,
		DeliveryAlert: deliveryAlertResponse(svc.BuildAdvertiserDeliveryAlert(balance, campaigns)),
	})
}

func deliveryAlertResponse(alert *svc.AdvertiserDeliveryAlert) *dto.DeliveryAlertResponse {
	if alert == nil {
		return nil
	}

	response := &dto.DeliveryAlertResponse{
		Level:             string(alert.Level),
		ActiveCampaigns:   alert.ActiveCampaigns,
		AffectedCampaigns: alert.AffectedCampaigns,
	}

	switch alert.Level {
	case svc.DeliveryAlertLowBalance:
		response.Title = "Баланс на исходе"
		response.Message = "Показы идут, но средств мало. Пополните баланс заранее, чтобы не прерывать рекламу."
	case svc.DeliveryAlertAtRisk:
		response.Title = "Реклама под риском остановки"
		response.Message = "Активные кампании еще работают, но при текущем балансе могут скоро остановиться."
	case svc.DeliveryAlertPartiallyStopped:
		response.Title = "Часть кампаний остановлена"
		response.Message = "У части активных кампаний уже недостаточно средств для продолжения показов."
	case svc.DeliveryAlertFullyStopped:
		response.Title = "Реклама остановлена из-за нехватки средств"
		response.Message = "У активных кампаний недостаточно средств. Пополните баланс, чтобы возобновить показы."
	}

	return response
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
// @Router       /api/advertisers/me/avatar [put]
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

	email, phone, canChangePassword, err := a.authClient.GetCredentials(ctx, int64(advertiserID))
	if err != nil {
		handler.HandleError(w, r, "getting advertiser credentials", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.AdvertiserWithContactsToProfile(adv, email, phone, canChangePassword))
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
