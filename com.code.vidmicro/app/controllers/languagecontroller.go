package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/app/models/jsonmodels"
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseinterfaces"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/baserouter"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basevalidators"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"com.code.vidmicro/com.code.vidmicro/settings/cache"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/serviceutils"
	"github.com/gin-gonic/gin"
)

type LanguageController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u LanguageController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u LanguageController) GetCollectionName() basetypes.CollectionName {
	return "language"
}

func (u LanguageController) DoIndexing() error {
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.Language{})
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

		tempLanguage := models.Language{}

		// Scan the row's values into the User struct.
		err := rows.Scan(&tempLanguage.Id, &tempLanguage.Name, &tempLanguage.Code)
		if err != nil {
			break
		}

		cache.GetInstance().Set(fmt.Sprintf("%d%s%s", tempLanguage.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguagePostfix), tempLanguage.EncodeRedisData())
	}
	return nil
}

func (u *LanguageController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *LanguageController) handleCreateLanguage() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelLanguage := models.Language{}
		if err := c.ShouldBind(&modelLanguage); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}
		err := basevalidators.GetInstance().GetValidator().Struct(modelLanguage)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		id, err := u.Add(u.GetDBName(), u.GetCollectionName(), modelLanguage, true)
		modelLanguage.Id = int(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		serviceutils.GetInstance().PublishEvent(modelLanguage, configmanager.GetInstance().MicroServiceName, "vidmicro.language.created")
		err = cache.GetInstance().Set(fmt.Sprintf("%d%s%s", modelLanguage.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguagePostfix), modelLanguage.EncodeRedisData())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_SUCCESS, err, modelLanguage))
	}
}

func (u *LanguageController) handleGetLanguage() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelLanguage := jsonmodels.IdEmpty{}
		id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
		modelLanguage.Id = int(id)

		err := basevalidators.GetInstance().GetValidator().Struct(modelLanguage)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		query := map[string]interface{}{"id": modelLanguage.Id}

		if modelLanguage.Id <= 0 {
			query = map[string]interface{}{}
		}

		rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), "", query, &modelLanguage, false, "", false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		defer rows.Close()

		languages := make([]models.Language, 0)
		// Iterate over the rows.
		for rows.Next() {
			tempLanguage := models.Language{}
			// Create a User struct to scan values into.

			// Scan the row's values into the User struct.
			err := rows.Scan(&tempLanguage.Id, &tempLanguage.Name, &tempLanguage.Code)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}

			languages = append(languages, tempLanguage)
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, err, languages))
	}
}

func (u *LanguageController) UpdateLanguage(modelLanguage models.Language) {
	u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET name = $1, code = $2 WHERE id = $3", []interface{}{modelLanguage.Name, modelLanguage.Code, modelLanguage.Id}, false)
}

func (u *LanguageController) DeleteLanguage(modelLanguage models.Language) {
	u.DeleteOne(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelLanguage.Id}, false, false)
}

func (u *LanguageController) handleUpdateLanguage() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelLanguage := jsonmodels.EditLanguage{}
		if err := c.ShouldBind(&modelLanguage); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := basevalidators.GetInstance().GetValidator().Struct(modelLanguage)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET name = $1, code = $2 WHERE id = $3", []interface{}{modelLanguage.Name, modelLanguage.Code, modelLanguage.Id}, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		err = cache.GetInstance().Set(fmt.Sprintf("%d%s%s", modelLanguage.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguagePostfix), modelLanguage.EncodeRedisData())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		serviceutils.GetInstance().PublishEvent(modelLanguage, configmanager.GetInstance().MicroServiceName, "vidmicro.language.updated")
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_SUCCESS, err, nil))
	}
}

func (u *LanguageController) handleDeleteLanguage() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelLanguage := jsonmodels.EditLanguage{}
		if err := c.ShouldBind(&modelLanguage); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := basevalidators.GetInstance().GetValidator().Struct(modelLanguage)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		err = u.DeleteOne(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelLanguage.Id}, false, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		cache.GetInstance().Del(fmt.Sprintf("%d%s%s", modelLanguage.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguagePostfix))
		serviceutils.GetInstance().PublishEvent(modelLanguage, configmanager.GetInstance().MicroServiceName, "vidmicro.language.deleted")
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_SUCCESS, err, nil))
	}
}

func (u LanguageController) RegisterApis() {
	baserouter.GetInstance().GetLoginRouter().PUT("/api/languages", u.handleCreateLanguage())
	baserouter.GetInstance().GetLoginRouter().GET("/api/languages", u.handleGetLanguage())
	baserouter.GetInstance().GetLoginRouter().POST("/api/languages", u.handleUpdateLanguage())
	baserouter.GetInstance().GetLoginRouter().DELETE("/api/languages", u.handleDeleteLanguage())
}
