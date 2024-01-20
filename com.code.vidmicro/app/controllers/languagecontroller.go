package controllers

import (
	"crypto/rand"
	"io"
	"net/http"
	"time"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseinterfaces"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/baserouter"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basevalidators"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/serviceutils"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid"
)

type LanguageController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
	entropy io.Reader
}

func (u LanguageController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u LanguageController) GetCollectionName() basetypes.CollectionName {
	return "language"
}

func (u LanguageController) DoIndexing() error {
	u.entropy = ulid.Monotonic(rand.Reader, 0)
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.Language{})
	return nil
}

func (u *LanguageController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *LanguageController) handleCreateLanguage() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelLanguage := models.Language{}
		if err := c.ShouldBind(&modelLanguage); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		ulid := ulid.MustNew(ulid.Timestamp(time.Now()), u.entropy)
		modelLanguage.Id = ulid.String()

		err := u.Validate(c.GetString("apiPath")+"/put", modelLanguage)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}
		_, err = u.Add(u.GetDBName(), u.GetCollectionName(), modelLanguage, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_FAILED, err, nil))
			return
		}

		serviceutils.GetInstance().PublishEvent(modelLanguage, configmanager.GetInstance().MicroServiceName, "vidmicro.language.created")
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_SUCCESS, err, modelLanguage))
	}
}

func (u *LanguageController) handleGetLanguage() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelLanguage := models.Language{}
		modelLanguage.Id = c.Query("id")

		err := u.Validate(c.GetString("apiPath")+"/get", modelLanguage)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		rows, err := u.FindOne(u.GetDBName(), u.GetCollectionName(), "", map[string]interface{}{"id": modelLanguage.Id}, &modelLanguage, false, " Limit 1", false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
			return
		}
		defer rows.Close()

		// Iterate over the rows.
		for rows.Next() {
			// Create a User struct to scan values into.

			// Scan the row's values into the User struct.
			err := rows.Scan(&modelLanguage.Id, &modelLanguage.Name, &modelLanguage.Code)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, err, modelLanguage))
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
		modelLanguage := models.Language{}
		if err := c.ShouldBind(&modelLanguage); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/post", modelLanguage)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET name = $1, code = $2 WHERE id = $3", []interface{}{modelLanguage.Name, modelLanguage.Code, modelLanguage.Id}, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_FAILED, err, nil))
			return
		}
		serviceutils.GetInstance().PublishEvent(modelLanguage, configmanager.GetInstance().MicroServiceName, "vidmicro.language.updated")
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_SUCCESS, err, nil))
	}
}

func (u *LanguageController) handleDeleteLanguage() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelLanguage := models.Language{}
		if err := c.ShouldBind(&modelLanguage); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/post", modelLanguage)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err = u.DeleteOne(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelLanguage.Id}, false, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_FAILED, err, nil))
			return
		}
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
