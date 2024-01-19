package middlewares

import (
	"encoding/base64"
	"net/http"
	"strings"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
}

func (u *AuthMiddleware) GetHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {

		auth := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)
		if len(auth) != 2 || auth[0] != "Basic" {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.BASIC_AUTH_FAILED, nil, nil))
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 || !u.validate(pair[0], pair[1], c) {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.BASIC_AUTH_FAILED, nil, nil))
		} else {
			c.Set("userId", strings.TrimSpace(pair[0]))
			c.Set("token", strings.TrimSpace(pair[1]))
			c.Next()
		}
	}
}

func (u *AuthMiddleware) validate(username string, password string, c *gin.Context) bool {
	//If it passes session middleware means session exists
	sessionGeneric, _ := c.Get("session")
	session := sessionGeneric.(models.Session)

	if session.UserId == username && session.Token == password {
		c.Set("role", int(session.Role))
		c.Set("email", strings.ToLower(session.Email))
		c.Set("phone", session.Phone)
		return true
	}

	return false
}
