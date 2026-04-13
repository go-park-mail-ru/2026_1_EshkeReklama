package errors

import (
	"errors"
)

var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrEmailTaken           = errors.New("email already registered")
	ErrPhoneTaken           = errors.New("phone already registered")
	ErrInvalidAdvertiserArg = errors.New("invalid argument")
)
