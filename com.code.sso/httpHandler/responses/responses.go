package responses

import (
	"net/http"
	"sync"
)

const (
	API_NOT_AVAILABLE = 1000
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
	u.responses[API_NOT_AVAILABLE] = "The Api is not available on current server"
}

// GetResponse returns the message for the particular response code
func (u *Responses) getResponse(code int) map[string]interface{} {
	message := make(map[string]interface{})
	message["code"] = code
	message["message"] = u.responses[code]

	return message
}

func (u *Responses) WriteJsonResponse(w http.ResponseWriter, r *http.Request, code int, err error, data interface{}) map[string]interface{} {
	// urlPath := r.URL
	returnMap := u.getResponse(code)

	// jsonData, _ := ioutil.ReadAll(r.Body)

	// auditTrial := models.AuditTrial{
	// 	Body:    string(jsonData),
	// 	Url:     urlPath,
	// 	Code:    returnMap["code"].(int),
	// 	Message: returnMap["message"].(string),
	// 	Email:   c.GetString("email"),
	// 	Phone:   c.GetString("phone"),
	// 	UserID:  c.GetString("userId"),
	// 	Method:  c.Request.Method,
	// 	IP:      c.ClientIP(),
	// }

	// if err != nil {
	// 	returnMap["error"] = err.Error()
	// 	auditTrial.Error = err.Error()
	// }
	// if data != nil {
	// 	returnMap["data"] = data
	// 	dataBytes, _ := json.Marshal(data)
	// 	auditTrial.Response = string(dataBytes)
	// }
	// log.Println("auditLogs:", auditTrial)

	return returnMap
}
