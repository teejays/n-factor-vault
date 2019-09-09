package validator

import (
	"strings"

	"gopkg.in/go-playground/validator.v9"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	validate.RegisterValidation("notblank", ValidateNotBlank)
}

// ValidateNotBlank validates that string is not only whitespace
func ValidateNotBlank(fl validator.FieldLevel) bool {
	return strings.TrimSpace(fl.Field().String()) != ""
}

// Validate validate the struct
func Validate(v interface{}) error {
	return validate.Struct(v)
}
