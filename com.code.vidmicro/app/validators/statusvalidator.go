package validators

import (
	"errors"

	"com.code.vidmicro/com.code.vidmicro/app/models"
)

type StatusValidator struct {
}

func (u *StatusValidator) Validate(apiName string, data interface{}) error {
	statusData := data.(models.Status)
	switch apiName {
	case "/api/genres/put":
		if statusData.Name == "" {
			return errors.New("genre name is required")
		}
	case "/api/genres/delete":
		fallthrough
	case "/api/genres/post":
		fallthrough
	case "/api/genres/get":
		if statusData.Id <= 0 {
			return errors.New("genre id is required")
		}
	}
	return nil
}
