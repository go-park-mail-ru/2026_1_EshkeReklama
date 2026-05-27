package dto

import (
	"time"

	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID                   int    `json:"id"`
	Email                string `json:"email"`
	Phone                string `json:"phone"`
	VerificationRequired bool   `json:"verification_required"`
	Message              string `json:"message"`
}

type VerifyRegistrationRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type VerifyRegistrationResponse struct {
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
	Message  string `json:"message"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type VKIDLoginRequest struct {
	AccessToken string `json:"access_token"`
	UserID      int64  `json:"user_id"`
}

type LoginResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type UpdateAdvertiserProfileRequest struct {
	Name    *string `json:"name"`
	Surname *string `json:"surname"`
	Email   *string `json:"email"`
	Phone   *string `json:"phone"`
	Company *string `json:"company"`
	City    *string `json:"city"`
	Tariff  *string `json:"tariff"`
}

func (u *UpdateAdvertiserProfileRequest) ToInput(advertiserID int) *serviceinput.UpdateAdvertiserProfile {
	return &serviceinput.UpdateAdvertiserProfile{
		AdvertiserID: advertiserID,
		Name:         u.Name,
		Surname:      u.Surname,
		Company:      u.Company,
		City:         u.City,
		Tariff:       u.Tariff,
	}
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}

// AdvertiserProfileResponse — публичные поля рекламодателя (без пароля).
type AdvertiserProfileResponse struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Surname           string `json:"surname"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	AvatarURL         string `json:"avatar_url,omitempty"`
	Balance           int64  `json:"balance"`
	Company           string `json:"company"`
	City              string `json:"city"`
	Tariff            string `json:"tariff"`
	Role              string `json:"role"`
	CanChangePassword bool   `json:"can_change_password"`
	CreatedAt         string `json:"created_at"`
}

type TopUpBalanceRequest struct {
	Amount int64 `json:"amount"`
}

type CreatePaymentRequest struct {
	Amount int64 `json:"amount"`
}

type CreatePaymentResponse struct {
	PaymentURL string `json:"payment_url"`
}

type AutopaySettingsRequest struct {
	Enabled   bool  `json:"enabled"`
	Threshold int64 `json:"threshold"`
	Limit     int64 `json:"limit"`
}

type AutopaySettingsResponse struct {
	Enabled   bool  `json:"enabled"`
	Threshold int64 `json:"threshold"`
	Limit     int64 `json:"limit"`
}

type NotificationSettingsRequest struct {
	EmailEnabled      bool  `json:"email_enabled"`
	WarningThreshold  int64 `json:"warning_threshold"`
	CriticalThreshold int64 `json:"critical_threshold"`
}

type NotificationSettingsResponse struct {
	EmailEnabled      bool  `json:"email_enabled"`
	WarningThreshold  int64 `json:"warning_threshold"`
	CriticalThreshold int64 `json:"critical_threshold"`
}

type DeliveryAlertResponse struct {
	Level             string `json:"level"`
	Title             string `json:"title"`
	Message           string `json:"message"`
	ActiveCampaigns   int    `json:"active_campaigns"`
	AffectedCampaigns int    `json:"affected_campaigns"`
}

type BalanceResponse struct {
	Balance                 int64                  `json:"balance"`
	SavedPaymentMethodID    string                 `json:"saved_payment_method_id,omitempty"`
	SavedPaymentMethodTitle string                 `json:"saved_payment_method_title,omitempty"`
	DeliveryAlert           *DeliveryAlertResponse `json:"delivery_alert,omitempty"`
}

func AdvertiserToProfile(adv *models.Advertiser) AdvertiserProfileResponse {
	return AdvertiserWithContactsToProfile(adv, "", "", false)
}

func AdvertiserWithContactsToProfile(adv *models.Advertiser, email, phone string, canChangePassword bool) AdvertiserProfileResponse {
	if adv == nil {
		return AdvertiserProfileResponse{}
	}
	return AdvertiserProfileResponse{
		ID:                adv.ID,
		Name:              adv.Name,
		Surname:           adv.Surname.String,
		Email:             email,
		Phone:             phone,
		AvatarURL:         adv.AvatarURL.String,
		Balance:           adv.Balance,
		Company:           adv.Company.String,
		City:              adv.City.String,
		Tariff:            string(adv.Tariff),
		Role:              string(adv.Role),
		CanChangePassword: canChangePassword,
		CreatedAt:         adv.CreatedAt.Format(time.RFC3339),
	}
}
