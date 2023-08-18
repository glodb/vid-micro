package middlewares

import "net/http"

type Middleware interface {
	GetHandlerFunc(next http.Handler) http.Handler
}
