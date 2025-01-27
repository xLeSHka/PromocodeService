package utils

import (
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

const TimeFormat = "2006-01-02"

var Validate = validator.New(validator.WithRequiredStructEnabled())

func PasswordValidationFunc(fl validator.FieldLevel) bool {
	var digit, upper, lower, symbol bool
	password := fl.Field().String()
	if len(password) < 8 && len(password) > 60 {
		return false
	}
	symbol = strings.ContainsAny(password, "@$!%*?&")
	digit = strings.ContainsAny(password, "0123456789")
	upper = strings.ContainsAny(password, "qwertyuiopasdfghjklzxcvbnm")
	lower = strings.ContainsAny(password, "QWERTYUIOPASDFGHJKLZXCVBNM")
	return digit && upper && lower && symbol
}

func DateValidationFunc(fl validator.FieldLevel) bool {
	date := fl.Field().String()
	if _, err := time.Parse(TimeFormat, date); err != nil {
		return false
	}
	return true
}
func CountryValidationFunc(fl validator.FieldLevel) bool {
	type Cntry struct {
		Cntry string `validate:"iso3166_1_alpha2"`
	}
	country := Cntry{
		Cntry: strings.ToUpper(fl.Field().String()),
	}
	err := Validate.Struct(country)
	return err == nil
}
