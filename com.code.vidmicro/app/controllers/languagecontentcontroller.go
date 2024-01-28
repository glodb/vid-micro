package controllers

import (
	"fmt"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseinterfaces"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basevalidators"
	"com.code.vidmicro/com.code.vidmicro/settings/cache"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
)

type LanguageContentController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u LanguageContentController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u LanguageContentController) GetCollectionName() basetypes.CollectionName {
	return "language"
}

func (u LanguageContentController) DoIndexing() error {
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.LanguageContent{})
	keys := cache.GetInstance().GetKeys("*" + configmanager.GetInstance().LanguagePostfix)
	cache.GetInstance().DelMany(keys)

	if len(keys) > 0 {
		cache.GetInstance().DelMany(keys)
	}

	rows, _ := u.Find(u.GetDBName(), u.GetCollectionName(), "", map[string]interface{}{}, &models.Language{}, false, "", false)

	defer rows.Close()
	// Iterate over the rows.
	for rows.Next() {
		// Create a User struct to scan values into.

		tempLanguage := models.LanguageContent{}

		// Scan the row's values into the User struct.
		err := rows.Scan(&tempLanguage.Id, &tempLanguage.Name, &tempLanguage.Code)
		if err != nil {
			break
		}

		cache.GetInstance().Set(fmt.Sprintf("%d%s%s", tempLanguage.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguagePostfix), tempLanguage.EncodeRedisData())
	}
	return nil
}

func (u *LanguageContentController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *LanguageContentController) UpdateLanguage(modelLanguage models.LanguageContent) {
	u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET name = $1, code = $2 WHERE id = $3", []interface{}{modelLanguage.Name, modelLanguage.Code, modelLanguage.Id}, false)
}

func (u *LanguageContentController) DeleteLanguage(modelLanguage models.LanguageContent) {
	u.DeleteOne(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelLanguage.Id}, false, false)
}

func (u LanguageContentController) RegisterApis() {
}
