package middlewares

import (
	"context"
	"errors"
	"net/http"

	"com.code.sso/com.code.sso/app/models"
	"com.code.sso/com.code.sso/httpHandler/responses"
	"com.code.sso/com.code.sso/httpHandler/sessionshandler"
)

type SessionMiddleware struct {
}

func (u *SessionMiddleware) GetHandlerFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionId := r.Header.Get("ssoSession")
		if sessionId == "" {
			responses.GetInstance().WriteJsonResponse(w, r, responses.SESSION_ID_NOT_PRESENT, errors.New("Session Id missing"), nil)
		}
		session, err := sessionshandler.GetInstance().GetSession().Get(r, sessionId)
		if err != nil {
			responses.GetInstance().WriteJsonResponse(w, r, responses.SESSION_NOT_VALID, err, nil)
		}

		if value, ok := session.Values["sessionId"]; ok {
			if value == sessionId {
				session := models.Session{
					SessionId: sessionId,
				}
				ctx := context.WithValue(r.Context(), "session", session)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		responses.GetInstance().WriteJsonResponse(w, r, responses.SESSION_NOT_VALID, errors.New("Missing in data store"), nil)
	})
}
