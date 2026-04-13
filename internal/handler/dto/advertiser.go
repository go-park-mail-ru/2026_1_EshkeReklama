package dto

import (
	"eshkere/internal/models"
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

// AdvertiserProfileResponse — публичные поля рекламодателя (без пароля).
type AdvertiserProfileResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
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
		Balance:   adv.Balance,
		CreatedAt: adv.CreatedAt.Format(time.RFC3339),
	}
}
