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

			languageMetaDetails := models.LanguageMetaDetails{LanguageId: titlesLanguage.LanguageId, StatusId: titlesLanguage.StatusId}
			languageData, err := cache.GetInstance().Get(fmt.Sprintf("%d%s%s", languageMetadata.LanguageId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguagePostfix))

			if err != nil && len(languageData) <= 0 {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, fmt.Errorf("one of the language is not found id:%d", languageMetadata.LanguageId), nil))
				return
			}

			lang := models.Language{}
			lang.DecodeRedisData(languageData)
			languageMetaDetails.LanguageCode = lang.Code
			languageMetaDetails.LanguageName = lang.Name

			statusData, err := cache.GetInstance().Get(fmt.Sprintf("%d%s%s", languageMetadata.StatusId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().StatusPostfix))

			if err != nil && len(statusData) <= 0 {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, fmt.Errorf("one of the sstatus is not found id:%d for language:%d", languageMetadata.StatusId, languageMetadata.LanguageId), nil))
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
		setPart := " SET "
		data := make([]interface{}, 0)
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

		file, err := c.FormFile("image")
		updateTitle := false

		if err == nil && file != nil {
			url, err := s3uploader.GetInstance().UploadToSCW(file)
			if err == nil {
				modelTitles.CoverUrl = url
				setPart += "thumbnail = $1"
				data = append(data, url)

			} else {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPLOADING_AVATAR_FAILED, err, nil))
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

		if modelTitles.Languages != "" {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("please use addLanguage and removeLanguage api for languages"), nil))
			return
		}

		if len(data) > 0 {
			lengthString := strconv.FormatInt(int64(len(data)+1), 10)
			setPart += " WHERE id =$" + lengthString
			data = append(data, modelTitles.Id)

			err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+setPart, data, false)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.FAILED_UPDATING_USER, err, nil))
				return
			}
			pattern := "*" + configmanager.GetInstance().ClassName + configmanager.GetInstance().RedisSeprator + configmanager.GetInstance().TitlesPostfix

			keys := cache.GetInstance().GetKeys(pattern)
			cache.GetInstance().DelMany(keys)

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
				serviceutils.GetInstance().PublishEvent(titleSummary, configmanager.GetInstance().ClassName, "vidmicro.title.updated")
			}

			tempMeiliTitle := models.MeilisearchTitle{
				Id:               modelTitles.Id,
				OriginalTitle:    modelTitles.OriginalTitle,
				CoverUrl:         modelTitles.CoverUrl,
				LanguagesDetails: modelTitles.LanguagesDetails,
				Year:             modelTitles.Year,
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
		// TODO:Delete meta data
		modelTitles := models.Titles{}
		if err := c.ShouldBind(&modelTitles); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/delete", modelTitles)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		modelMeiliTitles := models.MeilisearchTitle{}
		err = searchengine.GetInstance().DeleteDocuments(modelMeiliTitles)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_FAILED, err, nil))
			return
		}

		err = u.DeleteOne(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelTitles.Id}, false, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_FAILED, err, nil))
			return
		}

		languageMetaController, _ := u.BaseControllerFactory.GetController(baseconst.LanguageMeta)
		languageMetaController.DeleteOne(languageMetaController.GetDBName(), languageMetaController.GetCollectionName(), map[string]interface{}{"titles_id": modelTitles.Id}, false, false)

		pattern := "*" + configmanager.GetInstance().ClassName + configmanager.GetInstance().RedisSeprator + configmanager.GetInstance().TitlesPostfix

		keys := cache.GetInstance().GetKeys(pattern)
		cache.GetInstance().DelMany(keys)

		key := "1" + configmanager.GetInstance().RedisSeprator +
			fmt.Sprintf("%d", modelTitles.Id) +
			configmanager.GetInstance().RedisSeprator +
			configmanager.GetInstance().ClassName +
			configmanager.GetInstance().RedisSeprator +
			configmanager.GetInstance().SingleTitlePostfix +
			configmanager.GetInstance().TitlesPostfix

		cache.GetInstance().Del(key)

		titleSummary := models.TitlesSummary{Id: modelTitles.Id, OriginalTitle: modelTitles.OriginalTitle}

		serviceutils.GetInstance().PublishEvent(titleSummary, configmanager.GetInstance().ClassName, "vidmicro.title.deleted")

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_SUCCESS, err, nil))
	}
}

func (u *TitlesController) handleAddTitleLanguages() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelLanguagesMeta := models.LanguageMeta{}
		if err := c.ShouldBind(&modelLanguagesMeta); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		modelLanguagesMeta.Id = xid.New().String()
		if modelLanguagesMeta.LanguageId <= 0 {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("language id is required"), nil))
			return
		}

		if modelLanguagesMeta.StatusId <= 0 {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("status id is required"), nil))
			return
		}

		if modelLanguagesMeta.TitlesId <= 0 {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("title id is required"), nil))
			return
		}

		languageMetaController, _ := u.BaseControllerFactory.GetController(baseconst.LanguageMeta)
		count, err := languageMetaController.Count(languageMetaController.GetDBName(), languageMetaController.GetCollectionName(), map[string]interface{}{"language_id": modelLanguagesMeta.LanguageId, "titles_id": modelLanguagesMeta.TitlesId})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		if count >= 1 {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("language already exsist with this id for this title"), nil))
			return
		}

		if !cache.GetInstance().Exists(fmt.Sprintf("%d%s%s", modelLanguagesMeta.LanguageId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguagePostfix)) {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, fmt.Errorf("one of the language is not found id:%d", modelLanguagesMeta.LanguageId), nil))
			return
		}

		if !cache.GetInstance().Exists(fmt.Sprintf("%d%s%s", modelLanguagesMeta.StatusId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().StatusPostfix)) {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, fmt.Errorf("one of the sstatus is not found id:%d for language:%d", modelLanguagesMeta.StatusId, modelLanguagesMeta.LanguageId), nil))
			return
		}

		titleCount, err := u.Count(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"id": modelLanguagesMeta.TitlesId})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}
		if titleCount <= 0 {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("no title exists with this id"), nil))
			return
		}

		languageMetaController.Add(languageMetaController.GetDBName(), languageMetaController.GetCollectionName(), modelLanguagesMeta, false)

		updateQuery := "UPDATE " + string(u.GetCollectionName()) + " SET languages_meta = languages_meta || ARRAY[$1] WHERE id = $2"
		err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), updateQuery, []interface{}{modelLanguagesMeta.Id, modelLanguagesMeta.TitlesId}, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		serviceutils.GetInstance().PublishEvent(modelLanguagesMeta, configmanager.GetInstance().ClassName, "vidmicro.title.language.added")
		go u.updateForMeilisearch(modelLanguagesMeta.TitlesId)
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, nil, nil))
	}
}

func (u *TitlesController) handleDeleteLanguages() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelLanguagesMeta := models.LanguageMeta{}
		if err := c.ShouldBind(&modelLanguagesMeta); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		if modelLanguagesMeta.LanguageId <= 0 {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("language id is required"), nil))
			return
		}

		if modelLanguagesMeta.TitlesId <= 0 {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("title id is required"), nil))
			return
		}

		languageMetaController, _ := u.BaseControllerFactory.GetController(baseconst.LanguageMeta)
		rows, err := languageMetaController.Find(languageMetaController.GetDBName(), languageMetaController.GetCollectionName(), "", map[string]interface{}{"language_id": modelLanguagesMeta.LanguageId, "titles_id": modelLanguagesMeta.TitlesId}, models.LanguageMeta{}, false, "", false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		defer rows.Close()

		for rows.Next() {
			// Scan the row's values into the User struct.
			err := rows.Scan(&modelLanguagesMeta.Id, &modelLanguagesMeta.TitlesId, &modelLanguagesMeta.LanguageId, &modelLanguagesMeta.StatusId)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
				return
			}

		}

		if modelLanguagesMeta.Id == "" {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("id doesn't exist"), nil))
			return
		}

		updateQuery := "UPDATE " + u.GetCollectionName() + " SET languages_meta = array_remove(languages_meta, $1) WHERE id = $2"
		err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), string(updateQuery), []interface{}{modelLanguagesMeta.Id, modelLanguagesMeta.TitlesId}, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("updating title failed"), nil))
			return
		}

		err = languageMetaController.DeleteOne(languageMetaController.GetDBName(), languageMetaController.GetCollectionName(), map[string]interface{}{"id": modelLanguagesMeta.Id}, false, false)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("deleting language metadata failed"), nil))
			return
		}

		serviceutils.GetInstance().PublishEvent(modelLanguagesMeta, configmanager.GetInstance().ClassName, "vidmicro.title.language.deleted")
		go u.updateForMeilisearch(modelLanguagesMeta.TitlesId)
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_SUCCESS, nil, nil))
	}
}

func (u *TitlesController) updateForMeilisearch(titlesId int) {
	modelTitles := models.Titles{}
	query := map[string]interface{}{"id": modelTitles.Id}

	rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), "", query, &modelTitles, false, "", false)

	if err == nil {
		defer rows.Close()
	} else {
		return
	}
	tempTitle := models.Titles{}
	for rows.Next() {
		// Scan the row's values into the User struct.
		err := rows.Scan(&tempTitle.Id, &tempTitle.OriginalTitle, &tempTitle.Year, &tempTitle.CoverUrl, pq.Array(&tempTitle.LanguagesMeta))
		if err != nil {
			return
		}

		controller, _ := u.BaseControllerFactory.GetController(baseconst.LanguageMeta)
		languageController := controller.(*LanguageMetadataController)

		tempTitle.LanguagesDetails, err = languageController.GetLanguageDetails(tempTitle.LanguagesMeta)

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
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
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
