package handlers

import (
	"errors"
	errs "eshkere/internal/errors"
	"eshkere/pkg/httpx"
	lgr "eshkere/pkg/logger"
	"net/http"
)

func HandleError(w http.ResponseWriter, r *http.Request, desc string, err error) {
	logger := lgr.GetLoggerFromCtx(r.Context())

	switch {
	case errors.Is(err, errs.BadRequestError):
		logger.Debugf("Bad request error while %s: %v", desc, err.Error())
		httpx.BadRequest(w, err.Error())

	case errors.Is(err, errs.AlreadyExistsError):
		logger.Debugf("Already exists error while %s: %v", desc, err.Error())
		httpx.AlreadyExists(w, err.Error())

	case errors.Is(err, errs.BusinessLogicError):
		logger.Debugf("Business logic error while %s: %v", desc, err.Error())
		httpx.BusinessLogic(w, err.Error())

	case errors.Is(err, errs.UnauthorizedError):
		logger.Debugf("Not authorized error while %s: %v", desc, err.Error())
		httpx.Unauthorized(w, err.Error())

	case errors.Is(err, errs.NotFoundError):
		logger.Debugf("Not found error while %s: %v", desc, err.Error())
		httpx.NotFound(w, err.Error())

	case errors.Is(err, errs.NotImplementedError):
		logger.Errorf("Not implemented error while %s: %v", desc, err.Error())
		httpx.NotImplemented(w, err.Error())

	default:
		logger.Errorf("Internal error while %s: %v", desc, err.Error())
		httpx.InternalError(w, errs.InternalServiceError.Error())
	}
}
