package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

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
	"golang.org/x/oauth2"
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
		session, err := sessionshandler.GetInstance().GetSession().Get(r, sessionModel.SessionId)
		session.Values["token"] = sessionModel.Token
		data = sessionModel.Token
		session.Values["email"] = sessionModel.Email
		session.Values["firstName"] = sessionModel.FirstName
		session.Values["lastName"] = sessionModel.LastName
		session.Values["registrationType"] = sessionModel.RegistrationType
		session.Values["createdAt"] = sessionModel.CreatedAt
		session.Values["updatedAt"] = sessionModel.UpdatedAt
		err = session.Save(r, w)
		if err != nil {
			return nil, "", err
		}
		ctx := context.WithValue(r.Context(), "session", sessionModel)
		return ctx, data, nil
	}
	return nil, "", errors.New("Session value doesn't match")
}

func (u *UserController) HandleRegisterUser(w http.ResponseWriter, r *http.Request) {
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

	user.Salt, err = utils.GenerateSalt()
	if err != nil {
		responses.GetInstance().WriteJsonResponse(w, r, responses.VALIDATION_FAILED, err, nil)
		return
	}
	user.CreatedAt = int(time.Now().Unix())
	user.UpdatedAt = int(time.Now().Unix())
	user.Password = utils.HashPassword(user.Password, user.Salt)
	user.RegistrationType = 1

	if err != nil {
		responses.GetInstance().WriteJsonResponse(w, r, responses.CREATE_HASH_FAILED, err, nil)
		return
	}

	ctx, data, err := u.copySession(w, r, user)
	if err != nil {
		responses.GetInstance().WriteJsonResponse(w, r, responses.CREATE_HASH_FAILED, err, nil)
		return
	}
	err = u.Add(u.GetDBName(), u.GetCollectionName(), user)
	if err != nil {
		responses.GetInstance().WriteJsonResponse(w, r, responses.DB_ERROR, err, nil)
		return
	}
	responses.GetInstance().WriteJsonResponse(w, r.WithContext(ctx), responses.REGISTER_USER_SUCCESS, nil, data)
}

func (u *UserController) HandleGetUser(w http.ResponseWriter, r *http.Request) {

	session := models.Session{}
	sessionValue := r.Context().Value("session")

	if sessionValue != nil {
		session = sessionValue.(models.Session)
		responses.GetInstance().WriteJsonResponse(w, r, responses.GET_USER_SUCCESS, nil, session)
	} else {
		responses.GetInstance().WriteJsonResponse(w, r, responses.GET_USER_FAILED, nil, session)
	}
}

func (u *UserController) HandleLogin(w http.ResponseWriter, r *http.Request) {

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

	query := make(map[string]interface{})
	query["email"] = user.Email
	query["registrationType"] = 1

	rows, err := u.FindOne(u.GetDBName(), u.GetCollectionName(), query, &user)

	rowsCount := 0
	enteredPassword := user.Password

	for rows.Next() {
		rowsCount++
		err = rows.Scan(&user.Id, &user.Email, &user.Phone, &user.Password, &user.FirstName, &user.LastName, &user.Salt, &user.RegistrationType, &user.CreatedAt, &user.UpdatedAt)
	}

	if rowsCount == 0 {
		responses.GetInstance().WriteJsonResponse(w, r, responses.USERNAME_EXISTS_FAILED, errors.New("No rows returned"), nil)
		return
	}
	hashedEntered := utils.HashPassword(enteredPassword, user.Salt)

	if hashedEntered != user.Password {
		responses.GetInstance().WriteJsonResponse(w, r, responses.PASSWORD_INCORRECT, errors.New("Password doesn't matched"), nil)
		return
	}

	ctx, data, err := u.copySession(w, r, user)
	if err != nil {
		responses.GetInstance().WriteJsonResponse(w, r, responses.CREATE_HASH_FAILED, err, nil)
		return
	}
	responses.GetInstance().WriteJsonResponse(w, r.WithContext(ctx), responses.LOGIN_SUCCESS, nil, data)
}

var (
	googleOauthConfig = oauth2.Config{
		ClientID:     "359287556402-j1fchiuumr87kjvsh73oik2am4inoovv.apps.googleusercontent.com",
		ClientSecret: "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",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
	}
)

func (u *UserController) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	sessionId := r.URL.Query().Get("sessionId")

	session, _ := sessionshandler.GetInstance().GetSession().Get(r, sessionId)
	session.Values["sessionId"] = sessionId

	session.Save(r, w)

	url := googleOauthConfig.AuthCodeURL(sessionId, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (u *UserController) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	session, _ := sessionshandler.GetInstance().GetSession().Get(r, state)
	session.Values["sessionId"] = state
	session.Save(r, w)

	sessionModel := models.Session{
		SessionId: state,
	}
	ctx := context.WithValue(r.Context(), "session", sessionModel)

	token, err := googleOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Println("Token exchange error:", err)
		return
	}

	client := googleOauthConfig.Client(r.Context(), token)
	userInfo, err := getUserInfo(client)
	if err != nil {
		log.Println("Userinfo retrieval error:", err)
		return
	}
	user := models.User{}
	user.Email = userInfo.Email
	user.FirstName = userInfo.GivenName
	user.LastName = userInfo.FamilyName
	user.CreatedAt = int(time.Now().Unix())
	user.UpdatedAt = int(time.Now().Unix())
	user.Password = utils.HashPassword(user.Password, user.Salt)
	user.RegistrationType = 2

	ctx, localToken, err := u.copySession(w, r.WithContext(ctx), user)
	if err != nil {
		responses.GetInstance().WriteJsonResponse(w, r, responses.CREATE_HASH_FAILED, err, nil)
		return
	}

	query := make(map[string]interface{})
	query["email"] = user.Email
	query["registrationType"] = 2

	rows, err := u.FindOne(u.GetDBName(), u.GetCollectionName(), query, &user)

	if rows.Next() {
		http.Redirect(w, r, "/dashboard?email="+user.Email+"&token="+localToken+"&registrationType=2", http.StatusSeeOther)
	} else {
		u.Add(u.GetDBName(), u.GetCollectionName(), user)
		http.Redirect(w, r, "/updateprofile?email="+user.Email+"&token="+localToken+"&registrationType=2", http.StatusSeeOther)
	}
}

func getUserInfo(client *http.Client) (*models.UserInfo, error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo models.UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	return &userInfo, nil
}

func (u *UserController) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	user := models.User{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	defer r.Body.Close()
	if err != nil {
		responses.GetInstance().WriteJsonResponse(w, r, responses.MALFORMED_JSON, errors.New("illegal json format"), nil)
		return
	}

	query := "UPDATE " + string(u.GetCollectionName()) + " SET firstName = ?, lastName = ?, email = ?, phone = ? WHERE email = ?"
	sessionModel := models.Session{}
	sessionValue := r.Context().Value("session")

	if sessionValue != nil {
		sessionModel = sessionValue.(models.Session)
	}

	values := make([]interface{}, 0)
	values = append(values, user.FirstName)
	values = append(values, user.LastName)
	values = append(values, user.Email)
	values = append(values, user.Phone)
	values = append(values, sessionModel.Email)
	sessionModel.Phone = user.Phone

	session, err := sessionshandler.GetInstance().GetSession().Get(r, sessionModel.SessionId)
	session.Values["email"] = user.Email
	session.Values["firstName"] = user.FirstName
	session.Values["lastName"] = user.LastName
	session.Values["registrationType"] = sessionModel.RegistrationType
	session.Values["phone"] = user.Phone
	session.Values["updatedAt"] = int(time.Now().Unix())
	err = session.Save(r, w)
	ctx := context.WithValue(r.Context(), "session", sessionModel)

	err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), query, values, false)
	if err != nil {
		responses.GetInstance().WriteJsonResponse(w, r, responses.UPDATE_FAILED, err, nil)
	} else {
		responses.GetInstance().WriteJsonResponse(w, r.WithContext(ctx), responses.UPDATE_SUCCESS, nil, nil)
	}
}

func (u *UserController) HandleLogout(w http.ResponseWriter, r *http.Request) {
	sessionModel := models.Session{}
	sessionValue := r.Context().Value("session")

	if sessionValue != nil {
		sessionModel = sessionValue.(models.Session)
	}
	session, err := sessionshandler.GetInstance().GetSession().Get(r, sessionModel.SessionId)
	if err != nil {
		responses.GetInstance().WriteJsonResponse(w, r, responses.LOGOUT_FAILED, nil, nil)
	}
	session.Values["email"] = ""
	session.Values["firstName"] = ""
	session.Values["lastName"] = ""
	session.Values["registrationType"] = 0
	session.Values["phone"] = ""
	session.Values["updatedAt"] = 0
	err = session.Save(r, w)
	ctx := context.WithValue(r.Context(), "session", sessionModel)
	responses.GetInstance().WriteJsonResponse(w, r.WithContext(ctx), responses.LOGOUT_SUCCESS, nil, nil)
}
func (u *UserController) RegisterApis() {
	baserouter.GetInstance().GetOpenRouter().HandleFunc("/api/registerUser", u.HandleRegisterUser).Methods("POST")
	baserouter.GetInstance().GetLoginRouter().HandleFunc("/api/getUser", u.HandleGetUser).Methods("GET")
	baserouter.GetInstance().GetOpenRouter().HandleFunc("/api/login", u.HandleLogin).Methods("POST")
	baserouter.GetInstance().GetBaseRouter().HandleFunc("/api/googleLogin", u.HandleGoogleLogin).Methods("GET")
	baserouter.GetInstance().GetBaseRouter().HandleFunc("/callback", u.HandleCallback).Methods("GET")
	baserouter.GetInstance().GetLoginRouter().HandleFunc("/api/updateUser", u.HandleUpdateUser).Methods("POST")
	baserouter.GetInstance().GetLoginRouter().HandleFunc("/api/logout", u.HandleLogout).Methods("GET")
}
