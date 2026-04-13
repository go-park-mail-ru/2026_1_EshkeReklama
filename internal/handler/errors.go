package handlers

import (
	"errors"
	errs "eshkere/internal/errors"
)

func convertDomainError(err error) (statusCode int, clientMessage string, isExpected bool) {
	switch {
	case errors.Is(err, errs.ErrEmailTaken):
		return 400, "email already registered", true
	case errors.Is(err, errs.ErrPhoneTaken):
		return 400, "phone already registered", true
	case errors.Is(err, errs.ErrInvalidAdvertiserArg):
		return 400, err.Error(), true
	case errors.Is(err, errs.ErrInvalidCredentials):
		return 401, "invalid credentials", true
	default:
		return 500, "internal error", false
	}
}
