package validator

import (
	"unicode"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

const (
	StrongPassword string = "strong-password"
)

func RegisterCustomValidator(validate *validator.Validate, trans ut.Translator) {
	validate.RegisterValidation(StrongPassword, validateStrongPassword)
	validate.RegisterTranslation(StrongPassword, trans, validateStrongPasswordMessage, validateStrongPasswordField)
}

func validateStrongPassword(fl validator.FieldLevel) bool {
	input := fl.Field().String()
	var (
		hasUpper   = false
		hasLower   = false
		hasSpecial = false
	)

	for _, char := range input {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasSpecial
}

func validateStrongPasswordMessage(ut ut.Translator) error {
	return ut.Add(StrongPassword, "password at least contains one letter, one special character", true)
}

func validateStrongPasswordField(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T(StrongPassword, fe.Field())
	return t
}
