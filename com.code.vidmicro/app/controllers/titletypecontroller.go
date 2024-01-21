package controllers

import (
	"net/http"
	"strconv"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseinterfaces"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/baserouter"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basevalidators"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"github.com/gin-gonic/gin"
)

type TitleTypeController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u TitleTypeController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u TitleTypeController) GetCollectionName() basetypes.CollectionName {
	return "titletype"
}

func (u TitleTypeController) DoIndexing() error {
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.TitleType{})
	return nil
}

func (u *TitleTypeController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *TitleTypeController) handleCreateTitleType() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelTitleType := models.TitleType{}
		if err := c.ShouldBind(&modelTitleType); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/put", modelTitleType)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}
		id, err := u.Add(u.GetDBName(), u.GetCollectionName(), modelTitleType, false)
		modelTitleType.Id = int(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_FAILED, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_SUCCESS, err, modelTitleType))
	}
}

func (u *TitleTypeController) handleGetTitleType() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelTitleType := models.TitleType{}
		id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
		modelTitleType.Id = int(id)

		err := u.Validate(c.GetString("apiPath")+"/get", modelTitleType)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		query := map[string]interface{}{"id": modelTitleType.Id}

		if modelTitleType.Id <= 0 {
			query = map[string]interface{}{}
		}

		rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), "", query, &modelTitleType, false, "", false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
			return
		}
		defer rows.Close()

		titleTypes := make([]models.TitleType, 0)

		// Iterate over the rows.
		for rows.Next() {
			// Create a User struct to scan values into.
			tempTitleType := models.TitleType{}

			// Scan the row's values into the User struct.
			err := rows.Scan(&tempTitleType.Id, &tempTitleType.Name, &tempTitleType.Slug)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
				return
			}

			titleTypes = append(titleTypes, tempTitleType)
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, err, titleTypes))
	}
}

func (u *TitleTypeController) handleUpdateTitleType() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelTitleType := models.TitleType{}
		if err := c.ShouldBind(&modelTitleType); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/post", modelTitleType)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET name = $1, slug = $2 WHERE id = $3", []interface{}{modelTitleType.Name, modelTitleType.Slug, modelTitleType.Id}, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_FAILED, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_SUCCESS, err, nil))
	}
}

func (u *TitleTypeController) handleDeleteTitleType() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelTitleType := models.TitleType{}
		if err := c.ShouldBind(&modelTitleType); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/post", modelTitleType)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err = u.DeleteOne(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelTitleType.Id}, false, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_FAILED, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_SUCCESS, err, nil))
	}
}

func (u TitleTypeController) RegisterApis() {
	baserouter.GetInstance().GetLoginRouter().PUT("/api/title_type", u.handleCreateTitleType())
	baserouter.GetInstance().GetLoginRouter().GET("/api/title_type", u.handleGetTitleType())
	baserouter.GetInstance().GetLoginRouter().POST("/api/title_type", u.handleUpdateTitleType())
	baserouter.GetInstance().GetLoginRouter().DELETE("/api/title_type", u.handleDeleteTitleType())
}
