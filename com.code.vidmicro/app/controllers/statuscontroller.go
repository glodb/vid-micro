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
	"github.com/gin-gonic/gin"
)

type StatusController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u StatusController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u StatusController) GetCollectionName() basetypes.CollectionName {
	return "status"
}

func (u StatusController) DoIndexing() error {
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.Status{})

	keys := cache.GetInstance().GetKeys("*" + configmanager.GetInstance().StatusPostfix)
	if len(keys) > 0 {
		cache.GetInstance().DelMany(keys)
	}

	rows, _ := u.Find(u.GetDBName(), u.GetCollectionName(), "", map[string]interface{}{}, &models.Status{}, false, "", false)

	defer rows.Close()
	// Iterate over the rows.
	for rows.Next() {
		// Create a User struct to scan values into.

		tempStatus := models.Status{}

		// Scan the row's values into the User struct.
		err := rows.Scan(&tempStatus.Id, &tempStatus.Name)
		if err != nil {
			break
		}

		cache.GetInstance().Set(fmt.Sprintf("%d%s%s", tempStatus.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().StatusPostfix), tempStatus.EncodeRedisData())
	}
	return nil
}

func (u *StatusController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *StatusController) handleCreateStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelStatus := models.Status{}
		if err := c.ShouldBind(&modelStatus); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := basevalidators.GetInstance().GetValidator().Struct(modelStatus)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		id, err := u.Add(u.GetDBName(), u.GetCollectionName(), modelStatus, true)
		modelStatus.Id = int(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		err = cache.GetInstance().Set(fmt.Sprintf("%d%s%s", modelStatus.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().StatusPostfix), modelStatus.EncodeRedisData())

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_SUCCESS, err, modelStatus))
	}
}

func (u *StatusController) handleGetStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelStatus := jsonmodels.IdEmpty{}
		id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
		modelStatus.Id = int(id)

		err := basevalidators.GetInstance().GetValidator().Struct(modelStatus)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		query := map[string]interface{}{"id": modelStatus.Id}

		if modelStatus.Id <= 0 {
			query = map[string]interface{}{}
		}

		rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), "", query, &modelStatus, false, "", false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		defer rows.Close()

		statuses := make([]models.Status, 0)

		// Iterate over the rows.
		for rows.Next() {
			// Create a User struct to scan values into.
			tempStatus := models.Status{}

			// Scan the row's values into the User struct.
			err := rows.Scan(&tempStatus.Id, &tempStatus.Name)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}
			statuses = append(statuses, tempStatus)
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, err, statuses))
	}
}

func (u *StatusController) handleUpdateStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelStatus := jsonmodels.EditStatus{}
		if err := c.ShouldBind(&modelStatus); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := basevalidators.GetInstance().GetValidator().Struct(modelStatus)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET name = $1 WHERE id = $2", []interface{}{modelStatus.Name, modelStatus.Id}, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		err = cache.GetInstance().Set(fmt.Sprintf("%d%s%s", modelStatus.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().StatusPostfix), modelStatus.EncodeRedisData())

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_SUCCESS, err, nil))
	}
}

func (u *StatusController) handleDeleteStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelStatus := jsonmodels.Id{}
		if err := c.ShouldBind(&modelStatus); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := basevalidators.GetInstance().GetValidator().Struct(modelStatus)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		err = u.DeleteOne(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelStatus.Id}, false, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		cache.GetInstance().Del(fmt.Sprintf("%d%s%s", modelStatus.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().StatusPostfix))
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_SUCCESS, err, nil))
	}
}

func (u StatusController) RegisterApis() {
	baserouter.GetInstance().GetLoginRouter().PUT("/api/status", u.handleCreateStatus())
	baserouter.GetInstance().GetLoginRouter().GET("/api/status", u.handleGetStatus())
	baserouter.GetInstance().GetLoginRouter().POST("/api/status", u.handleUpdateStatus())
	baserouter.GetInstance().GetLoginRouter().DELETE("/api/status", u.handleDeleteStatus())
}
