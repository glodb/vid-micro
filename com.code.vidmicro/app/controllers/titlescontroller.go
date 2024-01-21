package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseconst"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseinterfaces"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/baserouter"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basevalidators"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/s3uploader"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

type TitlesController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
	entropy io.Reader
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
		// --Add language to language meta and save the id in title
		// Send the available languages and title to content service
		// Save the data in meilisearch
		// Clear paginated data from cache
		modelTitles := models.Titles{}
		if err := c.ShouldBind(&modelTitles); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		titlesLanguages := make([]models.TitlesLanguage, 0)

		err := json.Unmarshal([]byte(modelTitles.Languages), &titlesLanguages)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		if len(titlesLanguages) <= 0 {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("atleast one language is required"), nil))
			return
		}

		languageController, _ := u.BaseControllerFactory.GetController(baseconst.Language)
		statusController, _ := u.BaseControllerFactory.GetController(baseconst.Status)

		uniqueLanguages := make(map[string]bool)
		uniqueStatuses := make(map[int]bool)

		languageInQuery := " Where id IN ("
		statusInQuery := " Where id IN ("

		languagesMetadata := make([]interface{}, 0)

		for _, titlesLanguage := range titlesLanguages {

			languageMetadata := models.LanguageMeta{LanguageId: titlesLanguage.LanguageId, StatusId: titlesLanguage.StatusId}
			u.entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
			ulid := ulid.MustNew(ulid.Timestamp(time.Now()), u.entropy)
			languageMetadata.Id = ulid.String()
			modelTitles.LanguagesMeta = append(modelTitles.LanguagesMeta, languageMetadata.Id)

			if ok := uniqueLanguages[titlesLanguage.LanguageId]; !ok {
				if len(uniqueLanguages) != 0 {
					languageInQuery += ","
				}
				languageInQuery += "'" + titlesLanguage.LanguageId + "'"
				uniqueLanguages[titlesLanguage.LanguageId] = true
			}

			if ok := uniqueStatuses[titlesLanguage.StatusId]; !ok {
				if len(uniqueStatuses) != 0 {
					statusInQuery += ","
				}
				statusInQuery += strconv.FormatInt(int64(titlesLanguage.StatusId), 10)
				uniqueStatuses[titlesLanguage.StatusId] = true
			}
			languagesMetadata = append(languagesMetadata, languageMetadata)
		}

		languageInQuery += ")"
		statusInQuery += ")"

		languageRows, err := languageController.Find(languageController.GetDBName(), languageController.GetCollectionName(), "", map[string]interface{}{}, models.Language{}, false, languageInQuery, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
			return
		}
		defer languageRows.Close()

		languages := make([]models.Language, 0)

		// Iterate over the rows.
		for languageRows.Next() {
			// Create a User struct to scan values into.

			tempLanguage := models.Language{}

			// Scan the row's values into the User struct.
			err := languageRows.Scan(&tempLanguage.Id, &tempLanguage.Name, &tempLanguage.Code)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
				return
			}

			languages = append(languages, tempLanguage)
		}

		statusRows, err := statusController.Find(statusController.GetDBName(), statusController.GetCollectionName(), "", map[string]interface{}{}, models.Language{}, false, statusInQuery, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
			return
		}
		defer languageRows.Close()

		statuses := make([]models.Status, 0)

		// Iterate over the rows.
		for statusRows.Next() {
			// Create a User struct to scan values into.

			tempStatus := models.Status{}

			// Scan the row's values into the User struct.
			err := statusRows.Scan(&tempStatus.Id, &tempStatus.Name)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
				return
			}

			statuses = append(statuses, tempStatus)
		}

		defer statusRows.Close()

		if (len(statuses) != len(uniqueStatuses)) || len(languages) != len(uniqueLanguages) {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, errors.New("status id or language id not found"), nil))
			return
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

		id, err := u.Add(u.GetDBName(), u.GetCollectionName(), modelTitles, false)
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

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_SUCCESS, err, modelTitles))
	}
}

func (u *TitlesController) handleGetTitles() gin.HandlerFunc {
	return func(c *gin.Context) {
		// modelTitles := models.Titles{}
		// id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
		// modelTitles.Id = int(id)

		// err := u.Validate(c.GetString("apiPath")+"/get", modelTitles)
		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
		// 	return
		// }

		// rows, err := u.FindOne(u.GetDBName(), u.GetCollectionName(), "", map[string]interface{}{"id": modelTitles.Id}, &modelTitles, false, " Limit 1", false)

		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
		// 	return
		// }
		// defer rows.Close()

		// // Iterate over the rows.
		// for rows.Next() {
		// 	// Create a User struct to scan values into.

		// 	// Scan the row's values into the User struct.
		// 	err := rows.Scan(&modelTitles.Id, &modelTitles.Name, &modelTitles.Slug)
		// 	if err != nil {
		// 		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_FAILED, err, nil))
		// 		return
		// 	}
		// }
		// c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GETTING_SUCCESS, err, modelTitles))
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
