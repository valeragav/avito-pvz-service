package validation

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

type FieldError struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
}

func MapErrors(err error) []FieldError {
	var verrs validator.ValidationErrors
	ok := errors.As(err, &verrs)
	if !ok {
		return nil
	}

	out := make([]FieldError, 0, len(verrs))
	for _, e := range verrs {
		out = append(out, FieldError{
			Field: e.Field(),
			Rule:  e.Tag(),
		})
	}
	return out
}
