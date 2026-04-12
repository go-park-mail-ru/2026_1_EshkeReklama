package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Success struct {
	Data interface{} `json:"data,omitempty"`
}

type Error struct {
	Error string `json:"error"`
}

func DecodeJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return errors.New("decodeJSON: empty request body")
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(v)
}

func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(Success{
		Data: data,
	})
}

func ErrorJSON(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(Error{
		Error: message,
	})
}

func BadRequest(w http.ResponseWriter, message string) {
	ErrorJSON(w, http.StatusBadRequest, message)
}

func Unauthorized(w http.ResponseWriter, message string) {
	ErrorJSON(w, http.StatusUnauthorized, message)
}

func NotFound(w http.ResponseWriter, message string) {
	ErrorJSON(w, http.StatusNotFound, message)
}

func InternalError(w http.ResponseWriter, message string) {
	ErrorJSON(w, http.StatusInternalServerError, message)
}

func NotImplemented(w http.ResponseWriter, message string) {
	ErrorJSON(w, http.StatusNotImplemented, message)
}
