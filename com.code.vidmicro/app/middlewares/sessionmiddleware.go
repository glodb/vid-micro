package middlewares

import (
	"errors"
	"log"
	"net/http"
	"time"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"com.code.vidmicro/com.code.vidmicro/settings/cache"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/cookie"
	"github.com/gin-gonic/gin"
)

type SessionMiddleware struct {
}

// func (u *SessionMiddleware) createSession(c *gin.Context) (models.Session, error) {
// 	sessionId, err := utils.GenerateUUID()
// 	modelSession := models.Session{
// 		SessionId: sessionId,
// 	}

// 	if err != nil {
// 		return modelSession, err
// 	}
// 	now := time.Now().Unix()
// 	modelSession.CreatedAt = now
// 	modelSession.ExpiringAt = now + configmanager.GetInstance().SessionExpirySeconds
// 	err = cache.GetInstance().Set(sessionId, modelSession.EncodeRedisData())

// 	if err != nil {
// 		return modelSession, err
// 	}

// 	err = cache.GetInstance().Expire(sessionId, int(configmanager.GetInstance().SessionExpirySeconds))
// 	if err != nil {
// 		return modelSession, err
// 	}

// 	// Set a cookie
// 	c.SetCookie(configmanager.GetInstance().CookieName, sessionId, int(configmanager.GetInstance().SessionExpirySeconds), configmanager.GetInstance().CookiePath, configmanager.GetInstance().CookieDomain, false, true)
// 	return modelSession, nil
// }

// func (u *SessionMiddleware) getCookieSession(c *gin.Context) (models.Session, error) {
// 	cookieValue, err := c.Cookie(configmanager.GetInstance().CookieName)

// 	if err != nil {
// 		return u.createSession(c)
// 	}
// 	var session models.Session
// 	data, err := cache.GetInstance().Get(cookieValue)
// 	if err != nil || len(data) == 0 {
// 		return u.createSession(c)
// 	} else {
// 		session.DecodeRedisData(data)
// 	}
// 	return session, nil
// }

// func (u *SessionMiddleware) processSession(c *gin.Context, session models.Session) bool {
// 	if session.BlackListed {
// 		c.AbortWithStatusJSON(http.StatusUnauthorized, responses.GetInstance().WriteResponse(c, responses.SESSION_NOT_FOUND, errors.New("session is blacklisted"), nil))
// 		return false
// 	}
// 	c.Set("session", session)
// 	c.Set("session-id", session.SessionId)
// 	session.LastActivity = time.Now().Unix()
// 	err := cache.GetInstance().Set(session.SessionId, session.EncodeRedisData())

// 	if err != nil {
// 		c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
// 		return false
// 	}
// 	return true
// }

func (u *SessionMiddleware) GetHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {

		if localCookie, err := c.Cookie(configmanager.GetInstance().CookieName); err == nil {
			value := make([]byte, 0)
			if err = cookie.GetInstance().GetCookie().Decode(configmanager.GetInstance().CookieName, localCookie, &value); err == nil {
				log.Println(string(localCookie))
				var session models.Session
				data, err := cache.GetInstance().Get(string(localCookie))
				if err != nil || len(data) == 0 {
					c.AbortWithStatusJSON(http.StatusUnauthorized, responses.GetInstance().WriteResponse(c, responses.SESSION_NOT_PROVIDED, err, nil))
					return
				} else {
					session.DecodeRedisData(data)
				}
				if session.BlackListed {
					c.AbortWithStatusJSON(http.StatusUnauthorized, responses.GetInstance().WriteResponse(c, responses.SESSION_NOT_FOUND, errors.New("session is blacklisted"), nil))
					return
				}
				c.Set("session", session)
				c.Set("session-id", session.SessionId)
				session.LastActivity = time.Now().Unix()
				err = cache.GetInstance().Set(session.SessionId, session.EncodeRedisData())
				if err != nil {
					c.AbortWithStatusJSON(http.StatusUnauthorized, responses.GetInstance().WriteResponse(c, responses.SESSION_NOT_PROVIDED, err, nil))
					return
				}
				c.Next()
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, responses.GetInstance().WriteResponse(c, responses.SESSION_NOT_PROVIDED, err, nil))
			}
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, responses.GetInstance().WriteResponse(c, responses.SESSION_NOT_PROVIDED, err, nil))
		}
		// if session, err := u.getCookieSession(c); err == nil {

		// 	if u.processSession(c, session) {
		// 		c.Next()
		// 	}

		// } else if data, err := cache.GetInstance().Get(c.Request.Header.Get("jwt")); err == nil {
		// 	if err != nil || len(data) == 0 {
		// 		c.AbortWithStatusJSON(http.StatusUnauthorized, responses.GetInstance().WriteResponse(c, responses.SESSION_NOT_FOUND, err, nil))
		// 	} else {
		// 		var session models.Session
		// 		session.DecodeRedisData(data)

		// 		if u.processSession(c, session) {
		// 			c.Next()
		// 		}
		// 	}
		// } else {
		// 	c.AbortWithStatusJSON(http.StatusUnauthorized, responses.GetInstance().WriteResponse(c, responses.SESSION_NOT_PROVIDED, err, nil))
		// }
	}
}
