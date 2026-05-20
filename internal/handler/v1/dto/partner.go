package dto

import (
	"time"

	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

type PartnerRegisterRequest struct {
	LastName               string `json:"last_name" validate:"required"`
	FirstName              string `json:"first_name" validate:"required"`
	MiddleName             string `json:"middle_name"`
	BirthDate              string `json:"birth_date" validate:"required"`
	Email                  string `json:"email" validate:"required,email"`
	Phone                  string `json:"phone" validate:"required"`
	CountryCode            string `json:"country_code" validate:"required"`
	RegistrationRegionCode string `json:"registration_region_code" validate:"required"`
	CooperationForm        string `json:"cooperation_form" validate:"required,oneof=self_employed individual_entrepreneur legal_entity"`
	PayoutCurrency         string `json:"payout_currency" validate:"required,oneof=RUB USD EUR"`
	Password               string `json:"password" validate:"required,min=6"`
}

func (r *PartnerRegisterRequest) ToInput(id int64) *serviceinput.CreatePartnerProfile {
	return &serviceinput.CreatePartnerProfile{
		ID:                     id,
		LastName:               r.LastName,
		FirstName:              r.FirstName,
		MiddleName:             r.MiddleName,
		BirthDate:              r.BirthDate,
		CountryCode:            r.CountryCode,
		RegistrationRegionCode: r.RegistrationRegionCode,
		CooperationForm:        r.CooperationForm,
		PayoutCurrency:         r.PayoutCurrency,
	}
}

type PartnerLoginRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

type PartnerAuthResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type UpdatePartnerProfileRequest struct {
	LastName               *string `json:"last_name"`
	FirstName              *string `json:"first_name"`
	MiddleName             *string `json:"middle_name"`
	BirthDate              *string `json:"birth_date"`
	Email                  *string `json:"email" validate:"omitempty,email"`
	Phone                  *string `json:"phone"`
	CountryCode            *string `json:"country_code"`
	RegistrationRegionCode *string `json:"registration_region_code"`
	CooperationForm        *string `json:"cooperation_form" validate:"omitempty,oneof=self_employed individual_entrepreneur legal_entity"`
	PayoutCurrency         *string `json:"payout_currency" validate:"omitempty,oneof=RUB USD EUR"`
}

func (r *UpdatePartnerProfileRequest) ToInput(partnerID int) *serviceinput.UpdatePartnerProfile {
	return &serviceinput.UpdatePartnerProfile{
		PartnerID:              partnerID,
		LastName:               r.LastName,
		FirstName:              r.FirstName,
		MiddleName:             r.MiddleName,
		BirthDate:              r.BirthDate,
		CountryCode:            r.CountryCode,
		RegistrationRegionCode: r.RegistrationRegionCode,
		CooperationForm:        r.CooperationForm,
		PayoutCurrency:         r.PayoutCurrency,
	}
}

type PartnerProfileResponse struct {
	ID                     int    `json:"id"`
	LastName               string `json:"last_name"`
	FirstName              string `json:"first_name"`
	MiddleName             string `json:"middle_name"`
	BirthDate              string `json:"birth_date"`
	Email                  string `json:"email"`
	Phone                  string `json:"phone"`
	CountryCode            string `json:"country_code"`
	RegistrationRegionCode string `json:"registration_region_code"`
	CooperationForm        string `json:"cooperation_form"`
	PayoutCurrency         string `json:"payout_currency"`
	Balance                int64  `json:"balance"`
	CreatedAt              string `json:"created_at"`
}

func PartnerToProfile(partner *models.Partner, email, phone string) PartnerProfileResponse {
	if partner == nil {
		return PartnerProfileResponse{}
	}
	return PartnerProfileResponse{
		ID:                     partner.ID,
		LastName:               partner.LastName,
		FirstName:              partner.FirstName,
		MiddleName:             partner.MiddleName,
		BirthDate:              partner.BirthDate.Format("2006-01-02"),
		Email:                  email,
		Phone:                  phone,
		CountryCode:            partner.CountryCode,
		RegistrationRegionCode: partner.RegistrationRegionCode,
		CooperationForm:        string(partner.CooperationForm),
		PayoutCurrency:         string(partner.PayoutCurrency),
		Balance:                partner.Balance,
		CreatedAt:              partner.CreatedAt.Format(time.RFC3339),
	}
}
