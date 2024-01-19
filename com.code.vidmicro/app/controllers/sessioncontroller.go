package controllers

import (
	"log"
	"net/http"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseinterfaces"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/baserouter"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basevalidators"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
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
	return nil
}

func (u *SessionController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *SessionController) handleCreateSession() gin.HandlerFunc {
	return func(c *gin.Context) {

		sessionId, err := utils.GenerateUUID()
		modelSession := models.Session{
			SessionId: sessionId,
		}

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GENERATING_UUID_FAILED, err, nil))
			return
		}

		//ignoring err because library said it will always give a response
		// session, err := sessionshandler.GetInstance().GetSession().Get(r, sessionId)
		// session.Values["sessionId"] = sessionId

		//TODO: handle creating and saving session
		log.Println(modelSession)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.CREATE_SESSION_FAILED, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.CREATE_SESSION_SUCCESS, err, nil))
	}
}

func (u SessionController) RegisterApis() {
	baserouter.GetInstance().GetBaseRouter(configmanager.GetInstance().SessionKey).POST("/api/createSession", u.handleCreateSession())
}
