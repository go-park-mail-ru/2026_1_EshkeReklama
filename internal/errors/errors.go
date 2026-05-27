package errors

import (
	"errors"
)

var (
	BadRequestError         = errors.New("bad request")
	AlreadyExistsError      = errors.New("already exists")
	BusinessLogicError      = errors.New("business logic error")
	UnauthorizedError       = errors.New("unauthorized")
	ForbiddenError          = errors.New("forbidden")
	NotFoundError           = errors.New("not found")
	InternalServiceError    = errors.New("internal service error")
	NotImplementedError     = errors.New("not implemented")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrEmailTaken           = errors.New("email already registered")
	ErrPhoneTaken           = errors.New("phone already registered")
	ErrVKIDConflict         = errors.New("vk id account conflict")
	ErrInvalidAdvertiserArg = errors.New("invalid advertiser argument")
	ErrSessionNotFound      = errors.New("session not found")
	ErrPasswordUnavailable  = errors.New("password change is unavailable")
	ErrEmailNotVerified     = errors.New("email is not verified")
	ErrInvalidVerifyCode    = errors.New("invalid verification code")
)
