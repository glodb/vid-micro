package middlewares

import (
	"net/http"
	"time"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"com.code.vidmicro/com.code.vidmicro/settings/cache"
	"github.com/gin-gonic/gin"
)

type SessionMiddleware struct {
}

func (u *SessionMiddleware) GetHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := cache.GetInstance().Get(c.Request.Header.Get("vidmicroSession"))
		if err != nil || len(data) == 0 {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.SESSION_NOT_FOUND, err, nil))
		} else {
			var session models.Session
			session.DecodeRedisData(data)
			if session.BlackListed {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.SESSION_NOT_FOUND, err, nil))
				return
			}
			c.Set("session", session)
			c.Set("session-id", session.SessionId)
			session.LastActivity = time.Now().Unix()
			cache.GetInstance().Set(c.Request.Header.Get("vidmicroSession"), session.EncodeRedisData())
			c.Next()
		}
	}
}
