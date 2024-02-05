package responses

import (
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
)

const (
	WELCOME_TO_SSO                           = 1000
	NOT_FOUND                                = 1001
	OPTIONS_NOT_ALLOWED                      = 1002
	CREATE_SESSION_SUCCESS                   = 1003
	GENERATING_UUID_FAILED                   = 1004
	REGISTER_USER_SUCCESS                    = 1005
	SESSION_ID_NOT_PRESENT                   = 1006
	SESSION_NOT_VALID                        = 1007
	CREATE_SESSION_FAILED                    = 1008
	MALFORMED_JSON                           = 1009
	SERVER_ERROR                             = 1010
	DB_ERROR                                 = 1011
	CREATE_HASH_FAILED                       = 1012
	BASIC_AUTH_FAILED                        = 1013
	GET_USER_SUCCESS                         = 1014
	GET_USER_FAILED                          = 1015
	USERNAME_EXISTS_FAILED                   = 1016
	PASSWORD_INCORRECT                       = 1017
	LOGIN_SUCCESS                            = 1018
	UPDATE_FAILED                            = 1019
	UPDATE_SUCCESS                           = 1020
	LOGOUT_SUCCESS                           = 1021
	LOGOUT_FAILED                            = 1022
	SESSION_NOT_FOUND                        = 1023
	METHOD_NOT_AVAILABLE                     = 1024
	ADDING_DB_FAILED                         = 1025
	ERROR_READING_USER                       = 1026
	PASSWORD_MISMATCHED                      = 1027
	ERROR_CREATING_JWT                       = 1028
	USER_BLOCKED                             = 1029
	TOKEN_EXPIRED                            = 1030
	REFRESH_TOKEN_REQUIRED                   = 1031
	INVALID_REFRESH_TOKEN                    = 1032
	REFRESH_TOKEN_SUCCESS                    = 1033
	API_NOT_ACCESSABLE                       = 1034
	FAILED_BLACK_LISTING                     = 1035
	BLACK_LIST_SUCCESS                       = 1036
	UPLOADING_AVATAR_FAILED                  = 1037
	FAILED_UPDATING_USER                     = 1038
	UPDATING_USER_SUCCESS                    = 1039
	NOTHIN_TO_UPDATE                         = 1040
	PUTTING_SUCCESS                          = 1041
	PUTTING_FAILED                           = 1042
	GETTING_FAILED                           = 1043
	GETTING_SUCCESS                          = 1044
	UPDATING_SUCCESS                         = 1045
	UPDATING_FAILED                          = 1046
	DELETING_SUCCESS                         = 1047
	DELETING_FAILED                          = 1048
	LANGUAGE_NOT_ADDED_IN_TITLE              = 1049
	GENERATE_EMAIL_VERIFICATION_TOKEN_FAILED = 1050
	SEND_VERIFICATION_EMAIL_FAILED           = 1051
	INVALID_EMAIL_OR_TOKEN                   = 1052
	EMAIL_VERIFICATION_FAILED                = 1053
	EMAIL_VERIFICATION_SUCCESS               = 1054
	INVALID_PASSWORD_TOKEN                   = 1055
	NOT_VARIFIED_USER                        = 1056
	TOKEN_SENT_VIA_EMAIL                     = 1057
	TOKEN_ALREADY_SENT                       = 1058
	TOKEN_AND_NEW_PASSWORD_REQUIRED          = 1059
	TOKEN_VERIFICTION_SUCCESS                = 1060
	GOOGLE_LOGIN_FAILED                      = 1061
	FORBIDDEN                                = 1062
	BAD_REQUEST                              = 1063
	URL_GENERATED                            = 1064
	VALIDATION_FAILED                        = 1065
	USERNAME_OR_EMAIL_EXISTS                 = 1066
	SESSION_NOT_PROVIDED                     = 1067
	TWITTER_LOGIN_FAILED                     = 1068
	INVALID_OR_EXPIRED_OAUTH_TOKEN           = 1069
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
	u.responses[WELCOME_TO_SSO] = "Welcome to VidMicro"
	u.responses[NOT_FOUND] = "The Api is not available on current server"
	u.responses[OPTIONS_NOT_ALLOWED] = "Options are not allowed"
	u.responses[CREATE_SESSION_SUCCESS] = "Getting session successful"
	u.responses[GENERATING_UUID_FAILED] = "Generating UUID failed"
	u.responses[SESSION_ID_NOT_PRESENT] = "Session id is not present in header"
	u.responses[SESSION_NOT_VALID] = "Session not valid"
	u.responses[REGISTER_USER_SUCCESS] = "Register user success"
	u.responses[MALFORMED_JSON] = "Json Decoding failed"
	u.responses[SERVER_ERROR] = "Server Error"
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
	u.responses[SESSION_NOT_FOUND] = "SESSION_NOT_FOUND"
	u.responses[METHOD_NOT_AVAILABLE] = "METHOD_NOT_AVAILABLE"
	u.responses[ADDING_DB_FAILED] = "ADDING_DB_FAILED"
	u.responses[ERROR_READING_USER] = "ERROR_READING_USER"
	u.responses[ERROR_CREATING_JWT] = "ERROR_CREATING_JWT"
	u.responses[USER_BLOCKED] = "USER_BLOCKED"
	u.responses[TOKEN_EXPIRED] = "TOKEN_EXPIRED"
	u.responses[REFRESH_TOKEN_REQUIRED] = "REFRESH_TOKEN_REQUIRED"
	u.responses[INVALID_REFRESH_TOKEN] = "INVALID_REFRESH_TOKEN"
	u.responses[REFRESH_TOKEN_SUCCESS] = "REFRESH_TOKEN_SUCCESS"
	u.responses[API_NOT_ACCESSABLE] = "API_NOT_ACCESSABLE"
	u.responses[FAILED_BLACK_LISTING] = "FAILED_BLACK_LISTING"
	u.responses[BLACK_LIST_SUCCESS] = "BLACK_LIST_SUCCESS"
	u.responses[UPLOADING_AVATAR_FAILED] = "UPLOADING_AVATAR_FAILED"
	u.responses[FAILED_UPDATING_USER] = "FAILED_UPDATING_USER"
	u.responses[UPDATING_USER_SUCCESS] = "UPDATING_USER_SUCCESS"
	u.responses[NOTHIN_TO_UPDATE] = "NOTHIN_TO_UPDATE"
	u.responses[PUTTING_SUCCESS] = "PUTTING_SUCCESS"
	u.responses[PUTTING_FAILED] = "PUTTING_FAILED"
	u.responses[GETTING_FAILED] = "GETTING_FAILED"
	u.responses[GETTING_SUCCESS] = "GETTING_SUCCESS"
	u.responses[UPDATING_SUCCESS] = "UPDATING_SUCCESS"
	u.responses[UPDATING_FAILED] = "UPDATING_FAILED"
	u.responses[DELETING_SUCCESS] = "DELETING_SUCCESS"
	u.responses[DELETING_FAILED] = "DELETING_FAILED"
	u.responses[LANGUAGE_NOT_ADDED_IN_TITLE] = "LANGUAGE_NOT_ADDED_IN_TITLE"
	u.responses[GENERATE_EMAIL_VERIFICATION_TOKEN_FAILED] = "GENERATE_EMAIL_VERIFICATION_TOKEN_FAILED"
	u.responses[SEND_VERIFICATION_EMAIL_FAILED] = "SEND_VERIFICATION_EMAIL_FAILED"
	u.responses[INVALID_EMAIL_OR_TOKEN] = "INVALID_EMAIL_OR_TOKEN"
	u.responses[EMAIL_VERIFICATION_FAILED] = "EMAIL_VERIFICATION_FAILED"
	u.responses[EMAIL_VERIFICATION_SUCCESS] = "EMAIL_VERIFICATION_SUCCESS"
	u.responses[INVALID_PASSWORD_TOKEN] = "Invalid password reset token"
	u.responses[NOT_VARIFIED_USER] = "User not verified"
	u.responses[TOKEN_SENT_VIA_EMAIL] = "Token sent via email"
	u.responses[TOKEN_ALREADY_SENT] = "Token already sent and not expired"
	u.responses[PASSWORD_MISMATCHED] = "Password mismatched"
	u.responses[TOKEN_AND_NEW_PASSWORD_REQUIRED] = "Token and new password, both fields are required and should not be empty"
	u.responses[TOKEN_VERIFICTION_SUCCESS] = "Password token verified successfully"
	u.responses[GOOGLE_LOGIN_FAILED] = "GOOGLE_LOGIN_FAILED"
	u.responses[FORBIDDEN] = "FORBIDDEN"
	u.responses[BAD_REQUEST] = "BAD_REQUEST"
	u.responses[URL_GENERATED] = "URL_GENERATED"
	u.responses[USERNAME_EXISTS_FAILED] = "User name or email already exists"
	u.responses[VALIDATION_FAILED] = "Validation failed on the field"
}

// GetResponse returns the message for the particular response code
func (u *Responses) getResponse(code int) map[string]interface{} {
	message := make(map[string]interface{})
	message["code"] = code
	message["message"] = u.responses[code]

	return message
}

func (u *Responses) WriteResponse(c *gin.Context, code int, err interface{}, data interface{}) map[string]interface{} {
	returnMap := u.getResponse(code)
	queryBytes, _ := sonic.Marshal(c.Request.URL.Query())

	jsonData, _ := ioutil.ReadAll(c.Request.Body)
	// dst := bytes.Buffer{}
	// json.Compact(&dst, jsonData)

	auditTrial := models.AuditTrial{
		QueryParams: string(queryBytes),
		Body:        string(jsonData),
		Url:         c.Request.URL.Path,
		Code:        returnMap["code"].(int),
		Message:     returnMap["message"].(string),
		Email:       c.GetString("email"),
		Phone:       c.GetString("phone"),
		UserID:      c.GetString("userId"),
		Role:        c.GetInt("role"),
		Method:      c.Request.Method,
		IP:          c.ClientIP(),
		Version:     c.GetString("version"),
		Platform:    c.GetString("platform"),
	}

	if err != nil {

		switch val := err.(type) {
		case error:
			if configmanager.GetInstance().WriteError {
				returnMap["error"] = val.Error()
			}
			auditTrial.Error = val.Error()
		case map[string][]string:
			returnMap["errors"] = val
			jsonBytes, _ := sonic.Marshal(val)
			auditTrial.Error = string(jsonBytes)
		default:
			fmt.Println("Not supported type of errors")
		}

	}

	if data != nil {
		returnMap["data"] = data
		dataBytes, _ := sonic.Marshal(data)
		auditTrial.Response = string(dataBytes)
	}
	log.Println("auditLogs:", auditTrial)
	return returnMap
}

// func (u *Responses) WriteJsonResponse(w http.ResponseWriter, r *http.Request, code int, err error, data interface{}) {
// 	// urlPath := r.URL
// 	returnMap := u.getResponse(code)

// 	jsonData, _ := ioutil.ReadAll(r.Body)
// 	session := models.Session{}
// 	sessionValue := r.Context().Value("session")

// 	if sessionValue != nil {
// 		session = sessionValue.(models.Session)
// 	}
// 	auditTrial := models.AuditTrial{
// 		Body:    string(jsonData),
// 		Url:     r.URL.String(),
// 		Code:    returnMap["code"].(int),
// 		Message: returnMap["message"].(string),
// 		Session: session.SessionId,
// 		Email:   session.Email,
// 		Method:  r.Method,
// 		IP:      r.RemoteAddr,
// 	}

// 	status := http.StatusOK
// 	if err != nil {
// 		status = http.StatusNotAcceptable
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Header().Set("Cache-Control", "no-store")

// 	w.WriteHeader(status)

// 	if err != nil {
// 		returnMap["error"] = err.Error()
// 		auditTrial.Error = err.Error()
// 		status = http.StatusNotModified
// 	}

// 	if data != nil {
// 		returnMap["data"] = data
// 		dataBytes, _ := sonic.Marshal(data)
// 		auditTrial.Response = string(dataBytes)
// 	}
// 	err = sonic.ConfigDefault.NewEncoder(w).Encode(returnMap)
// 	if err != nil {
// 		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
// 		return
// 	}

// 	log.Println("auditLogs:", auditTrial)
// }
