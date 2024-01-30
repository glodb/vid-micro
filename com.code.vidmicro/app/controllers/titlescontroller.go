package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/app/models/jsonmodels"
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
	"com.code.vidmicro/com.code.vidmicro/settings/searchengine"
	"com.code.vidmicro/com.code.vidmicro/settings/serviceutils"
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/rs/xid"
	"golang.org/x/sync/semaphore"
)

type TitlesController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
	semaphore *semaphore.Weighted
}

func (u TitlesController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u TitlesController) GetCollectionName() basetypes.CollectionName {
	return "titles"
}

func (u TitlesController) DoIndexing() error {
	u.semaphore = semaphore.NewWeighted(int64(configmanager.GetInstance().MaxMeiliSearchUpdates))
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.Titles{})
	return nil
}

func (u *TitlesController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *TitlesController) handleCreateTitles() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelTitles := models.Titles{}
		if err := c.ShouldBind(&modelTitles); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := basevalidators.GetInstance().GetValidator().Struct(modelTitles)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		titlesLanguages := make([]models.TitlesLanguage, 0)

		sonic.Unmarshal([]byte(modelTitles.Languages), &titlesLanguages)

		languagesMetadata := make([]interface{}, 0)

		titlesSummary := models.TitlesSummary{}

		for _, titlesLanguage := range titlesLanguages {

			languageMetadata := models.LanguageMeta{LanguageId: titlesLanguage.LanguageId, StatusId: titlesLanguage.StatusId}
			newXID := xid.New()
			languageMetadata.Id = newXID.String()

			languageMetaDetails := models.LanguageMetaDetails{LanguageId: titlesLanguage.LanguageId, StatusId: titlesLanguage.StatusId}
			languageData, err := cache.GetInstance().Get(fmt.Sprintf("%d%s%s", languageMetadata.LanguageId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguagePostfix))

			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, nil, nil))
				return
			}
			if len(languageData) <= 0 {
				c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, fmt.Errorf("one of the language is not found id:%d", languageMetadata.LanguageId), nil))
				return
			}

			lang := models.Language{}
			lang.DecodeRedisData(languageData)
			languageMetaDetails.LanguageCode = lang.Code
			languageMetaDetails.LanguageName = lang.Name

			statusData, err := cache.GetInstance().Get(fmt.Sprintf("%d%s%s", languageMetadata.StatusId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().StatusPostfix))

			if err != nil && len(statusData) <= 0 {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, nil, nil))
				return
			}

			if len(statusData) <= 0 {
				c.AbortWithStatusJSON(http.StatusBadGateway, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, fmt.Errorf("one of the status is not found id:%d for language:%d", languageMetadata.StatusId, languageMetadata.LanguageId), nil))
				return
			}

			stat := models.Status{}
			stat.DecodeRedisData(statusData)
			languageMetaDetails.StatusId = stat.Id
			languageMetaDetails.StatusName = stat.Name

			modelTitles.LanguagesDetails = append(modelTitles.LanguagesDetails, languageMetaDetails)
			modelTitles.LanguagesMeta = append(modelTitles.LanguagesMeta, languageMetadata.Id)
			languagesMetadata = append(languagesMetadata, languageMetadata)
			titlesSummary.Languages = append(titlesSummary.Languages, titlesLanguage.LanguageId)
		}

		file, err := c.FormFile("image")

		if err == nil && file != nil {
			url, statusCode, err := s3uploader.GetInstance().UploadToSCW(file)
			if err == nil {
				modelTitles.CoverUrl = url
			} else {
				c.AbortWithStatusJSON(statusCode, responses.GetInstance().WriteResponse(c, responses.UPLOADING_AVATAR_FAILED, err, nil))
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
		if len(keys) > 0 {
			cache.GetInstance().DelMany(keys)
		}

		tempMeiliTitle := models.MeilisearchTitle{
			Id:               modelTitles.Id,
			OriginalTitle:    modelTitles.OriginalTitle,
			CoverUrl:         modelTitles.CoverUrl,
			LanguagesDetails: modelTitles.LanguagesDetails,
			Year:             modelTitles.Year,
		}

		searchengine.GetInstance().ProcessTitleDocuments(tempMeiliTitle)

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_SUCCESS, err, modelTitles))
	}
}

func (u *TitlesController) GetSingleTitle(id int) (models.Titles, int, error) {
	titles := make([]models.Titles, 0)
	key := "1" + configmanager.GetInstance().RedisSeprator +
		fmt.Sprintf("%d", id) +
		configmanager.GetInstance().RedisSeprator +
		configmanager.GetInstance().ClassName +
		configmanager.GetInstance().RedisSeprator +
		configmanager.GetInstance().SingleTitlePostfix +
		configmanager.GetInstance().TitlesPostfix

	if data, err := cache.GetInstance().Get(key); err == nil && len(data) > 0 {
		pr := models.PaginationResults{}
		pr.DecodeRedisData(data)
		jsonData, err := sonic.Marshal(pr.Data)

		if err != nil {
			return models.Titles{}, http.StatusInternalServerError, err
		}

		err = sonic.Unmarshal(jsonData, &titles)
		if err != nil {
			return models.Titles{}, http.StatusInternalServerError, err
		}
	} else {
		query := map[string]interface{}{"id": id}

		rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), "", query, &models.Titles{}, false, " LIMIT 1", false)
		if err != nil {
			return models.Titles{}, http.StatusInternalServerError, err
		}

		for rows.Next() {
			tempTitle := models.Titles{}
			coverUrl := sql.NullString{}
			err := rows.Scan(&tempTitle.Id, &tempTitle.OriginalTitle, &tempTitle.Year, &coverUrl, pq.Array(&tempTitle.LanguagesMeta))
			if err != nil {
				return models.Titles{}, http.StatusInternalServerError, err
			}
			tempTitle.CoverUrl = coverUrl.String
			controller, _ := u.BaseControllerFactory.GetController(baseconst.LanguageMeta)
			languageController := controller.(*LanguageMetadataController)

			langDetails, statusCode, err := languageController.GetLanguageDetails(tempTitle.LanguagesMeta)

			if err != nil {
				return models.Titles{}, statusCode, err
			}
			tempTitle.LanguagesDetails = langDetails
			titles = append(titles, tempTitle)
		}

		pagination := models.NewPagination(1, int(configmanager.GetInstance().PageSize), 1)
		pr := models.PaginationResults{Pagination: pagination, Data: titles}

		err = cache.GetInstance().SetEx(key, pr.EncodeRedisData(), configmanager.GetInstance().TitleExpiryTime)
		if err != nil {
			return models.Titles{}, http.StatusInternalServerError, err
		}
	}
	return titles[0], http.StatusOK, nil
}

func (u *TitlesController) handleGetTitles() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelTitles := jsonmodels.IdEmpty{}
		idString := c.Query("id")
		id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
		page := int64(1)
		pageString := c.Query("page")
		modelTitles.Id = int(id)

		err := basevalidators.GetInstance().GetValidator().Struct(modelTitles)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
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
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		defer rows.Close()

		titles := make([]models.Titles, 0)
		// Iterate over the rows.
		for rows.Next() {
			// Create a User struct to scan values into.
			tempTitle := models.Titles{}

			coverUrl := sql.NullString{}
			// Scan the row's values into the User struct.
			err := rows.Scan(&tempTitle.Id, &tempTitle.OriginalTitle, &tempTitle.Year, &coverUrl, pq.Array(&tempTitle.LanguagesMeta))
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}
			tempTitle.CoverUrl = coverUrl.String

			controller, _ := u.BaseControllerFactory.GetController(baseconst.LanguageMeta)
			languageController := controller.(*LanguageMetadataController)

			langDetails, statusCode, err := languageController.GetLanguageDetails(tempTitle.LanguagesMeta)
			tempTitle.LanguagesDetails = langDetails

			if err != nil {
				responseMessage := responses.SERVER_ERROR
				if statusCode == http.StatusBadRequest {
					responseMessage = responses.BAD_REQUEST
				}
				c.AbortWithStatusJSON(statusCode, responses.GetInstance().WriteResponse(c, responseMessage, err, nil))
				return
			}

			titles = append(titles, tempTitle)
		}

		pagination := models.NewPagination(count, int(pageSize), int(page))
		pr := models.PaginationResults{Pagination: pagination, Data: titles}

		if modelTitles.Id <= 0 {
			err = cache.GetInstance().Set(key, pr.EncodeRedisData())
		} else {
			err = cache.GetInstance().SetEx(key, pr.EncodeRedisData(), configmanager.GetInstance().TitleExpiryTime)
		}

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, err, pr))
	}
}

func (u *TitlesController) handleUpdateTitles() gin.HandlerFunc {
	return func(c *gin.Context) {
		setPart := " SET "
		data := make([]interface{}, 0)
		modelTitles := jsonmodels.EditTitles{}
		if err := c.ShouldBind(&modelTitles); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := basevalidators.GetInstance().GetValidator().Struct(modelTitles)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		file, err := c.FormFile("image")
		updateTitle := false

		if err == nil && file != nil {
			url, statusCode, err := s3uploader.GetInstance().UploadToSCW(file)
			if err == nil {
				modelTitles.CoverUrl = url
				setPart += "thumbnail = $1"
				data = append(data, url)

			} else {
				c.AbortWithStatusJSON(statusCode, responses.GetInstance().WriteResponse(c, responses.UPLOADING_AVATAR_FAILED, err, nil))
				return
			}
		}

		if modelTitles.OriginalTitle != "" {
			lengthString := strconv.FormatInt(int64(len(data)+1), 10)
			if len(data) > 0 {
				setPart += ","
			}
			setPart += "original_title = $" + lengthString
			data = append(data, modelTitles.OriginalTitle)
			updateTitle = true
		}

		if modelTitles.Year >= 0 {
			lengthString := strconv.FormatInt(int64(len(data)+1), 10)
			if len(data) > 0 {
				setPart += ","
			}
			setPart += "year = $" + lengthString
			data = append(data, modelTitles.Year)
		}

		if len(data) > 0 {
			lengthString := strconv.FormatInt(int64(len(data)+1), 10)
			setPart += " WHERE id =$" + lengthString
			data = append(data, modelTitles.Id)

			err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+setPart, data, false)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}
			pattern := "*" + configmanager.GetInstance().ClassName + configmanager.GetInstance().RedisSeprator + configmanager.GetInstance().TitlesPostfix

			keys := cache.GetInstance().GetKeys(pattern)
			if len(keys) > 0 {
				cache.GetInstance().DelMany(keys)
			}

			key := "1" + configmanager.GetInstance().RedisSeprator +
				fmt.Sprintf("%d", modelTitles.Id) +
				configmanager.GetInstance().RedisSeprator +
				configmanager.GetInstance().ClassName +
				configmanager.GetInstance().RedisSeprator +
				configmanager.GetInstance().SingleTitlePostfix +
				configmanager.GetInstance().TitlesPostfix

			cache.GetInstance().Del(key)

			if updateTitle {
				titleSummary := models.TitlesSummary{Id: modelTitles.Id, OriginalTitle: modelTitles.OriginalTitle}

				controller, _ := u.BaseControllerFactory.GetController(baseconst.TitleMeta)
				metaController := controller.(*TitleMetaController)
				metaController.ChangeTitleName(modelTitles.Id, modelTitles.OriginalTitle)

				serviceutils.GetInstance().PublishEvent(titleSummary, configmanager.GetInstance().ClassName, "vidmicro.title.updated")
			}

			tempMeiliTitle := models.MeilisearchTitle{
				Id:            modelTitles.Id,
				OriginalTitle: modelTitles.OriginalTitle,
				CoverUrl:      modelTitles.CoverUrl,
				Year:          modelTitles.Year,
			}
			searchengine.GetInstance().ProcessTitleDocuments(tempMeiliTitle)

			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_SUCCESS, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.NOTHIN_TO_UPDATE, err, nil))
	}
}

func (u *TitlesController) handleDeleteTitles() gin.HandlerFunc {
	return func(c *gin.Context) {

		modelTitles := jsonmodels.Id{}
		if err := c.ShouldBind(&modelTitles); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}
		err := basevalidators.GetInstance().GetValidator().Struct(modelTitles)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		modelMeiliTitles := models.MeilisearchTitle{}
		err = searchengine.GetInstance().DeleteDocuments(modelMeiliTitles)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		err = u.DeleteOne(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelTitles.Id}, false, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		languageMetaController, _ := u.BaseControllerFactory.GetController(baseconst.LanguageMeta)
		err = languageMetaController.DeleteOne(languageMetaController.GetDBName(), languageMetaController.GetCollectionName(), map[string]interface{}{"titles_id": modelTitles.Id}, false, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		titleMetaController, _ := u.BaseControllerFactory.GetController(baseconst.TitleMeta)
		err = titleMetaController.DeleteOne(titleMetaController.GetDBName(), titleMetaController.GetCollectionName(), map[string]interface{}{"title_id": modelTitles.Id}, false, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		pattern := "*" + configmanager.GetInstance().ClassName + configmanager.GetInstance().RedisSeprator + configmanager.GetInstance().TitlesPostfix

		keys := cache.GetInstance().GetKeys(pattern)
		if len(keys) > 0 {
			cache.GetInstance().DelMany(keys)
		}

		key := "1" + configmanager.GetInstance().RedisSeprator +
			fmt.Sprintf("%d", modelTitles.Id) +
			configmanager.GetInstance().RedisSeprator +
			configmanager.GetInstance().ClassName +
			configmanager.GetInstance().RedisSeprator +
			configmanager.GetInstance().SingleTitlePostfix +
			configmanager.GetInstance().TitlesPostfix

		err = cache.GetInstance().Del(key)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		titleSummary := models.TitlesSummary{Id: modelTitles.Id}

		searchengine.GetInstance().DeleteDocuments(models.MeilisearchTitle{Id: modelTitles.Id})
		serviceutils.GetInstance().PublishEvent(titleSummary, configmanager.GetInstance().ClassName, "vidmicro.title.deleted")

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_SUCCESS, err, nil))
	}
}

func (u *TitlesController) handleAddTitleLanguages() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelLanguagesMeta := models.LanguageMeta{}
		if err := c.ShouldBind(&modelLanguagesMeta); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		modelLanguagesMeta.Id = xid.New().String()
		err := basevalidators.GetInstance().GetValidator().Struct(modelLanguagesMeta)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		languageMetaController, _ := u.BaseControllerFactory.GetController(baseconst.LanguageMeta)
		count, err := languageMetaController.Count(languageMetaController.GetDBName(), languageMetaController.GetCollectionName(), map[string]interface{}{"language_id": modelLanguagesMeta.LanguageId, "titles_id": modelLanguagesMeta.TitlesId}, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		if count >= 1 {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, errors.New("language already exsist with this id for this title"), nil))
			return
		}
		languageExists, err := cache.GetInstance().Exists(fmt.Sprintf("%d%s%s", modelLanguagesMeta.LanguageId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguagePostfix))
		if !languageExists {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, fmt.Errorf("one of the language is not found id:%d", modelLanguagesMeta.LanguageId), nil))
			return
		}

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, fmt.Errorf("redis connection failed:%d", modelLanguagesMeta.LanguageId), nil))
			return
		}

		statusExist, err := cache.GetInstance().Exists(fmt.Sprintf("%d%s%s", modelLanguagesMeta.StatusId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().StatusPostfix))
		if !statusExist {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, fmt.Errorf("one of the sstatus is not found id:%d for language:%d", modelLanguagesMeta.StatusId, modelLanguagesMeta.LanguageId), nil))
			return
		}

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, fmt.Errorf("redis connection failed:%d", modelLanguagesMeta.LanguageId), nil))
			return
		}

		titleCount, err := u.Count(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelLanguagesMeta.TitlesId}, false)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		if titleCount <= 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, errors.New("no title exists with this id"), nil))
			return
		}

		languageMetaController.Add(languageMetaController.GetDBName(), languageMetaController.GetCollectionName(), modelLanguagesMeta, false)

		updateQuery := "UPDATE " + string(u.GetCollectionName()) + " SET languages_meta = languages_meta || ARRAY[$1] WHERE id = $2"
		err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), updateQuery, []interface{}{modelLanguagesMeta.Id, modelLanguagesMeta.TitlesId}, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		serviceutils.GetInstance().PublishEvent(modelLanguagesMeta, configmanager.GetInstance().ClassName, "vidmicro.title.language.added")
		go u.updateForMeilisearch(modelLanguagesMeta.TitlesId)
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, nil, nil))
	}
}

func (u *TitlesController) handleDeleteLanguages() gin.HandlerFunc {
	return func(c *gin.Context) {
		modLangMeta := models.EditLanguageMeta{}
		if err := c.ShouldBind(&modLangMeta); err != nil {
			c.AbortWithStatusJSON(http.StatusBadGateway, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := basevalidators.GetInstance().GetValidator().Struct(modLangMeta)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		languageMetaController, _ := u.BaseControllerFactory.GetController(baseconst.LanguageMeta)
		rows, err := languageMetaController.Find(languageMetaController.GetDBName(), languageMetaController.GetCollectionName(), "", map[string]interface{}{"language_id": modLangMeta.LanguageId, "titles_id": modLangMeta.TitlesId}, models.LanguageMeta{}, false, "", false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		defer rows.Close()
		modelLanguagesMeta := models.LanguageMeta{}
		for rows.Next() {
			// Scan the row's values into the User struct.
			err := rows.Scan(&modelLanguagesMeta.Id, &modelLanguagesMeta.TitlesId, &modelLanguagesMeta.LanguageId, &modelLanguagesMeta.StatusId)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}

		}

		if modelLanguagesMeta.Id == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, errors.New("id doesn't exist"), nil))
			return
		}

		updateQuery := "UPDATE " + u.GetCollectionName() + " SET languages_meta = array_remove(languages_meta, $1) WHERE id = $2"
		err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), string(updateQuery), []interface{}{modelLanguagesMeta.Id, modelLanguagesMeta.TitlesId}, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, errors.New("updating title failed"), nil))
			return
		}

		err = languageMetaController.DeleteOne(languageMetaController.GetDBName(), languageMetaController.GetCollectionName(), map[string]interface{}{"id": modelLanguagesMeta.Id}, false, false)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, errors.New("deleting language metadata failed"), nil))
			return
		}

		serviceutils.GetInstance().PublishEvent(modelLanguagesMeta, configmanager.GetInstance().ClassName, "vidmicro.title.language.deleted")
		go u.updateForMeilisearch(modelLanguagesMeta.TitlesId)
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_SUCCESS, nil, nil))
	}
}

func (u *TitlesController) updateForMeilisearch(titlesId int) {
	modelTitles := models.Titles{}
	query := map[string]interface{}{"id": titlesId}

	rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), "", query, &modelTitles, false, " LIMIT 1", false)

	if err == nil {
		defer rows.Close()
	} else {
		return
	}
	tempTitle := models.Titles{}
	for rows.Next() {

		coverUrl := sql.NullString{}

		// Scan the row's values into the User struct.
		err := rows.Scan(&tempTitle.Id, &tempTitle.OriginalTitle, &tempTitle.Year, &coverUrl, pq.Array(&tempTitle.LanguagesMeta))
		if err != nil {
			return
		}
		tempTitle.CoverUrl = coverUrl.String

		controller, _ := u.BaseControllerFactory.GetController(baseconst.LanguageMeta)
		languageController := controller.(*LanguageMetadataController)

		tempTitle.LanguagesDetails, _, err = languageController.GetLanguageDetails(tempTitle.LanguagesMeta)

		if err != nil {
			tempTitle.LanguagesDetails = make([]models.LanguageMetaDetails, 0)
		}
	}

	tempMeiliTitle := models.MeilisearchTitle{
		Id:               modelTitles.Id,
		OriginalTitle:    tempTitle.OriginalTitle,
		CoverUrl:         tempTitle.CoverUrl,
		LanguagesDetails: tempTitle.LanguagesDetails,
		Year:             modelTitles.Year,
	}

	searchengine.GetInstance().ProcessTitleDocuments(tempMeiliTitle)
}

func (u *TitlesController) handleSearchTitles() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelTitles := models.MeilisearchTitle{}
		if c.Query("search") != "" {
			modelTitles.OriginalTitle = c.Query("search")
		}

		filter := "id > 1"

		if c.Query("year") != "" {
			filter += " AND year = " + c.Query("year")
		}

		if c.Query("type_name") != "" {
			filter += " AND type_name = " + c.Query("type_name")
		}

		if c.Query("genre") != "" {
			filter += " AND genres = " + c.Query("genre")
		}

		pageSize := configmanager.GetInstance().PageSize
		if c.Query("limit") != "" {
			pageSizeInt, _ := strconv.ParseInt(c.Query("limit"), 10, 64)
			pageSize = pageSizeInt
		}

		page := int64(1)
		if c.Query("page") != "" {
			pageInt, _ := strconv.ParseInt(c.Query("page"), 10, 64)
			page = pageInt
		}

		pr, err := searchengine.GetInstance().SearchDocuments(modelTitles, pageSize, page, filter)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, nil, pr))
	}
}
func (u TitlesController) RegisterApis() {
	baserouter.GetInstance().GetLoginRouter().PUT("/api/titles", u.handleCreateTitles())
	baserouter.GetInstance().GetLoginRouter().GET("/api/titles", u.handleGetTitles())
	baserouter.GetInstance().GetLoginRouter().POST("/api/titles", u.handleUpdateTitles())
	baserouter.GetInstance().GetLoginRouter().DELETE("/api/titles", u.handleDeleteTitles())
	baserouter.GetInstance().GetLoginRouter().POST("/api/add_title_language", u.handleAddTitleLanguages())
	baserouter.GetInstance().GetLoginRouter().POST("/api/delete_title_language", u.handleDeleteLanguages())
	baserouter.GetInstance().GetLoginRouter().GET("/api/search_titles", u.handleSearchTitles())
}
