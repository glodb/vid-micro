package controllers

import (
	"errors"
	"net/http"

	"com.code.sso/com.code.sso/app/models"
	"com.code.sso/com.code.sso/config"
	"com.code.sso/com.code.sso/database/basefunctions"
	"com.code.sso/com.code.sso/database/basetypes"
	"com.code.sso/com.code.sso/httpHandler/basecontrollers/baseinterfaces"
	"com.code.sso/com.code.sso/httpHandler/baserouter"
	"com.code.sso/com.code.sso/httpHandler/basevalidators"
	"com.code.sso/com.code.sso/httpHandler/responses"
	"com.code.sso/com.code.sso/httpHandler/sessionshandler"
	"com.code.sso/com.code.sso/utils"
)

type SessionController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u SessionController) GetDBName() basetypes.DBName {
	return basetypes.DBName(config.GetInstance().Database.DBName)
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

func (u *SessionController) HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	sessionId, err := utils.GenerateUUID()
	modelSession := models.Session{
		SessionId: sessionId,
	}

	if err != nil {
		responses.GetInstance().WriteJsonResponse(w, r, responses.GENERATING_UUID_FAILED, errors.New("Error in generating UUID"), nil)
		return
	}

	//ignoring err because library said it will always give a response
	session, err := sessionshandler.GetInstance().GetSession().Get(r, sessionId)
	session.Values["sessionId"] = sessionId

	err = session.Save(r, w)
	if err != nil {
		responses.GetInstance().WriteJsonResponse(w, r, responses.CREATE_SESSION_FAILED, err, nil)
		return
	}

	responses.GetInstance().WriteJsonResponse(w, r, responses.CREATE_SESSION_SUCCESS, nil, modelSession)
}

func (u SessionController) RegisterApis() {
	baserouter.GetInstance().GetBaseRouter().HandleFunc("/api/createSession", u.HandleCreateSession).Methods("GET")
}
