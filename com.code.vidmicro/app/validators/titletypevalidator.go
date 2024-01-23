package validators

import (
	"errors"

	"com.code.vidmicro/com.code.vidmicro/app/models"
)

type TitleTypeValidator struct {
}

func (u *TitleTypeValidator) Validate(apiName string, data interface{}) error {
	titleTypeData := data.(models.TitleType)
	switch apiName {
	case "/api/title_type/put":
		if titleTypeData.Name == "" {
			return errors.New("genre name is required")
		}
	case "/api/title_type/delete":
		fallthrough
	case "/api/title_type/post":
		if titleTypeData.Id <= 0 {
			return errors.New("genre id is required")
		}
	}
	return nil
}
