package validation

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	v *validator.Validate
}

func New() *Validator {
	v := validator.New()
	return &Validator{v: v}
}

func (v *Validator) Struct(s any) error {
	err := v.v.Struct(s)
	if err == nil {
		return nil
	}

	var errs validator.ValidationErrors
	if errors.As(err, &errs) {
		first := errs[0]
		return fmt.Errorf("field '%s' failed on the '%s' validation", first.Field(), first.Tag())
	}

	return fmt.Errorf("validation failed: %w", err)
}
