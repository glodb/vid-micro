package validators

import (
	"errors"
	"regexp"
	"unicode/utf8"

	"com.code.vidmicro/com.code.vidmicro/app/models"
)

type UserValidator struct {
}

func (u *UserValidator) Validate(apiName string, data interface{}) error {
	userData := data.(models.User)
	switch apiName {
	case "/api/login":
		fallthrough
	case "/api/registerUser":
		pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,4}$`
		matched, err := regexp.MatchString(pattern, userData.Email)

		if !matched {
			return errors.New("email address validation failed")
		}
		if err != nil {
			return err
		}

		length := utf8.RuneCountInString(userData.Password)
		if length < 8 || length > 64 {
			return errors.New("password length needs to be in 8 to 64 characters")
		}
	case "/api/blackListUser":
		if userData.Id <= 0 {
			return errors.New("username is must for black listing a user")
		}
	}
	return nil
}
