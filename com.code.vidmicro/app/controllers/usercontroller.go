package controllers

import (
	"context"
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
	"com.code.vidmicro/com.code.vidmicro/httpHandler/sessionshandler"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/utils"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u *UserController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u *UserController) GetCollectionName() basetypes.CollectionName {
	return "users"
}

func (u *UserController) DoIndexing() error {
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.User{})
	return nil
}

func (u *UserController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *UserController) copySession(w http.ResponseWriter, r *http.Request, user models.User) (context.Context, string, error) {
	sessionModel := models.Session{}
	sessionValue := r.Context().Value("session")
	data := ""

	if sessionValue != nil {
		sessionModel = sessionValue.(models.Session)
		sessionModel.Token, _ = utils.GenerateToken()
		sessionModel.Email = user.Email
		sessionModel.FirstName = user.FirstName
		sessionModel.LastName = user.LastName
		sessionModel.RegistrationType = user.RegistrationType
		sessionModel.UpdatedAt = user.UpdatedAt
		sessionModel.CreatedAt = user.CreatedAt
		//ignoring err because library said it will always give a response
		session, _ := sessionshandler.GetInstance().GetSession().Get(r, sessionModel.SessionId)
		session.Values["token"] = sessionModel.Token
		data = sessionModel.Token
		session.Values["email"] = sessionModel.Email
		session.Values["firstName"] = sessionModel.FirstName
		session.Values["lastName"] = sessionModel.LastName
		session.Values["registrationType"] = sessionModel.RegistrationType
		session.Values["createdAt"] = sessionModel.CreatedAt
		session.Values["updatedAt"] = sessionModel.UpdatedAt
		err := session.Save(r, w)
		if err != nil {
			return nil, "", err
		}
		ctx := context.WithValue(r.Context(), "session", sessionModel)
		return ctx, data, nil
	}
	return nil, "", errors.New("Session value doesn't match")
}

func (u *UserController) handleRegisterUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := models.User{}
		modelUser := models.User{}
		if err := c.ShouldBindJSON(&modelUser); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("path"), user)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		user.Salt, err = utils.GenerateSalt()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}
		user.CreatedAt = int(time.Now().Unix())
		user.UpdatedAt = int(time.Now().Unix())
		user.Password = utils.HashPassword(user.Password, user.Salt)
		user.RegistrationType = 1

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.CREATE_HASH_FAILED, err, nil))
			return
		}
		//TODO: have to manage session here
		//TODO: Add to db
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.REGISTER_USER_SUCCESS, err, nil))
	}
}

func (u *UserController) RegisterApis() {
	baserouter.GetInstance().GetOpenRouter().POST("/api/registerUser", u.handleRegisterUser())
}
