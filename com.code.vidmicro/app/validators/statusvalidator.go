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
	case "/api/status/put":
		if statusData.Name == "" {
			return errors.New("genre name is required")
		}
	case "/api/status/delete":
		fallthrough
	case "/api/status/post":
		if statusData.Id <= 0 {
			return errors.New("genre id is required")
		}
	}
	return nil
}