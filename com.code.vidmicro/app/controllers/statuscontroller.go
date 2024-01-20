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
	return nil
}

func (u *StatusController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *StatusController) handleCreateStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelStatus := models.Status{}
		if err := c.ShouldBind(&modelStatus); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/put", modelStatus)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}
		id, err := u.Add(u.GetDBName(), u.GetCollectionName(), modelStatus, true)
		modelStatus.Id = int(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_FAILED, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_SUCCESS, err, modelStatus))
	}
}

func (u *StatusController) handleGetStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelStatus := models.Status{}
		id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
		modelStatus.Id = int(id)

		err := u.Validate(c.GetString("apiPath")+"/get", modelStatus)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		rows, err := u.FindOne(u.GetDBName(), u.GetCollectionName(), "", map[string]interface{}{"id": modelStatus.Id}, &modelStatus, false, " Limit 1", false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
			return
		}
		defer rows.Close()

		// Iterate over the rows.
		for rows.Next() {
			// Create a User struct to scan values into.

			// Scan the row's values into the User struct.
			err := rows.Scan(&modelStatus.Id, &modelStatus.Name)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, err, modelStatus))
	}
}

func (u *StatusController) handleUpdateStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelStatus := models.Status{}
		if err := c.ShouldBind(&modelStatus); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/post", modelStatus)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET name = $1 WHERE id = $2", []interface{}{modelStatus.Name, modelStatus.Id}, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_FAILED, err, nil))
			return
		}

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_SUCCESS, err, nil))
	}
}

func (u *StatusController) handleDeleteStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelStatus := models.Status{}
		if err := c.ShouldBind(&modelStatus); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/post", modelStatus)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err = u.DeleteOne(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelStatus.Id}, false, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_FAILED, err, nil))
			return
		}

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_SUCCESS, err, nil))
	}
}

func (u StatusController) RegisterApis() {
	baserouter.GetInstance().GetLoginRouter().PUT("/api/status", u.handleCreateStatus())
	baserouter.GetInstance().GetLoginRouter().GET("/api/status", u.handleGetStatus())
	baserouter.GetInstance().GetLoginRouter().POST("/api/status", u.handleUpdateStatus())
	baserouter.GetInstance().GetLoginRouter().DELETE("/api/status", u.handleDeleteStatus())
}
