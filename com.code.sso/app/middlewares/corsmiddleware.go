package middlewares

import (
	"errors"
	"net/http"

	"com.code.sso/com.code.sso/httpHandler/responses"
)

type CORSMiddleware struct {
}

func (u *CORSMiddleware) GetHandlerFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, mode, Access-Control-Allow-Origin, x-access-token, ssoSession, url")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if r.Method == "OPTIONS" {
			responses.GetInstance().WriteJsonResponse(w, r, responses.OPTIONS_NOT_ALLOWED, errors.New("Method Options is not allowed in Cors Middleware"), nil)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
