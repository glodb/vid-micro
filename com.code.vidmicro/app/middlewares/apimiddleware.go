package middlewares

import (
	"log"
	"net/http"
	"strings"

	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"github.com/gin-gonic/gin"
)

type ApiMiddleware struct {
}

func (u *ApiMiddleware) GetHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		splittedString := strings.Split(c.Request.URL.Path, "/")
		if len(splittedString) < 3 {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.API_NOT_AVAILABLE, nil, nil))
			return
		}
		urlPath := "/" + splittedString[1] + "/" + splittedString[2]

		log.Println(urlPath, c.Request.Method, configmanager.GetInstance().Apis[urlPath])

		if methods, ok := configmanager.GetInstance().Apis[urlPath]; ok {
			if u.containsMethod(methods, c.Request.Method) {
				c.Next()
			} else {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.METHOD_NOT_AVAILABLE, nil, nil))
				return
			}
		} else {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.API_NOT_AVAILABLE, nil, nil))
			return
		}
	}
}

func (u *ApiMiddleware) containsMethod(methods []string, target string) bool {

	if len(methods) > 20 { //An api can't have more than 20 methods
		return false
	}

	for _, method := range methods {
		if method == target {
			return true
		}
	}
	return false
}
