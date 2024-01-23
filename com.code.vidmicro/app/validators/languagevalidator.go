package validators

import (
	"errors"

	"com.code.vidmicro/com.code.vidmicro/app/models"
)

type LanguageValidator struct {
}

func (u *LanguageValidator) Validate(apiName string, data interface{}) error {
	gnresData := data.(models.Language)
	switch apiName {
	case "/api/language/put":
		if gnresData.Name == "" {
			return errors.New("language name is required")
		}
	case "/api/language/delete":
		fallthrough
	case "/api/language/post":
		if gnresData.Id <= 0 {
			return errors.New("language id is required")
		}
	}
	return nil
}
