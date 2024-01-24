package validators

import (
	"errors"

	"com.code.vidmicro/com.code.vidmicro/app/models"
)

type ContentValidator struct {
}

func (u *ContentValidator) Validate(apiName string, data interface{}) error {
	contentData := data.(models.Contents)
	switch apiName {
	case "/api/content/put":
		if contentData.Name == "" {
			return errors.New("content type name is required")
		}
		if contentData.LanguageId <= 0 {
			return errors.New("content language id is required")
		}
		if contentData.TypeId <= 0 {
			return errors.New("content type id is required")
		}
		if contentData.AssociatedTitle <= 0 {
			return errors.New("associated title is required")
		}
	case "/api/content/delete":
		if contentData.Id <= 0 {
			return errors.New("content id is required")
		}
	case "/api/content/post":
		if contentData.Id <= 0 {
			return errors.New("content id is required")
		}
		if contentData.AssociatedTitle <= 0 {
			return errors.New("associated title is required")
		}
	case "/api/content/get":
		if contentData.AssociatedTitle <= 0 {
			return errors.New("associated title is required to get the data")
		}
	}
	return nil
}
