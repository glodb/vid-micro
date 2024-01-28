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
	keys := cache.GetInstance().GetKeys("*" + configmanager.GetInstance().GenresPostfix)
	if len(keys) > 0 {
		cache.GetInstance().DelMany(keys)
	}

	rows, _ := u.Find(u.GetDBName(), u.GetCollectionName(), "", map[string]interface{}{}, &models.Language{}, false, "", false)

	defer rows.Close()
	// Iterate over the rows.
	for rows.Next() {
		// Create a User struct to scan values into.

		tempGenre := models.Genres{}

		// Scan the row's values into the User struct.
		err := rows.Scan(&tempGenre.Id, &tempGenre.Name)
		if err != nil {
			break
		}

		cache.GetInstance().Set(fmt.Sprintf("%d%s%s", tempGenre.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().GenresPostfix), tempGenre.EncodeRedisData())
	}
	return nil
}

func (u *GenresController) GetGenre(id int) (models.Genres, int, error) {
	data, err := cache.GetInstance().Get(fmt.Sprintf("%d%s%s", id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().GenresPostfix))
	if err != nil {
		return models.Genres{}, http.StatusInternalServerError, err
	}
	if len(data) <= 0 {
		return models.Genres{}, http.StatusBadRequest, errors.New("record not available")
	}
	genre := models.Genres{}
	genre.DecodeRedisData(data)

	return genre, http.StatusOK, nil
}

func (u *GenresController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *GenresController) handleCreateGenre() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelGenre := models.Genres{}
		if err := c.ShouldBind(&modelGenre); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/put", modelGenre)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}
		id, err := u.Add(u.GetDBName(), u.GetCollectionName(), modelGenre, true)
		modelGenre.Id = int(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		err = cache.GetInstance().Set(fmt.Sprintf("%d%s%s", modelGenre.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().GenresPostfix), modelGenre.EncodeRedisData())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
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
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		query := map[string]interface{}{"id": modelGenre.Id}

		if modelGenre.Id <= 0 {
			query = map[string]interface{}{}
		}

		rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), "", query, &modelGenre, false, "", false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		defer rows.Close()

		genres := make([]models.Genres, 0)

		// Iterate over the rows.
		for rows.Next() {
			// Create a User struct to scan values into.
			tempGenre := models.Genres{}

			// Scan the row's values into the User struct.
			err := rows.Scan(&tempGenre.Id, &tempGenre.Name)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}
			genres = append(genres, tempGenre)
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, err, genres))
	}
}

func (u *GenresController) handleUpdateGenre() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelGenre := models.Genres{}
		if err := c.ShouldBind(&modelGenre); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/post", modelGenre)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET name = $1 WHERE id = $2", []interface{}{modelGenre.Name, modelGenre.Id}, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		err = cache.GetInstance().Set(fmt.Sprintf("%d%s%s", modelGenre.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().GenresPostfix), modelGenre.EncodeRedisData())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_SUCCESS, err, nil))
	}
}

func (u *GenresController) handleDeleteGenre() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelGenre := models.Genres{}
		if err := c.ShouldBind(&modelGenre); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/post", modelGenre)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err = u.DeleteOne(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelGenre.Id}, false, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		cache.GetInstance().Del(fmt.Sprintf("%d%s%s", modelGenre.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().GenresPostfix))
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_SUCCESS, err, nil))
	}
}

func (u GenresController) RegisterApis() {
	baserouter.GetInstance().GetLoginRouter().PUT("/api/genres", u.handleCreateGenre())
	baserouter.GetInstance().GetLoginRouter().GET("/api/genres", u.handleGetGenre())
	baserouter.GetInstance().GetLoginRouter().POST("/api/genres", u.handleUpdateGenre())
	baserouter.GetInstance().GetLoginRouter().DELETE("/api/genres", u.handleDeleteGenre())
}
