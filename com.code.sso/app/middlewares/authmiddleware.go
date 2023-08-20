package middlewares

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"strings"

	"com.code.sso/com.code.sso/app/models"
	"com.code.sso/com.code.sso/httpHandler/responses"
	"com.code.sso/com.code.sso/httpHandler/sessionshandler"
)

type AuthMiddleware struct {
}

func (u *AuthMiddleware) GetHandlerFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(auth) != 2 || auth[0] != "Basic" {
			responses.GetInstance().WriteJsonResponse(w, r, responses.BASIC_AUTH_FAILED, errors.New("Authorization header is not basic"), nil)
			return
		}
		log.Println(auth)

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)

		sessionValue := r.Context().Value("session")

		if sessionValue != nil {
			sessionModel := sessionValue.(models.Session)
			session, _ := sessionshandler.GetInstance().GetSession().Get(r, sessionModel.SessionId)

			if val, ok := session.Values["email"]; ok {
				sessionModel.Email = val.(string)
			}

			if val, ok := session.Values["token"]; ok {
				sessionModel.Token = val.(string)
			}

			if val, ok := session.Values["firstName"]; ok {
				sessionModel.FirstName = val.(string)
			}

			if val, ok := session.Values["lastName"]; ok {
				sessionModel.LastName = val.(string)
			}

			if val, ok := session.Values["registrationType"]; ok {
				sessionModel.RegistrationType = (val.(int))
			}

			if val, ok := session.Values["phone"]; ok {
				sessionModel.Phone = (val.(string))
			}

			if val, ok := session.Values["createdAt"]; ok {
				sessionModel.CreatedAt = (val.(int))
			}

			if val, ok := session.Values["updatedAt"]; ok {
				sessionModel.UpdatedAt = (val.(int))
			}

			log.Println(sessionModel, pair)
			if pair[0] == sessionModel.Email && pair[1] == sessionModel.Token {
				ctx := context.WithValue(r.Context(), "session", sessionModel)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		responses.GetInstance().WriteJsonResponse(w, r, responses.BASIC_AUTH_FAILED, errors.New("Failed in matching user name and password"), nil)
		return
	})
}
