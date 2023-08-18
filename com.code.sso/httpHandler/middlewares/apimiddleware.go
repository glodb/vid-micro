package middlewares

import (
	"net/http"
)

type ApiMiddleware struct {
}

func (u *ApiMiddleware) GetHandlerFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// urlPath := "/" + strings.Split(r.URL.Path, "/")[1] + "/" + strings.Split(r.URL.Path, "/")[2]
		// if _, ok := config.GetSet("apis")[urlPath]; ok {
		// 	next.ServeHTTP(w, r)
		// } else {
		// 	c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.API_NOT_AVAILABLE, nil, nil))
		// }
	})
}
