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

type GenresController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u GenresController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u GenresController) GetCollectionName() basetypes.CollectionName {
	return "genres"
}

func (u GenresController) DoIndexing() error {
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.Genres{})
	return nil
}

func (u *GenresController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *GenresController) handleCreateGenre() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelGenre := models.Genres{}
		if err := c.ShouldBind(&modelGenre); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/put", modelGenre)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}
		id, err := u.Add(u.GetDBName(), u.GetCollectionName(), modelGenre, true)
		modelGenre.Id = int(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_FAILED, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_SUCCESS, err, modelGenre))
	}
}

func (u *GenresController) handleGetGenre() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelGenre := models.Genres{}

		id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
		modelGenre.Id = int(id)

		err := u.Validate(c.GetString("apiPath")+"/get", modelGenre)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		rows, err := u.FindOne(u.GetDBName(), u.GetCollectionName(), "", map[string]interface{}{"id": modelGenre.Id}, &modelGenre, false, " Limit 1", false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
			return
		}
		defer rows.Close()

		// Iterate over the rows.
		for rows.Next() {
			// Create a User struct to scan values into.

			// Scan the row's values into the User struct.
			err := rows.Scan(&modelGenre.Id, &modelGenre.Name)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, err, modelGenre))
	}
}

func (u *GenresController) handleUpdateGenre() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelGenre := models.Genres{}
		if err := c.ShouldBind(&modelGenre); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/post", modelGenre)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET name = $1 WHERE id = $2", []interface{}{modelGenre.Name, modelGenre.Id}, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_FAILED, err, nil))
			return
		}

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_SUCCESS, err, nil))
	}
}

func (u *GenresController) handleDeleteGenre() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelGenre := models.Genres{}
		if err := c.ShouldBind(&modelGenre); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/post", modelGenre)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err = u.DeleteOne(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelGenre.Id}, false, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_FAILED, err, nil))
			return
		}

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_SUCCESS, err, nil))
	}
}

func (u GenresController) RegisterApis() {
	baserouter.GetInstance().GetLoginRouter().PUT("/api/genres", u.handleCreateGenre())
	baserouter.GetInstance().GetLoginRouter().GET("/api/genres", u.handleGetGenre())
	baserouter.GetInstance().GetLoginRouter().POST("/api/genres", u.handleUpdateGenre())
	baserouter.GetInstance().GetLoginRouter().DELETE("/api/genres", u.handleDeleteGenre())
}
