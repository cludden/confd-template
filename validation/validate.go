package validation

import (
	validator "gopkg.in/go-playground/validator.v9"
)

// Validate exports a sharable validator value
var Validate *validator.Validate

func init() {
	Validate = validator.New()
}
