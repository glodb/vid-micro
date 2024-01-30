package basevalidators

import (
	"reflect"
	"sync"
	"unicode"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"github.com/bytedance/sonic"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var (
	once     sync.Once
	instance *CustomValidator
)

type CustomValidator struct {
	specialChars map[rune]bool
	trans        ut.Translator
	v            *validator.Validate
}

func GetInstance() *CustomValidator {

	once.Do(func() {
		instance = &CustomValidator{}
		instance.specialChars = map[rune]bool{'!': true, '@': true, '#': true, '$': true, '%': true, '^': true, '&': true, '*': true, '(': true, ')': true, '-': true, '_': true, '+': true, '=': true, '<': true, '>': true, '?': true, '/': true, '{': true, '}': true, '[': true, ']': true, '|': true}
		en := en.New()
		uni := ut.New(en, en)
		instance.trans, _ = uni.GetTranslator("en")
		instance.v = validator.New(validator.WithRequiredStructEnabled())
		instance.RegisterCustomValidators()
	})
	return instance
}

func (cv *CustomValidator) GetTrans() ut.Translator {
	return cv.trans
}

func (cv *CustomValidator) GetValidator() *validator.Validate {
	return cv.v
}

func (cv *CustomValidator) RegisterCustomValidators() {
	cv.v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Tag.Get("field")
	})

	cv.v.RegisterValidation("password", cv.PasswordValidator)
	cv.v.RegisterValidation("arraylength", cv.JsonArrayValidator)

	cv.v.RegisterTranslation("required", cv.trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} must have a value!", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())

		return t
	})

	cv.v.RegisterTranslation("required_without", cv.trans, func(ut ut.Translator) error {
		return ut.Add("required_without", "{0} atleast must have a value!", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required_without", fe.Field())

		return t
	})

	cv.v.RegisterTranslation("min", cv.trans, func(ut ut.Translator) error {
		return ut.Add("min", "{0} must have minimum values! {1}", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("min", fe.Field(), fe.Param())

		return t
	})

	cv.v.RegisterTranslation("max", cv.trans, func(ut ut.Translator) error {
		return ut.Add("max", "{0} must have maximum values! {1}", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("max", fe.Field(), fe.Param())

		return t
	})

	cv.v.RegisterTranslation("password", cv.trans, func(ut ut.Translator) error {
		return ut.Add("password", "{0} must be 8 characters long and should contain combination of special characters, digits, small and capital letters", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("password", fe.Field())

		return t
	})

	cv.v.RegisterTranslation("email", cv.trans, func(ut ut.Translator) error {
		return ut.Add("email", "{0} syntax is not correct", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("email", fe.Field())

		return t
	})

	cv.v.RegisterTranslation("gt", cv.trans, func(ut ut.Translator) error {
		return ut.Add("gt", "{0} must be greater than! {1}", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("gt", fe.Field(), fe.Param())

		return t
	})

	cv.v.RegisterTranslation("arraylength", cv.trans, func(ut ut.Translator) error {
		return ut.Add("arraylength", "{0} must be json array with at least one value", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("arraylength", fe.Field())

		return t
	})

	cv.v.RegisterTranslation("oneof", cv.trans, func(ut ut.Translator) error {
		return ut.Add("oneof", "{0} can only be {1}", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("oneof", fe.Field(), fe.Param())

		return t
	})

	cv.v.RegisterTranslation("lte", cv.trans, func(ut ut.Translator) error {
		return ut.Add("lte", "{0} must be lesser than equal to {1}", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("lte", fe.Field(), fe.Param())

		return t
	})

}

func (cv *CustomValidator) JsonArrayValidator(fl validator.FieldLevel) bool {
	jsonValidator := fl.Field().String()

	titlesLanguages := make([]models.TitlesLanguage, 0)

	err := sonic.Unmarshal([]byte(jsonValidator), &titlesLanguages)
	if err != nil {
		return false
	}

	if len(titlesLanguages) <= 0 {
		return false
	}
	return true
}

func (cv *CustomValidator) PasswordValidator(fl validator.FieldLevel) bool {
	// Use a regular expression to enforce password criteria
	password := fl.Field().String()
	var (
		hasLower   bool
		hasUpper   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		case cv.isSpecialCharacter(char):
			hasSpecial = true
		}
	}
	return hasLower && hasUpper && hasDigit && hasSpecial && len(password) >= 8
}

func (cv *CustomValidator) isSpecialCharacter(char rune) bool {
	_, ok := cv.specialChars[char]
	return ok
}

func (cv *CustomValidator) CreateErrors(err error) map[string][]string {
	returnMap := make(map[string][]string)
	errs := err.(validator.ValidationErrors)

	for _, e := range errs {
		returnMap[e.Field()] = []string{e.Translate(cv.GetTrans())}
	}
	return returnMap
}
