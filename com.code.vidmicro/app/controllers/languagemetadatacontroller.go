package controllers

import (
	"errors"
	"fmt"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseinterfaces"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basevalidators"
	"com.code.vidmicro/com.code.vidmicro/settings/cache"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
)

type LanguageMetadataController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u LanguageMetadataController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u LanguageMetadataController) GetCollectionName() basetypes.CollectionName {
	return "language_metadata"
}

func (u LanguageMetadataController) DoIndexing() error {
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.LanguageMeta{})
	return nil
}

func (u *LanguageMetadataController) GetLanguageDetails(ids []string) ([]models.LanguageMetaDetails, error) {
	languagesMeta := []models.LanguageMetaDetails{}
	for _, language := range ids {

		langMetaDataBytes, err := cache.GetInstance().Get(fmt.Sprintf("%s%s%s", language, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguageMetadataPostfix))

		if err == nil && len(langMetaDataBytes) > 0 {
			tempMetadata := models.LanguageMetaDetails{}
			tempMetadata.DecodeRedisData(langMetaDataBytes)
			languagesMeta = append(languagesMeta, tempMetadata)
		} else {

			rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), "", map[string]interface{}{"id": language}, &models.LanguageMeta{}, false, "", false)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			tempMeta := models.LanguageMeta{}
			for rows.Next() {
				// Scan the row's values into the User struct.
				err := rows.Scan(&tempMeta.Id, &tempMeta.TitlesId, &tempMeta.LanguageId, &tempMeta.StatusId)
				if err != nil {
					return nil, err
				}

			}
			langData, err := cache.GetInstance().Get(fmt.Sprintf("%d%s%s", tempMeta.LanguageId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguagePostfix))

			if err != nil || len(langData) == 0 {
				return nil, errors.New("one of the language is not found")
			}

			statusData, err := cache.GetInstance().Get(fmt.Sprintf("%d%s%s", tempMeta.StatusId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().StatusPostfix))

			if err != nil || len(statusData) == 0 {
				return nil, errors.New("one of the status is not found")
			}

			statusObject := models.Status{}
			statusObject.DecodeRedisData(statusData)

			languageObject := models.Language{}
			languageObject.DecodeRedisData(langData)

			languageMeta := models.LanguageMetaDetails{LanguageId: languageObject.Id, LanguageName: languageObject.Name, LanguageCode: languageObject.Code, StatusId: statusObject.Id, StatusName: statusObject.Name}
			languagesMeta = append(languagesMeta, languageMeta)
			err = cache.GetInstance().SetEx(fmt.Sprintf("%s%s%s", language, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguageMetadataPostfix), languageMeta.EncodeRedisData(), configmanager.GetInstance().LanguageMetaExpiryTime)

			if err != nil {
				return nil, err
			}
		}
		if len(languagesMeta) <= 0 {
			return nil, errors.New("language meta not found")
		}
	}
	return languagesMeta, nil
}

func (u *LanguageMetadataController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u LanguageMetadataController) RegisterApis() {
}
