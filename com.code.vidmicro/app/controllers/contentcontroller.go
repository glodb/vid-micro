package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseconst"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseinterfaces"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/baserouter"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basevalidators"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"com.code.vidmicro/com.code.vidmicro/settings/cache"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/s3uploader"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type ContentController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u ContentController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u ContentController) GetCollectionName() basetypes.CollectionName {
	return "content"
}

func (u ContentController) DoIndexing() error {
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.Contents{})
	return nil
}

func (u *ContentController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *ContentController) validateTitleLanguage(content models.Contents) bool {
	languageExist, _ := cache.GetInstance().Exists(fmt.Sprintf("%d%s%s", content.AssociatedTitle, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTitleLanguagesPostfix))
	if languageExist {
		return cache.GetInstance().SetISMember(fmt.Sprintf("%d%s%s", content.AssociatedTitle, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTitleLanguagesPostfix), fmt.Sprintf("%d", content.LanguageId))
	} else {
		titleSummary, err := u.BaseControllerFactory.GetController(baseconst.TitlesSummary)
		if err != nil {
			return false
		}

		titleRows, err := titleSummary.Find(titleSummary.GetDBName(), titleSummary.GetCollectionName(), "languages_meta", map[string]interface{}{}, &models.TitlesSummary{}, false, "", false)

		if err != nil {
			return false
		}
		tempTitle := models.TitlesSummary{}

		var arrayData pq.Int32Array
		for titleRows.Next() {
			// Scan the row's values into the User struct.
			err := titleRows.Scan(&arrayData)
			if err != nil {
				return false
			}
		}

		contains := false
		contents := make([]interface{}, 0)
		contents = append(contents, fmt.Sprintf("%d%s%s", content.AssociatedTitle, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTitleLanguagesPostfix))

		tempTitle.Languages = make([]int, len(arrayData))
		for i, v := range arrayData {
			tempTitle.Languages[i] = int(v)
			contents = append(contents, int(v))
			if int(v) == content.Id {
				contains = true
			}
		}

		cache.GetInstance().SAdd(contents)
		cache.GetInstance().Expire(fmt.Sprintf("%d%s%s", content.AssociatedTitle, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTitleLanguagesPostfix), configmanager.GetInstance().TitlesLanguageExpirationTime)
		return contains
	}

}

func (u *ContentController) handleCreateContent() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelContent := models.Contents{}
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

		if !u.validateTitleLanguage(modelContent) {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.LANGUAGE_NOT_ADDED_IN_TITLE, errors.New("validating language failed"), nil))
			return
		}

		statusExist, err := cache.GetInstance().Exists(fmt.Sprintf("%d%s%s", modelContent.TypeId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTypePostfix))

		if !statusExist {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, errors.New("status id is not correct"), nil))
			return
		}

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		id, err := u.Add(u.GetDBName(), u.GetCollectionName(), modelContent, true)
		modelContent.Id = int(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		file, err := c.FormFile("image")

		if err == nil && file != nil {
			url, httpCode, err := s3uploader.GetInstance().UploadToSCW(file)
			if err == nil {
				modelContent.Thumbnail = url
			} else {
				c.AbortWithStatusJSON(httpCode, responses.GetInstance().WriteResponse(c, responses.UPLOADING_AVATAR_FAILED, err, nil))
				return
			}
		}
		keys := cache.GetInstance().GetKeys(fmt.Sprintf("*%d%s%s", modelContent.AssociatedTitle, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentPostFix))
		if len(keys) > 0 {
			cache.GetInstance().DelMany(keys)
		}

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_SUCCESS, err, nil))
	}
}

func (u *ContentController) handleGetContent() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelContent := models.Contents{}
		page := int64(1)
		pageString := c.Query("page")

		query := map[string]interface{}{}
		key := ""

		if c.Query("associated_title") != "" {
			associatedTitle := c.Query("associated_title")
			associatedId, _ := strconv.ParseInt(associatedTitle, 10, 64)
			query["associated_title"] = associatedId
			key = associatedTitle + configmanager.GetInstance().RedisSeprator
			modelContent.AssociatedTitle = int(associatedId)
		}

		if c.Query("lang_id") != "" {
			languageTitle := c.Query("lang_id")
			languageId, _ := strconv.ParseInt(languageTitle, 10, 64)
			query["language_id"] = languageId
			key = languageTitle + configmanager.GetInstance().RedisSeprator + "_lang_id_" + key
			modelContent.LanguageId = int(languageId)
		}

		if c.Query("content_type_id") != "" {
			contentTypeIdString := c.Query("content_type_id")
			contentType, _ := strconv.ParseInt(contentTypeIdString, 10, 64)
			query["type_id"] = contentType
			key = contentTypeIdString + configmanager.GetInstance().RedisSeprator + "_type_id_" + key
			modelContent.TypeId = int(contentType)
		}

		if c.Query("id") != "" {
			idString := c.Query("id")
			idType, _ := strconv.ParseInt(idString, 10, 64)
			query["id"] = idType
			key = idString + configmanager.GetInstance().RedisSeprator + "_id_" + key
			modelContent.TypeId = int(idType)
		}

		// TODO:
		// err := u.Validate(c.GetString("apiPath")+"/get", modelContent)

		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
		// 	return
		// }

		if pageString != "" {
			pageInt, _ := strconv.ParseInt(c.Query("page"), 10, 64)
			page = pageInt
		} else {
			pageString = "1"
		}

		key = pageString + configmanager.GetInstance().RedisSeprator + key
		key = key + configmanager.GetInstance().ContentPostFix

		if data, err := cache.GetInstance().Get(key); err == nil && len(data) > 0 {
			pr := models.PaginationResults{}
			pr.DecodeRedisData(data)
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, err, pr))
			return //Found in cache
		}

		pageSize := configmanager.GetInstance().PageSize

		if c.Query("limit") != "" {
			pageSizeInt, _ := strconv.ParseInt(c.Query("limit"), 10, 64)
			pageSize = pageSizeInt
		}

		rows, count, err := u.Paginate(u.GetDBName(), u.GetCollectionName(), "", query, &modelContent, false, "", false, int(pageSize), int(page))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		defer rows.Close()

		contents := make([]models.Contents, 0)
		// Iterate over the rows.
		for rows.Next() {
			// Create a User struct to scan values into.
			tempContent := models.Contents{}

			// Scan the row's values into the User struct.
			err := rows.Scan(&tempContent.Id, &tempContent.Name, &tempContent.AlternativeName, &tempContent.Thumbnail, &tempContent.Description, &tempContent.TypeId, &tempContent.LanguageId, &tempContent.AssociatedTitle)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}

			contentTypeData, err := cache.GetInstance().Get(fmt.Sprintf("%d%s%s", tempContent.TypeId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTypePostfix))
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}

			if len(contentTypeData) <= 0 {
				c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, errors.New("content data is not available"), nil))
				return
			}

			languageTypeData, err := cache.GetInstance().Get(fmt.Sprintf("%d%s%s", tempContent.LanguageId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguagePostfix))

			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}

			if len(languageTypeData) <= 0 {
				c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, errors.New("languae data is not available"), nil))
				return
			}

			modelLanguage := models.LanguageContent{}
			modelLanguage.DecodeRedisData(languageTypeData)

			modelContentType := models.ContentType{}
			modelContentType.DecodeRedisData(contentTypeData)

			tempContent.LanguageCode = modelLanguage.Code
			tempContent.LanguageName = modelLanguage.Name

			tempContent.TypeName = modelContentType.Name
			contents = append(contents, tempContent)
		}

		pagination := models.NewPagination(count, int(pageSize), int(page))
		pr := models.PaginationResults{Pagination: pagination, Data: contents}

		err = cache.GetInstance().Set(key, pr.EncodeRedisData())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, err, pr))
	}
}

func (u *ContentController) handleUpdateContent() gin.HandlerFunc {
	return func(c *gin.Context) {

		setPart := " SET "
		data := make([]interface{}, 0)
		modelContent := models.Contents{}
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

		file, err := c.FormFile("image")

		if err == nil && file != nil {
			url, httpCode, err := s3uploader.GetInstance().UploadToSCW(file)
			if err == nil {
				modelContent.Thumbnail = url
				setPart += "thumbnail = $1"
				data = append(data, url)

			} else {
				c.AbortWithStatusJSON(httpCode, responses.GetInstance().WriteResponse(c, responses.UPLOADING_AVATAR_FAILED, err, nil))
				return
			}
		}

		if modelContent.Name != "" {
			lengthString := strconv.FormatInt(int64(len(data)+1), 10)
			if len(data) > 0 {
				setPart += ","
			}
			setPart += "name = $" + lengthString
			data = append(data, modelContent.Name)
		}

		if modelContent.AlternativeName != "" {
			lengthString := strconv.FormatInt(int64(len(data)+1), 10)
			if len(data) > 0 {
				setPart += ","
			}
			setPart += "alternative_name = $" + lengthString
			data = append(data, modelContent.AlternativeName)
		}

		if modelContent.Description != "" {
			lengthString := strconv.FormatInt(int64(len(data)+1), 10)
			if len(data) > 0 {
				setPart += ","
			}
			setPart += "description = $" + lengthString
			data = append(data, modelContent.Description)
		}

		if modelContent.TypeId >= 0 {
			typeExist, err := cache.GetInstance().Exists(fmt.Sprintf("%d%s%s", modelContent.TypeId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTypePostfix))
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}
			if typeExist {
				lengthString := strconv.FormatInt(int64(len(data)+1), 10)
				if len(data) > 0 {
					setPart += ","
				}
				setPart += "type_id = $" + lengthString
				data = append(data, modelContent.TypeId)
			} else {
				c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, errors.New("one of the type doesn't exists"), nil))
				return
			}
		}

		if modelContent.LanguageId >= 0 {
			if u.validateTitleLanguage(modelContent) {
				lengthString := strconv.FormatInt(int64(len(data)+1), 10)
				if len(data) > 0 {
					setPart += ","
				}
				setPart += "language_id = $" + lengthString
				data = append(data, modelContent.LanguageId)
			}
		}

		if len(data) > 0 {
			lengthString := strconv.FormatInt(int64(len(data)+1), 10)
			setPart += " WHERE id =$" + lengthString
			data = append(data, modelContent.Id)

			lengthString = strconv.FormatInt(int64(len(data)+1), 10)
			setPart += " and associated_title =$" + lengthString
			data = append(data, modelContent.AssociatedTitle)

			err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+setPart, data, false)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}
			keys := cache.GetInstance().GetKeys(fmt.Sprintf("*%d%s%s", modelContent.AssociatedTitle, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentPostFix))
			if len(keys) > 0 {
				cache.GetInstance().DelMany(keys)
			}
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_SUCCESS, err, nil))
			return
		}

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.NOTHIN_TO_UPDATE, err, nil))
	}
}

func (u *ContentController) handleDeleteContent() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelContent := models.Contents{}
		if err := c.ShouldBind(&modelContent); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		// TODO:
		// err := u.Validate(c.GetString("apiPath")+"/delete", modelContent)
		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
		// 	return
		// }

		err := u.DeleteOne(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelContent.Id}, false, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		keys := cache.GetInstance().GetKeys(fmt.Sprintf("*%d%s%s", modelContent.AssociatedTitle, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentPostFix))
		if len(keys) > 0 {
			cache.GetInstance().DelMany(keys)
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_SUCCESS, err, nil))
	}
}

func (u ContentController) RegisterApis() {
	baserouter.GetInstance().GetLoginRouter().PUT("/api/content", u.handleCreateContent())
	baserouter.GetInstance().GetLoginRouter().GET("/api/content", u.handleGetContent())
	baserouter.GetInstance().GetLoginRouter().POST("/api/content", u.handleUpdateContent())
	baserouter.GetInstance().GetLoginRouter().DELETE("/api/content", u.handleDeleteContent())
}
