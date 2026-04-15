package v1

import (
	"net/http"

	"eshkere/pkg/httpx"

	"github.com/go-playground/validator/v10"
)

type validatableRequest interface {
	Validate() error
}

var requestValidator = validator.New()

func newJSONRequest[T any](r *http.Request) (*T, error) {
	req := new(T)

	if err := httpx.DecodeJSON(r, req); err != nil {
		return nil, err
	}

	if err := requestValidator.Struct(req); err != nil {
		return nil, err
	}

	if validatable, ok := any(req).(validatableRequest); ok {
		if err := validatable.Validate(); err != nil {
			return nil, err
		}
	}

	return req, nil
}
