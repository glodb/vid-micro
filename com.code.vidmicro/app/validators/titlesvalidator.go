package validators

import (
	"errors"

	"com.code.vidmicro/com.code.vidmicro/app/models"
)

type TitlesValidator struct {
}

func (u *TitlesValidator) Validate(apiName string, data interface{}) error {
	titles := data.(models.Titles)
	switch apiName {
	case "/api/titles/put":
		if titles.OriginalTitle == "" {
			return errors.New("original title can't be null")
		}
		if titles.Year <= 1900 {
			return errors.New("titles year can't be null")
		}
	case "/api/titles/delete":
		fallthrough
	case "/api/titles/post":
		if titles.Id <= 0 {
			return errors.New("title id is required")
		}
	}
	return nil
}
