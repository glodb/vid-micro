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
	"com.code.vidmicro/com.code.vidmicro/settings/searchengine"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type TitleMetaController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u TitleMetaController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u TitleMetaController) GetCollectionName() basetypes.CollectionName {
	return "title_meta"
}

func (u TitleMetaController) DoIndexing() error {
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.TitleMetaData{})
	return nil
}

func (u *TitleMetaController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *TitleMetaController) GetMetaDataRecordData(titlesId int, genreData []int, titleTypeId int) (models.Titles, models.TitleType, []string, map[int]string, error) {
	controller, _ := u.BaseControllerFactory.GetController(baseconst.Titles)
	titlesController := controller.(*TitlesController)

	title, err := titlesController.GetSingleTitle(titlesId)

	if err != nil {
		return models.Titles{}, models.TitleType{}, []string{}, map[int]string{}, errors.New("didn't find the titles attached")

	}

	titleTypeGenController, _ := u.BaseControllerFactory.GetController(baseconst.TitleType)
	titleTypeController := titleTypeGenController.(*TitleTypeController)

	titleType, err := titleTypeController.GetTitleType(titleTypeId)

	if err != nil {
		return models.Titles{}, models.TitleType{}, []string{}, map[int]string{}, errors.New("type id is not valid")
	}

	genreTypeGenController, _ := u.BaseControllerFactory.GetController(baseconst.Genres)
	genreTypeController := genreTypeGenController.(*GenresController)

	genres := make([]string, 0)
	genreObject := make(map[int]string)
	for _, genre := range genreData {
		genreData, err := genreTypeController.GetGenre(genre)
		if err != nil {
			return models.Titles{}, models.TitleType{}, []string{}, map[int]string{}, fmt.Errorf("one of genre id is not valid, %d", genre)
		}
		genres = append(genres, genreData.Name)
		genreObject[genre] = genreData.Name
	}
	return title, titleType, genres, genreObject, nil
}

func (u *TitleMetaController) handleCreateTitleMeta() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelTitleMeta := models.TitleMetaData{}
		if err := c.ShouldBind(&modelTitleMeta); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/put", modelTitleMeta)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		title, titleType, genres, genreObject, err := u.GetMetaDataRecordData(modelTitleMeta.TitleId, modelTitleMeta.Genres, modelTitleMeta.TypeId)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		modelTitleMeta.Title = title.OriginalTitle
		modelTitleMeta.Year = title.Year

		meilisearchRecord := models.MeilisearchTitle{
			Id:               title.Id,
			OriginalTitle:    title.OriginalTitle,
			Year:             title.Year,
			CoverUrl:         title.CoverUrl,
			LanguagesDetails: title.LanguagesDetails,
			AlternativeName:  modelTitleMeta.AlternativeName,
			Sequence:         modelTitleMeta.Sequence,
			TypeId:           modelTitleMeta.TypeId,
			TypeName:         titleType.Name,
			Genres:           genres,
			GenresObject:     genreObject,
		}

		id, err := u.Add(u.GetDBName(), u.GetCollectionName(), modelTitleMeta, false)
		modelTitleMeta.Id = int(id)

		searchengine.GetInstance().ProcessTitleDocuments(meilisearchRecord)
		cache.GetInstance().SetEx(fmt.Sprintf("%d%s%s", modelTitleMeta.TitleId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().TitlesMetaPostfix), meilisearchRecord.EncodeRedisData(), configmanager.GetInstance().TitleExpiryTime)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_FAILED, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_SUCCESS, err, modelTitleMeta))
	}
}

func (u *TitleMetaController) handleGetTitleMeta() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelTitleMeta := models.TitleMetaData{}
		id, _ := strconv.ParseInt(c.Query("title_id"), 10, 64)
		modelTitleMeta.TitleId = int(id)

		err := u.Validate(c.GetString("apiPath")+"/get", modelTitleMeta)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		meilisearchRecord := models.MeilisearchTitle{}
		data, err := cache.GetInstance().Get(fmt.Sprintf("%d%s%s", modelTitleMeta.TitleId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().TitlesMetaPostfix))

		if err == nil && len(data) > 0 {
			meilisearchRecord.DecodeRedisData(data)
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, err, meilisearchRecord))
			return
		} else {
			query := map[string]interface{}{"title_id": modelTitleMeta.TitleId}

			rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), "", query, &modelTitleMeta, false, " LIMIT 1", false)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
				return
			}
			defer rows.Close()

			tempTitleMeta := models.TitleMetaData{}

			for rows.Next() {
				var arrayData pq.Int32Array
				err := rows.Scan(&tempTitleMeta.Id, &tempTitleMeta.TitleId, &tempTitleMeta.Title, &tempTitleMeta.AlternativeName, &tempTitleMeta.Sequence, &tempTitleMeta.TypeId, &tempTitleMeta.Year, &tempTitleMeta.Score, &arrayData)
				if err != nil {
					c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
					return
				}
				tempTitleMeta.Genres = make([]int, len(arrayData))
				for i, v := range arrayData {
					tempTitleMeta.Genres[i] = int(v)
				}
			}

			title, titleType, genres, genreObject, err := u.GetMetaDataRecordData(modelTitleMeta.TitleId, tempTitleMeta.Genres, tempTitleMeta.TypeId)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
				return
			}

			meilisearchRecord := models.MeilisearchTitle{
				Id:               title.Id,
				OriginalTitle:    title.OriginalTitle,
				Year:             title.Year,
				CoverUrl:         title.CoverUrl,
				LanguagesDetails: title.LanguagesDetails,
				AlternativeName:  tempTitleMeta.AlternativeName,
				Sequence:         tempTitleMeta.Sequence,
				TypeId:           tempTitleMeta.TypeId,
				TypeName:         titleType.Name,
				Genres:           genres,
				GenresObject:     genreObject,
			}
			cache.GetInstance().SetEx(fmt.Sprintf("%d%s%s", meilisearchRecord.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().TitlesMetaPostfix), meilisearchRecord.EncodeRedisData(), configmanager.GetInstance().TitleExpiryTime)
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, err, meilisearchRecord))
		}
	}
}
func (u *TitleMetaController) ChangeTitleName(title_id int, name string) {
	cache.GetInstance().Del(fmt.Sprintf("%d%s%s", title_id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().TitlesMetaPostfix))
	u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET title = $1 WHERE title_id = $2", []interface{}{name, title_id}, false)
}

func (u *TitleMetaController) handleUpdateTitleMeta() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelTitleMeta := models.TitleMetaData{}
		meiliSearchUpdate := models.MeilisearchTitle{}
		if err := c.ShouldBind(&modelTitleMeta); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/post", modelTitleMeta)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		meiliSearchUpdate.Id = modelTitleMeta.TitleId
		data := make([]interface{}, 0)
		setPart := " SET "

		if modelTitleMeta.AlternativeName != "" {
			lengthString := strconv.FormatInt(int64(len(data)+1), 10)
			if len(data) > 0 {
				setPart += ","
			}
			setPart += "alternative_name = $" + lengthString
			data = append(data, modelTitleMeta.AlternativeName)
			meiliSearchUpdate.AlternativeName = modelTitleMeta.AlternativeName
		}

		if modelTitleMeta.Sequence > 0 {
			lengthString := strconv.FormatInt(int64(len(data)+1), 10)
			if len(data) > 0 {
				setPart += ","
			}
			setPart += "sequence = $" + lengthString
			data = append(data, modelTitleMeta.Sequence)
			meiliSearchUpdate.Sequence = modelTitleMeta.Sequence
		}

		if modelTitleMeta.TypeId > 0 {
			lengthString := strconv.FormatInt(int64(len(data)+1), 10)
			if len(data) > 0 {
				setPart += ","
			}
			setPart += "type_id = $" + lengthString
			data = append(data, modelTitleMeta.Sequence)
			meiliSearchUpdate.TypeId = modelTitleMeta.TypeId

			titleTypeGenController, _ := u.BaseControllerFactory.GetController(baseconst.TitleType)
			titleTypeController := titleTypeGenController.(*TitleTypeController)

			titleType, err := titleTypeController.GetTitleType(modelTitleMeta.TypeId)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("type id not in the system"), nil))
				return
			}
			meiliSearchUpdate.TypeId = titleType.Id
			meiliSearchUpdate.TypeName = titleType.Name
		}

		if modelTitleMeta.Score > 0 {
			lengthString := strconv.FormatInt(int64(len(data)+1), 10)
			if len(data) > 0 {
				setPart += ","
			}
			setPart += "score = $" + lengthString
			data = append(data, modelTitleMeta.Score)
			meiliSearchUpdate.Score = modelTitleMeta.Score
		}

		if len(modelTitleMeta.Genres) > 0 {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("to update genres use, add_title_genres, delete_title_genres"), nil))
			return
		}

		if len(data) > 0 {
			lengthString := strconv.FormatInt(int64(len(data)+1), 10)
			setPart += " WHERE id =$" + lengthString
			data = append(data, modelTitleMeta.TitleId)

			err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+setPart, data, false)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_FAILED, err, nil))
				return
			}

			searchengine.GetInstance().ProcessTitleDocuments(meiliSearchUpdate)
			cache.GetInstance().Del(fmt.Sprintf("%d%s%s", modelTitleMeta.TitleId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().TitlesMetaPostfix))
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_SUCCESS, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.NOTHIN_TO_UPDATE, err, nil))
	}
}

func (u *TitleMetaController) handleDeleteTitleMeta() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelTitleMeta := models.TitleMetaData{}
		if err := c.ShouldBind(&modelTitleMeta); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/post", modelTitleMeta)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err = u.DeleteOne(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"title_id": modelTitleMeta.TitleId}, false, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_FAILED, err, nil))
			return
		}
		searchengine.GetInstance().DeleteDocumentsMeta(models.MeilisearchTitle{Id: modelTitleMeta.TitleId})
		cache.GetInstance().Del(fmt.Sprintf("%d%s%s", modelTitleMeta.TitleId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().TitlesMetaPostfix))
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.DELETING_SUCCESS, err, nil))
	}
}

func (u *TitleMetaController) handleAddTitleGenre() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelGenreData := models.GenresData{}
		if err := c.ShouldBind(&modelGenreData); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		if modelGenreData.GenreId <= 0 {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("genre id is required"), nil))
			return
		}

		if modelGenreData.TitleId <= 0 {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("title id is required"), nil))
			return
		}

		genreTypeGenController, _ := u.BaseControllerFactory.GetController(baseconst.Genres)
		genreTypeController := genreTypeGenController.(*GenresController)
		_, err := genreTypeController.GetGenre(modelGenreData.GenreId)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		updateQuery := "UPDATE " + string(u.GetCollectionName()) + " SET genres = array_append(genres, $1) WHERE title_id = $2 AND NOT $1 = ANY(genres)"
		err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), updateQuery, []interface{}{modelGenreData.GenreId, modelGenreData.TitleId}, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}
		cache.GetInstance().Del(fmt.Sprintf("%d%s%s", modelGenreData.TitleId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().TitlesMetaPostfix))
		go u.updateForMeilisearch(modelGenreData.TitleId)
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_SUCCESS, nil, nil))
	}
}

func (u *TitleMetaController) handleDeleteTitleGenre() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelGenreData := models.GenresData{}
		if err := c.ShouldBind(&modelGenreData); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		if modelGenreData.GenreId <= 0 {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("genre id is required"), nil))
			return
		}

		if modelGenreData.TitleId <= 0 {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("title id is required"), nil))
			return
		}

		updateQuery := "UPDATE " + u.GetCollectionName() + " SET genres = array_remove(genres, $1) WHERE title_id = $2  AND $1 = ANY(genres)"
		err := u.UpdateOne(u.GetDBName(), u.GetCollectionName(), string(updateQuery), []interface{}{modelGenreData.GenreId, modelGenreData.TitleId}, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		cache.GetInstance().Del(fmt.Sprintf("%d%s%s", modelGenreData.TitleId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().TitlesMetaPostfix))
		go u.updateForMeilisearch(modelGenreData.TitleId)
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATE_SUCCESS, nil, nil))
	}
}

func (u *TitleMetaController) updateForMeilisearch(titlesId int) {
	modelTitlesMetaData := models.TitleMetaData{}
	query := map[string]interface{}{"title_id": titlesId}

	rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), "genres", query, &modelTitlesMetaData, false, " LIMIT 1", false)

	if err == nil {
		defer rows.Close()
	} else {
		return
	}

	for rows.Next() {
		var arrayData pq.Int32Array
		err := rows.Scan(&arrayData)
		if err != nil {
			return
		}
		modelTitlesMetaData.Genres = make([]int, len(arrayData))
		for i, v := range arrayData {
			modelTitlesMetaData.Genres[i] = int(v)
		}
	}

	genreTypeGenController, _ := u.BaseControllerFactory.GetController(baseconst.Genres)
	genreTypeController := genreTypeGenController.(*GenresController)

	genres := make([]string, 0)
	genreObject := make(map[int]string)
	for _, genre := range modelTitlesMetaData.Genres {
		genreData, err := genreTypeController.GetGenre(genre)
		if err != nil {
			return
		}
		genres = append(genres, genreData.Name)
		genreObject[genre] = genreData.Name
	}

	meilisearchRecord := models.MeilisearchTitle{
		Id:           titlesId,
		Genres:       genres,
		GenresObject: genreObject,
	}
	searchengine.GetInstance().ProcessTitleDocuments(meilisearchRecord)
}

func (u TitleMetaController) RegisterApis() {
	baserouter.GetInstance().GetLoginRouter().PUT("/api/title_meta", u.handleCreateTitleMeta())
	baserouter.GetInstance().GetLoginRouter().GET("/api/title_meta", u.handleGetTitleMeta())
	baserouter.GetInstance().GetLoginRouter().POST("/api/title_meta", u.handleUpdateTitleMeta())
	baserouter.GetInstance().GetLoginRouter().DELETE("/api/title_meta", u.handleDeleteTitleMeta())
	baserouter.GetInstance().GetLoginRouter().POST("/api/add_title_genre", u.handleAddTitleGenre())
	baserouter.GetInstance().GetLoginRouter().POST("/api/delete_title_genre", u.handleDeleteTitleGenre())
}