package validate

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	_validator *Validator
	_once      sync.Once
)

type Validator struct {
	Valid *validator.Validate
}

func GetValidator() *Validator {
	_once.Do(func() {
		_validator = &Validator{
			Valid: validator.New(),
		}
	})
	return _validator
}

func (v *Validator) Validate(i interface{}) error {
	err := v.Valid.Struct(i)
	if err != nil {
		return err
	}
	return nil
}
