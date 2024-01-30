package controllers

import (
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

type ContentTypeController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u ContentTypeController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u ContentTypeController) GetCollectionName() basetypes.CollectionName {
	return "content_type"
}

func (u ContentTypeController) DoIndexing() error {
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.Contents{})

	keys := cache.GetInstance().GetKeys("*" + configmanager.GetInstance().ContentTypePostfix)
	if len(keys) > 0 {
		cache.GetInstance().DelMany(keys)
	}

	rows, _ := u.Find(u.GetDBName(), u.GetCollectionName(), "", map[string]interface{}{}, &models.ContentType{}, false, "", false)

	defer rows.Close()
	// Iterate over the rows.
	for rows.Next() {
		// Create a User struct to scan values into.

		tempContent := models.ContentType{}

		// Scan the row's values into the User struct.
		err := rows.Scan(&tempContent.Id, &tempContent.Name)
		if err != nil {
			break
		}

		cache.GetInstance().Set(fmt.Sprintf("%d%s%s", tempContent.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTypePostfix), tempContent.EncodeRedisData())
	}
	return nil
}

func (u *ContentTypeController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *ContentTypeController) handleCreateContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelContent := models.ContentType{}
		if err := c.ShouldBind(&modelContent); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		// TODO:
		// err := u.Validate(c.GetString("apiPath")+"/put", modelContent)
		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
		// 	return
		// }
		id, err := u.Add(u.GetDBName(), u.GetCollectionName(), modelContent, true)
		modelContent.Id = int(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		err = cache.GetInstance().Set(fmt.Sprintf("%d%s%s", modelContent.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTypePostfix), modelContent.EncodeRedisData())

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_SUCCESS, err, modelContent))
	}
}

func (u *ContentTypeController) handleGetContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelContent := models.ContentType{}
		id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
		modelContent.Id = int(id)

		// TODO:
		// err := u.Validate(c.GetString("apiPath")+"/get", modelContent)
		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
		// 	return
		// }

		query := map[string]interface{}{"id": modelContent.Id}

		if modelContent.Id <= 0 {
			query = map[string]interface{}{}
		}

		rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), "", query, &modelContent, false, "", false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		defer rows.Close()

		content_typees := make([]models.ContentType, 0)

		// Iterate over the rows.
		for rows.Next() {
			// Create a User struct to scan values into.
			tempContent := models.ContentType{}

			// Scan the row's values into the User struct.
			err := rows.Scan(&tempContent.Id, &tempContent.Name)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}
			content_typees = append(content_typees, tempContent)
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, err, content_typees))
	}
}

func (u *ContentTypeController) handleUpdateContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelContent := models.ContentType{}
		if err := c.ShouldBind(&modelContent); err != nil {
			c.AbortWithStatusJSON(http.StatusBadGateway, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		// TODO:
		// err := u.Validate(c.GetString("apiPath")+"/post", modelContent)
		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
		// 	return
		// }

		err := u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET name = $1 WHERE id = $2", []interface{}{modelContent.Name, modelContent.Id}, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		err = cache.GetInstance().Set(fmt.Sprintf("%d%s%s", modelContent.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTypePostfix), modelContent.EncodeRedisData())

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_SUCCESS, err, nil))
	}
}

func (u *ContentTypeController) handleDeleteContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelContent := models.ContentType{}
		if err := c.ShouldBind(&modelContent); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		// TODO:
		// err := u.Validate(c.GetString("apiPath")+"/post", modelContent)
		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
		// 	return
		// }

		err := u.DeleteOne(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelContent.Id}, false, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		cache.GetInstance().Del(fmt.Sprintf("%d%s%s", modelContent.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTypePostfix))
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_SUCCESS, err, nil))
	}
}

func (u ContentTypeController) RegisterApis() {
	baserouter.GetInstance().GetLoginRouter().PUT("/api/content_type", u.handleCreateContentType())
	baserouter.GetInstance().GetLoginRouter().GET("/api/content_type", u.handleGetContentType())
	baserouter.GetInstance().GetLoginRouter().POST("/api/content_type", u.handleUpdateContentType())
	baserouter.GetInstance().GetLoginRouter().DELETE("/api/content_type", u.handleDeleteContentType())
}
