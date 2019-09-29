package validator

import (
	"reflect"
	"strings"

	"gopkg.in/go-playground/validator.v9"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterValidation("notblank", NotBlank)
}

// validateNotBlank validates that string is not only whitespace
//
// Copying this function from gopkg.in/go-playground/validator.v9/non-standard/validators since the big fix in there
// is not being pulled yet by go mod (may be caching issues).
// NotBlank is the validation function for validating if the current field
// has a value or length greater than zero, or is not a space only string.
func NotBlank(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		return len(strings.TrimSpace(field.String())) > 0
	case reflect.Chan, reflect.Map, reflect.Slice, reflect.Array:
		return field.Len() > 0
	case reflect.Ptr, reflect.Interface, reflect.Func:
		return !field.IsNil()
	default:
		return field.IsValid() && field.Interface() != reflect.Zero(field.Type()).Interface()
	}
}

// validateUniqueStrings should only be called on slices, arrays and maps,
// otherwise it should fail
func validateUniqueStrings(fl validator.FieldLevel) bool {
	var v = fl.Field()
	// The reflect.Value is not even valid, it's weird and we can't pass
	if !v.IsValid() {
		return false
	}
	// The reflect.Value is the zero value for it's type, it's unique (since no elements)
	if v.IsZero() {
		return true
	}
	// if it's not a slice, it's unique
	if v.Kind() != reflect.Array &&
		v.Kind() != reflect.Map &&
		v.Kind() != reflect.Slice {
		return false
	}
	// Iterate over slice to make sure it's unique
	len := v.Len()
	witness := make(map[reflect.Value]bool)
	for i := 0; i < len; i++ {
		e := v.Index(i)
		if witness[e] {
			// We're already seen e, so not unique
			return false
		}
		witness[e] = true
	}

	return true
}

// Validate validate the struct
func Validate(v interface{}) error {
	return validate.Struct(v)
}
