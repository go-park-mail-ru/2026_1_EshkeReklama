package handler

import (
	"context"
	"encoding/json"
	"errors"
	errs "eshkere/internal/errors"
	lgr "eshkere/pkg/logger"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestHandleError_StatusMapping(t *testing.T) {
	logger := zap.Must(zap.NewDevelopment()).Sugar()

	tests := []struct {
		name    string
		err     error
		status  int
		message string
	}{
		{name: "bad request", err: errs.BadRequestError, status: http.StatusBadRequest, message: errs.BadRequestError.Error()},
		{name: "invalid advertiser arg", err: errs.ErrInvalidAdvertiserArg, status: http.StatusBadRequest, message: errs.ErrInvalidAdvertiserArg.Error()},
		{name: "unauthorized", err: errs.UnauthorizedError, status: http.StatusUnauthorized, message: errs.UnauthorizedError.Error()},
		{name: "invalid credentials", err: errs.ErrInvalidCredentials, status: http.StatusUnauthorized, message: errs.ErrInvalidCredentials.Error()},
		{name: "not found", err: errs.NotFoundError, status: http.StatusNotFound, message: errs.NotFoundError.Error()},
		{name: "already exists", err: errs.AlreadyExistsError, status: http.StatusConflict, message: errs.AlreadyExistsError.Error()},
		{name: "email taken", err: errs.ErrEmailTaken, status: http.StatusConflict, message: errs.ErrEmailTaken.Error()},
		{name: "phone taken", err: errs.ErrPhoneTaken, status: http.StatusConflict, message: errs.ErrPhoneTaken.Error()},
		{name: "business logic", err: errs.BusinessLogicError, status: http.StatusUnprocessableEntity, message: errs.BusinessLogicError.Error()},
		{name: "not implemented", err: errs.NotImplementedError, status: http.StatusNotImplemented, message: errs.NotImplementedError.Error()},
		{name: "wrapped internal", err: errors.New("db down"), status: http.StatusInternalServerError, message: errs.InternalServiceError.Error()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = req.WithContext(lgr.CtxWithLogger(context.Background(), logger))

			HandleError(rr, req, "testing", tt.err)

			if rr.Code != tt.status {
				t.Fatalf("expected %d got %d", tt.status, rr.Code)
			}

			var resp map[string]string
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if resp["error"] != tt.message {
				t.Fatalf("expected %q got %q", tt.message, resp["error"])
			}
		})
	}
}
