package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorHelpers(t *testing.T) {
	tests := []struct {
		name    string
		fn      func(http.ResponseWriter, string)
		status  int
		message string
	}{
		{name: "bad request", fn: BadRequest, status: http.StatusBadRequest, message: "bad"},
		{name: "already exists", fn: AlreadyExists, status: http.StatusConflict, message: "exists"},
		{name: "business logic", fn: BusinessLogic, status: http.StatusUnprocessableEntity, message: "logic"},
		{name: "unauthorized", fn: Unauthorized, status: http.StatusUnauthorized, message: "unauth"},
		{name: "not found", fn: NotFound, status: http.StatusNotFound, message: "missing"},
		{name: "internal", fn: InternalError, status: http.StatusInternalServerError, message: "boom"},
		{name: "not implemented", fn: NotImplemented, status: http.StatusNotImplemented, message: "todo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			tt.fn(rr, tt.message)

			if rr.Code != tt.status {
				t.Fatalf("expected %d got %d", tt.status, rr.Code)
			}

			var resp Error
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if resp.Error != tt.message {
				t.Fatalf("expected %q got %q", tt.message, resp.Error)
			}
		})
	}
}
