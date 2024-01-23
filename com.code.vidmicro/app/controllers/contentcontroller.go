package controllers

import (
	"errors"
	"fmt"
	"net/http"

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
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.ContentType{})
	return nil
}

func (u *ContentController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *ContentController) validateTitleLanguage(content models.Contents) bool {

	if cache.GetInstance().Exists(fmt.Sprintf("%d%s%s", content.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTitleLanguagesPostfix)) {
		return cache.GetInstance().SetISMember(fmt.Sprintf("%d%s%s", content.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTitleLanguagesPostfix), fmt.Sprintf("%d", content.Id))
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

		for titleRows.Next() {
			// Scan the row's values into the User struct.
			err := titleRows.Scan(pq.Array(&tempTitle.Languages))
			if err != nil {
				return false
			}
		}

		contains := false

		contents := make([]interface{}, 0)

		contents = append(contents, fmt.Sprintf("%d%s%s", tempTitle.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTitleLanguagesPostfix))

		for language := range tempTitle.Languages {
			contents = append(contents, language)

			if language == content.Id {
				contains = true
			}
		}
		cache.GetInstance().SAdd(contents)
		cache.GetInstance().Expire(fmt.Sprintf("%d%s%s", tempTitle.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTitleLanguagesPostfix), configmanager.GetInstance().TitlesLanguageExpirationTime)
		return contains
	}

}

func (u *ContentController) handleCreateContent() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelContent := models.Contents{}
		if err := c.ShouldBind(&modelContent); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath")+"/put", modelContent)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		if !u.validateTitleLanguage(modelContent) {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.LANGUAGE_NOT_ADDED_IN_TITLE, errors.New("validating language failed"), nil))
			return
		}

		if !cache.GetInstance().Exists(fmt.Sprintf("%d%s%s", modelContent.TypeId, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentTypePostfix)) {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, errors.New("status id is not correct"), nil))
			return
		}

		id, err := u.Add(u.GetDBName(), u.GetCollectionName(), modelContent, true)
		modelContent.Id = int(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		file, err := c.FormFile("image")

		if err == nil && file != nil {
			url, err := s3uploader.GetInstance().UploadToSCW(file)
			if err == nil {
				modelContent.Thumbnail = url
			} else {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPLOADING_AVATAR_FAILED, err, nil))
				return
			}
		}
		// TODO: Invalidate all the cache for this title
		// TODO: Invalidate all the filters from cache for this title
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PUTTING_SUCCESS, err, nil))
	}
}

func (u *ContentController) handleGetContent() gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

func (u *ContentController) handleUpdateContent() gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

func (u *ContentController) handleDeleteContent() gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

func (u ContentController) RegisterApis() {
	baserouter.GetInstance().GetLoginRouter().PUT("/api/content", u.handleCreateContent())
	baserouter.GetInstance().GetLoginRouter().GET("/api/content", u.handleGetContent())
	baserouter.GetInstance().GetLoginRouter().POST("/api/content", u.handleUpdateContent())
	baserouter.GetInstance().GetLoginRouter().DELETE("/api/content", u.handleDeleteContent())
}
