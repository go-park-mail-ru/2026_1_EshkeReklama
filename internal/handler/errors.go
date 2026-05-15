package handler

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
	// 400 Bad Request: Ошибки входных данных
	case errors.Is(err, errs.BadRequestError),
		errors.Is(err, errs.ErrInvalidAdvertiserArg):
		logger.Debugf("bad request during %s: %v", desc, err)
		httpx.BadRequest(w, err.Error())

	// 401 Unauthorized: Ошибки авторизации
	case errors.Is(err, errs.UnauthorizedError),
		errors.Is(err, errs.ErrInvalidCredentials):
		logger.Debugf("unauthorized during %s: %v", desc, err)
		httpx.Unauthorized(w, err.Error())

	// 403 Forbidden: Пользователь авторизован, но у него нет нужных прав
	case errors.Is(err, errs.ForbiddenError):
		logger.Debugf("forbidden during %s: %v", desc, err)
		httpx.Forbidden(w, err.Error())

	// 404 Not Found: Ресурс не найден
	case errors.Is(err, errs.NotFoundError):
		logger.Debugf("not found during %s: %v", desc, err)
		httpx.NotFound(w, err.Error())

	// 409 Conflict: Дубликаты (email, телефон и т.д.)
	case errors.Is(err, errs.AlreadyExistsError),
		errors.Is(err, errs.ErrEmailTaken),
		errors.Is(err, errs.ErrPhoneTaken),
		errors.Is(err, errs.ErrVKIDConflict):
		logger.Debugf("resource conflict during %s: %v", desc, err)
		httpx.AlreadyExists(w, err.Error())

	// 422 Unprocessable Entity: Нарушение бизнес-правил
	case errors.Is(err, errs.BusinessLogicError):
		logger.Debugf("business logic error during %s: %v", desc, err)
		httpx.BusinessLogic(w, err.Error())

	// 501 Not Implemented: Функционал еще не готов
	case errors.Is(err, errs.NotImplementedError):
		logger.Errorf("not implemented during %s: %v", desc, err)
		httpx.NotImplemented(w, err.Error())

	// 500 Internal Server Error: Непредвиденные ошибки (БД, сеть и т.д.)
	default:
		logger.Errorf("internal error during %s: %v", desc, err)
		httpx.InternalError(w, errs.InternalServiceError.Error())
	}
}
