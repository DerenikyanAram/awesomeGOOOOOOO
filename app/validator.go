package main

import "github.com/go-playground/validator/v10"

var validate = validator.New()

func ValidateOrder(o Order) error {
	return validate.Struct(o)
}
