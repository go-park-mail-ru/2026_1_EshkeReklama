package dto

import (
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"time"
)

type RegisterRequest struct {
	Name     string `json:"name,omitempty"`
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

type LoginResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type UpdateAdvertiserProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

func (u *UpdateAdvertiserProfileRequest) ToInput(advertiserID int) *serviceinput.UpdateAdvertiserProfile {
	return &serviceinput.UpdateAdvertiserProfile{
		AdvertiserID: advertiserID,
		Name:         u.Name,
		Email:        u.Email,
		Phone:        u.Phone,
	}
}

// AdvertiserProfileResponse — публичные поля рекламодателя (без пароля).
type AdvertiserProfileResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Balance   int64  `json:"balance"`
	CreatedAt string `json:"created_at"`
}

type TopUpBalanceRequest struct {
	Amount int64 `json:"amount"`
}

type BalanceResponse struct {
	Balance int64 `json:"balance"`
}

func AdvertiserToProfile(adv *models.Advertiser) AdvertiserProfileResponse {
	if adv == nil {
		return AdvertiserProfileResponse{}
	}
	return AdvertiserProfileResponse{
		ID:        adv.ID,
		Name:      adv.Name,
		Email:     adv.Email,
		Phone:     adv.Phone,
		AvatarURL: adv.AvatarURL.String,
		Balance:   adv.Balance,
		CreatedAt: adv.CreatedAt.Format(time.RFC3339),
	}
}
