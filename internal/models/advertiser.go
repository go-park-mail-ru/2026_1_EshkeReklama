package model

import "time"

type Advertiser struct {
	ID           int
	Email        string
	Phone        string
	PasswordHash string
	CreatedAt    time.Time
	// бла бла бла
}

type RegisterAdvertiserInput struct {
	Email    string
	Phone    string
	Password string
}

type LoginAdvertiserInput struct {
	Identifier string
	Password   string
}
