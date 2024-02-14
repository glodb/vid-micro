package middlewares

import (
	"net/http"
	"strings"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"com.code.vidmicro/com.code.vidmicro/settings/utils"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
}

func (u *AuthMiddleware) GetHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {

		auth := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)
		if len(auth) != 2 || auth[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, responses.GetInstance().WriteResponse(c, responses.BEARER_AUTH_FAILED, nil, nil))
		}

		if u.validate(auth[1], c) {
			ok, err := utils.IsTokenValid(strings.TrimSpace(auth[1]))

			if ok && err == nil {

				c.Set("token", strings.TrimSpace(auth[1]))
				c.Next()
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, responses.GetInstance().WriteResponse(c, responses.TOKEN_EXPIRED, nil, nil))
			}
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, responses.GetInstance().WriteResponse(c, responses.BEARER_AUTH_FAILED, nil, nil))
		}
	}
}

func (u *AuthMiddleware) validate(password string, c *gin.Context) bool {
	//If it passes session middleware means session exists
	sessionGeneric, _ := c.Get("session")
	session := sessionGeneric.(models.Session)

	if session.Token == password {
		c.Set("userId", session.UserId)
		c.Set("role", int(session.Role))
		c.Set("email", strings.ToLower(session.Email))
		c.Set("username", strings.ToLower(session.Username))
		c.Set("roleName", strings.ToLower(session.RoleName))
		return true
	}

	return false
}
