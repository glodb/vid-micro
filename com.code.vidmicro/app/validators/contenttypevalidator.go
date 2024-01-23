package validators

import (
	"errors"

	"com.code.vidmicro/com.code.vidmicro/app/models"
)

type ContentTypeValidator struct {
}

func (u *ContentTypeValidator) Validate(apiName string, data interface{}) error {
	contentType := data.(models.ContentType)
	switch apiName {
	case "/api/content_type/put":
		if contentType.Name == "" {
			return errors.New("genre name is required")
		}
	case "/api/content_type/delete":
		fallthrough
	case "/api/content_type/post":
		if contentType.Id <= 0 {
			return errors.New("genre id is required")
		}
	}
	return nil
}
