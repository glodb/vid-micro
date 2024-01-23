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
	"com.code.vidmicro/com.code.vidmicro/settings/serviceutils"
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/rs/xid"
)

type TitlesController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u TitlesController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u TitlesController) GetCollectionName() basetypes.CollectionName {
	return "titles"
}

func (u TitlesController) DoIndexing() error {
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.Titles{})
	return nil
}

func (u *TitlesController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *TitlesController) handleCreateTitles() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Save the data in meilisearch
		modelTitles := models.Titles{}
		if err := c.ShouldBind(&modelTitles); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		titlesLanguages := make([]models.TitlesLanguage, 0)

		err := sonic.Unmarshal([]byte(modelTitles.Languages), &titlesLanguages)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		if len(titlesLanguages) <= 0 {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("atleast one language is required"), nil))
			return
		}

		languagesMetadata := make([]interface{}, 0)

		titlesSummary := models.TitlesSummary{}

		for _, titlesLanguage := range titlesLanguages {

			languageMetadata := models.LanguageMeta{LanguageId: titlesLanguage.LanguageId, StatusId: titlesLanguage.StatusId}
			newXID := xid.New()
			languageMetadata.Id = newXID.String()

			if !cache.GetInstance().Exists(fmt.Sprintf("%d%s%s", languageMetadata.LanguageId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguagePostfix)) {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, fmt.Errorf("one of the language is not found id:%d", languageMetadata.LanguageId), nil))
				return
			}

			if !cache.GetInstance().Exists(fmt.Sprintf("%d%s%s", languageMetadata.StatusId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().StatusPostfix)) {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, fmt.Errorf("one of the sstatus is not found id:%d for language:%d", languageMetadata.StatusId, languageMetadata.LanguageId), nil))
				return
			}

			modelTitles.LanguagesMeta = append(modelTitles.LanguagesMeta, languageMetadata.Id)
			languagesMetadata = append(languagesMetadata, languageMetadata)
			titlesSummary.Languages = append(titlesSummary.Languages, titlesLanguage.LanguageId)
		}

		err = u.Validate(c.GetString("apiPath")+"/put", modelTitles)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		file, err := c.FormFile("image")

		if err == nil && file != nil {
			url, err := s3uploader.GetInstance().UploadToSCW(file)
			if err == nil {
				modelTitles.CoverUrl = url
			} else {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPLOADING_AVATAR_FAILED, err, nil))
				return
			}
		}

		id, err := u.Add(u.GetDBName(), u.GetCollectionName(), modelTitles, true)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_FAILED, err, nil))
			return
		}
		for i := range languagesMetadata {
			metadataObject := languagesMetadata[i].(models.LanguageMeta)
			metadataObject.TitlesId = int(id)
			languagesMetadata[i] = metadataObject
		}

		modelTitles.Id = int(id)

		languageMetadataController, _ := u.BaseControllerFactory.GetController(baseconst.LanguageMeta)
		_, err = languageMetadataController.AddMany(languageMetadataController.GetDBName(), languageMetadataController.GetCollectionName(), languagesMetadata, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_FAILED, err, nil))
			return
		}

		titlesSummary.Id = int(id)
		titlesSummary.OriginalTitle = modelTitles.OriginalTitle

		serviceutils.GetInstance().PublishEvent(titlesSummary, configmanager.GetInstance().ClassName, "vidmicro.title.created")

		pattern := "*" + configmanager.GetInstance().ClassName + configmanager.GetInstance().RedisSeprator + configmanager.GetInstance().TitlesPostfix

		keys := cache.GetInstance().GetKeys(pattern)
		cache.GetInstance().DelMany(keys)

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_SUCCESS, err, modelTitles))
	}
}

func (u *TitlesController) handleGetTitles() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelTitles := models.Titles{}
		idString := c.Query("id")
		id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
		page := int64(1)
		pageString := c.Query("page")
		modelTitles.Id = int(id)

		err := u.Validate(c.GetString("apiPath")+"/get", modelTitles)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		query := map[string]interface{}{"id": modelTitles.Id}

		if pageString != "" {
			pageInt, _ := strconv.ParseInt(c.Query("page"), 10, 64)
			page = pageInt
		} else {
			pageString = "1"
		}

		key := pageString + configmanager.GetInstance().RedisSeprator + configmanager.GetInstance().ClassName + configmanager.GetInstance().RedisSeprator + configmanager.GetInstance().TitlesPostfix

		if modelTitles.Id <= 0 {
			query = map[string]interface{}{}
		} else {
			key = pageString + configmanager.GetInstance().RedisSeprator +
				idString +
				configmanager.GetInstance().RedisSeprator +
				configmanager.GetInstance().ClassName +
				configmanager.GetInstance().RedisSeprator +
				configmanager.GetInstance().SingleTitlePostfix +
				configmanager.GetInstance().TitlesPostfix
		}

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

		rows, count, err := u.Paginate(u.GetDBName(), u.GetCollectionName(), "", query, &modelTitles, false, "", false, int(pageSize), int(page))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
			return
		}
		defer rows.Close()

		titles := make([]models.Titles, 0)
		// Iterate over the rows.
		for rows.Next() {
			// Create a User struct to scan values into.
			tempTitle := models.Titles{}

			// Scan the row's values into the User struct.
			err := rows.Scan(&tempTitle.Id, &tempTitle.OriginalTitle, &tempTitle.Year, &tempTitle.CoverUrl, pq.Array(&tempTitle.LanguagesMeta))
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
				return
			}

			controller, _ := u.BaseControllerFactory.GetController(baseconst.LanguageMeta)
			languageController := controller.(*LanguageMetadataController)

			tempTitle.LanguagesDetails, err = languageController.GetLanguageDetails(tempTitle.LanguagesMeta)

			if err != nil {
				tempTitle.LanguagesDetails = make([]models.LanguageMetaDetails, 0)
			}

			titles = append(titles, tempTitle)
		}

		pagination := models.NewPagination(count, int(pageSize), int(page))
		pr := models.PaginationResults{Pagination: pagination, Data: titles}

		if modelTitles.Id <= 0 {
			cache.GetInstance().Set(key, pr.EncodeRedisData())
		} else {
			cache.GetInstance().SetEx(key, pr.EncodeRedisData(), configmanager.GetInstance().TitleExpiryTime)
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, err, pr))
	}
}

func (u *TitlesController) handleUpdateTitles() gin.HandlerFunc {
	return func(c *gin.Context) {
		// modelTitles := models.Titles{}
		// if err := c.ShouldBind(&modelTitles); err != nil {
		// 	c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
		// 	return
		// }

		// err := u.Validate(c.GetString("apiPath")+"/post", modelTitles)
		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
		// 	return
		// }

		// err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET name = $1, slug = $2 WHERE id = $3", []interface{}{modelTitles.Name, modelTitles.Slug, modelTitles.Id}, false)

		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_FAILED, err, nil))
		// 	return
		// }
		// c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_SUCCESS, err, nil))
	}
}

func (u *TitlesController) handleDeleteTitles() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelTitles := models.Titles{}
		if err := c.ShouldBind(&modelTitles); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/post", modelTitles)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err = u.DeleteOne(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelTitles.Id}, false, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_FAILED, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_SUCCESS, err, nil))
	}
}

func (u TitlesController) RegisterApis() {
	baserouter.GetInstance().GetLoginRouter().PUT("/api/titles", u.handleCreateTitles())
	baserouter.GetInstance().GetLoginRouter().GET("/api/titles", u.handleGetTitles())
	baserouter.GetInstance().GetLoginRouter().POST("/api/titles", u.handleUpdateTitles())
	baserouter.GetInstance().GetLoginRouter().DELETE("/api/titles", u.handleDeleteTitles())
}
