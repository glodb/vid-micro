package validators

import (
	"errors"

	"com.code.vidmicro/com.code.vidmicro/app/models"
)

type TitleMetaValidator struct {
}

func (u *TitleMetaValidator) Validate(apiName string, data interface{}) error {
	titleMetaData := data.(models.TitleMetaData)
	switch apiName {
	case "/api/title_meta/put":
		if titleMetaData.TitleId <= 0 {
			return errors.New("title id is required")
		}

		if len(titleMetaData.Genres) == 0 {
			return errors.New("one of the genre is must for title")
		}

		if titleMetaData.TypeId <= 0 {
			return errors.New("type id is must for title")
		}
	case "/api/title_meta/get":
		fallthrough
	case "/api/title_meta/delete":
		fallthrough
	case "/api/title_meta/post":
		if titleMetaData.TitleId <= 0 {
			return errors.New("titles id is required")
		}
	}
	return nil
}
