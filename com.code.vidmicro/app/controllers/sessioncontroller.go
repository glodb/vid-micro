package controllers

import (
	"errors"
	"net/http"
	"time"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseinterfaces"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/baserouter"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basevalidators"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"com.code.vidmicro/com.code.vidmicro/settings/cache"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/utils"
	"github.com/gin-gonic/gin"
)

type SessionController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u SessionController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u SessionController) GetCollectionName() basetypes.CollectionName {
	return "sessions"
}

func (u SessionController) DoIndexing() error {
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.Session{})
	return nil
}

func (u *SessionController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *SessionController) handleCreateSession() gin.HandlerFunc {
	return func(c *gin.Context) {

		savedModel, err := u.retrieveCookie(c)

		if err == nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.CREATE_SESSION_SUCCESS, err, savedModel))
			return
		}

		sessionId, err := utils.GenerateUUID()
		modelSession := models.Session{
			SessionId: sessionId,
		}

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		now := time.Now().Unix()
		modelSession.CreatedAt = now
		modelSession.ExpiringAt = now + configmanager.GetInstance().SessionExpirySeconds
		err = cache.GetInstance().Set(sessionId, modelSession.EncodeRedisData())

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		err = cache.GetInstance().Expire(sessionId, int(configmanager.GetInstance().SessionExpirySeconds))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		// Set a cookie
		c.SetCookie(configmanager.GetInstance().CookieName, sessionId, int(configmanager.GetInstance().SessionExpirySeconds), configmanager.GetInstance().CookiePath, configmanager.GetInstance().CookieDomain, false, true)

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.CREATE_SESSION_SUCCESS, err, modelSession))
	}
}

func (u *SessionController) retrieveCookie(c *gin.Context) (models.Session, error) {
	cookieValue, err := c.Cookie(configmanager.GetInstance().CookieName)

	if err != nil {
		return models.Session{}, err
	}
	var session models.Session
	data, err := cache.GetInstance().Get(cookieValue)
	if err != nil || len(data) == 0 {
		return models.Session{}, errors.New("cookie not valid")
	} else {
		session.DecodeRedisData(data)
	}

	return models.Session{SessionId: session.SessionId, CreatedAt: session.CreatedAt, ExpiringAt: session.ExpiringAt}, nil
}

func (u SessionController) RegisterApis() {
	baserouter.GetInstance().GetBaseRouter(configmanager.GetInstance().SessionKey).GET("/api/getSession", u.handleCreateSession())
}
