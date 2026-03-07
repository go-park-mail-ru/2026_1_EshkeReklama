package dto

import (
	"errors"
)

type RegisterRequest struct {
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

func (r RegisterRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}

	if r.Phone == "" {
		return errors.New("phone is required")
	}

	if r.Password == "" {
		return errors.New("password is required")
	}

	return nil
}

func (r LoginRequest) Validate() error {
	if r.Identifier == "" {
		return errors.New("identifier is required")
	}

	if r.Password == "" {
		return errors.New("password is required")
	}

	return nil
}
