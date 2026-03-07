package handlers

import (
	"eshkere/internal/handler/dto"
	"eshkere/internal/models"
)

func ToRegisterAdvertiserModel(r dto.RegisterRequest) model.RegisterAdvertiserInput {
	return model.RegisterAdvertiserInput{
		Email:    r.Email,
		Phone:    r.Phone,
		Password: r.Password,
	}
}

func ToLoginAdvertiserModel(r dto.LoginRequest) model.LoginAdvertiserInput {
	return model.LoginAdvertiserInput{
		Identifier: r.Identifier,
		Password:   r.Password,
	}
}
