package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseinterfaces"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/baserouter"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basevalidators"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"com.code.vidmicro/com.code.vidmicro/settings/cache"
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
	keys := cache.GetInstance().GetKeys("*" + configmanager.GetInstance().TypePostfix)
	cache.GetInstance().DelMany(keys)

	rows, _ := u.Find(u.GetDBName(), u.GetCollectionName(), "", map[string]interface{}{}, &models.Language{}, false, "", false)

	defer rows.Close()
	// Iterate over the rows.
	for rows.Next() {
		// Create a User struct to scan values into.

		tempTitle := models.TitleType{}

		// Scan the row's values into the User struct.
		err := rows.Scan(&tempTitle.Id, &tempTitle.Name, &tempTitle.Slug)
		if err != nil {
			break
		}

		cache.GetInstance().Set(fmt.Sprintf("%d%s%s", tempTitle.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().TypePostfix), tempTitle.EncodeRedisData())
	}
	return nil
}

func (u *TitleTypeController) GetTitleType(id int) (models.TitleType, int, error) {
	data, err := cache.GetInstance().Get(fmt.Sprintf("%d%s%s", id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().TypePostfix))
	if err != nil {
		return models.TitleType{}, http.StatusInternalServerError, err
	}
	if len(data) <= 0 {
		return models.TitleType{}, http.StatusBadRequest, errors.New("record not available")
	}
	titleType := models.TitleType{}
	titleType.DecodeRedisData(data)

	return titleType, http.StatusOK, nil
}

func (u *TitleTypeController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *TitleTypeController) handleCreateTitleType() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelTitleType := models.TitleType{}
		if err := c.ShouldBind(&modelTitleType); err != nil {
			c.AbortWithStatusJSON(http.StatusBadGateway, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/put", modelTitleType)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}
		id, err := u.Add(u.GetDBName(), u.GetCollectionName(), modelTitleType, true)
		modelTitleType.Id = int(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		err = cache.GetInstance().Set(fmt.Sprintf("%d%s%s", modelTitleType.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().TypePostfix), modelTitleType.EncodeRedisData())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
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
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		query := map[string]interface{}{"id": modelTitleType.Id}

		if modelTitleType.Id <= 0 {
			query = map[string]interface{}{}
		}

		rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), "", query, &modelTitleType, false, "", false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
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
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
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
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/post", modelTitleType)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET name = $1, slug = $2 WHERE id = $3", []interface{}{modelTitleType.Name, modelTitleType.Slug, modelTitleType.Id}, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		err = cache.GetInstance().Set(fmt.Sprintf("%d%s%s", modelTitleType.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().TypePostfix), modelTitleType.EncodeRedisData())

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_SUCCESS, err, nil))
	}
}

func (u *TitleTypeController) handleDeleteTitleType() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelTitleType := models.TitleType{}
		if err := c.ShouldBind(&modelTitleType); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/post", modelTitleType)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err = u.DeleteOne(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelTitleType.Id}, false, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		cache.GetInstance().Del(fmt.Sprintf("%d%s%s", modelTitleType.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().TypePostfix))
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_SUCCESS, err, nil))
	}
}

func (u TitleTypeController) RegisterApis() {
	baserouter.GetInstance().GetLoginRouter().PUT("/api/title_type", u.handleCreateTitleType())
	baserouter.GetInstance().GetLoginRouter().GET("/api/title_type", u.handleGetTitleType())
	baserouter.GetInstance().GetLoginRouter().POST("/api/title_type", u.handleUpdateTitleType())
	baserouter.GetInstance().GetLoginRouter().DELETE("/api/title_type", u.handleDeleteTitleType())
}
