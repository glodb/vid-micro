package validators

import (
	"errors"

	"com.code.vidmicro/com.code.vidmicro/app/models"
)

type GenresValidator struct {
}

func (u *GenresValidator) Validate(apiName string, data interface{}) error {
	gnresData := data.(models.Genres)
	switch apiName {
	case "/api/genres/put":
		if gnresData.Name == "" {
			return errors.New("genre name is required")
		}
	case "/api/genres/delete":
		fallthrough
	case "/api/genres/post":
		fallthrough
	case "/api/genres/get":
		if gnresData.Id <= 0 {
			return errors.New("genre id is required")
		}
	}
	return nil
}
