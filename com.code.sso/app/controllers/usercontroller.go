package controllers

import (
	"encoding/json"
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
)

type UserController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u *UserController) GetDBName() basetypes.DBName {
	return basetypes.DBName(config.GetInstance().Database.DBName)
}

func (u *UserController) GetCollectionName() basetypes.CollectionName {
	return "users"
}

func (u *UserController) DoIndexing() error {
	return nil
}

func (u *UserController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *UserController) handleRegisterUser(w http.ResponseWriter, r *http.Request) {
	user := models.User{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	defer r.Body.Close()
	if err != nil {
		responses.GetInstance().WriteJsonResponse(w, r, responses.MALFORMED_JSON, errors.New("illegal json format"), nil)
		return
	}

	err = u.Validate(r.URL.Path, user)
	if err != nil {
		responses.GetInstance().WriteJsonResponse(w, r, responses.VALIDATION_FAILED, err, nil)
		return
	}
	err = u.Add(u.GetDBName(), u.GetCollectionName(), user)
	if err != nil {
		responses.GetInstance().WriteJsonResponse(w, r, responses.DB_ERROR, err, nil)
		return
	}
	responses.GetInstance().WriteJsonResponse(w, r, responses.REGISTER_USER_SUCCESS, nil, nil)
}

func (u *UserController) RegisterApis() {
	baserouter.GetInstance().GetOpenRouter().HandleFunc("/api/registerUser", u.handleRegisterUser).Methods("POST")
	// baserouter.GetInstance().GetOpenRouter().POST("/api/register", u.registerUser())
	// baserouter.GetInstance().GetOpenRouter().POST("/api/login", u.loginUser())
	// baserouter.GetInstance().GetLoginRouter().GET("/api/getUser", u.getUser())
	// baserouter.GetInstance().GetLoginRouter().POST("/api/logout", u.logout())
	// baserouter.GetInstance().GetLoginRouter().POST("/api/setFCMKey", u.setFCMKey())
	// baserouter.GetInstance().GetOpenRouter().POST("/api/loginAdmin", u.loginAdmin())
	// baserouter.GetInstance().GetLoginRouter().POST("/api/registerAdmin", u.registerAdmin())
	// baserouter.GetInstance().GetLoginRouter().GET("/api/getAllSuspendedUsers", u.getAllSuspendedUsers())
	// baserouter.GetInstance().GetLoginRouter().GET("/api/getAllUnsuspendedUsers", u.getAllUnsuspendedUsers())
	// baserouter.GetInstance().GetLoginRouter().POST("/api/suspendUser", u.suspendUser())
	// baserouter.GetInstance().GetLoginRouter().POST("/api/unsuspendUser", u.unsuspendUser())
	// baserouter.GetInstance().GetLoginRouter().GET("/api/getAllUsers", u.getAllUsers())
	// baserouter.GetInstance().GetLoginRouter().POST("/api/updateAvatar", u.updateAvatar())
	// baserouter.GetInstance().GetLoginRouter().POST("/api/updateEmail", u.updateEmail())
	// baserouter.GetInstance().GetLoginRouter().POST("/api/updatePhone", u.updatePhone())
	// baserouter.GetInstance().GetLoginRouter().POST("/api/updateRole", u.updateRole())
	// baserouter.GetInstance().GetServerRouter().GET("/api/checkLogin", u.checkLogin())
}
