package validator

import (
	"context"
	"errors"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

type GoValidator struct {
	validate *validator.Validate
	uni      ut.Translator
}

func NewGoValidator() Validator {
	v := validator.New()
	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")

	_ = en_translations.RegisterDefaultTranslations(v, trans)
	RegisterCustomValidator(v, trans)
	return &GoValidator{validate: v, uni: trans}
}

func (v *GoValidator) Validate(ctx context.Context, data interface{}) error {
	err := v.validate.StructCtx(ctx, data)
	if err == nil {
		return nil
	}

	var invalidErr *validator.InvalidValidationError
	if errors.As(err, &invalidErr) {
		return err
	}

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) && len(validationErrs) > 0 {
		mapErr := make(map[string]error, len(validationErrs))
		for _, fe := range validationErrs {
			mapErr[fe.Field()] = errors.New(fe.Translate(v.uni))
		}
		return NewErrorMap(mapErr)
	}

	return nil
}
