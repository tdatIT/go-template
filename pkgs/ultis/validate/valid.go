package validate

import (
	"github.com/go-playground/validator/v10"
)

var _validator *Validator

type Validator struct {
	Valid *validator.Validate
}

func (v *Validator) Validate(i interface{}) error {
	err := v.Valid.Struct(i)
	if err != nil {
		return err
	}
	return nil
}

func GetValidator() *Validator {
	if _validator != nil {
		return _validator
	}

	_validator = &Validator{
		Valid: validator.New(),
	}

	return _validator
}
