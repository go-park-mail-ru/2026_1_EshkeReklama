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
	ID    int    `json:"id"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type VKIDLoginRequest struct {
	Code         string `json:"code"`
	DeviceID     string `json:"device_id"`
	CodeVerifier string `json:"code_verifier"`
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

// AdvertiserProfileResponse — публичные поля рекламодателя (без пароля).
type AdvertiserProfileResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Surname   string `json:"surname"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Balance   int64  `json:"balance"`
	Company   string `json:"company"`
	City      string `json:"city"`
	Tariff    string `json:"tariff"`
	CreatedAt string `json:"created_at"`
}

type TopUpBalanceRequest struct {
	Amount int64 `json:"amount"`
}

type BalanceResponse struct {
	Balance int64 `json:"balance"`
}

func AdvertiserToProfile(adv *models.Advertiser) AdvertiserProfileResponse {
	return AdvertiserWithContactsToProfile(adv, "", "")
}

func AdvertiserWithContactsToProfile(adv *models.Advertiser, email, phone string) AdvertiserProfileResponse {
	if adv == nil {
		return AdvertiserProfileResponse{}
	}
	return AdvertiserProfileResponse{
		ID:        adv.ID,
		Name:      adv.Name,
		Surname:   adv.Surname.String,
		Email:     email,
		Phone:     phone,
		AvatarURL: adv.AvatarURL.String,
		Balance:   adv.Balance,
		Company:   adv.Company.String,
		City:      adv.City.String,
		Tariff:    string(adv.Tariff),
		CreatedAt: adv.CreatedAt.Format(time.RFC3339),
	}
}
