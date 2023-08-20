package responses

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"com.code.sso/com.code.sso/app/models"
)

const (
	WELCOME_TO_SSO         = 1000
	API_NOT_AVAILABLE      = 1001
	OPTIONS_NOT_ALLOWED    = 1002
	CREATE_SESSION_SUCCESS = 1003
	GENERATING_UUID_FAILED = 1004
	REGISTER_USER_SUCCESS  = 1005
	SESSION_ID_NOT_PRESENT = 1006
	SESSION_NOT_VALID      = 1007
	CREATE_SESSION_FAILED  = 1008
	MALFORMED_JSON         = 1009
	VALIDATION_FAILED      = 1010
	DB_ERROR               = 1011
	CREATE_HASH_FAILED     = 1012
	BASIC_AUTH_FAILED      = 1013
	GET_USER_SUCCESS       = 1014
	GET_USER_FAILED        = 1015
	USERNAME_EXISTS_FAILED = 1016
	PASSWORD_INCORRECT     = 1017
	LOGIN_SUCCESS          = 1018
	UPDATE_FAILED          = 1019
	UPDATE_SUCCESS         = 1020
	LOGOUT_SUCCESS         = 1021
	LOGOUT_FAILED          = 1022
)

type Responses struct {
	responses map[int]string
}

var (
	instance *Responses
	once     sync.Once
)

// Singleton. Returns a single object of Factory
func GetInstance() *Responses {
	// var instance
	once.Do(func() {
		instance = &Responses{}
		instance.InitResponses()
	})
	return instance
}

// InitResponses function just initialise the response headers to be sent
func (u *Responses) InitResponses() {
	u.responses = make(map[int]string)
	u.responses[WELCOME_TO_SSO] = "Welcome to SSO"
	u.responses[API_NOT_AVAILABLE] = "The Api is not available on current server"
	u.responses[OPTIONS_NOT_ALLOWED] = "Options are not allowed"
	u.responses[CREATE_SESSION_SUCCESS] = "Creating session successful"
	u.responses[GENERATING_UUID_FAILED] = "Generating UUID failed"
	u.responses[SESSION_ID_NOT_PRESENT] = "Session id is not present in header"
	u.responses[SESSION_NOT_VALID] = "Session not valid"
	u.responses[REGISTER_USER_SUCCESS] = "Register user success"
	u.responses[MALFORMED_JSON] = "Json Decoding failed"
	u.responses[VALIDATION_FAILED] = "Validation failed"
	u.responses[DB_ERROR] = "DB Error in query"
	u.responses[CREATE_HASH_FAILED] = "Failed creating hash"
	u.responses[BASIC_AUTH_FAILED] = "Basic auth failed"
	u.responses[GET_USER_SUCCESS] = "Success in getting user"
	u.responses[GET_USER_FAILED] = "Failure in getting user"
	u.responses[USERNAME_EXISTS_FAILED] = "Username entered does not exist"
	u.responses[PASSWORD_INCORRECT] = "Password is incorrect"
	u.responses[LOGIN_SUCCESS] = "Login Success"
	u.responses[UPDATE_FAILED] = "Updation Failed"
	u.responses[UPDATE_SUCCESS] = "Updateion Success"
}

// GetResponse returns the message for the particular response code
func (u *Responses) getResponse(code int) map[string]interface{} {
	message := make(map[string]interface{})
	message["code"] = code
	message["message"] = u.responses[code]

	return message
}

func (u *Responses) WriteJsonResponse(w http.ResponseWriter, r *http.Request, code int, err error, data interface{}) {
	// urlPath := r.URL
	returnMap := u.getResponse(code)

	jsonData, _ := ioutil.ReadAll(r.Body)
	session := models.Session{}
	sessionValue := r.Context().Value("session")

	if sessionValue != nil {
		session = sessionValue.(models.Session)
	}
	auditTrial := models.AuditTrial{
		Body:    string(jsonData),
		Url:     r.URL.String(),
		Code:    returnMap["code"].(int),
		Message: returnMap["message"].(string),
		Session: session.SessionId,
		Email:   session.Email,
		Phone:   session.Phone,
		Method:  r.Method,
		IP:      r.RemoteAddr,
	}

	status := http.StatusOK
	if err != nil {
		status = http.StatusNotAcceptable
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	w.WriteHeader(status)

	if err != nil {
		returnMap["error"] = err.Error()
		auditTrial.Error = err.Error()
		status = http.StatusNotModified
	}

	if data != nil {
		returnMap["data"] = data
		dataBytes, _ := json.Marshal(data)
		auditTrial.Response = string(dataBytes)
	}
	err = json.NewEncoder(w).Encode(returnMap)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	log.Println("auditLogs:", auditTrial)
}
