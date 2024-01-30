package controllers

import (
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
	"github.com/gin-gonic/gin"
)

type TitlesSummaryController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u TitlesSummaryController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u TitlesSummaryController) GetCollectionName() basetypes.CollectionName {
	return "titles"
}

func (u TitlesSummaryController) DoIndexing() error {
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.TitlesSummary{})
	return nil
}

func (u *TitlesSummaryController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *TitlesSummaryController) UpdateTitleLanguage(languageMeta models.LanguageMeta) {

	updateQuery := "UPDATE " + string(u.GetCollectionName()) + " SET languages_meta = array_append(languages_meta, $1) WHERE id = $2 AND $1 = ANY(languages_meta)"
	err := u.UpdateOne(u.GetDBName(), u.GetCollectionName(), updateQuery, []interface{}{languageMeta.LanguageId, languageMeta.TitlesId}, false)
	if err == nil {
		cache.GetInstance().SAdd([]interface{}{fmt.Sprintf("%d%s%s", languageMeta.TitlesId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTitleLanguagesPostfix), fmt.Sprintf("%d", languageMeta.LanguageId)})
	}
}

func (u *TitlesSummaryController) DeleteTitleLanguage(languageMeta models.LanguageMeta) {

	updateQuery := "UPDATE " + u.GetCollectionName() + " SET languages_meta = array_remove(languages_meta, $1) WHERE id = $2  AND $1 = ANY(languages_meta)"
	err := u.UpdateOne(u.GetDBName(), u.GetCollectionName(), string(updateQuery), []interface{}{languageMeta.LanguageId, languageMeta.TitlesId}, false)

	if err == nil {
		cache.GetInstance().SRem([]interface{}{fmt.Sprintf("%d%s%s", languageMeta.TitlesId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTitleLanguagesPostfix), fmt.Sprintf("%d", languageMeta.LanguageId)})
	}

	contentController, _ := u.BaseControllerFactory.GetController(baseconst.Content)
	contentController.DeleteOne(contentController.GetDBName(), contentController.GetCollectionName(), map[string]interface{}{"associated_title": languageMeta.TitlesId, "language_id": languageMeta.LanguageId}, false, false)

	keys := cache.GetInstance().GetKeys(fmt.Sprintf("*%d%s%s", languageMeta.TitlesId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentPostFix))
	if len(keys) > 0 {
		cache.GetInstance().DelMany(keys)
	}
}

func (u *TitlesSummaryController) handleGetTitles() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelTitles := models.TitlesSummary{}
		idString := c.Query("id")
		id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
		page := int64(1)
		pageString := c.Query("page")
		modelTitles.Id = int(id)

		// TODO:
		// err := u.Validate(c.GetString("apiPath")+"/get", modelTitles)
		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
		// 	return
		// }

		query := map[string]interface{}{"id": modelTitles.Id}

		if pageString != "" {
			pageInt, _ := strconv.ParseInt(c.Query("page"), 10, 64)
			page = pageInt
		} else {
			pageString = "1"
		}

		key := pageString + configmanager.GetInstance().RedisSeprator + configmanager.GetInstance().ClassName + configmanager.GetInstance().RedisSeprator + configmanager.GetInstance().TitlesContentPostfix

		if modelTitles.Id <= 0 {
			query = map[string]interface{}{}
		} else {
			key = pageString + configmanager.GetInstance().RedisSeprator +
				idString +
				configmanager.GetInstance().RedisSeprator +
				configmanager.GetInstance().ClassName +
				configmanager.GetInstance().RedisSeprator +
				configmanager.GetInstance().SingleTitlePostfix +
				configmanager.GetInstance().TitlesContentPostfix
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

		rows, count, err := u.Paginate(u.GetDBName(), u.GetCollectionName(), "id,original_title", query, &modelTitles, false, "", false, int(pageSize), int(page))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		defer rows.Close()

		titles := make([]models.TitlesSummary, 0)
		// Iterate over the rows.
		for rows.Next() {
			// Create a User struct to scan values into.
			tempTitle := models.TitlesSummary{}

			// Scan the row's values into the User struct.
			err := rows.Scan(&tempTitle.Id, &tempTitle.OriginalTitle)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
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

func (u TitlesSummaryController) RegisterApis() {
	baserouter.GetInstance().GetLoginRouter().GET("/api/titles_summary", u.handleGetTitles())
}
