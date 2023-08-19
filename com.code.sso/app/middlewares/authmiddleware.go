package middlewares

import (
	"net/http"
)

type AuthMiddleware struct {
}

func (u *AuthMiddleware) GetHandlerFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// auth := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)
		// if len(auth) != 2 || auth[0] != "Basic" {
		// 	c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.BASIC_AUTH_FAILED, nil, nil))
		// }

		// payload, _ := base64.StdEncoding.DecodeString(auth[1])
		// pair := strings.SplitN(string(payload), ":", 2)
		// if len(pair) != 2 || !u.validate(pair[0], pair[1], c) {
		// 	c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.BASIC_AUTH_FAILED, nil, nil))
		// } else {
		// 	c.Set("userId", strings.TrimSpace(pair[0]))
		// 	c.Set("token", strings.TrimSpace(pair[1]))
		// 	c.Next()
		// }
	})
}

func (u *AuthMiddleware) validate(username string, password string) bool {
	//If it passes session middleware means session exists
	// sessionGeneric, _ := c.Get("session")
	// session := sessionGeneric.(models.Session)

	// if session.UserId == username && session.Token == password {
	// 	c.Set("role", int(session.Role))
	// 	c.Set("email", strings.ToLower(session.Email))
	// 	c.Set("phone", session.Phone)
	// 	c.Set("loginType", session.LoginType)
	// 	c.Set("masterAccount", bool(session.MasterAccount))
	// 	c.Set("currentUserId", session.CurrentUserId)
	// 	return true
	// }

	return false
}
