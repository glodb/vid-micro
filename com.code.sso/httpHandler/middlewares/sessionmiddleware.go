package middlewares

import (
	"net/http"
)

type SessionMiddleware struct {
}

func (u *SessionMiddleware) GetHandlerFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// data, err := cache.GetInstance().Get(c.Request.Header.Get("etSession"))
		// if err != nil || len(data) == 0 {
		// 	c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.SESSION_NOT_FOUND, err, nil))
		// } else {
		// 	var session models.Session
		// 	session.DecodeRedisData(data)
		// 	c.Set("session", session)
		// 	c.Set("fcmKey", session.FcmKey)
		// 	c.Set("appID", int(session.AppID))
		// 	c.Set("platform", strings.ToLower(session.Platform))
		// 	c.Set("version", session.Version)
		// 	c.Set("session-id", session.SessionId)
		// 	session.LastActivity = time.Now().Unix()
		// 	cache.GetInstance().Set(c.Request.Header.Get("etSession"), session.EncodeRedisData())
		// 	c.Next()
		// }
	})
}
