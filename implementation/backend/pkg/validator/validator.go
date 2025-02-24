package myvalidator

import (
	"errors"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entranslations "github.com/go-playground/validator/v10/translations/en"
	myerror "tugas-akhir/backend/pkg/error"
)

type TranslatedValidator struct {
	Validator  *validator.Validate
	Translator ut.Translator
}

func (tv *TranslatedValidator) Validate(i interface{}) ([]myerror.FieldError, error) {
	err := tv.Validator.Struct(i)

	var ve validator.ValidationErrors

	if errors.As(err, &ve) {
		out := make([]myerror.FieldError, len(ve))
		for i, fe := range ve {
			out[i] = myerror.FieldError{Field: fe.Field(), Message: fe.Translate(tv.Translator), Tag: fe.Tag()}
		}

		return out, nil
	}

	return nil, err
}

func NewTranslastedValidator() *TranslatedValidator {
	enLocale := en.New()
	universalTranslator := ut.New(enLocale, enLocale)

	translator, _ := universalTranslator.GetTranslator("en")

	validate := validator.New()

	_ = entranslations.RegisterDefaultTranslations(validate, translator)

	return &TranslatedValidator{
		Validator:  validate,
		Translator: translator,
	}
}
